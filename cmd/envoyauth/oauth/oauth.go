package oauth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

// Config contains our configuration
type Config struct {
	Logger shared.Logger `yaml:"logging"` // log configuration of webadmin accesslog
	Listen string        `yaml:"listen"`  // OAuth Address and port to listen
	TLS    struct {
		certFile string `yaml:"certfile"` // TLS certifcate file
		keyFile  string `yaml:"keyfile"`  // TLS certifcate key file
	} `yaml:"tls"`
	TokenIssuePath string `yaml:"tokenissuepath"` // Path to request access tokens (e.g. "/oauth2/token")
	TokenInfoPath  string `yaml:"tokeninfopath"`  // Path to request info about token (e.g. "/oauth2/info")
}

// Server is an oauth server instance
type Server struct {
	config      Config
	router      *gin.Engine
	db          *db.Database
	oauthserver *server.Server
	logger      *zap.Logger
	metrics     *metrics
}

// New returns a new oauth server instance
func New(config Config, db *db.Database, logger *zap.Logger) *Server {

	return &Server{
		config:  config,
		db:      db,
		logger:  logger.With(zap.String("system", "oauth")),
		metrics: newMetrics(),
	}
}

// Start starts OAuth2 public endpoints to request new access token
// or get info about an access info
func (oauth *Server) Start(applicationName string) error {
	// Do not start oauth system if we do not have a listenport
	if oauth.config.Listen == "" {
		return nil
	}
	if oauth.config.TokenIssuePath == "" {
		return errors.New("OAuth TokenIssuePath needs to be configured")
	}

	oauth.logger = shared.NewLogger(&oauth.config.Logger)

	oauth.metrics.RegisterWithPrometheus(applicationName)
	oauth.prepareOAuthInstance()

	gin.SetMode(gin.ReleaseMode)
	oauth.router = gin.New()
	oauth.router.Use(webadmin.LogHTTPRequest(oauth.logger))
	oauth.router.Use(webadmin.SetRequestID())

	oauth.router.POST(oauth.config.TokenIssuePath, oauth.handleTokenIssueRequest)
	// TokenInfo is an optional endpoint
	if oauth.config.TokenInfoPath != "" {
		oauth.router.GET(oauth.config.TokenInfoPath, oauth.handleTokenInfo)
	}

	oauth.logger.Info("OAuth2 listening on " + oauth.config.Listen)
	if oauth.config.TLS.certFile != "" &&
		oauth.config.TLS.keyFile != "" {

		oauth.logger.Fatal("error starting tls webadmin",
			zap.Error(oauth.router.RunTLS(
				oauth.config.Listen,
				oauth.config.TLS.certFile,
				oauth.config.TLS.keyFile)))
	}

	oauth.logger.Fatal("error starting webadmin",
		zap.Error(oauth.router.Run(oauth.config.Listen)))
	return nil
}

// prepareOAuthInstance build OAuth server instance with client and token storage backends
func (oauth *Server) prepareOAuthInstance() {

	manager := manage.NewDefaultManager()

	// Set our token storage engine for access tokens
	manager.MapTokenStorage(NewOAuthTokenStore(oauth.db, oauth.metrics, oauth.logger))

	// Set client id engine for client ids
	manager.MapClientStorage(NewOAuthClientTokenStore(oauth.db, oauth.metrics, oauth.logger))

	// Set default token ttl
	manager.SetClientTokenCfg(&manage.Config{AccessTokenExp: 1 * time.Hour})

	config := &server.Config{
		TokenType: "Bearer",
		// We do not allow retrieving token using HTTP GET method
		AllowGetAccessRequest: false,
		AllowedResponseTypes: []oauth2.ResponseType{
			oauth2.Token,
		},
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.ClientCredentials,
		},
	}
	oauth.oauthserver = server.NewServer(config, manager)

	// Setup extracting POSTed clientId/Secret
	oauth.oauthserver.SetClientInfoHandler(server.ClientFormHandler)
}

// LoadAccessToken returns the details of token
func (oauth *Server) LoadAccessToken(accessToken string) (oauth2.TokenInfo, error) {

	return oauth.oauthserver.Manager.LoadAccessToken(accessToken)
}

// handleTokenIssueRequest handles a POST request for a new OAuth token
func (oauth *Server) handleTokenIssueRequest(c *gin.Context) {

	if err := oauth.oauthserver.HandleTokenRequest(c.Writer, c.Request); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
}

// tokenInfoAnswer is returned by public OAuth Token Info endpoint
type tokenInfoAnswer struct {
	Valid     bool      `json:"valid"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	Scope     string    `json:"scope"`
}

// handleTokenInfo shows information about temporary token
func (oauth *Server) handleTokenInfo(c *gin.Context) {

	tokenInfo, err := oauth.oauthserver.ValidationBearerToken(c.Request)
	if err != nil {
		_ = c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	// Copy over the information we want to return back as "public info"
	// We must not show client_id / client_secret as they are secret.
	status := tokenInfoAnswer{
		Valid:     true,
		CreatedAt: tokenInfo.GetAccessCreateAt().UTC(),
		ExpiresAt: tokenInfo.GetAccessCreateAt().Add(tokenInfo.GetAccessExpiresIn()).UTC(),
		Scope:     tokenInfo.GetScope(),
	}
	c.JSON(http.StatusOK, status)
}
