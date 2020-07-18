package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
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
	Listen string `yaml:"listen"`
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

// StartOAuthServer runs our public endpoint
func StartOAuthServer(a *authorizationServer) {
	// shared.StartLogging(myName, version, buildTime)

	server := oauthServer{
		config: &a.config.OAuth,
		db:     a.db,
		cache:  a.cache,
	}
	a.oauth = &server

	// shared.SetLoggingConfiguration(server.config.LogLevel)
	// server.readiness.RegisterMetrics(myName)

	registerOAuthMetrics(&server)

	server.prepareOAuthInstance()

	gin.SetMode(gin.ReleaseMode)

	server.ginEngine = gin.New()
	server.ginEngine.Use(gin.LoggerWithFormatter(shared.LogHTTPRequest))
	server.ginEngine.Use(shared.AddRequestID())

	server.ginEngine.POST("/oauth2/token", server.handleTokenIssueRequest)
	server.ginEngine.GET("/oauth2/info", server.handleTokenInfo)

	log.Panic(server.ginEngine.Run(a.config.OAuth.Listen))
}

// prepareOAuthInstance build OAuth server instance with client and token storage backends
func (o *oauthServer) prepareOAuthInstance() {

	manager := manage.NewDefaultManager()

	// Set our token storage engine for access tokens
	manager.MapTokenStorage(NewOAuthTokenStore(o.db, o.cache))

	// Set client id engine for client ids
	manager.MapClientStorage(NewOAuthClientTokenStore(o.db, o.cache))

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
	o.server = server.NewServer(config, manager)

	// Setup extracting POSTed clientId/Secret
	o.server.SetClientInfoHandler(server.ClientFormHandler)
}

// handleTokenIssueRequest handles a POST request for a new OAuth token
func (o *oauthServer) handleTokenIssueRequest(c *gin.Context) {

	if err := o.server.HandleTokenRequest(c.Writer, c.Request); err != nil {
		o.tokenIssueRequests.WithLabelValues("400").Inc()
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	statusCode := fmt.Sprintf("%d", c.Writer.Status())
	o.tokenIssueRequests.WithLabelValues(statusCode).Inc()
}

type tokenInfoAnswer struct {
	Valid     bool
	CreatedAt time.Time
	ExpiresAt time.Time
	Scope     string
}

// handleTokenInfo shows information about temporary token
func (o *oauthServer) handleTokenInfo(c *gin.Context) {

	tokenInfo, err := o.server.ValidationBearerToken(c.Request)
	if err != nil {
		o.tokenInfoRequests.WithLabelValues("401").Inc()
		_ = c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	// Copy over the information we want to show back:
	// We want to prevent showing client_id / client_secret for example.
	status := tokenInfoAnswer{
		Valid:     true,
		CreatedAt: tokenInfo.GetAccessCreateAt(),
		ExpiresAt: tokenInfo.GetAccessCreateAt().Add(tokenInfo.GetAccessExpiresIn()),
		Scope:     tokenInfo.GetScope(),
	}

	o.tokenInfoRequests.WithLabelValues("200").Inc()
	c.JSON(http.StatusOK, status)
}

func registerOAuthMetrics(o *oauthServer) {

	o.tokenIssueRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicatioName,
			Name:      "token_issue_total",
			Help:      "Number of OAuth token issue requests.",
		}, []string{"status"})
	prometheus.MustRegister(o.tokenIssueRequests)

	o.tokenInfoRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: applicatioName,
			Name:      "token_info_total",
			Help:      "Number of OAuth token info requests.",
		}, []string{"status"})
	prometheus.MustRegister(o.tokenInfoRequests)
}
