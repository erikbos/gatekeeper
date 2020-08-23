package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

type webAdminConfig struct {
	Listen      string `yaml:"listen"`      // Address and port to listen
	IPACL       string `yaml:"ipacl"`       // ip accesslist (e.g. "10.0.0.0/8,192.168.0.0/16")
	LogFileName string `yaml:"logfilename"` // Filename for writing admin access logs
}

// StartWebAdminServer starts the admin web UI
func StartWebAdminServer(s *server) {

	if logFile, err := os.Create(s.config.WebAdmin.LogFileName); err == nil {
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
	s.registerVirtualHostRoutes(s.ginEngine)

	s.ginEngine.GET("/", s.ShowWebAdminHomePage)
	s.ginEngine.GET(shared.LivenessCheckPath, shared.LivenessProbe)
	s.ginEngine.GET(shared.ReadinessCheckPath, s.readiness.ReadinessProbe)
	s.ginEngine.GET(shared.MetricsPath, gin.WrapH(promhttp.Handler()))
	s.ginEngine.GET(shared.ConfigDumpPath, s.showConfiguration)
	s.ginEngine.GET("show_http_forwarding", s.showHTTPForwarding)

	log.Info("Webadmin listening on ", s.config.WebAdmin.Listen)
	if err := s.ginEngine.Run(s.config.WebAdmin.Listen); err != nil {
		log.Fatal(err)
	}
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

	virtualhosts, err := s.db.Virtualhost.GetAll()
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

	templateFunctions := template.FuncMap{
		"ISO8601": shared.TimeMillisecondsToString,
		"OrderedList": func(stringToSplit string) string {
			out := "<ol>"
			for _, policy := range strings.Split(stringToSplit, ",") {
				out += fmt.Sprintf("<li>%s</li>", strings.TrimSpace(policy))
			}
			out += "</ol>\n"
			return out
		},
	}

	t, err := template.New("page").Funcs(templateFunctions).Parse(pageTemplate)
	if err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	templateVariables := struct {
		V []shared.VirtualHost
		R []shared.Route
		C []shared.Cluster
		P []shared.APIProduct
	}{
		virtualhosts, routes, clusters, apiproducts,
	}
	c.Header("Content-type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	t.Execute(c.Writer, templateVariables)
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

<h1>Virtualhosts</h1>

<table border=1>
<tr>
<th>Organization</th>
<th>Name</th>
<th>DisplayName</th>
<th>Port</th>
<th>Vhosts</th>
<th>Attributes</th>
<th>Policies</th>
<th>RouteGroup</th>
<th>LastmodifiedAt</th>
<th>LastmodifiedBy</th>
</tr>

{{range .V }}
<tr>
<td><a href="/v1/organizations/{{ .OrganizationName }}">{{ .OrganizationName }}</a>
<td><a href="/v1/virtualhosts/{{ .Name}}">{{ .Name }}</a>
<td>{{ .DisplayName }}</td>
<td>{{ .Port }}</td>
<td>
<ul>
{{ range $vhost := .VirtualHosts }}
<li>{{ $vhost }}</li>
{{ end }}
</ul>
</td>
<td>
<ul>
{{ range $attribute := .Attributes }}
<li>
{{ if or (eq $attribute.Name "TLSCertificate") (eq $attribute.Name "TLSCertificateKey") }}
{{ $attribute.Name }} = [redacted]
{{ else }}
{{ $attribute.Name }} = {{ $attribute.Value }}
{{ end }}
</li>
{{ end }}
</ul>
</td>
<td>{{ OrderedList .Policies }}</td>
<td>{{ .RouteGroup }}</td>
<td>{{ ISO8601 .LastmodifiedAt }}</td>
<td>{{ .LastmodifiedBy }}</td>
</tr>
{{ end }}

</table>

<h1>Routes</h1>

<table border=1>
<tr>
<th>RouteName</th>
<th>DisplayName</th>
<th>RouteGroup</th>
<th>Path</th>
<th>PathType</th>
<th>Cluster</th>
<th>Attributes</th>
<th>LastmodifiedAt</th>
<th>LastmodifiedBy</th>
</tr>

{{range .R }}
<tr>
<td><a href="/v1/routes/{{ .Name }}">{{ .Name }}</a>
<td>{{ .DisplayName }}</td>
<td>{{ .RouteGroup }}</td>
<td>{{ .Path }}</td>
<td>{{ .PathType }}</td>
<td><a href="/v1/clusters/{{ .Cluster }}">{{ .Cluster }}</a>
<td>
<ul>
{{range $attribute := .Attributes }}
<li>{{ $attribute.Name }} = {{ $attribute.Value }}</li>
{{end}}
</ul>
</td>
<td>{{ ISO8601 .LastmodifiedAt }}</td>
<td>{{ .LastmodifiedBy }}</td>
</tr>
{{ end }}
</table>

<h1>Clusters</h1>

<table border=1>
<tr>
<th>ClusterName</th>
<th>DisplayName</th>
<th>HostName</th>
<th>Port</th>
<th>Attributes</th>
<th>LastmodifiedAt</th>
<th>LastmodifiedBy</th>
</tr>

{{ range .C }}
<tr>
<td><a href="/v1/clusters/{{ .Name }}">{{ .Name }}</a>
<td>{{ .DisplayName }}</td>
<td>{{ .HostName }}</td>
<td>{{ .Port }}</td>
<td>
<ul>
{{ range $attribute := .Attributes }}
<li>{{ $attribute.Name }} = {{ $attribute.Value }}</li>
{{ end }}
</ul>
</td>
<td>{{ ISO8601 .LastmodifiedAt}}</td>
<td>{{ .LastmodifiedBy}}</td>
</tr>
{{ end }}
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
<th>LastmodifiedAt</th>
<th>LastmodifiedBy</th>
</tr>

{{range .P}}
<tr>
<td>{{ .OrganizationName }}</td>
<td><a href="/v1/organizations/{{ .OrganizationName }}/apiproducts/{{ .Name }}">{{ .Name }}</a>
<td>{{ .DisplayName }}</td>
<td>{{ .Description }}</td>
<td>{{ .RouteGroup }}</td>
<td>
<ul>
{{range $path := .Paths}}
<li>{{ $path }}</li>
{{ end }}
</ul>
</td>
<td>
<ul>
{{ range $attribute := .Attributes }}
<li>{{ $attribute.Name }} = {{ $attribute.Value }}</li>
{{ end }}
</ul>
</td>
<td>{{ OrderedList .Policies }}</td>
<td>{{ ISO8601 .LastmodifiedAt }}</td>
<td>{{ .LastmodifiedBy }}</td>
</tr>
{{end}}
</table>
`
