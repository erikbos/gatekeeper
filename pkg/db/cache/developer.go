package cache

import (
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// DeveloperCache holds our database config
type DeveloperCache struct {
	developer db.Developer
	cache     *Cache
}

// NewDeveloperCache creates developer instance
func NewDeveloperCache(cache *Cache, developer db.Developer) *DeveloperCache {
	return &DeveloperCache{
		developer: developer,
		cache:     cache,
	}
}

// GetAll retrieves all developer
func (s *DeveloperCache) GetAll(organizationName string) (types.Developers, types.Error) {

	getDevelopers := func() (interface{}, types.Error) {
		return s.developer.GetAll(organizationName)
	}
	var developers types.Developers
	// TODO/FIXME
	if err := s.cache.fetchEntity(types.TypeDeveloperName, "--all-developers--", &developers, getDevelopers); err != nil {
		return nil, err
	}
	return developers, nil
}

// GetByEmail retrieves a developer from database
func (s *DeveloperCache) GetByEmail(organizationName, developerEmail string) (*types.Developer, types.Error) {

	getDeveloperByEmail := func() (interface{}, types.Error) {
		return s.developer.GetByEmail(organizationName, developerEmail)
	}
	var developer types.Developer
	if err := s.cache.fetchEntity(types.TypeDeveloperName, developerEmail, &developer, getDeveloperByEmail); err != nil {
		return nil, err
	}
	return &developer, nil
}

// GetByID retrieves a developer from database
func (s *DeveloperCache) GetByID(organizationName, developerID string) (*types.Developer, types.Error) {

	getDeveloperByID := func() (interface{}, types.Error) {
		return s.developer.GetByID(organizationName, developerID)
	}
	var developer types.Developer
	if err := s.cache.fetchEntity(types.TypeDeveloperName, developerID, &developer, getDeveloperByID); err != nil {
		return nil, err
	}
	return &developer, nil
}

// Update UPSERTs a developer in database
func (s *DeveloperCache) Update(organizationName string, d *types.Developer) types.Error {

	s.cache.deleteEntry(types.TypeDeveloperName, d.DeveloperID)
	s.cache.deleteEntry(types.TypeDeveloperName, d.Email)
	return s.developer.Update(organizationName, d)
}

// DeleteByID deletes a developer
func (s *DeveloperCache) DeleteByID(organizationName, developerID string) types.Error {

	s.cache.deleteEntry(types.TypeDeveloperName, developerID)
	return s.developer.DeleteByID(organizationName, developerID)
}
