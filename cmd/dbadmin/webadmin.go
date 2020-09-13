package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

type webAdminConfig struct {
	Listen  string `yaml:"listen"`  // Address and port to listen
	IPACL   string `yaml:"ipacl"`   // ip accesslist (e.g. "10.0.0.0/8,192.168.0.0/16")
	LogFile string `yaml:"logfile"` // File for writing admin access logs
	TLS     struct {
		certFile string `yaml:"certfile"` // TLS certifcate file
		keyFile  string `yaml:"keyfile"`  // TLS certifcate key file
	} `yaml:"tls"`
}

// StartWebAdminServer starts the admin web UI
func StartWebAdminServer(s *server) {

	if logFile, err := os.Create(s.config.WebAdmin.LogFile); err == nil {
		gin.DefaultWriter = io.MultiWriter(logFile)
	}

	// disable debuglogging
	gin.SetMode(gin.ReleaseMode)

	// Enable strict checking of posted JSON fields
	gin.EnableJsonDecoderDisallowUnknownFields()

	s.ginEngine = gin.New()
	s.ginEngine.Use(gin.LoggerWithFormatter(shared.LogHTTPRequest))
	s.ginEngine.Use(shared.AddRequestID())
	s.ginEngine.Use(shared.WebAdminCheckIPACL(s.config.WebAdmin.IPACL))

	s.registerOrganizationRoutes(s.ginEngine)
	s.registerDeveloperRoutes(s.ginEngine)
	s.registerDeveloperAppRoutes(s.ginEngine)
	s.registerCredentialRoutes(s.ginEngine)
	s.registerAPIProductRoutes(s.ginEngine)
	s.registerClusterRoutes(s.ginEngine)
	s.registerRouteRoutes(s.ginEngine)
	s.registerListenerRoutes(s.ginEngine)

	s.ginEngine.GET("/", s.ShowWebAdminHomePage)
	s.ginEngine.GET(shared.LivenessCheckPath, shared.LivenessProbe)
	s.ginEngine.GET(shared.ReadinessCheckPath, s.readiness.ReadinessProbe)
	s.ginEngine.GET(shared.MetricsPath, gin.WrapH(promhttp.Handler()))
	s.ginEngine.GET(shared.ConfigDumpPath, s.showConfiguration)
	s.ginEngine.GET("show_http_forwarding", s.showHTTPForwarding)

	log.Info("Webadmin listening on ", s.config.WebAdmin.Listen)
	if s.config.WebAdmin.TLS.certFile != "" &&
		s.config.WebAdmin.TLS.keyFile != "" {

		log.Fatal(s.ginEngine.RunTLS(s.config.WebAdmin.Listen,
			s.config.WebAdmin.TLS.certFile, s.config.WebAdmin.TLS.keyFile))
	}
	log.Fatal(s.ginEngine.Run(s.config.WebAdmin.Listen))
}

// ShowWebAdminHomePage shows home page
func (s *server) ShowWebAdminHomePage(c *gin.Context) {
	// FIXME feels like hack, is there a better way to pass gin engine context?
	shared.ShowIndexPage(c, s.ginEngine, applicationName)
}

// showConfiguration pretty prints the active configuration
func (s *server) showConfiguration(c *gin.Context) {

	c.Header("Content-type", "text/yaml")
	c.String(http.StatusOK, fmt.Sprint(s.config))
}

// showForwarding pretty prints the current forwarding table from database
func (s *server) showHTTPForwarding(c *gin.Context) {

	// Retrieve all configuration entities
	listeners, err := s.db.Listener.GetAll()
	if err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	routes, err := s.db.Route.GetAll()
	if err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	clusters, err := s.db.Cluster.GetAll()
	if err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	apiproducts, err := s.db.APIProduct.GetAll()
	if err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}

	// Supporting functions embedded in template with {{value | <functioname}}
	templateFunctions := template.FuncMap{
		"ISO8601":            shared.TimeMillisecondsToString,
		"OrderedList":        HMTLOrderedList,
		"CertificateDetails": HTMLCertificateDetails,
	}

	// Order all entries to make the overview more readable
	listeners.Sort()
	routes.Sort()
	clusters.Sort()

	t, err := template.New("page").Funcs(templateFunctions).Parse(pageTemplate)
	if err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	templateVariables := struct {
		Listeners   types.Listeners
		Routes      types.Routes
		Clusters    types.Clusters
		APIProducts types.APIProducts
	}{
		listeners, routes, clusters, apiproducts,
	}
	c.Header("Content-type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	_ = t.Execute(c.Writer, templateVariables)
}

const pageTemplate string = `
<!DOCTYPE html>
<html>
<head>
<title>HTTP forwarding configuration</title>
<link rel="icon" type="image/svg+xml" href="data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBzdGFuZGFsb25lPSJubyI/Pgo8IURPQ1RZUEUgc3ZnIFBVQkxJQyAiLS8vVzNDLy9EVEQgU1ZHIDIwMDEwOTA0Ly9FTiIKICJodHRwOi8vd3d3LnczLm9yZy9UUi8yMDAxL1JFQy1TVkctMjAwMTA5MDQvRFREL3N2ZzEwLmR0ZCI+CjxzdmcgdmVyc2lvbj0iMS4wIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciCiB3aWR0aD0iNDguMDAwMDAwcHQiIGhlaWdodD0iNDguMDAwMDAwcHQiIHZpZXdCb3g9IjAgMCA0OC4wMDAwMDAgNDguMDAwMDAwIgogcHJlc2VydmVBc3BlY3RSYXRpbz0ieE1pZFlNaWQgbWVldCI+CjxtZXRhZGF0YT4KQ3JlYXRlZCBieSBwb3RyYWNlIDEuMTUsIHdyaXR0ZW4gYnkgUGV0ZXIgU2VsaW5nZXIgMjAwMS0yMDE3CjwvbWV0YWRhdGE+CjxnIHRyYW5zZm9ybT0idHJhbnNsYXRlKDAuMDAwMDAwLDQ4LjAwMDAwMCkgc2NhbGUoMC4xMDAwMDAsLTAuMTAwMDAwKSIKZmlsbD0iIzAwMDAwMCIgc3Ryb2tlPSJub25lIj4KPHBhdGggZD0iTTIwOSAzOTIgYy0xMTYgLTc0IC0xODQgLTIzNSAtMTI5IC0zMDUgMTEgLTE0IDI5IC0yOSA0MCAtMzIgMzIgLTEwCjk0IDEzIDEyNSA0NiBsMzAgMzEgNSAtMzkgYzQgLTMzIDkgLTM4IDMzIC00MSAyMiAtMyAzNyA2IDY3IDM2IDIyIDIyIDQwIDQ4CjQwIDU4IDAgMTUgLTYgMTIgLTI5IC0xNSAtMTUgLTE5IC0zMiAtMzIgLTM4IC0zMCAtNiAyIDcgNjMgMzMgMTU0IDI0IDgzIDQ0CjE1MyA0NCAxNTggMCAxNCAtNTkgNyAtNjggLTkgLTggLTE0IC0xMCAtMTQgLTIxIDAgLTIxIDI2IC04MSAyMCAtMTMyIC0xMnoKbTExOCAtMjMgYzMzIC02OCAtNzkgLTI2OSAtMTQ5IC0yNjkgLTg4IDAgLTM1IDIwNiA3MSAyNzggNDIgMjcgNjIgMjUgNzggLTl6Ii8+CjwvZz4KPC9zdmc+Cg==">
<style>
table {
	font-family: sans-serif;
	font-size: medium;
	border-collapse: collapse;
	text-align: left;
}
th {
	border: 1px solid #000000;
	text-align: left;
	padding: 8px;
}
tr:nth-child(even) {
	background-color: #dddddd;
	border: 1px solid #dddddd;
}
td {
	border: 1px solid #000000;
	text-align: left;
	padding: 8px;
}
ul {
	list-style-type: none;
	margin: 0px;
	padding: 0px;
}
ol {
	padding: 15px;
}
</style>
</head>
<body>

{{/* We put these in vars to be able to do nested ranges */}}
{{$listeners := .Listeners}}
{{$routes := .Routes}}
{{$clusters := .Clusters}}
{{$apiproducts := .APIProducts}}

<h1>Listeners</h1>
<table border=1>
<tr>
<th>Organization</th>
<th>Name</th>
<th>DisplayName</th>
<th>Port</th>
<th>VirtualHosts</th>
<th>Attributes</th>
<th>Policies</th>
<th>RouteGroup</th>
<th>Lastmodified</th>
</tr>

{{range $listener := $listeners}}
<tr>
<td><a href="/v1/organizations/{{$listener.OrganizationName}}">{{$listener.OrganizationName}}</a>
<td><a href="/v1/listeners/{{$listener.Name}}">{{$listener.Name}}</a>
<td>{{$listener.DisplayName}}</td>
<td>{{$listener.Port}}</td>
<td>
<ul>
{{range $hostname := $listener.VirtualHosts}}
<li>{{$hostname}}</li>
{{end}}
</ul>
</td>
<td>
<ul>
{{range $attribute := $listener.Attributes}}
<li>
{{if or (eq $attribute.Name "TLSCertificate") (eq $attribute.Name "TLSCertificateKey")}}
{{$attribute.Name}} = {{$attribute | CertificateDetails}}
{{else}}
{{$attribute.Name}} = {{$attribute.Value}}
{{end}}
</li>
{{end}}
</ul>
</td>
<td>{{$listener.Policies | OrderedList}}</td>
<td>{{$listener.RouteGroup}}</td>
<td>{{$listener.LastmodifiedAt | ISO8601}} <br> {{$listener.LastmodifiedBy}}</td>
</tr>
{{end}}

</table>



<h1>Routes</h1>
<table border=1>
<tr>
<th>RouteName</th>
<th>DisplayName</th>
<th>RouteGroup</th>
<th>Path</th>
<th>PathType</th>
<th>Attributes</th>
<th>Lastmodified</th>
</tr>

{{range $r := $routes}}
<tr>
<td><a href="/v1/routes/{{$r.Name}}">{{$r.Name}}</a>
<td>{{$r.DisplayName}}</td>
<td>{{$r.RouteGroup}}</td>
<td>{{$r.Path}}</td>
<td>{{$r.PathType}}</td>
<td>
<ul>
{{range $attribute := $r.Attributes}}

<li>
{{if eq $attribute.Name "Cluster"}}
{{$attribute.Name}} = <a href="/v1/clusters/{{$attribute.Value}}">{{$attribute.Value}}</a>
{{else}}
{{$attribute.Name}} = {{$attribute.Value}}
{{end}}
</li>

{{end}}
</ul>
</td>
<td>{{$r.LastmodifiedAt | ISO8601}} <br> {{$r.LastmodifiedBy}}</td>
</tr>
{{end}}
</table>



<h1>Clusters</h1>
<table border=1>
<tr>
<th>ClusterName</th>
<th>DisplayName</th>
<th>HostName</th>
<th>Port</th>
<th>Attributes</th>
<th>Lastmodified</th>
</tr>

{{range $c := $clusters}}
<tr>
<td><a href="/v1/clusters/{{$c.Name}}">{{$c.Name}}</a>
<td>{{$c.DisplayName}}</td>
<td>{{$c.HostName}}</td>
<td>{{$c.Port}}</td>
<td>
<ul>
{{range $attribute := $c.Attributes}}
<li>{{$attribute.Name}} = {{$attribute.Value}}</li>
{{end}}
</ul>
</td>
<td>{{$c.LastmodifiedAt | ISO8601}} <br> {{$c.LastmodifiedBy}}</td>
</tr>
{{end}}
</table>



<h1>API Products</h1>
<table border=1>
<tr>
<th>Organization</th>
<th>ProductName</th>
<th>DisplayName</th>
<th>Description</th>
<th>RouteGroup</th>
<th>Paths</th>
<th>Attributes</th>
<th>Policies</th>
<th>Lastmodified</th>
</tr>

{{range $a := $apiproducts}}
<tr>
<td>{{$a.OrganizationName}}</td>
<td><a href="/v1/organizations/{{$a.OrganizationName}}/apiproducts/{{$a.Name}}">{{$a.Name}}</a>
<td>{{$a.DisplayName}}</td>
<td>{{$a.Description}}</td>
<td>{{$a.RouteGroup}}</td>
<td>
<ul>
{{range $path := $a.Paths}}
<li>{{$path}}</li>
{{end}}
</ul>
</td>
<td>
<ul>
{{range $attribute := .Attributes}}
<li>{{$attribute.Name}} = {{$attribute.Value}}</li>
{{end}}
</ul>
</td>
<td>{{$a.Policies | OrderedList}}</td>
<td>{{$a.LastmodifiedAt | ISO8601}} <br> {{$a.LastmodifiedBy}}</td>
</tr>
{{end}}
</table>
`

// HMTLOrderedList prints a comma separated string as HTML ordered and numbered list
func HMTLOrderedList(stringToSplit string) string {

	out := "<ol>"
	for _, value := range strings.Split(stringToSplit, ",") {
		out += fmt.Sprintf("<li>%s</li>", strings.TrimSpace(value))
	}
	out += "</ol>\n"

	return out
}

// HTMLCertificateDetails prints summary of certificate attributes
func HTMLCertificateDetails(attribute types.Attribute) string {

	switch attribute.Name {
	case "TLSCertificateKey":
		// We never shown a private key
		return "[redacted]"
	case "TLSCertificate":
		return certDetails([]byte(attribute.Value))
	}
	return "unknown"
}

func certDetails(certificate []byte) string {

	block, rest := pem.Decode(certificate)
	if block == nil || len(rest) > 0 {
		return fmt.Sprintf("[Cannot parse '%s' as .pem certificate]", certificate)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Sprintf("[Cannot parse asn.1 data in '%s']", certificate)
	}

	return fmt.Sprintf("[Serial=%s, CN=%s, DNS=%s, NotAfter=%s]",
		cert.SerialNumber.Text(16), cert.Subject.CommonName,
		cert.DNSNames, cert.NotAfter.UTC().Format(time.RFC3339))

}
