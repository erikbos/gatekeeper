package handler

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/erikbos/gatekeeper/cmd/dbadmin/service"
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
	"github.com/erikbos/gatekeeper/pkg/webadmin"
)

var (
	errBasicAuthRequired = types.NewUnauthorizedError(errors.New("Basic Authorization required"))
	errUnknownUser       = types.NewUnauthorizedError(errors.New("Unknown user"))
	errPasswordMismatch  = types.NewUnauthorizedError(errors.New("Password mismatch"))
	errPathNotAllowed    = types.NewUnauthorizedError(errors.New("Path not allowed"))
)

//
const (

	// key in request context holding name & connection details of requestor
	RequestorKey = "Requestor"

	// interval between entities refresh loads
	entityRefreshInterval = 3 * time.Second
)

// AuthHandler provides authentication and authorization middleware
type AuthHandler struct {
	user        *service.User
	role        *service.Role
	entityCache *db.EntityCache
	logger      *zap.Logger
}

// newAuth setups new AuthHandler entity
func newAuth(user *service.User, role *service.Role, logger *zap.Logger) *AuthHandler {

	return &AuthHandler{
		user:   user,
		role:   role,
		logger: logger,
	}
}

// Start auth background jobs such as entity cache loading
func (a *AuthHandler) Start(database *db.Database, logger *zap.Logger) {

	entityCacheConf := db.EntityCacheConfig{
		RefreshInterval: entityRefreshInterval,
		User:            true,
		Role:            true,
	}
	a.entityCache = db.NewEntityCache(database, entityCacheConf, logger)
	a.entityCache.Start()
}

// AuthenticateAndAuthorize validates username and password supplied via HTTP Basic authentication
// and checks whether requested path is allowed according to the assigned roles of the user
func (a *AuthHandler) AuthenticateAndAuthorize(c *gin.Context) {

	username, password, err := decodeBasicAuthorizationHeader(
		c.Request.Header.Get("Authorization"))
	if err != nil {
		// Parsing of Authorization header failed
		abortAuthorizationRequired(c, err)
		return
	}
	// Store provided username in request context so we log this afterwards
	webadmin.StoreUser(c, username)

	user, err := a.ValidatePassword(username, password)
	if err != nil {
		// Credentials doesn't match, we return 401 and abort request.
		abortAuthorizationRequired(c, err)
		return
	}

	roleWhichAllowsAccess := a.IsPathAllowedByUser(user, c.Request.Method, c.Request.URL.Path)
	if roleWhichAllowsAccess == "" {
		// Path not allowed by none of the roles of the user, we return 401 and abort request.
		abortAuthorizationRequired(c, errPathNotAllowed)
	}

	// Store user's role in the request context so we can use it later on while logging request
	webadmin.StoreRole(c, roleWhichAllowsAccess)
}

// decodeBasicAuthorizationHeader decodes a HTTP Authorization header value
func decodeBasicAuthorizationHeader(authorizationHeader string) (username,
	password string, e types.Error) {

	auth := strings.SplitN(authorizationHeader, " ", 2)
	// No credentials or no Authorization basic prefix?
	if len(auth) != 2 || auth[0] != "Basic" {
		return "", "", errBasicAuthRequired
	}

	// Decode basic auth header
	payload, err := base64.StdEncoding.DecodeString(auth[1])
	usernameAndPassword := strings.SplitN(string(payload), ":", 2)
	if err != nil || len(usernameAndPassword) != 2 {
		return "", "", errBasicAuthRequired
	}
	return usernameAndPassword[0], usernameAndPassword[1], nil
}

// ValidatePassword confirm if user supplied a valid password
func (a *AuthHandler) ValidatePassword(username, password string) (user *types.User, e types.Error) {

	user, err := a.entityCache.GetUser(username)
	if err != nil {
		return nil, types.NewUnauthorizedError(errUnknownUser)
	}
	if !service.CheckPasswordHash(password, user.Password) {
		return nil, types.NewUnauthorizedError(errPasswordMismatch)
	}
	return user, nil
}

// IsPathAllowedByUser checks whether user is allowed to access a path,
// if yes returns the name of the allowing access
func (a *AuthHandler) IsPathAllowedByUser(user *types.User, method, path string) string {

	a.logger.Debug("IsPathAllowedByUser",
		zap.String("user", user.Name), zap.String("method", method), zap.String("path", path))

	for _, roleName := range user.Roles {
		if role, err := a.entityCache.GetRole(roleName); err == nil {
			if role.PathAllowed(method, path) {
				return role.Name
			}
		}
	}
	// in case nothing matched we do not allow access
	return ""
}

func abortAuthorizationRequired(c *gin.Context, errorDetails types.Error) {

	realm := "\"Authorization required\""
	c.Header("WWW-Authenticate", "Basic realm="+strconv.Quote(realm))
	showErrorMessageAndAbort(c, http.StatusUnauthorized, errorDetails)
}

// who returns name of authenticated user requesting this API call
func (h *Handler) who(c *gin.Context) service.Requester {

	// Store more details, changelog will use these records
	return service.Requester{
		RemoteAddr: c.Request.RemoteAddr,
		Header:     c.Request.Header,
		Username:   webadmin.GetUser(c),
		RequestID:  webadmin.GetRequestID(c),
	}
}
