package cache

import (
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// UserCache holds our database config
type UserCache struct {
	user  db.User
	cache *Cache
}

// NewUserCache creates user instance
func NewUserCache(cache *Cache, user db.User) *UserCache {
	return &UserCache{
		user:  user,
		cache: cache,
	}
}

// GetAll retrieves all users
func (s *UserCache) GetAll() (types.Users, types.Error) {

	getAll := func() (interface{}, types.Error) {
		return s.user.GetAll()
	}
	var users types.Users
	if err := s.cache.fetchEntry(db.EntityTypeUser, "", &users, getAll); err != nil {
		return nil, err
	}
	return users, nil
}

// Get retrieves a user from database
func (s *UserCache) Get(userName string) (*types.User, types.Error) {

	getUser := func() (interface{}, types.Error) {
		return s.user.Get(userName)
	}
	var user types.User
	if err := s.cache.fetchEntry(db.EntityTypeUser, userName, &user, getUser); err != nil {
		return nil, err
	}
	return &user, nil
}

// Update UPSERTs an user in database
func (s *UserCache) Update(c *types.User) types.Error {

	s.cache.deleteEntry(db.EntityTypeUser, c.Name)
	return s.user.Update(c)
}

// Delete deletes a user
func (s *UserCache) Delete(userToDelete string) types.Error {

	s.cache.deleteEntry(db.EntityTypeUser, userToDelete)
	return s.user.Delete(userToDelete)
}
