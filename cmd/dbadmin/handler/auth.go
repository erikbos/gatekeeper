package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/dbadmin/service"
	"github.com/erikbos/gatekeeper/pkg/types"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

var (
	errBasicAuthRequired = types.NewUnauthorizedError(errors.New("basic authorization required"))
	errUnknownUser       = types.NewUnauthorizedError(errors.New("unknown user"))
	errPasswordMismatch  = types.NewUnauthorizedError(errors.New("password mismatch"))
	errPathNotAllowed    = types.NewForbiddenError(errors.New("method or path not allowed"))
)

//
const (
	// key in request context holding name & connection details of requestor
	RequestorKey = "Requestor"
)

// AuthHandler provides authentication and authorization middleware
type AuthHandler struct {
	user   service.User
	role   service.Role
	logger *zap.Logger
}

// newAuth setups new AuthHandler entity which can authenticate HTTP requests
func newAuth(user service.User, role service.Role, logger *zap.Logger) *AuthHandler {

	return &AuthHandler{
		user:   user,
		role:   role,
		logger: logger,
	}
}

// AuthenticateAndAuthorize validates username and password supplied via HTTP Basic authentication
// and checks whether requested path is allowed according to the assigned roles of the user
func (a *AuthHandler) AuthenticateAndAuthorize(c *gin.Context) {

	username, password, ok := c.Request.BasicAuth()
	if !ok {
		abortAuthorizationRequired(c, errBasicAuthRequired)
		return
	}
	// Store provided username in request context so we log this afterwards
	webadmin.StoreUser(c, username)

	user, err := a.ValidatePassword(username, password)
	if err != nil {
		// Unknown user or pw mismatch, we return 401 and abort request.
		abortAuthorizationRequired(c, err)
		return
	}

	roleWhichAllowsAccess, err := a.IsPathAllowedByUser(user, c.Request.Method, c.Request.URL.Path)
	if err != nil {
		// Path not allowed by none of the roles of the user, we return 401 and abort request.
		abortAuthorizationRequired(c, err)
	}

	// Store user's role in the request context so we can use it later on while logging request
	webadmin.StoreRole(c, roleWhichAllowsAccess)
}

// ValidatePassword confirm if user supplied a valid password
func (a *AuthHandler) ValidatePassword(username, password string) (user *types.User, e types.Error) {

	user, err := a.user.Get(username)
	if err != nil {
		return nil, types.NewUnauthorizedError(errUnknownUser)
	}
	if !service.CheckPasswordHash(password, user.Password) {
		return nil, types.NewUnauthorizedError(errPasswordMismatch)
	}
	return user, nil
}

// IsPathAllowedByUser checks whether user is allowed to access a path,
// if allowed return name of allowing role
func (a *AuthHandler) IsPathAllowedByUser(user *types.User, method, path string) (
	roleName string, e types.Error) {

	a.logger.Debug("IsPathAllowedByUser",
		zap.String("user", user.Name), zap.String("method", method), zap.String("path", path))

	for _, roleName := range user.Roles {
		if role, err := a.role.Get(roleName); err == nil {
			if role.IsPathAllowed(method, path) {
				return role.Name, nil
			}
		}
	}
	// in case nothing matched we do not allow access
	return "", errPathNotAllowed
}

func abortAuthorizationRequired(c *gin.Context, errorDetails types.Error) {

	realm := "Authorization required"
	c.Header("WWW-Authenticate", "Basic realm="+strconv.Quote(realm))
	responseError(c, types.NewUnauthorizedError(errorDetails))
}

// who returns name of authenticated user requesting this API call
func (h *Handler) who(c *gin.Context) service.Requester {

	// Store details, changelog will use these records
	return service.Requester{
		RemoteAddr: c.ClientIP(),
		Header:     c.Request.Header,
		User:       webadmin.GetUser(c),
		Role:       webadmin.GetRole(c),
		RequestID:  webadmin.GetRequestID(c),
	}
}
