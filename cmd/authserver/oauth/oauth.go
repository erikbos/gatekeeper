package oauth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/authserver/metrics"
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

// Config contains our configuration
type Config struct {
	Logger shared.Logger // log configuration of webadmin accesslog
	Listen string        // OAuth Address and port to listen
	TLS    struct {
		certFile string // TLS certifcate file
		keyFile  string // TLS certifcate key file
	}
	TokenIssuePath string // Path to request access tokens (e.g. "/oauth2/token")
	TokenInfoPath  string // Path to request info about token (e.g. "/oauth2/info")
}

// Server is an oauth server instance
type Server struct {
	config      Config
	router      *gin.Engine
	db          *db.Database
	oauthserver *server.Server
	logger      *zap.Logger
	metrics     *metrics.Metrics
}

// New returns a new oauth server instance
func New(config Config, db *db.Database, metrics *metrics.Metrics,
	logger *zap.Logger) *Server {

	return &Server{
		config:  config,
		db:      db,
		metrics: metrics,
		logger:  logger.With(zap.String("system", "oauth")),
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
		return errors.New("oauth TokenIssuePath needs to be configured")
	}

	oauth.logger = shared.NewLogger("oauth", &oauth.config.Logger)

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

		err := oauth.router.RunTLS(oauth.config.Listen,
			oauth.config.TLS.certFile, oauth.config.TLS.keyFile)
		if err != nil {
			oauth.logger.Fatal("error starting tls webadmin", zap.Error(err))
		}
	}

	err := oauth.router.Run(oauth.config.Listen)
	if err != nil {
		oauth.logger.Fatal("error starting webadmin", zap.Error(err))
	}
	return err
}

// prepareOAuthInstance build OAuth server instance with client and token storage backends
func (oauth *Server) prepareOAuthInstance() {

	manager := manage.NewDefaultManager()

	// Set our token storage engine for access tokens
	manager.MapTokenStorage(NewTokenStore(oauth.db, oauth.metrics, oauth.logger))

	// Set client id engine for client ids
	manager.MapClientStorage(NewClientTokenStore(oauth.db, oauth.metrics, oauth.logger))

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
func (oauth *Server) LoadAccessToken(ctx context.Context, accessToken string) (oauth2.TokenInfo, error) {

	return oauth.oauthserver.Manager.LoadAccessToken(ctx, accessToken)
}

// handleTokenIssueRequest handles a POST request for a new OAuth token
func (oauth *Server) handleTokenIssueRequest(c *gin.Context) {

	if err := oauth.oauthserver.HandleTokenRequest(c.Writer, c.Request); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
}

// tokenInfoAnswer is returned by public OAuth Token Info endpoint
// fields according to https://tools.ietf.org/html/rfc7662
// OAuth 2.0 Token Introspection
type tokenInfoAnswer struct {
	Active    bool   `json:"active"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	Scope     string `json:"scope"`
}

// handleTokenInfo shows information about temporary token (RFC7662)
func (oauth *Server) handleTokenInfo(c *gin.Context) {

	tokenInfo, err := oauth.oauthserver.ValidationBearerToken(c.Request)
	if err != nil {
		_ = c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	// Copy over the information we want to return back as "public info"
	// We must not show client_id / client_secret as they are secret.
	//
	// Fields named according to https://tools.ietf.org/html/rfc7662
	//
	status := tokenInfoAnswer{
		Active:   true,
		IssuedAt: tokenInfo.GetAccessCreateAt().UTC().Unix(),
		ExpiresAt: tokenInfo.GetAccessCreateAt().
			Add(tokenInfo.GetAccessExpiresIn()).UTC().Unix(),
		Scope: tokenInfo.GetScope(),
	}
	c.IndentedJSON(http.StatusOK, status)
}
