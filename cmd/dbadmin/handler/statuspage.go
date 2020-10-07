package handler

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	showHTTPForwardingPath = "show/http_forwarding"
	showUserRolesPath      = "show/user_role"

	contentType     = "content-type"
	contentTypeYAML = "text/yaml; charset=utf-8"
	contentTypeHTML = "text/html; charset=utf-8"
)

// showHTTPForwardingPage pretty prints the current forwarding table from database
func (h *Handler) showHTTPForwardingPage(c *gin.Context) {

	// Retrieve all configuration entities
	listeners, err := h.service.Listener.GetAll()
	if err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	routes, err := h.service.Route.GetAll()
	if err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	clusters, err := h.service.Cluster.GetAll()
	if err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	apiproducts := types.NullAPIProducts
	// apiproducts, err := h.service.APIProduct.GetAll()
	// if err != nil {
	// 	returnJSONMessage(c, http.StatusServiceUnavailable, err)
	// 	return
	// }
	// Order all entries to make page more readable
	listeners.Sort()
	routes.Sort()
	clusters.Sort()

	wholePageTemplate := pageHeading("HTTP forwarding configuration") + templateHTTPForwarding
	templateEngine, templateError := template.New("page").
		Funcs(embeddedTemplateFunctions()).Parse(wholePageTemplate)
	if templateError != nil {
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
	c.Header(contentType, contentTypeHTML)
	c.Status(http.StatusOK)
	_ = templateEngine.Execute(c.Writer, templateVariables)
}

const templateHTTPForwarding string = `
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
{{if (eq $attribute.Name "TLSCertificate" "TLSCertificateKey" "AccessLogFileFields")}}
{{$attribute.Name}} = {{$attribute | PrettyPrint}}
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
<th>Attributes</th>
<th>Lastmodified</th>
</tr>

{{range $c := $clusters}}
<tr>
<td><a href="/v1/clusters/{{$c.Name}}">{{$c.Name}}</a>
<td>{{$c.DisplayName}}</td>
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
<th>Policies</th>
<th>Attributes</th>
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
<td>{{$a.Policies | OrderedList}}</td>
<td>
<ul>
{{range $attribute := .Attributes}}
<li>{{$attribute.Name}} = {{$attribute.Value}}</li>
{{end}}
</ul>
</td>
<td>{{$a.LastmodifiedAt | ISO8601}} <br> {{$a.LastmodifiedBy}}</td>
</tr>
{{end}}
</table>
</bod>
`

// showUserRolesPath pretty prints user and roles from database
func (h *Handler) showUserRolePage(c *gin.Context) {

	// Retrieve all user entities
	users, err := h.service.User.GetAll()
	if err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	roles, err := h.service.Role.GetAll()
	if err != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	// Order all entries to make page more readable
	users.Sort()
	roles.Sort()

	wholePageTemplate := pageHeading("User and Roles") + pageTemplateUsersAndRoles
	templateEngine, templateError := template.New("page").
		Funcs(embeddedTemplateFunctions()).Parse(wholePageTemplate)
	if templateError != nil {
		returnJSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	templateVariables := struct {
		Users types.Users
		Roles types.Roles
	}{
		users, roles,
	}
	c.Header(contentType, contentTypeHTML)
	c.Status(http.StatusOK)
	_ = templateEngine.Execute(c.Writer, templateVariables)
}

const pageTemplateUsersAndRoles string = `
<body>

{{/* We put these in vars to be able to do nested ranges */}}
{{$users := .Users}}
{{$roles := .Roles}}

<h1>Users</h1>
<table border=1>

<tr>
<th>Name</th>
<th>DisplayName</th>
<th>Status</th>
<th>Roles</th>
<th>CreatedBy</th>
<th>CreatedAt</th>
<th>LastmodifiedBy</th>
<th>LastmodifiedAt</th>
</tr>

{{range $user := $users}}
<tr>
<td><a href="/v1/users/{{$user.Name}}">{{$user.Name}}</a>
<td>{{$user.DisplayName}}</td>
<td>{{$user.Status}}</td>
<td><ul>{{range $role := $user.Roles}}<li>{{$role}}</li>{{end}}</ul></td>
<td>{{$user.CreatedBy}}</td>
<td>{{$user.CreatedAt | ISO8601}}</td>
<td>{{$user.LastmodifiedBy}}</td>
<td>{{$user.LastmodifiedAt | ISO8601}}</td>
</tr>
{{end}}

</table>

<h1>Roles</h1>
<table border=1>

<tr>
<th>Name</th>
<th>DisplayName</th>
<th>Allowed</th>
<th>CreatedBy</th>
<th>CreatedAt</th>
<th>LastmodifiedBy</th>
<th>LastmodifiedAt</th>
</tr>

{{range $role := $roles}}
<tr>
<td><a href="/v1/roles/{{$role.Name}}">{{$role.Name}}</a>
<td>{{$role.DisplayName}}</td>

<td>
<table>
<tr><th>Methods</th><th>Paths</th></tr>
{{range $allow := $role.Allows}}
<tr>
<td><ul>{{range $methods := $allow.Methods}}<li>{{$methods}}</li>{{end}}</ul></td>
<td><ul>{{range $paths := $allow.Paths}}<li>{{$paths}}</li>{{end}}</ul></td>
{{end}}
</table>
</td>

<td>{{$role.CreatedBy}}</td>
<td>{{$role.CreatedAt | ISO8601}}</td>
<td>{{$role.LastmodifiedBy}}</td>
<td>{{$role.LastmodifiedAt | ISO8601}}</td>
</tr>
{{end}}

</table>
</body>
`

func pageHeading(title string) string {

	return fmt.Sprintf(pageTemplateHeading, title)
}

const pageTemplateHeading string = `
<!DOCTYPE html>
<html>
<head>
<title>%s</title>
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
`

// embeddedTemplateFunctions returns configuration for all custom template functions
func embeddedTemplateFunctions() template.FuncMap {

	// Supporting functions embedded in template invoked with "{{value | <functioname}}"
	return template.FuncMap{
		"ISO8601":     shared.TimeMillisecondsToString,
		"OrderedList": HMTLOrderedList,
		"PrettyPrint": prettyPrintAttribute,
	}
}

// HMTLOrderedList prints a comma separated string as HTML ordered and numbered list
func HMTLOrderedList(stringToSplit string) string {

	out := "<ol>"
	for _, value := range strings.Split(stringToSplit, ",") {
		out += fmt.Sprintf("<li>%s</li>", strings.TrimSpace(value))
	}
	out += "</ol>\n"

	return out
}

// prettyPrintAttribute prints summary of length attribute values
func prettyPrintAttribute(attribute types.Attribute) string {

	switch attribute.Name {
	case types.AttributeAccessLogFileFields:
		return "[log fields]"
	case types.AttributeTLSCertificateKey:
		// We never shown private key itself
		return "[redacted]"
	case types.AttributeTLSCertificate:
		return certDetails([]byte(attribute.Value))
	}
	return "unknown"
}

// certDetails prints summary of a few key public certificate attributes
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

// returnJSONMessage returns an public error message, it should not leak any internal details
func returnJSONMessage(c *gin.Context, statusCode int, errorMessage error) {

	c.IndentedJSON(statusCode,
		gin.H{
			"message": fmt.Sprintf("%s", errorMessage),
		})
}
