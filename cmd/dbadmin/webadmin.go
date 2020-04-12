package main

import (
	"github.com/erikbos/apiauth/pkg/types"
	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

//StartWebAdminServer starts the admin web UI
//
func StartWebAdminServer(e *env) {
	if e.config.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	e.ginEngine = gin.New()
	e.ginEngine.Use(gin.LoggerWithFormatter(types.LogHTTPRequest))

	e.registerOrganizationRoutes(e.ginEngine)
	e.registerDeveloperRoutes(e.ginEngine)
	e.registerDeveloperAppRoutes(e.ginEngine)
	e.registerCredentialRoutes(e.ginEngine)
	e.registerAPIProductRoutes(e.ginEngine)
	e.registerClusterRoutes(e.ginEngine)

	e.ginEngine.Static("/assets", "./assets")
	e.ginEngine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	e.ginEngine.GET("/", e.ShowWebAdminHomePage)
	e.ginEngine.GET("/ready", e.readyness.DisplayReadyness)

	e.readyness.Up()

	log.Info("Webadmin listening on ", e.config.WebAdminListen)
	e.ginEngine.Run(e.config.WebAdminListen)
}

// ShowWebAdminHomePage shows home page
func (e *env) ShowWebAdminHomePage(c *gin.Context) {
	// FIXME feels like hack, is there a better way to pass gin engine context?
	types.ShowIndexPage(c, e.ginEngine, myName)
}
