package statuspage

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/cmd/managementserver/service"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	contentType     = "content-type"
	contentTypeHTML = "text/html; charset=utf-8"
)

type Status struct {
	service *service.Service
}

func New(s *service.Service) Status {

	return Status{
		service: s,
	}
}

func (s *Status) RegisterRoutes(c *gin.Engine) {

	c.GET("/show/apiproducts", s.ShowAPIProducts)
	c.GET("/show/http_forwarding", s.ShowHTTPForwarding)
	c.GET("/show/user_role", s.ShowUserRole)
	c.GET("/show/developer", s.ShowDevelopers)
}

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
