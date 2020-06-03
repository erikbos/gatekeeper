package main

import (
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
	server             *server.Server
	tokenIssueRequests prometheus.Counter
	tokenInfoRequests  prometheus.Counter
}

// StartOAuthServer runs our public endpoint
func StartOAuthServer(a *authorizationServer) {
	// shared.StartLogging(myName, version, buildTime)

	server := oauthServer{
		config: &a.config.Oauth,
		db:     a.db,
	}

	// shared.SetLoggingConfiguration(server.config.LogLevel)
	// server.readiness.RegisterMetrics(myName)

	registerOAuthMetrics(&server)

	server.prepareOAuthInstance()

	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.Use(gin.LoggerWithFormatter(shared.LogHTTPRequest))
	g.Use(shared.AddRequestID())

	g.POST("/oauth2/token", server.handleTokenIssueRequest)
	g.GET("/oauth2/info", server.handleTokenInfo)

	g.Run(a.config.Oauth.Listen)
}

// prepareOAuthInstance build OAuth server instance with client and token storage backends
func (o *oauthServer) prepareOAuthInstance() {

	manager := manage.NewDefaultManager()

	// Set our token storage engine for access tokens
	manager.MapTokenStorage(NewOAuthTokenStore(o.db))

	// Set client id engine for client ids
	manager.MapClientStorage(NewOAuthClientTokenStore(o.db))

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
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
}

// handleTokenInfo shows information about token
func (o *oauthServer) handleTokenInfo(c *gin.Context) {

	tokenInfo, err := o.server.ValidationBearerToken(c.Request)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	// FIX ME we probably do not want to show all fields
	c.JSON(http.StatusOK, tokenInfo)
}

func registerOAuthMetrics(o *oauthServer) {

	o.tokenIssueRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "token_issue_total",
			Help:      "Number of OAuth token issue requests.",
		})
	prometheus.MustRegister(o.tokenIssueRequests)

	o.tokenInfoRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: myName,
			Name:      "token_info_total",
			Help:      "Number of OAuth token info requests.",
		})
	prometheus.MustRegister(o.tokenInfoRequests)
}
