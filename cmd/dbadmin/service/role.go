package service

import (
	"fmt"

	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// RoleService is
type RoleService struct {
	db        *db.Database
	Changelog *Changelog
}

// NewRole returns a new role instance
func NewRole(database *db.Database, c *Changelog) *RoleService {

	return &RoleService{
		db:        database,
		Changelog: c,
	}
}

// GetAll returns all roles
func (rs *RoleService) GetAll() (roles types.Roles, err types.Error) {

	return rs.db.Role.GetAll()
}

// Get returns details of an role
func (rs *RoleService) Get(roleName string) (role *types.Role, err types.Error) {

	return rs.db.Role.Get(roleName)
}

// Create creates an role
func (rs *RoleService) Create(newRole types.Role, who Requester) (*types.Role, types.Error) {

	if _, err := rs.db.Role.Get(newRole.Name); err == nil {
		return nil, types.NewBadRequestError(
			fmt.Errorf("role '%s' already exists", newRole.Name))
	}
	// Automatically set default fields
	newRole.CreatedAt = shared.GetCurrentTimeMilliseconds()
	newRole.CreatedBy = who.User

	if err := rs.updateRole(&newRole, who); err != nil {
		return nil, err
	}
	rs.Changelog.Create(newRole, who)
	return &newRole, nil
}

// Update updates an existing role
func (rs *RoleService) Update(updatedRole types.Role, who Requester) (*types.Role, types.Error) {

	currentRole, err := rs.db.Role.Get(updatedRole.Name)
	if err != nil {
		return nil, err
	}
	// Populate fields which are not updateable
	updatedRole.Name = currentRole.Name
	updatedRole.CreatedAt = currentRole.CreatedAt
	updatedRole.CreatedBy = currentRole.CreatedBy

	if err = rs.updateRole(&updatedRole, who); err != nil {
		return nil, err
	}
	rs.Changelog.Update(currentRole, updatedRole, who)
	return &updatedRole, nil
}

// updateRole updates last-modified field(s) and updates role in database
func (rs *RoleService) updateRole(updatedRole *types.Role, who Requester) types.Error {

	updatedRole.LastModifiedAt = shared.GetCurrentTimeMilliseconds()
	updatedRole.LastModifiedBy = who.User
	return rs.db.Role.Update(updatedRole)
}

// Delete deletes an role
func (rs *RoleService) Delete(roleName string, who Requester) (deletedRole *types.Role, e types.Error) {

	role, err := rs.db.Role.Get(roleName)
	if err != nil {
		return nil, err
	}
	userWithRoleCount := rs.countUserWithRole(roleName)
	if userWithRoleCount > 0 {
		return nil, types.NewForbiddenError(
			fmt.Errorf("cannot delete role '%s' still assigned to %d users",
				roleName, userWithRoleCount))
	}
	if err = rs.db.Role.Delete(roleName); err != nil {
		return nil, err
	}
	rs.Changelog.Delete(role, who)
	return role, nil
}

// counts number of users with a specific role
func (rs *RoleService) countUserWithRole(role string) int {

	users, err := rs.db.User.GetAll()
	if err != nil {
		return 0
	}
	var count int
	for _, user := range users {
		for _, userRole := range user.Roles {
			if role == userRole {
				count++
			}
		}
	}
	return count
}
