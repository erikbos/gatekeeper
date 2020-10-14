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

// GetByOrganization retrieves all developer belonging to an organization
func (s *DeveloperCache) GetByOrganization(organizationName string) (types.Developers, types.Error) {

	getDeveloperByOrganization := func() (interface{}, types.Error) {
		return s.developer.GetByOrganization(organizationName)
	}
	var developers types.Developers
	if err := s.cache.fetchEntity(db.EntityTypeDeveloper, organizationName, &developers, getDeveloperByOrganization); err != nil {
		return nil, err
	}
	return developers, nil
}

// GetCountByOrganization retrieves number of developer belonging to an organization
func (s *DeveloperCache) GetCountByOrganization(organizationName string) (int, types.Error) {

	return s.developer.GetCountByOrganization(organizationName)
}

// GetByEmail retrieves a developer from database
func (s *DeveloperCache) GetByEmail(developerOrganization, developerEmail string) (*types.Developer, types.Error) {

	getDeveloperByEmail := func() (interface{}, types.Error) {
		return s.developer.GetByEmail(developerOrganization, developerEmail)
	}
	var developer types.Developer
	if err := s.cache.fetchEntity(db.EntityTypeDeveloper, developerEmail, &developer, getDeveloperByEmail); err != nil {
		return nil, err
	}
	return &developer, nil
}

// GetByID retrieves a developer from database
func (s *DeveloperCache) GetByID(developerID string) (*types.Developer, types.Error) {

	getDeveloperByID := func() (interface{}, types.Error) {
		return s.developer.GetByID(developerID)
	}
	var developer types.Developer
	if err := s.cache.fetchEntity(db.EntityTypeDeveloper, developerID, &developer, getDeveloperByID); err != nil {
		return nil, err
	}
	return &developer, nil
}

// Update UPSERTs a developer in database
func (s *DeveloperCache) Update(d *types.Developer) types.Error {

	s.cache.deleteEntry(db.EntityTypeDeveloper, d.DeveloperID)
	s.cache.deleteEntry(db.EntityTypeDeveloper, d.Email)
	return s.developer.Update(d)
}

// DeleteByID deletes a developer
func (s *DeveloperCache) DeleteByID(organizationName, developerID string) types.Error {

	s.cache.deleteEntry(db.EntityTypeDeveloper, developerID)
	return s.developer.DeleteByID(organizationName, developerID)
}
