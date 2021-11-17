package status

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

func (s *Status) ShowAPIProducts(c *gin.Context) {

	organizations, err := s.service.Organization.GetAll()
	if err != nil {
		webadmin.JSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}

	c.Header(contentType, contentTypeHTML)
	c.Status(http.StatusOK)
	fmt.Fprint(c.Writer, pageHeading("API Products"))

	for _, organization := range organizations {
		s.ShowAPIProductsOrganization(c, organization.Name)
	}
}

func (s *Status) ShowAPIProductsOrganization(c *gin.Context, organization string) {

	fmt.Fprintf(c.Writer, "<h1>Organization: %s</h1>\n", organization)

	apiproducts, err := s.service.APIProduct.GetAll(organization)
	if err != nil {
		webadmin.JSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}

	templateEngine, templateError := template.New("page").
		Funcs(embeddedTemplateFunctions()).Parse(templateAPIProducts)
	if templateError != nil {
		webadmin.JSONMessage(c, http.StatusServiceUnavailable, templateError)
		return
	}
	templateVariables := struct {
		APIProducts types.APIProducts
	}{
		APIProducts: apiproducts,
	}
	if err := templateEngine.Execute(c.Writer, templateVariables); err != nil {
		_ = c.Error(err)
	}
}

const templateAPIProducts string = `
{{/* We put these in vars to be able to do nested ranges */}}
{{$apiproducts := .APIProducts}}

<h1>API Products</h1>
<table border=1>
<tr>
<th>ProductName</th>
<th>DisplayName</th>
<th>Description</th>
<th>RouteGroup</th>
<th>APIResources</th>
<th>Policies</th>
<th>Attributes</th>
<th>Lastmodified</th>
</tr>

{{range $a := $apiproducts}}
<tr>
<td><a href="/v1/apiproducts/{{$a.Name}}">{{$a.Name}}</a>
<td>{{$a.DisplayName}}</td>
<td>{{$a.Description}}</td>
<td>{{$a.RouteGroup}}</td>
<td>
<ul>
{{range $apiresource := $a.APIResources}}
<li>{{$apiresource}}</li>
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
<td>{{$a.LastModifiedAt | ISO8601}} <br> {{$a.LastModifiedBy}}</td>
</tr>
{{end}}
</table>
</body>
`
