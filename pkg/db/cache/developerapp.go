package cache

import (
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// DeveloperAppCache holds our database config
type DeveloperAppCache struct {
	developerapp db.DeveloperApp
	cache        *Cache
}

// NewDeveloperAppCache creates developer instance
func NewDeveloperAppCache(cache *Cache, developerapp db.DeveloperApp) *DeveloperAppCache {
	return &DeveloperAppCache{
		developerapp: developerapp,
		cache:        cache,
	}
}

// GetByOrganization retrieves all developer apps belonging to an organization
func (s *DeveloperAppCache) GetByOrganization(organizationName string) (types.DeveloperApps, types.Error) {

	getDeveloperAppByOrganization := func() (interface{}, types.Error) {
		return s.developerapp.GetByOrganization(organizationName)
	}
	var developerApps types.DeveloperApps
	if err := s.cache.fetchEntry(db.EntityTypeDeveloperApp, organizationName, &developerApps, getDeveloperAppByOrganization); err != nil {
		return nil, err
	}
	return developerApps, nil
}

// GetByName returns a developer app
func (s *DeveloperAppCache) GetByName(organization, developerAppName string) (*types.DeveloperApp, types.Error) {

	getDeveloperAppByName := func() (interface{}, types.Error) {
		return s.developerapp.GetByID(organization, developerAppName)
	}
	var developerApp types.DeveloperApp
	if err := s.cache.fetchEntry(db.EntityTypeDeveloperApp, developerAppName, &developerApp, getDeveloperAppByName); err != nil {
		return nil, err
	}
	return &developerApp, nil
}

// GetByID returns a developer app
func (s *DeveloperAppCache) GetByID(organization, developerAppID string) (*types.DeveloperApp, types.Error) {

	getDeveloperAppByID := func() (interface{}, types.Error) {
		return s.developerapp.GetByID(organization, developerAppID)
	}
	var developerApp types.DeveloperApp
	if err := s.cache.fetchEntry(db.EntityTypeDeveloperApp, developerAppID, &developerApp, getDeveloperAppByID); err != nil {
		return nil, err
	}
	return &developerApp, nil
}

// GetCountByDeveloperID retrieves number of apps belonging to a developer
func (s *DeveloperAppCache) GetCountByDeveloperID(developerID string) (int, types.Error) {

	return s.developerapp.GetCountByDeveloperID(developerID)
}

// Update UPSERTs a developer app
func (s *DeveloperAppCache) Update(app *types.DeveloperApp) types.Error {

	s.cache.deleteEntry(db.EntityTypeDeveloperApp, app.AppID)
	return s.developerapp.Update(app)
}

// DeleteByID deletes a developer app
func (s *DeveloperAppCache) DeleteByID(organizationName, developerAppID string) types.Error {

	s.cache.deleteEntry(db.EntityTypeDeveloperApp, developerAppID)
	return s.developerapp.DeleteByID(organizationName, developerAppID)
}
