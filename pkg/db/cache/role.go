package cache

import (
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// RoleCache holds our database config
type RoleCache struct {
	role  db.Role
	cache *Cache
}

// NewRoleCache creates role instance
func NewRoleCache(cache *Cache, role db.Role) *RoleCache {
	return &RoleCache{
		role:  role,
		cache: cache,
	}
}

// GetAll retrieves all roles
func (s *RoleCache) GetAll() (types.Roles, types.Error) {

	getAll := func() (interface{}, types.Error) {
		return s.role.GetAll()
	}
	var roles types.Roles
	if err := s.cache.fetchEntity(types.TypeRoleName, "", &roles, getAll); err != nil {
		return nil, err
	}
	return roles, nil
}

// Get retrieves a role from database
func (s *RoleCache) Get(roleName string) (*types.Role, types.Error) {

	getRole := func() (interface{}, types.Error) {
		return s.role.Get(roleName)
	}
	var role types.Role
	if err := s.cache.fetchEntity(types.TypeRoleName, roleName, &role, getRole); err != nil {
		return nil, err
	}
	return &role, nil
}

// Update UPSERTs an role in database
func (s *RoleCache) Update(c *types.Role) types.Error {

	s.cache.deleteEntry(types.TypeRoleName, c.Name)
	return s.role.Update(c)
}

// Delete deletes a role
func (s *RoleCache) Delete(roleToDelete string) types.Error {

	s.cache.deleteEntry(types.TypeRoleName, roleToDelete)
	return s.role.Delete(roleToDelete)
}
