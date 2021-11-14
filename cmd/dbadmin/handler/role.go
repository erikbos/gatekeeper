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
		responseError(c, types.NewBadRequestError(err))
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
		responseError(c, types.NewBadRequestError(err))
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

	deletedRole, err := h.service.Role.Delete(string(roleName), h.who(c))
	if err != nil {
		responseError(c, err)
		return
	}
	h.responseRole(c, deletedRole)
}

// Returns API response all user details
func (h *Handler) responseRoles(c *gin.Context, roles types.Roles) {

	all_roles := make([]Role, len(roles))
	for i := range roles {
		all_roles[i] = h.ToRoleResponse(&roles[i])
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
	if l.Allows != nil {
		role.Allow = ToRoleAllowsResponse(l.Allows)
	}
	return role
}

func ToRoleAllowsResponse(allows types.Allows) *[]RoleAllow {

	allowed_paths := make([]RoleAllow, len(allows))
	for i := range allows {
		allowed_paths[i] = RoleAllow{
			Methods: &allows[i].Methods,
			Paths:   &allows[i].Paths,
		}
	}
	return &allowed_paths
}

func fromRole(u Role) types.Role {

	role := types.Role{}
	if u.Allow != nil {
		role.Allows = fromRoleAllow(u.Allow)
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

func fromRoleAllow(a *[]RoleAllow) types.Allows {

	if a == nil {
		return types.NullAllows
	}
	all_attributes := make([]types.Allow, len(*a))
	for i, a := range *a {
		all_attributes[i] = types.Allow{
			Methods: *a.Methods,
			Paths:   *a.Paths,
		}
	}
	return all_attributes
}
