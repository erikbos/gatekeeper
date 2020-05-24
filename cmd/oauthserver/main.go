package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"

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
	// manager.MustTokenStorage(store.NewFileTokenStore("/tmp/data.db"))
	manager.MapTokenStorage(oauthtoken.NewTokenStore(os.db))

	// client store
	clientStore := store.NewClientStore()
	clientStore.Set("000000", &models.Client{
		ID:     "000000",
		Secret: "999999",
		Domain: "http://localhost",
	})
	manager.MapClientStorage(clientStore)

	// Initialize the oauth2 service
	os.server = server.NewDefaultServer(manager)
	os.server.SetClientInfoHandler(server.ClientFormHandler)
}

func (os *oauthServer) handleToken(c *gin.Context) {

	err := os.server.HandleTokenRequest(c.Writer, c.Request)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
}

func (os *oauthServer) handeTestPath(c *gin.Context) {

	ti, err := os.server.ValidationBearerToken(c.Request)

	log.Printf("ti: %+v, err: %+v", ti, err)

	c.JSON(http.StatusOK, ti)
}
