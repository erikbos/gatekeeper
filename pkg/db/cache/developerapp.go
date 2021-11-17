package cache

import (
	"fmt"

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
func (s *DeveloperAppCache) GetAll(organizationName string) (types.DeveloperApps, types.Error) {

	getDeveloperApps := func() (interface{}, types.Error) {
		return s.developerapp.GetAll(organizationName)
	}
	var developerApps types.DeveloperApps
	// TODO/FIXME
	if err := s.cache.fetchEntity(types.TypeDeveloperAppName, "--all-apps--", &developerApps, getDeveloperApps); err != nil {
		return nil, err
	}
	return developerApps, nil
}

// GetAllByDeveloperID retrieves all developer apps from one developer
func (s *DeveloperAppCache) GetAllByDeveloperID(organizationName, developerID string) (types.DeveloperApps, types.Error) {

	getDeveloperApps := func() (interface{}, types.Error) {
		return s.developerapp.GetAllByDeveloperID(organizationName, developerID)
	}
	var developerApps types.DeveloperApps
	cacheKey := fmt.Sprintf("--all-apps-%s-", developerID)
	if err := s.cache.fetchEntity(types.TypeDeveloperAppName, cacheKey, &developerApps, getDeveloperApps); err != nil {
		return nil, err
	}
	return developerApps, nil
}

// GetByName returns a developer app
func (s *DeveloperAppCache) GetByName(organizationName string, developerEmail, developerAppName string) (*types.DeveloperApp, types.Error) {

	getDeveloperAppByName := func() (interface{}, types.Error) {
		return s.developerapp.GetByName(organizationName, developerEmail, developerAppName)
	}
	var developerApp types.DeveloperApp
	if err := s.cache.fetchEntity(types.TypeDeveloperAppName, developerAppName, &developerApp, getDeveloperAppByName); err != nil {
		return nil, err
	}
	return &developerApp, nil
}

// GetByID returns a developer app
func (s *DeveloperAppCache) GetByID(organizationName, developerAppID string) (*types.DeveloperApp, types.Error) {

	getDeveloperAppByID := func() (interface{}, types.Error) {
		return s.developerapp.GetByID(organizationName, developerAppID)
	}
	var developerApp types.DeveloperApp
	if err := s.cache.fetchEntity(types.TypeDeveloperAppName, developerAppID, &developerApp, getDeveloperAppByID); err != nil {
		return nil, err
	}
	return &developerApp, nil
}

// GetCountByDeveloperID retrieves number of apps belonging to a developer
func (s *DeveloperAppCache) GetCountByDeveloperID(organizationName, developerID string) (int, types.Error) {

	return s.developerapp.GetCountByDeveloperID(organizationName, developerID)
}

// Update UPSERTs a developer app
func (s *DeveloperAppCache) Update(organizationName string, app *types.DeveloperApp) types.Error {

	s.cache.deleteEntry(types.TypeDeveloperAppName, app.AppID)
	return s.developerapp.Update(organizationName, app)
}

// DeleteByID deletes a developer app
func (s *DeveloperAppCache) DeleteByID(organizationName, developerAppID string) types.Error {

	s.cache.deleteEntry(types.TypeDeveloperAppName, developerAppID)
	return s.developerapp.DeleteByID(organizationName, developerAppID)
}
