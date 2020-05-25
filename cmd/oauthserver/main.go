package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"

	"github.com/erikbos/gatekeeper/pkg/db"
	oauthtoken "github.com/erikbos/gatekeeper/pkg/oauthtoken"
	"github.com/erikbos/gatekeeper/pkg/shared"
)

var (
	version   string
	buildTime string
)

const (
	myName = "oauthserver"
)

type oauthServer struct {
	config    *OAuthServerConfig
	ginEngine *gin.Engine
	db        *db.Database
	server    *server.Server
	readiness shared.Readiness
}

func main() {
	shared.StartLogging(myName, version, buildTime)

	server := oauthServer{
		config: loadConfiguration(),
	}

	shared.SetLoggingConfiguration(server.config.LogLevel)
	server.readiness.RegisterMetrics(myName)

	var err error
	server.db, err = db.Connect(server.config.Database, &server.readiness, myName)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
	}

	server.prepOAuth()

	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.Use(gin.LoggerWithFormatter(shared.LogHTTPRequest))
	g.Use(shared.AddRequestID())

	g.POST("/oauth2/token", server.handleToken)
	g.GET("/oauth2/test", server.handeTestPath)

	g.Run(":1000")
}

func (os *oauthServer) prepOAuth() {

	manager := manage.NewDefaultManager()

	// Set our token storage engine for access tokens
	manager.MapTokenStorage(oauthtoken.NewTokenStore(os.db))

	// Set client id engine for client ids
	manager.MapClientStorage(oauthtoken.NewClientTokenStore(os.db))

	// Set token ttl
	// manager.SetClientTokenCfg(&manage.Config{AccessTokenExp: 1 * time.Hour})

	config := &server.Config{
		TokenType:             "Bearer",
		AllowGetAccessRequest: false,
		AllowedResponseTypes: []oauth2.ResponseType{
			// oauth2.Code,
			oauth2.Token,
		},
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.ClientCredentials,
		},
	}
	os.server = server.NewServer(config, manager)

	// Setup extracting POSTed clientId/Secret
	os.server.SetClientInfoHandler(server.ClientFormHandler)
}

func (os *oauthServer) handleToken(c *gin.Context) {

	if err := os.server.HandleTokenRequest(c.Writer, c.Request); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
}

func (os *oauthServer) handeTestPath(c *gin.Context) {

	tokenInfo, err := os.server.ValidationBearerToken(c.Request)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	// FIX ME we probably do not want to show all fields
	c.JSON(http.StatusOK, tokenInfo)
}
