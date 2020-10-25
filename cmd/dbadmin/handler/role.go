package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// registerRoleRoutes registers all routes we handle
func (h *Handler) registerRoleRoutes(r *gin.RouterGroup) {
	r.GET("/roles", h.handler(h.getAllRoles))
	r.POST("/roles", h.handler(h.createRole))

	r.GET("/roles/:role", h.handler(h.getRole))
	r.POST("/roles/:role", h.handler(h.updateRole))
	r.DELETE("/roles/:role", h.handler(h.deleteRole))
}

const (
	// Name of role parameter in the route definition
	roleParameter = "role"
)

// getAllRoles returns all roles
func (h *Handler) getAllRoles(c *gin.Context) handlerResponse {

	roles, err := h.service.Role.GetAll()
	if err != nil {
		return handleError(err)
	}
	return handleOK(StringMap{"roles": roles})
}

// getRole returns details of an role
func (h *Handler) getRole(c *gin.Context) handlerResponse {

	role, err := h.service.Role.Get(c.Param(roleParameter))
	if err != nil {
		return handleError(err)
	}
	return handleOK(role)
}

// createRole creates an role
func (h *Handler) createRole(c *gin.Context) handlerResponse {

	var newRole types.Role
	if err := c.ShouldBindJSON(&newRole); err != nil {
		return handleBadRequest(err)
	}
	storedRole, err := h.service.Role.Create(newRole, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleCreated(storedRole)
}

// updateRole updates an existing role
func (h *Handler) updateRole(c *gin.Context) handlerResponse {

	var updatedRole types.Role
	if err := c.ShouldBindJSON(&updatedRole); err != nil {
		return handleBadRequest(err)
	}
	if updatedRole.Name != c.Param(roleParameter) {
		return handleNameMismatch()
	}
	storedRole, err := h.service.Role.Update(updatedRole, h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(storedRole)
}

// deleteRole deletes an role
func (h *Handler) deleteRole(c *gin.Context) handlerResponse {

	deletedRole, err := h.service.Role.Delete(c.Param(roleParameter), h.who(c))
	if err != nil {
		return handleError(err)
	}
	return handleOK(deletedRole)
}
