package cache

import (
	"github.com/erikbos/gatekeeper/pkg/db"
	"github.com/erikbos/gatekeeper/pkg/types"
)

// APIProductCache holds our database config
type APIProductCache struct {
	apiproduct db.APIProduct
	cache      *Cache
}

// NewAPIProductCache creates apiproduct instance
func NewAPIProductCache(cache *Cache, apiproduct db.APIProduct) *APIProductCache {
	return &APIProductCache{
		apiproduct: apiproduct,
		cache:      cache,
	}
}

// GetAll retrieves all api products
func (s *APIProductCache) GetAll() (types.APIProducts, types.Error) {

	getAll := func() (interface{}, types.Error) {
		return s.apiproduct.GetAll()
	}
	var apiproducts types.APIProducts
	if err := s.cache.fetchEntry("", &apiproducts, getAll); err != nil {
		return nil, err
	}
	return apiproducts, nil
}

// GetByOrganization retrieves all api products belonging to an organization
func (s *APIProductCache) GetByOrganization(organizationName string) (types.APIProducts, types.Error) {

	getByOrg := func() (interface{}, types.Error) {
		return s.apiproduct.GetByOrganization(organizationName)
	}
	var apiproducts types.APIProducts
	if err := s.cache.fetchEntry("", &apiproducts, getByOrg); err != nil {
		return nil, err
	}
	return apiproducts, nil
}

// Get returns an apiproduct
func (s *APIProductCache) Get(organizationName, apiproductName string) (*types.APIProduct, types.Error) {

	getAPIProduct := func() (interface{}, types.Error) {
		return s.apiproduct.Get(organizationName, apiproductName)
	}
	var apiproduct types.APIProduct
	if err := s.cache.fetchEntry(apiproductName, &apiproduct, getAPIProduct); err != nil {
		return nil, err
	}
	return &apiproduct, nil
}

// Update UPSERTs an apiproduct in database
func (s *APIProductCache) Update(p *types.APIProduct) types.Error {

	s.cache.deleteEntry(p.Name, p)
	return s.apiproduct.Update(p)
}

// Delete deletes an apiproduct
func (s *APIProductCache) Delete(organizationName, apiProduct string) types.Error {

	s.cache.deleteEntry(apiProduct, types.NullAPIProduct)
	return s.apiproduct.Delete(organizationName, apiProduct)
}
