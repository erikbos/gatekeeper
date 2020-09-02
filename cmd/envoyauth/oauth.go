package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
)

// - [OAuth 2 Simplified](https://aaronparecki.com/oauth-2-simplified/)
// - [An introduction to OAuth](https://www.digitalocean.com/community/tutorials/an-introduction-to-oauth-2)
// - [OAuth 2.0 RFC](https://tools.ietf.org/html/rfc6749)
// - [OAuth 2.0 Bearer Token Usage RFC](https://tools.ietf.org/html/rfc6750)
// - [Go OAuth package](https://pkg.go.dev/gopkg.in/oauth2.v3/server)

// oauthServerConfig contains our configuration
type oauthServerConfig struct {
	Listen string `yaml:"listen"` // OAuth Address and port to listen
	TLS    struct {
		certFile string `yaml:"certfile"` // TLS certifcate file
		keyFile  string `yaml:"keyfile"`  // TLS certifcate key file
	} `yaml:"tls"`
	TokenIssuePath string `yaml:"tokenissuepath"` // Path to request access tokens (e.g. "/oauth2/token")
	TokenInfoPath  string `yaml:"tokeninfopath"`  // Path to request info about token (e.g. "/oauth2/info")
}

type oauthServer struct {
	config             *oauthServerConfig
	ginEngine          *gin.Engine
	db                 *db.Database
	cache              *Cache
	server             *server.Server
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

func newOAuthServer(config *oauthServerConfig, db *db.Database, cache *Cache) *oauthServer {

	return &oauthServer{
		config: config,
		db:     db,
		cache:  cache,
	}
}

// Start starts OAuth2 public endpoints to request new access token
// or get info about an access info
func (oauth *oauthServer) Start() {
	// Do not start oauth system if we do not have a listenport
	if oauth.config.Listen == "" {
		return
	}
	if oauth.config.TokenIssuePath == "" {
		log.Warn("OAuth TokenIssuePath needs to be configured")
		return
	}

	oauth.registerMetrics()
	oauth.prepareOAuthInstance()

	gin.SetMode(gin.ReleaseMode)
	oauth.ginEngine = gin.New()
	oauth.ginEngine.Use(gin.LoggerWithFormatter(shared.LogHTTPRequest))
	oauth.ginEngine.Use(shared.AddRequestID())

	oauth.ginEngine.POST(oauth.config.TokenIssuePath, oauth.handleTokenIssueRequest)
	// TokenInfo is an optional endpoint
	if oauth.config.TokenInfoPath != "" {
		oauth.ginEngine.GET(oauth.config.TokenInfoPath, oauth.handleTokenInfo)
	}

	log.Info("OAuth2 listening on ", oauth.config.Listen)
	if oauth.config.TLS.certFile != "" && oauth.config.TLS.keyFile != "" {
		log.Fatal(oauth.ginEngine.RunTLS(oauth.config.Listen,
			oauth.config.TLS.certFile, oauth.config.TLS.keyFile))
	}
	log.Panic(oauth.ginEngine.Run(oauth.config.Listen))
}

// prepareOAuthInstance build OAuth server instance with client and token storage backends
func (oauth *oauthServer) prepareOAuthInstance() {

	manager := manage.NewDefaultManager()

	// Set our token storage engine for access tokens
	manager.MapTokenStorage(NewOAuthTokenStore(oauth.db, oauth.cache))

	// Set client id engine for client ids
	manager.MapClientStorage(NewOAuthClientTokenStore(oauth.db, oauth.cache))

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
func (oauth *oauthServer) handleTokenIssueRequest(c *gin.Context) {

	if err := oauth.server.HandleTokenRequest(c.Writer, c.Request); err != nil {
		oauth.tokenIssueRequests.WithLabelValues("400").Inc()
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	statusCode := fmt.Sprintf("%d", c.Writer.Status())
	oauth.tokenIssueRequests.WithLabelValues(statusCode).Inc()
}

// handleTokenInfo shows information about temporary token
func (oauth *oauthServer) handleTokenInfo(c *gin.Context) {

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

func (oauth *oauthServer) registerMetrics() {

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
