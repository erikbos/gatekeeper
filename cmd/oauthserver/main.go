package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"

	"github.com/erikbos/gatekeeper/pkg/oauthtokenstore"
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
	server *server.Server
}

func main() {
	shared.StartLogging(myName, version, buildTime)
	shared.SetLoggingConfiguration("info")

	os := oauthServer{}

	manager := manage.NewDefaultManager()

	// manager.MustTokenStorage(store.NewFileTokenStore("data.db"))
	manager.MustTokenStorage(oauthtokenstore.NewTokenStore("x"))

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

	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.Use(gin.LoggerWithFormatter(shared.LogHTTPRequest))
	g.Use(shared.AddRequestID())

	g.POST("/oauth2/token", os.handleToken)
	g.GET("/oauth2/test", os.handeTestPath)

	g.Run(":1000")
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
