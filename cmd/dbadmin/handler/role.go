package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// returns all roles
// (GET /v1/roles)
func (h *Handler) GetV1Roles(c *gin.Context) {

	roles, err := h.service.Role.GetAll()
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseRoles(c, roles)
}

//  creates an role
// (POST /v1/roles)
func (h *Handler) PostV1Roles(c *gin.Context) {

	var receivedRole Role
	if err := c.ShouldBindJSON(&receivedRole); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	newRole := fromRole(receivedRole)
	storedRole, err := h.service.Role.Create(newRole, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseRoleCreated(c, storedRole)
}

// returns details of a role
// (GET /v1/roles/{role_name})
func (h *Handler) GetV1RolesRoleName(c *gin.Context, roleName RoleName) {

	role, err := h.service.Role.Get(string(roleName))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseRole(c, role)
}

// updates an existing role
// (POST /v1/roles/{role_name})
func (h *Handler) PostV1RolesRoleName(c *gin.Context, roleName RoleName) {

	_, err := h.service.Role.Get(string(roleName))
	if err != nil {
		responseError(c, err)
		return
	}
	var receivedRole Role
	if err := c.ShouldBindJSON(&receivedRole); err != nil {
		responseErrorBadRequest(c, err)
		return
	}
	updatedRole := fromRole(receivedRole)
	if updatedRole.Name != string(roleName) {
		responseErrorNameValueMisMatch(c)
		return
	}
	storedRole, err := h.service.Role.Update(updatedRole, h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseRoleUpdated(c, storedRole)
}

// deleteRole deletes a role
// (DELETE /v1/roles/{role_name})
func (h *Handler) DeleteV1RolesRoleName(c *gin.Context, roleName RoleName) {

	deletedRole, err := h.service.Role.Get(string(roleName))
	if err != nil {
		responseError(c, err)
		return
	}
	if err := h.service.Role.Delete(string(roleName), h.who(c)); err != nil {
		responseError(c, err)
		return
	}
	h.responseRole(c, deletedRole)
}

// Retrieve users assigned to role
// (GET /v1/roles/{role_name}/users)
func (h *Handler) GetV1RolesRoleNameUsers(c *gin.Context, roleName RoleName) {

	usersWithRole, err := h.service.GetUsersByRole(string(roleName))
	if err != nil {
		responseError(c, err)
		return
	}
	c.JSON(http.StatusOK, usersWithRole)
}

// API responses

func (h *Handler) responseRoles(c *gin.Context, roles types.Roles) {

	all_roles := make([]Role, len(roles))
	for i, v := range roles {
		all_roles[i] = h.ToRoleResponse(&v)
	}
	c.IndentedJSON(http.StatusOK, Roles{
		Role: &all_roles,
	})
}

func (h *Handler) responseRole(c *gin.Context, user *types.Role) {

	c.IndentedJSON(http.StatusOK, h.ToRoleResponse(user))
}

func (h *Handler) responseRoleCreated(c *gin.Context, role *types.Role) {

	c.IndentedJSON(http.StatusCreated, h.ToRoleResponse(role))
}

func (h *Handler) responseRoleUpdated(c *gin.Context, role *types.Role) {

	c.IndentedJSON(http.StatusOK, h.ToRoleResponse(role))
}

// type conversion

func (h *Handler) ToRoleResponse(l *types.Role) Role {

	role := Role{
		CreatedAt:      &l.CreatedAt,
		CreatedBy:      &l.CreatedBy,
		DisplayName:    &l.DisplayName,
		LastModifiedBy: &l.LastModifiedBy,
		LastModifiedAt: &l.LastModifiedAt,
		Name:           l.Name,
	}
	if l.Permissions != nil {
		role.Permissions = ToRolePermissionsResponse(l.Permissions)
	}
	return role
}

func ToRolePermissionsResponse(permissions types.Permissions) *[]RolePermissions {

	allowed_paths := make([]RolePermissions, len(permissions))
	for i := range permissions {
		allowed_paths[i] = RolePermissions{
			Methods: &permissions[i].Methods,
			Paths:   &permissions[i].Paths,
		}
	}
	return &allowed_paths
}

func fromRole(u Role) types.Role {

	role := types.Role{}
	if u.Permissions != nil {
		role.Permissions = fromRolePermissions(u.Permissions)
	}
	if u.CreatedAt != nil {
		role.CreatedAt = *u.CreatedAt
	}
	if u.CreatedBy != nil {
		role.CreatedBy = *u.CreatedBy
	}
	if u.DisplayName != nil {
		role.DisplayName = *u.DisplayName
	}
	if u.LastModifiedBy != nil {
		role.LastModifiedBy = *u.LastModifiedBy
	}
	if u.LastModifiedAt != nil {
		role.LastModifiedAt = *u.LastModifiedAt
	}
	if u.Name != "" {
		role.Name = u.Name
	}
	return role
}

func fromRolePermissions(a *[]RolePermissions) types.Permissions {

	if a == nil {
		return types.NullPermissions
	}
	all_attributes := make([]types.Permission, len(*a))
	for i, a := range *a {
		all_attributes[i] = types.Permission{
			Methods: *a.Methods,
			Paths:   *a.Paths,
		}
	}
	return all_attributes
}
