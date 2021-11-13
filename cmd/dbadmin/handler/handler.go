package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/dbadmin/service"
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// Generate REST API handlers from OpenAPI specification
//go:generate oapi-codegen -package handler -generate types,gin -o dbadmin.gen.go ../../../openapi/gatekeeper.yaml

// Handler has implements all methods of oapi-codegen's ServiceInterface
type Handler struct {
	service *service.Service
	logger  *zap.Logger
}

// NewHandler sets up all API endpoint routes
func NewHandler(g *gin.Engine, db *db.Database, s *service.Service, applicationName,
	organizationName string, disableAPIAuthentication bool, logger *zap.Logger) *Handler {

	handler := &Handler{
		service: s,
		logger:  logger,
	}

	registerMetricsRoute(g, applicationName)

	g.GET(showHTTPForwardingPath, handler.showHTTPForwardingPage)
	g.GET(showDevelopersPath, handler.showDevelopersPage)
	g.GET(showUserRolesPath, handler.showUserRolePage)

	// Insert authentication middleware for endpoint we are registering next
	auth := newAuth(s.User, s.Role, logger)
	g.Use(auth.AuthenticateAndAuthorize)

	// Register all API endpoint routes
	RegisterHandlers(g, handler)

	return handler
}

// POSTwithoutContentTypeJSON returns boolean indicating whether
// request has POST method without content-type = application/json
// func P2OSTwithoutContentTypeJSON(c *gin.Context) bool {

// 	if c.Request.Method == http.MethodPost {
// 		if c.Request.Header.Get("content-type") != "application/json" {
// 			return true
// 		}
// 	}
// 	return false
// }

// responseError returns formated error back to API client
func responseError(c *gin.Context, e types.Error) {

	code := types.HTTPStatusCode(e)
	msg := e.ErrorDetails()

	// Save internal error details in request context so we can write it in access log later
	_ = c.Error(errors.New(msg))

	c.IndentedJSON(code, ErrorMessage{
		Code:    &code,
		Message: &msg,
	})
	c.Abort()
}

// responseError returns formated error back to API client
func responseErrorBadRequest(c *gin.Context, e error) {

	responseError(c, types.NewBadRequestError(e))
}

// responseErrorNameValueMisMatch when an entity update request has a name mismatch
// between name of entity in url path vs name of entity in POSTed JSON name field
func responseErrorNameValueMisMatch(c *gin.Context) {

	responseErrorBadRequest(c, errors.New("name field value mismatch"))
}
