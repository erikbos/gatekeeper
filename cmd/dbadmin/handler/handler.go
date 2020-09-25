package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/cmd/dbadmin/service"
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// Handler contains our runtime parameters
type Handler struct {
	// db      *db.Database
	service *service.Service
}

// NewHandler returns a newly setup Handler
func NewHandler(ginEngine *gin.Engine, db *db.Database, s *service.Service) *Handler {

	handler := &Handler{
		// db:      db,
		service: s,
	}

	handler.registerListenerRoutes(ginEngine)
	handler.registerRouteRoutes(ginEngine)
	handler.registerClusterRoutes(ginEngine)
	handler.registerOrganizationRoutes(ginEngine)
	handler.registerDeveloperRoutes(ginEngine)
	handler.registerDeveloperAppRoutes(ginEngine)
	handler.registerCredentialRoutes(ginEngine)
	handler.registerAPIProductRoutes(ginEngine)

	return handler
}

//
const (
	// Error msg for no JSON
	POSTWithoutNoJSONError = "Content-type application/json required when submitting data"

	// SessionUserKey parameter key containing authenticated user doing this single API request
	SessionUserKey = "AuthenticatedUser"
)

// StringMap is a shortcut for map[string]interface{}
type StringMap map[string]interface{}

type handlerResponse struct {
	error        types.Error
	created      bool
	responseBody interface{}
}

func handleOK(body interface{}) handlerResponse {
	return handlerResponse{error: nil, responseBody: body}
}

func handleCreated(body interface{}) handlerResponse {
	return handlerResponse{error: nil, created: true, responseBody: body}
}

func handleOKAttribute(a types.Attribute) handlerResponse {
	return handleOK(types.Attribute{Name: a.Name, Value: a.Value})
}

func handleOKAttributes(a types.Attributes) handlerResponse {
	return handleOK(StringMap{"attribute": a})
}

func handleError(e types.Error) handlerResponse {
	return handlerResponse{error: e}
}

func handleBadRequest(e error) handlerResponse {
	return handlerResponse{error: types.NewBadRequestError(e)}
}

func handlePermissionDenied(e error) handlerResponse {
	return handlerResponse{error: types.NewPermissionDeniedError(e)}
}

func (h *Handler) handler(function func(c *gin.Context) handlerResponse) gin.HandlerFunc {

	return func(c *gin.Context) {

		// Set authenticated user for this session, placeholder for now
		h.SetSessionUser(c, "rest-api@test")

		// a POST needs to have content-type = application/json
		if POSTwithoutContentTypeJSON(c) {
			ErrorMessage(c, http.StatusUnsupportedMediaType,
				POSTWithoutNoJSONError, types.ErrBadRequest.Error())
			return
		}

		response := function(c)
		if response.error != nil {
			ErrorMessage(c, httpErrorStatusCode(response),
				response.error.TypeString(), response.error.ErrorDetails())
			return
		}
		if response.created {
			c.IndentedJSON(http.StatusCreated, response.responseBody)
			return
		}
		c.IndentedJSON(http.StatusOK, response.responseBody)
	}
}

// POSTwithoutContentTypeJSON returns bool if request is POST with content-type = application/json
func POSTwithoutContentTypeJSON(c *gin.Context) bool {

	if c.Request.Method == "POST" {
		if c.Request.Header.Get("content-type") != "application/json" {
			return true
		}
	}
	return false
}

// httpErrorStatusCode returns HTTP status code for type
func httpErrorStatusCode(response handlerResponse) int {

	switch response.error.Type() {
	case types.ErrBadRequest:
		return http.StatusBadRequest

	case types.ErrItemNotFound:
		return http.StatusNotFound

	default:
		return http.StatusServiceUnavailable
	}
}

// ErrorMessage returns a format error message back to requestor
func ErrorMessage(c *gin.Context, statusCode int, err, details string) {

	c.IndentedJSON(statusCode, gin.H{
		"error":   err,
		"details": details,
	})
}

// SetSessionUser sets authenticated user for this session
func (h *Handler) SetSessionUser(c *gin.Context, authenticateUser string) {

	//, placeholder for now
	c.Set(SessionUserKey, authenticateUser)
}

// GetSessionUser returns name of authenticated user requesting this API call
func (h *Handler) GetSessionUser(c *gin.Context) string {

	user, exists := c.Get(SessionUserKey)
	if !exists {
		user = ""
	}
	return fmt.Sprintf("%s", user)
}

func setLastModifiedHeader(c *gin.Context, timeStamp int64) {

	c.Header("Last-Modified",
		time.Unix(0, timeStamp*int64(time.Millisecond)).UTC().Format(http.TimeFormat))
}

// returnJSONMessage returns an error message
func returnJSONMessage(c *gin.Context, statusCode int, errorMessage error) {

	c.IndentedJSON(statusCode,
		gin.H{
			"message": fmt.Sprintf("%s", errorMessage),
		})
}

func returnCanNotFindAttribute(c *gin.Context, name string) {

	returnJSONMessage(c,
		http.StatusNotFound,
		fmt.Errorf("Could not find attribute '%s'", name),
	)
}
