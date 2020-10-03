package shared

import (
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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

	// Key of RequestID in request context
	RequestIDKey = "RequestId"

	// RequestID header
	RequestIDHeader = "Request-Id"
)

// returnJSONMessage returns an error message
func returnJSONMessage(c *gin.Context, statusCode int, msg error) {

	c.IndentedJSON(statusCode,
		gin.H{
			"message": fmt.Sprint(msg),
		})
}

// returnJSONMessage returns an error message, and aborts request
func returnJSONMessageAndAbort(c *gin.Context, statusCode int, msg error) {

	returnJSONMessage(c, statusCode, msg)
	c.Abort()
}

// setLastModified adds Last-Modified timestamp to response
func setLastModified(c *gin.Context, timestamp int64) {

	c.Header("Last-Modified",
		time.Unix(0, timestamp*int64(time.Millisecond)).Format(time.RFC822))
}

// WebAdminCheckIPACL checks if requestor's ip address matches ACL
func WebAdminCheckIPACL(ipAccessList string) gin.HandlerFunc {

	return func(c *gin.Context) {
		if ipAccessList == "" {
			returnJSONMessageAndAbort(c, http.StatusForbidden,
				errors.New("Permission denied, No IP ACL configured"))
			return
		}
		if !CheckIPinAccessList(net.ParseIP(c.ClientIP()), ipAccessList) {
			returnJSONMessageAndAbort(c, http.StatusForbidden,
				errors.New("Permission denied, IP ACL denied request"))
			return
		}
		// no hit, we allow request
	}
}

// LivenessProbe answer with OK
func LivenessProbe(c *gin.Context) {

	returnJSONMessage(c, http.StatusOK, errors.New("Liveness OK"))
}

// LogHTTPRequest logs details of an HTTP request
func LogHTTPRequest(param gin.LogFormatterParams) string {

	// Do not log k8s health probes
	if param.Path == LivenessCheckPath || param.Path == ReadinessCheckPath {
		return ""
	}

	// Get username of requestor
	var user string
	if value, ok := param.Keys[gin.AuthUserKey]; ok {
		user = fmt.Sprint(value)
	} else {
		user = "-"
	}

	// Get requestID from context
	var requestID string
	if value, ok := param.Keys[RequestIDKey]; ok {
		requestID = fmt.Sprint(value)
	}

	return fmt.Sprintf("%s - %s %s \"%s %s %s\" %d %d \"%s\" \"%s\" \"%s\"\n",
		param.TimeStamp.Format(time.RFC3339),
		user,
		param.ClientIP,
		param.Method,
		param.Path,
		param.Request.Proto,
		param.StatusCode,
		param.Latency/time.Millisecond,
		param.Request.UserAgent(),
		requestID,
		// Remove any whitespace clutter in error messages to keep logs tidy
		strings.TrimSpace(param.ErrorMessage),
	)
}

// AddRequestID adds a Request-Id HTTP header for tracking purposes
func AddRequestID() gin.HandlerFunc {

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

	if requestID, exists := c.Get(RequestIDKey); exists {
		return fmt.Sprint(requestID)
	}
	return ""
}

//ShowIndexPage produces the index page based upon all registered routes
func ShowIndexPage(c *gin.Context, e *gin.Engine, applicationName string) {

	t, err := template.New("indexpage").Parse(adminIndexPageTemplate)
	if err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	templateVariables := struct {
		Name   string
		Routes gin.RoutesInfo
	}{
		applicationName, e.Routes(),
	}
	c.Status(http.StatusOK)
	c.Header("Content-type", "text/html; charset=utf-8")
	_ = t.Execute(c.Writer, templateVariables)
}

const adminIndexPageTemplate = `
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
	<th class='home-data'>Description</th>
</thead>
<tbody>
{{range .Routes}}
	<tr class='home-row'>
	<td class='home-data'>{{ .Method }}</td>
	<td class='home-data'><a href='{{ .Path }}'>{{ .Path }}</a></td>
	<td class='home-data'></td>
</tr>
{{ end }}
</tbody>
</table>
</body>
`
