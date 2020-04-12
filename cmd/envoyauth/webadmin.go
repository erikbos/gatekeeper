package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

//StartWebAdminServer starts the admin web UI
//
func StartWebAdminServer(a *authorizationServer) {
	if a.config.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	a.ginEngine = gin.New()
	a.ginEngine.Use(gin.LoggerWithFormatter(types.LogHTTPRequest))

	a.ginEngine.GET("/", a.ShowWebAdminHomePage)
	a.ginEngine.GET("/ready", a.readyness.DisplayReadyness)
	a.ginEngine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	a.ginEngine.GET("/config_dump", a.configDump)

	log.Info("Webadmin listening on ", a.config.WebAdminListen)
	go func() {
		a.ginEngine.Run(a.config.WebAdminListen)
	}()
}

// ShowWebAdminHomePage shows home page
func (a *authorizationServer) ShowWebAdminHomePage(c *gin.Context) {
	// FIXME feels like hack, is there a better way to pass gin engine context?
	types.ShowIndexPage(c, a.ginEngine, myName)
}

//configDump pretty prints the active configuration
//
func (a *authorizationServer) configDump(c *gin.Context) {
	// We must remove db password from configuration struct before showing
	configToPrint := a.config
	configToPrint.Database.Password = ""

	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "\t")
	err := encoder.Encode(configToPrint)
	if err != nil {
		return
	}

	c.Header("Content-type", "text/json")
	c.String(http.StatusOK, buffer.String())
}
