package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// registerUserRoutes registers all routes we handle
func (h *Handler) registerUserRoutes(r *gin.RouterGroup) {
	r.GET("/users", h.handler(h.getAllUsers))
	r.POST("/users", h.handler(h.createUser))

	r.GET("/users/:user", h.handler(h.getUser))
	r.POST("/users/:user", h.handler(h.updateUser))
	r.DELETE("/users/:user", h.handler(h.deleteUser))
}

const (
	// Name of user parameter in the route definition
	userParameter = "user"

	// name of array in JSON response
	userArray = "user"
)

// getAllUsers returns all users
func (h *Handler) getAllUsers(c *gin.Context) handlerResponse {

	users, err := h.service.User.GetAll()
	if err != nil {
		return handleError(err)
	}
	removePasswords(users)
	return handleOK(StringMap{userArray: users})
}

// removePasswords set passwords of all users in slice to empty string
func removePasswords(users types.Users) {

	for index := range users {
		users[index].Password = ""
	}
}

// getUser returns details of an user
func (h *Handler) getUser(c *gin.Context) handlerResponse {

	user, err := h.service.User.Get(c.Param(userParameter))
	if err != nil {
		return handleError(err)
	}
	user.Password = ""
	return handleOK(user)
}

// createUser creates an user
func (h *Handler) createUser(c *gin.Context) handlerResponse {

	var newUser types.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		return handleBadRequest(err)
	}
	storedUser, err := h.service.User.Create(newUser, h.who(c))
	if err != nil {
		return handleError(err)
	}
	// Remove password so we do not show in response
	storedUser.Password = ""
	return handleCreated(storedUser)
}

// updateUser updates an existing user
func (h *Handler) updateUser(c *gin.Context) handlerResponse {

	var updatedUser types.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		return handleBadRequest(err)
	}
	if updatedUser.Name != c.Param(userParameter) {
		return handleNameMismatch()
	}
	storedUser, err := h.service.User.Update(updatedUser, h.who(c))
	if err != nil {
		return handleError(err)
	}
	// Remove password so we do not show in response
	storedUser.Password = ""
	return handleOK(storedUser)
}

// deleteUser deletes an user
func (h *Handler) deleteUser(c *gin.Context) handlerResponse {

	deletedUser, err := h.service.User.Delete(c.Param(userParameter), h.who(c))
	if err != nil {
		return handleError(err)
	}
	// Remove password so we do not show in response
	deletedUser.Password = ""
	return handleOK(deletedUser)
}
