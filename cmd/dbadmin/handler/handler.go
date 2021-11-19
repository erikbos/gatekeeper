package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/dbadmin/handler/status"
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

var (
	errFieldMisMatch = errors.New("path and name field value mismatch")
)

// NewHandler sets up all API endpoint routes
func NewHandler(router *gin.Engine, db *db.Database, s *service.Service,
	applicationName string, disableAPIAuthentication bool, logger *zap.Logger) *Handler {

	handler := &Handler{
		service: s,
		logger:  logger,
	}

	registerMetricsRoute(router, applicationName)

	status := status.New(s)
	status.RegisterRoutes(router)

	// Register routes and add authentication middleware
	auth := newAuth(s.User, s.Role, logger)
	router.Use(auth.AuthenticateAndAuthorize)

	RegisterHandlersWithOptions(router, handler, GinServerOptions{
		// TODO specifying auth.AuthenticateAndAuthorize as middleware here does not work
		// Middlewares: []MiddlewareFunc{
		// 	auth.AuthenticateAndAuthorize,
		// },
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

// responseErrorBadRequest returns BadRequest error back to API client
func responseErrorBadRequest(c *gin.Context, e error) {

	responseError(c, types.NewBadRequestError(e))
}

// responseErrorNameValueMisMatch when an entity update request has a name mismatch
// between name of entity in url path vs name of entity in POSTed JSON name field
func responseErrorNameValueMisMatch(c *gin.Context) {

	responseErrorBadRequest(c, errFieldMisMatch)
}
