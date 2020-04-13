package shared

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

//ShowIndexPage produces the index page based upon all registered routes
func ShowIndexPage(c *gin.Context, e *gin.Engine, name string) {
	body := fmt.Sprintf(adminIndexHTMLheader, name, name)
	for _, v := range e.Routes() {
		body += fmt.Sprintf(`<tr class='home-row'>
		<td class='home-data'>%s</td>
		<td class='home-data'><a href='%s'>%s</a></td>
		<td class='home-data'>%s</td>
	</tr>`, v.Method, v.Path, v.Path, "")
	}
	body += adminIndexHTMLend

	c.Header("Content-type", "text/html")
	c.String(http.StatusOK, body)
}

const adminIndexHTMLheader = `
<head>
<title>%s</title>
<link rel="icon" type="image/svg+xml" href="data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBzdGFuZGFsb25lPSJubyI/Pgo8IURPQ1RZUEUgc3ZnIFBVQkxJQyAiLS8vVzNDLy9EVEQgU1ZHIDIwMDEwOTA0Ly9FTiIKICJodHRwOi8vd3d3LnczLm9yZy9UUi8yMDAxL1JFQy1TVkctMjAwMTA5MDQvRFREL3N2ZzEwLmR0ZCI+CjxzdmcgdmVyc2lvbj0iMS4wIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciCiB3aWR0aD0iNDguMDAwMDAwcHQiIGhlaWdodD0iNDguMDAwMDAwcHQiIHZpZXdCb3g9IjAgMCA0OC4wMDAwMDAgNDguMDAwMDAwIgogcHJlc2VydmVBc3BlY3RSYXRpbz0ieE1pZFlNaWQgbWVldCI+CjxtZXRhZGF0YT4KQ3JlYXRlZCBieSBwb3RyYWNlIDEuMTUsIHdyaXR0ZW4gYnkgUGV0ZXIgU2VsaW5nZXIgMjAwMS0yMDE3CjwvbWV0YWRhdGE+CjxnIHRyYW5zZm9ybT0idHJhbnNsYXRlKDAuMDAwMDAwLDQ4LjAwMDAwMCkgc2NhbGUoMC4xMDAwMDAsLTAuMTAwMDAwKSIKZmlsbD0iIzAwMDAwMCIgc3Ryb2tlPSJub25lIj4KPHBhdGggZD0iTTIwOSAzOTIgYy0xMTYgLTc0IC0xODQgLTIzNSAtMTI5IC0zMDUgMTEgLTE0IDI5IC0yOSA0MCAtMzIgMzIgLTEwCjk0IDEzIDEyNSA0NiBsMzAgMzEgNSAtMzkgYzQgLTMzIDkgLTM4IDMzIC00MSAyMiAtMyAzNyA2IDY3IDM2IDIyIDIyIDQwIDQ4CjQwIDU4IDAgMTUgLTYgMTIgLTI5IC0xNSAtMTUgLTE5IC0zMiAtMzIgLTM4IC0zMCAtNiAyIDcgNjMgMzMgMTU0IDI0IDgzIDQ0CjE1MyA0NCAxNTggMCAxNCAtNTkgNyAtNjggLTkgLTggLTE0IC0xMCAtMTQgLTIxIDAgLTIxIDI2IC04MSAyMCAtMTMyIC0xMnoKbTExOCAtMjMgYzMzIC02OCAtNzkgLTI2OSAtMTQ5IC0yNjkgLTg4IDAgLTM1IDIwNiA3MSAyNzggNDIgMjcgNjIgMjUgNzggLTl6Ii8+CjwvZz4KPC9zdmc+Cg==">
<style>
.home-table {
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
<h1>%s</h1>
<table class='home-table'>
<thead>
	<th class='home-data'>Method</th>
	<th class='home-data'>Path</th>
	<th class='home-data'>Description</th>
</thead>
<tbody>
`

const adminIndexHTMLend = `
</tbody>
</table>
</body>
`
