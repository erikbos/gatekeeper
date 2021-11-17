// showDevelopersPage pretty prints all developers and developer apps
package status

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

func (s *Status) ShowDevelopers(c *gin.Context) {

	organizations, err := s.service.Organization.GetAll()
	if err != nil {
		webadmin.JSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}

	c.Header(contentType, contentTypeHTML)
	c.Status(http.StatusOK)
	fmt.Fprint(c.Writer, pageHeading("Developers"))

	for _, organization := range organizations {
		s.ShowDevelopersInOrganization(c, organization.Name)
	}
}

func (s *Status) ShowDevelopersInOrganization(c *gin.Context, organizationName string) {

	fmt.Fprintf(c.Writer, "<h1>Organization: %s</h1>\n", organizationName)

	developers, err := s.service.Developer.GetAll(organizationName)
	if err != nil {
		webadmin.JSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}

	type AppEntry struct {
		App  types.DeveloperApp
		Keys types.Keys
	}
	type AllApps map[string][]AppEntry
	apps := make(AllApps)

	for _, developer := range developers {
		appDetails := make([]AppEntry, 0, 10)

		for _, appName := range developer.Apps {
			app, err := s.service.DeveloperApp.GetByName(string(organizationName), developer.Email, appName)
			if err != nil {
				webadmin.JSONMessage(c, http.StatusServiceUnavailable, err)
				return
			}
			keys, err := s.service.Key.GetByDeveloperAppID(string(organizationName), app.AppID)
			if err != nil {
				webadmin.JSONMessage(c, http.StatusServiceUnavailable, err)
				return
			}
			appDetails = append(appDetails, AppEntry{
				App:  *app,
				Keys: keys,
			})
		}
		apps[developer.Email] = appDetails
	}

	wholePageTemplate := pageHeading("Developer overview") + templateDeveloper
	templateEngine, templateError := template.New("page").
		Funcs(embeddedTemplateFunctions()).Parse(wholePageTemplate)
	if templateError != nil {
		webadmin.JSONMessage(c, http.StatusServiceUnavailable, templateError)
		return
	}
	templateVariables := struct {
		Developers types.Developers
		Apps       AllApps
	}{
		Developers: developers,
		Apps:       apps,
	}
	c.Header(contentType, contentTypeHTML)
	c.Status(http.StatusOK)
	if err := templateEngine.Execute(c.Writer, templateVariables); err != nil {
		_ = c.Error(err)
	}
}

const templateDeveloper string = `
<body>

{{/* We put these in vars to be able to do nested ranges */}}
{{$developers := .Developers}}
{{$apps := .Apps}}

<h1>Developers</h1>
<table border=1>
<tr>
<th>Developer</th>
<th>Application</th>
<th>Keys</th>
<th>Lastmodified</th>
</tr>

{{range $developer := $developers}}
<tr>
<td><a href="/v1/developers/{{$developer.Email}}">{{$developer.Email}}</a>

<ul>
{{range $attribute := $developer.Attributes}}
<li>{{$attribute.Name}} = {{$attribute.Value}}</li>
{{end}}
</ul>

</td>
{{range $app := index $apps $developer.Email}}
<td>
<a href="/v1/developers/{{$developer.Email}}/apps/{{$app.App.Name}}">{{$app.App.Name}}</a>

<ul>
{{range $attribute := $app.App.Attributes}}
<li>{{$attribute.Name}} = {{$attribute.Value}}</li>
{{end}}
</ul>
</td>

<td>
<table>
<tr>
<th>consumer key</th>
<th>consumer secret</th>
<th>products</th>
</tr>
{{range $key := index $app.Keys}}
<tr>
<td><a href="/v1/developers/{{$developer.Email}}/apps/{{$app.App.Name}}/keys/{{$key.ConsumerKey}}">{{$key.ConsumerKey}}</a>
<td>{{$key.ConsumerSecret}}</td>
<td>
<ul>
{{range $product := $key.APIProducts}}
<li><a href="/v1/apiproducts/{{$product.Apiproduct}}">{{$product.Apiproduct}}</a> ({{$product.Status}})
{{end}}
</ul>
</td>
</tr>
{{end}}
</table>

</td>
{{end}}

<td>{{$developer.LastModifiedAt | ISO8601}} <br> {{$developer.LastModifiedBy}}</td>
</tr>
{{end}}

</table>
</body>
`
