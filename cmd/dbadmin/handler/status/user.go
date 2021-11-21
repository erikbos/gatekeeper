package status

import (
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

// showUserRolesPath pretty prints user and roles from database
func (s *Status) ShowUserRole(c *gin.Context) {

	// Retrieve all user entities
	users, err := s.service.User.GetAll()
	if err != nil {
		webadmin.JSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	roles, err := s.service.Role.GetAll()
	if err != nil {
		webadmin.JSONMessage(c, http.StatusServiceUnavailable, err)
		return
	}
	// Order all entries to make page more readable
	users.Sort()
	roles.Sort()

	wholePageTemplate := pageHeading("User and Roles") + pageTemplateUsersAndRoles
	templateEngine, templateError := template.New("page").
		Funcs(embeddedTemplateFunctions()).Parse(wholePageTemplate)
	if templateError != nil {
		webadmin.JSONMessage(c, http.StatusServiceUnavailable, templateError)
		return
	}
	templateVariables := struct {
		Users types.Users
		Roles types.Roles
	}{
		Users: users,
		Roles: roles,
	}
	c.Header(contentType, contentTypeHTML)
	c.Status(http.StatusOK)
	if err := templateEngine.Execute(c.Writer, templateVariables); err != nil {
		_ = c.Error(err)
	}
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
<th>LastModifiedBy</th>
<th>LastModifiedAt</th>
</tr>

{{range $user := $users}}
<tr>
<td><a href="/v1/users/{{$user.Name}}">{{$user.Name}}</a>
<td>{{$user.DisplayName}}</td>
<td>{{$user.Status}}</td>
<td><ul>{{range $role := $user.Roles}}<li><a href="/v1/roles/{{$role}}">{{$role}}</li>{{end}}</ul></td>
<td>{{$user.CreatedBy}}</td>
<td>{{$user.CreatedAt | ISO8601}}</td>
<td>{{$user.LastModifiedBy}}</td>
<td>{{$user.LastModifiedAt | ISO8601}}</td>
</tr>
{{end}}

</table>

<h1>Roles</h1>
<table border=1>

<tr>
<th>Name</th>
<th>DisplayName</th>
<th>Permissions</th>
<th>CreatedBy</th>
<th>CreatedAt</th>
<th>LastModifiedBy</th>
<th>LastModifiedAt</th>
</tr>

{{range $role := $roles}}
<tr>
<td><a href="/v1/roles/{{$role.Name}}">{{$role.Name}}</a>
<td>{{$role.DisplayName}}</td>

<td>
<table>
<tr><th>Methods</th><th>Paths</th></tr>
{{range $allow := $role.Permissions}}
<tr>
<td><ul>{{range $methods := $allow.Methods}}<li>{{$methods}}</li>{{end}}</ul></td>
<td><ul>{{range $paths := $allow.Paths}}<li>{{$paths}}</li>{{end}}</ul></td>
{{end}}
</table>
</td>

<td>{{$role.CreatedBy}}</td>
<td>{{$role.CreatedAt | ISO8601}}</td>
<td>{{$role.LastModifiedBy}}</td>
<td>{{$role.LastModifiedAt | ISO8601}}</td>
</tr>
{{end}}

</table>
</body>
`
