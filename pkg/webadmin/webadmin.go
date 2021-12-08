package webadmin

import (
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Config holds the configuration of a webadmin
type Config struct {
	Logger shared.Logger `yaml:"logging"` // log configuration of webadmin accesslog
	Listen string        `yaml:"listen"`  // Address and port to listen
	IPACL  string        `yaml:"ipacl"`   // ip accesslist (e.g. "10.0.0.0/8,192.168.0.0/16")
	TLS    struct {
		CertFile string `yaml:"certfile"` // TLS certifcate file
		KeyFile  string `yaml:"keyfile"`  // TLS certifcate key file
	} `yaml:"tls"`
}

// Webadmin is an instance of our admin interface that provides operational information
type Webadmin struct {
	config Config
	Router *gin.Engine
	logger *zap.Logger
}

//
const (
	// Path to be used by k8s liveness check
	LivenessCheckPath = "/liveness"

	// Path to be used by k8s readiness check
	ReadinessCheckPath = "/readiness"

	// Prometheus metrics endpoint
	MetricsPath = "/metrics"

	// Path endpoint for showing running configuration
	ConfigDumpPath = "/config_dump"

	// Key of user's role in request context
	RoleContextKey = "Role"

	// Key of RequestID in request context
	RequestIDKey = "RequestId"

	// RequestID header
	RequestIDHeader = "request-id"

	contentTypeHeader = "content-type"
	contentTypeYAML   = "text/yaml; charset=utf-8"
	contentTypeHTML   = "text/html; charset=utf-8"
)

// New returns a new webadmin
func New(config Config, applicationName string, logger *zap.Logger) *Webadmin {

	gin.SetMode(gin.ReleaseMode)

	// Enable strict field checking of POSTed JSON
	gin.EnableJsonDecoderDisallowUnknownFields()

	router := gin.New()

	// Enable access logging
	router.Use(LogHTTPRequest(logger))

	// Enable adding setting a request-id per request
	router.Use(SetRequestID())

	// Enable source ip addressing checking per request
	router.Use(CheckIPACL(config.IPACL))

	return &Webadmin{
		config: config,
		Router: router,
		logger: logger,
	}
}

// Start starts a web admin instance
func (w *Webadmin) Start() {

	w.logger.Info("Webadmin listening on " + w.config.Listen)
	if w.config.TLS.CertFile != "" &&
		w.config.TLS.KeyFile != "" {

		if err := w.Router.RunTLS(w.config.Listen,
			w.config.TLS.CertFile, w.config.TLS.KeyFile); err != nil {
			w.logger.Fatal("error starting webadmin",
				zap.Error(err))
		}
	}
	if err := w.Router.Run(w.config.Listen); err != nil {
		w.logger.Fatal("error starting webadmin", zap.Error(err))
	}
}

// CheckIPACL checks if requestor's ip address matches ACL
func CheckIPACL(ipAccessList string) gin.HandlerFunc {

	return func(c *gin.Context) {
		if ipAccessList == "" {
			JSONMessageAndAbort(c, http.StatusForbidden,
				errors.New("permission denied, No IP ACL configured"))
			return
		}
		if !shared.CheckIPinAccessList(net.ParseIP(c.ClientIP()), ipAccessList) {
			JSONMessageAndAbort(c, http.StatusForbidden,
				errors.New("permission denied, Access denied by IP ACL"))
			return
		}
		// no hit, we allow request
	}
}

// StoreUser stores provided username in request context
func StoreUser(c *gin.Context, user string) {

	c.Set(gin.AuthUserKey, user)
}

// GetUser returns name of requestor
func GetUser(c *gin.Context) string {

	return c.GetString(gin.AuthUserKey)
}

// StoreRole stores provided role in request context
func StoreRole(c *gin.Context, role string) {

	c.Set(RoleContextKey, role)
}

// GetRole returns role of requestor
func GetRole(c *gin.Context) string {

	return c.GetString(RoleContextKey)
}

// ShowStartupConfiguration prints configuration object as yaml
func ShowStartupConfiguration(configObject interface{}) gin.HandlerFunc {

	config := fmt.Sprint(configObject)

	return func(c *gin.Context) {
		c.Header(contentTypeHeader, contentTypeYAML)
		c.String(http.StatusOK, config)
	}
}

// SetRequestID adds a Request-Id HTTP header for tracking purposes
func SetRequestID() gin.HandlerFunc {

	return func(c *gin.Context) {
		requestID := uuid.New().String()
		// Save in request context
		c.Set(RequestIDKey, requestID)
		// Set http header
		c.Writer.Header().Set(RequestIDHeader, requestID)
		c.Next()
	}
}

// GetRequestID returns RequestID from request context
func GetRequestID(c *gin.Context) string {

	return c.GetString(RequestIDKey)
}

// ShowAllRoutes shows HTML page based with all registered routes
func ShowAllRoutes(e *gin.Engine, applicationName string) gin.HandlerFunc {

	return func(c *gin.Context) {

		routes := e.Routes()
		sort.SliceStable(routes, func(i, j int) bool {
			return routes[i].Path < routes[j].Path
		})

		t, err := template.New("indexpage").Parse(showAllRoutesPageTemplate)
		if err != nil {
			JSONMessage(c, http.StatusServiceUnavailable, err)
			return
		}
		templateVariables := struct {
			Name   string
			Routes gin.RoutesInfo
		}{
			applicationName,
			routes,
		}
		c.Status(http.StatusOK)
		c.Header(contentTypeHeader, contentTypeHTML)
		if err := t.Execute(c.Writer, templateVariables); err != nil {
			_ = c.Error(err)
		}
	}
}

// JSONMessage returns an error message
func JSONMessage(c *gin.Context, statusCode int, errorMessage error) {

	// Store error in request context so it ends up in access log
	if errorMessage != nil {
		_ = c.Error(errorMessage)
	}

	c.JSON(statusCode,
		gin.H{
			"message": fmt.Sprint(errorMessage),
		})
}

// JSONMessageAndAbort returns an error message, and aborts request
func JSONMessageAndAbort(c *gin.Context, statusCode int, errorMessage error) {

	JSONMessage(c, statusCode, errorMessage)
	c.Abort()
}

// setLastModified adds Last-Modified timestamp to response
// func setLastModified(c *gin.Context, timestamp int64) {

// 	c.Header("Last-Modified",
// 		time.Unix(0, timestamp*int64(time.Millisecond)).Format(time.RFC822))
// }

// LivenessProbe answer with OK
func LivenessProbe(c *gin.Context) {

	JSONMessage(c, http.StatusOK, errors.New("liveness OK"))
}

// LogHTTPRequest logs details of an HTTP request
func LogHTTPRequest(logger *zap.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {

		requesturi := c.Request.URL.RequestURI()

		// Do not log k8s health probes
		if requesturi == LivenessCheckPath || requesturi == ReadinessCheckPath {
			return
		}

		start := time.Now()
		c.Next()
		servicetime := time.Since(start)

		// Get errors that might have occured during request handling
		var allErrors string
		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				allErrors += strings.TrimSpace(e)
			}
		}
		logger.Info("http",
			zap.String("ip", c.ClientIP()),
			zap.String("xff", c.Request.Header.Get("x-forwarded-for")),
			zap.String("method", c.Request.Method),
			zap.String("uri", requesturi),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
			zap.String("user", GetUser(c)),
			zap.String("role", GetRole(c)),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("servicetime", servicetime),
			zap.String("requestid", GetRequestID(c)),
			zap.String("error", allErrors))
	}
}

const showAllRoutesPageTemplate = `
<!DOCTYPE html>
<html>
<head>
<title>{{ .Name }}</title>
<link rel="icon" type="image/svg+xml" href="data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBzdGFuZGFsb25lPSJubyI/Pgo8IURPQ1RZUEUgc3ZnIFBVQkxJQyAiLS8vVzNDLy9EVEQgU1ZHIDIwMDEwOTA0Ly9FTiIKICJodHRwOi8vd3d3LnczLm9yZy9UUi8yMDAxL1JFQy1TVkctMjAwMTA5MDQvRFREL3N2ZzEwLmR0ZCI+CjxzdmcgdmVyc2lvbj0iMS4wIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciCiB3aWR0aD0iNDguMDAwMDAwcHQiIGhlaWdodD0iNDguMDAwMDAwcHQiIHZpZXdCb3g9IjAgMCA0OC4wMDAwMDAgNDguMDAwMDAwIgogcHJlc2VydmVBc3BlY3RSYXRpbz0ieE1pZFlNaWQgbWVldCI+CjxtZXRhZGF0YT4KQ3JlYXRlZCBieSBwb3RyYWNlIDEuMTUsIHdyaXR0ZW4gYnkgUGV0ZXIgU2VsaW5nZXIgMjAwMS0yMDE3CjwvbWV0YWRhdGE+CjxnIHRyYW5zZm9ybT0idHJhbnNsYXRlKDAuMDAwMDAwLDQ4LjAwMDAwMCkgc2NhbGUoMC4xMDAwMDAsLTAuMTAwMDAwKSIKZmlsbD0iIzAwMDAwMCIgc3Ryb2tlPSJub25lIj4KPHBhdGggZD0iTTIwOSAzOTIgYy0xMTYgLTc0IC0xODQgLTIzNSAtMTI5IC0zMDUgMTEgLTE0IDI5IC0yOSA0MCAtMzIgMzIgLTEwCjk0IDEzIDEyNSA0NiBsMzAgMzEgNSAtMzkgYzQgLTMzIDkgLTM4IDMzIC00MSAyMiAtMyAzNyA2IDY3IDM2IDIyIDIyIDQwIDQ4CjQwIDU4IDAgMTUgLTYgMTIgLTI5IC0xNSAtMTUgLTE5IC0zMiAtMzIgLTM4IC0zMCAtNiAyIDcgNjMgMzMgMTU0IDI0IDgzIDQ0CjE1MyA0NCAxNTggMCAxNCAtNTkgNyAtNjggLTkgLTggLTE0IC0xMCAtMTQgLTIxIDAgLTIxIDI2IC04MSAyMCAtMTMyIC0xMnoKbTExOCAtMjMgYzMzIC02OCAtNzkgLTI2OSAtMTQ5IC0yNjkgLTg4IDAgLTM1IDIwNiA3MSAyNzggNDIgMjcgNjIgMjUgNzggLTl6Ii8+CjwvZz4KPC9zdmc+Cg==">
<style>
table {
  font-family: sans-serif;
  font-size: medium;
  border-collapse: collapse;
}
.home-row:nth-child(even) {
  background-color: #dddddd;
}
.home-data {
  border: 1px solid #dddddd;
  text-align: left;
  padding: 8px;
}
.home-form {
  margin-bottom: 0;
}
</style>
</head>
<body>
<h1>{{ .Name }}</h1>
<table class='home-table'>
<thead>
	<th class='home-data'>Method</th>
	<th class='home-data'>Path</th>
</thead>
<tbody>
{{range .Routes}}
	<tr class='home-row'>
	<td class='home-data'>{{ .Method }}</td>
	<td class='home-data'><a href='{{ .Path }}'>{{ .Path }}</a></td>
</tr>
{{ end }}
</tbody>
</table>
</body>
`
