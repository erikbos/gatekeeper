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
func NewHandler(router *gin.Engine, db *db.Database, s *service.Service, applicationName,
	organizationName string, disableAPIAuthentication bool, logger *zap.Logger) *Handler {

	handler := &Handler{
		service: s,
		logger:  logger,
	}

	registerMetricsRoute(router, applicationName)

	router.GET(showHTTPForwardingPath, handler.showHTTPForwardingPage)
	router.GET(showDevelopersPath, handler.showDevelopersPage)
	router.GET(showUserRolesPath, handler.showUserRolePage)

	// Insert authentication middleware for endpoint we are registering next
	auth := newAuth(s.User, s.Role, logger)

	RegisterHandlersWithOptions(router, handler, GinServerOptions{
		Middlewares: []MiddlewareFunc{
			auth.AuthenticateAndAuthorize,
		},
	})

	return handler
}

// func CheckAccept(c *gin.Context) {

// 	if c.Request.Method == http.MethodGet &&
// 		!strings.HasPrefix(c.Request.Header.Get("content-type"), "application/json") {
// 		responseError(c, types.NewNotAcceptable(errors.New(c.Request.Header.Get("content-type"))))
// 	}
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
