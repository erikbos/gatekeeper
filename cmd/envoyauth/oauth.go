package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

// OAuthServerConfig contains our configuration
type OAuthServerConfig struct {
	Logger shared.Logger `yaml:"logging"` // log configuration of webadmin accesslog
	Listen string        `yaml:"listen"`  // OAuth Address and port to listen
	TLS    struct {
		certFile string `yaml:"certfile"` // TLS certifcate file
		keyFile  string `yaml:"keyfile"`  // TLS certifcate key file
	} `yaml:"tls"`
	TokenIssuePath string `yaml:"tokenissuepath"` // Path to request access tokens (e.g. "/oauth2/token")
	TokenInfoPath  string `yaml:"tokeninfopath"`  // Path to request info about token (e.g. "/oauth2/info")
}

// OAuthServer is an oauth server instance
type OAuthServer struct {
	config             *OAuthServerConfig
	router             *gin.Engine
	db                 *db.Database
	server             *server.Server
	logger             *zap.Logger
	tokenIssueRequests *prometheus.CounterVec
	tokenInfoRequests  *prometheus.CounterVec
}

// tokenInfoAnswer is returned by public OAuth Token Info endpoint
type tokenInfoAnswer struct {
	Valid     bool      `json:"valid"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	Scope     string    `json:"scope"`
}

func newOAuthServer(config *OAuthServerConfig, db *db.Database) *OAuthServer {

	return &OAuthServer{
		config: config,
		db:     db,
		// cache: newCache(&cacheConfig{
		// 	Size:        100 * 1024 * 1025,
		// 	TTL:         15,
		// 	NegativeTTL: 5,
		// }),
	}
}

// Start starts OAuth2 public endpoints to request new access token
// or get info about an access info
func (oauth *OAuthServer) Start() error {
	// Do not start oauth system if we do not have a listenport
	if oauth.config.Listen == "" {
		return nil
	}
	if oauth.config.TokenIssuePath == "" {
		return errors.New("OAuth TokenIssuePath needs to be configured")
	}

	oauth.logger = shared.NewLogger(&oauth.config.Logger, false)

	oauth.registerMetrics()
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
func (oauth *OAuthServer) prepareOAuthInstance() {

	manager := manage.NewDefaultManager()

	// Set our token storage engine for access tokens
	manager.MapTokenStorage(NewOAuthTokenStore(oauth.db, oauth.logger))

	// Set client id engine for client ids
	manager.MapClientStorage(NewOAuthClientTokenStore(oauth.db, oauth.logger))

	// Set default token ttl
	manager.SetClientTokenCfg(&manage.Config{AccessTokenExp: 1 * time.Hour})

	config := &server.Config{
		TokenType: "Bearer",
		// We do not allow token-by-GET requests
		AllowGetAccessRequest: false,
		AllowedResponseTypes: []oauth2.ResponseType{
			oauth2.Token,
		},
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.ClientCredentials,
		},
	}
	oauth.server = server.NewServer(config, manager)

	// Setup extracting POSTed clientId/Secret
	oauth.server.SetClientInfoHandler(server.ClientFormHandler)
}

// handleTokenIssueRequest handles a POST request for a new OAuth token
func (oauth *OAuthServer) handleTokenIssueRequest(c *gin.Context) {

	if err := oauth.server.HandleTokenRequest(c.Writer, c.Request); err != nil {
		oauth.tokenIssueRequests.WithLabelValues("400").Inc()
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	statusCode := fmt.Sprintf("%d", c.Writer.Status())
	oauth.tokenIssueRequests.WithLabelValues(statusCode).Inc()
}

// handleTokenInfo shows information about temporary token
func (oauth *OAuthServer) handleTokenInfo(c *gin.Context) {

	tokenInfo, err := oauth.server.ValidationBearerToken(c.Request)
	if err != nil {
		oauth.tokenInfoRequests.WithLabelValues("401").Inc()
		_ = c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	// Copy over the information we want to show back:
	// We want to prevent showing client_id / client_secret for example.
	status := tokenInfoAnswer{
		Valid:     true,
		CreatedAt: tokenInfo.GetAccessCreateAt().UTC(),
		ExpiresAt: tokenInfo.GetAccessCreateAt().Add(tokenInfo.GetAccessExpiresIn()).UTC(),
		Scope:     tokenInfo.GetScope(),
	}

	oauth.tokenInfoRequests.WithLabelValues("200").Inc()
	c.JSON(http.StatusOK, status)
}

func (oauth *OAuthServer) registerMetrics() {

	oauth.tokenIssueRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "token_issue_total",
			Help:      "Number of OAuth token issue requests.",
		}, []string{"status"})
	prometheus.MustRegister(oauth.tokenIssueRequests)

	oauth.tokenInfoRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicationName,
			Name:      "token_info_total",
			Help:      "Number of OAuth token info requests.",
		}, []string{"status"})
	prometheus.MustRegister(oauth.tokenInfoRequests)
}
