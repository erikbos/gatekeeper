package status

import (
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

// showHTTPForwardingPage pretty prints the current forwarding table from database
func (s *Status) ShowHTTPForwarding(c *gin.Context) {

	// Retrieve all configuration entities
	listeners, err := s.service.Listener.GetAll()
	if err != nil {
		webadmin.JSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	routes, err := s.service.Route.GetAll()
	if err != nil {
		webadmin.JSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	clusters, err := s.service.Cluster.GetAll()
	if err != nil {
		webadmin.JSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	// Order all entries to make page more readable
	listeners.Sort()
	routes.Sort()
	clusters.Sort()

	wholePageTemplate := pageHeading("HTTP forwarding configuration") + templateHTTPForwarding
	templateEngine, templateError := template.New("page").
		Funcs(embeddedTemplateFunctions()).Parse(wholePageTemplate)
	if templateError != nil {
		webadmin.JSONMessage(c, http.StatusServiceUnavailable, templateError)
		return
	}
	templateVariables := struct {
		Listeners types.Listeners
		Routes    types.Routes
		Clusters  types.Clusters
	}{
		Listeners: listeners,
		Routes:    routes,
		Clusters:  clusters,
	}
	c.Header(contentType, contentTypeHTML)
	c.Status(http.StatusOK)
	if err := templateEngine.Execute(c.Writer, templateVariables); err != nil {
		_ = c.Error(err)
	}
}

const templateHTTPForwarding string = `
<body>

{{/* We put these in vars to be able to do nested ranges */}}
{{$listeners := .Listeners}}
{{$routes := .Routes}}
{{$clusters := .Clusters}}

<h1>Listeners</h1>
<table border=1>
<tr>
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
<td>{{$listener.LastModifiedAt | ISO8601}} <br> {{$listener.LastModifiedBy}}</td>
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
<td>{{$r.LastModifiedAt | ISO8601}} <br> {{$r.LastModifiedBy}}</td>
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
<td>{{$c.LastModifiedAt | ISO8601}} <br> {{$c.LastModifiedBy}}</td>
</tr>
{{end}}
</table>

</body>
`
