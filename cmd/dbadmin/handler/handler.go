package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/dbadmin/service"
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// Handler contains our runtime parameters
type Handler struct {
	service *service.Service
}

// NewHandler sets up all API endpoint routes
func NewHandler(g *gin.Engine, db *db.Database, s *service.Service, logger *zap.Logger, enableAPIAuthentication bool) *Handler {

	// m := &Metrics{}
	// m.register("Qqq")

	// Insert authentication middleware for every /v1 prefix'ed API endpoint
	apiRoutes := g.Group("/v1")
	// apiRoutes.Use(metricsMiddleware(m))

	if enableAPIAuthentication {
		auth := newAuth(&s.User, &s.Role, logger)
		auth.Start(db, logger)
		apiRoutes.Use(auth.AuthenticateAndAuthorize)
	}

	handler := &Handler{
		service: s,
	}
	g.GET(showHTTPForwardingPath, handler.showHTTPForwardingPage)
	g.GET(showUserRolesPath, handler.showUserRolePage)

	// Register all API endpoint routes
	handler.registerListenerRoutes(apiRoutes)
	handler.registerRouteRoutes(apiRoutes)
	handler.registerClusterRoutes(apiRoutes)
	handler.registerOrganizationRoutes(apiRoutes)
	handler.registerDeveloperRoutes(apiRoutes)
	handler.registerDeveloperAppRoutes(apiRoutes)
	handler.registerCredentialRoutes(apiRoutes)
	handler.registerAPIProductRoutes(apiRoutes)
	handler.registerUserRoutes(apiRoutes)
	handler.registerRoleRoutes(apiRoutes)
	return handler
}

// handler wraps an API endpoint function and returns any returned object as
// indented json, or in case of error generates an structured error message.
func (h *Handler) handler(function func(c *gin.Context) handlerResponse) gin.HandlerFunc {

	return func(c *gin.Context) {
		// a POST request must have content-type = application/json
		if POSTwithoutContentTypeJSON(c) {
			showErrorMessageAndAbort(c, http.StatusUnsupportedMediaType,
				types.NewBadRequestError(errors.New(
					"Content-type application/json required when submitting data")))
			return
		}

		// Invoke actual API endpoint function
		response := function(c)
		if response.error != nil {
			showErrorMessageAndAbort(c, types.HTTPErrorStatusCode(response.error), response.error)
			return
		}
		if response.created {
			c.IndentedJSON(http.StatusCreated, response.responseBody)
			return
		}
		c.IndentedJSON(http.StatusOK, response.responseBody)
	}
}

// POSTwithoutContentTypeJSON returns boolean indicating whether
// request has POST method without content-type = application/json
func POSTwithoutContentTypeJSON(c *gin.Context) bool {

	if c.Request.Method == http.MethodPost {
		if c.Request.Header.Get("content-type") != "application/json" {
			return true
		}
	}
	return false
}

// handlerResponse is used as return type for an HTTP endpoint to indicate
// whether request successed or not, and to indicate the type of error.
type handlerResponse struct {
	error        types.Error
	created      bool
	responseBody interface{}
}

// StringMap is a shortcut for map[string]interface{}
type StringMap map[string]interface{}

// handleOK returns 200 + json contents
func handleOK(body interface{}) handlerResponse {
	return handlerResponse{error: nil, responseBody: body}
}

// handleOKAttribute returns 200 + json contents of single attribute
func handleOKAttribute(a types.Attribute) handlerResponse {
	return handleOK(types.Attribute{Name: a.Name, Value: a.Value})
}

// handleOKAttributes returns 200 + json contents of multiple attributes
func handleOKAttributes(a types.Attributes) handlerResponse {
	return handleOK(StringMap{"attribute": a})
}

// handleCreated returns 201 + json contents
func handleCreated(body interface{}) handlerResponse {
	return handlerResponse{error: nil, created: true, responseBody: body}
}

func handleError(e types.Error) handlerResponse {
	return handlerResponse{error: e}
}

func handleBadRequest(e error) handlerResponse {
	return handlerResponse{error: types.NewBadRequestError(e)}
}

// func handleUnauthorized(e error) handlerResponse {
// 	return handlerResponse{error: types.NewUnauthorizedError(e)}
// }

// handleNameMismatch when an entity update request has a name mismatch between name of entity in url path
// vs name of entity in POSTed JSON name field
func handleNameMismatch() handlerResponse {

	return handlerResponse{
		error: types.NewBadRequestError(errors.New("Name field value mismatch")),
	}
}

func showErrorMessageAndAbort(c *gin.Context, statusCode int, e types.Error) {

	// Save internal error details in request context so we can write it in access log later
	c.Error(errors.New(e.ErrorDetails()))

	// Show (public) error message
	c.IndentedJSON(statusCode, StringMap{"error": e.ErrorDetails()})
	c.Abort()
}

// func metricsMiddleware(m *Metrics) gin.HandlerFunc {

// 	return func(c *gin.Context) {

// 		c.Next()

// 		status := strconv.Itoa(c.Writer.Status())
// 		m.QueryHit(c.Request.Method, c.FullPath(), status)
// 	}
// }
