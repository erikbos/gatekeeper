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

// GetAll retrieves all developer apps
func (s *DeveloperAppCache) GetAll() (types.DeveloperApps, types.Error) {

	getDeveloperApps := func() (interface{}, types.Error) {
		return s.developerapp.GetAll()
	}
	var developerApps types.DeveloperApps
	// TODO/FIXME
	if err := s.cache.fetchEntity(types.TypeDeveloperAppName, "--all-apps--", &developerApps, getDeveloperApps); err != nil {
		return nil, err
	}
	return developerApps, nil
}

// GetByName returns a developer app
func (s *DeveloperAppCache) GetByName(developerAppName string) (*types.DeveloperApp, types.Error) {

	getDeveloperAppByName := func() (interface{}, types.Error) {
		return s.developerapp.GetByName(developerAppName)
	}
	var developerApp types.DeveloperApp
	if err := s.cache.fetchEntity(types.TypeDeveloperAppName, developerAppName, &developerApp, getDeveloperAppByName); err != nil {
		return nil, err
	}
	return &developerApp, nil
}

// GetByID returns a developer app
func (s *DeveloperAppCache) GetByID(developerAppID string) (*types.DeveloperApp, types.Error) {

	getDeveloperAppByID := func() (interface{}, types.Error) {
		return s.developerapp.GetByID(developerAppID)
	}
	var developerApp types.DeveloperApp
	if err := s.cache.fetchEntity(types.TypeDeveloperAppName, developerAppID, &developerApp, getDeveloperAppByID); err != nil {
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

	s.cache.deleteEntry(types.TypeDeveloperAppName, app.AppID)
	return s.developerapp.Update(app)
}

// DeleteByID deletes a developer app
func (s *DeveloperAppCache) DeleteByID(developerAppID string) types.Error {

	s.cache.deleteEntry(types.TypeDeveloperAppName, developerAppID)
	return s.developerapp.DeleteByID(developerAppID)
}
