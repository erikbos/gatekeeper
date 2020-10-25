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
	if err := s.cache.fetchEntity(types.TypeAPIProductName, "", &apiproducts, getAll); err != nil {
		return nil, err
	}
	return apiproducts, nil
}

// Get returns an apiproduct
func (s *APIProductCache) Get(apiproductName string) (*types.APIProduct, types.Error) {

	getAPIProduct := func() (interface{}, types.Error) {
		return s.apiproduct.Get(apiproductName)
	}
	var apiproduct types.APIProduct
	if err := s.cache.fetchEntity(types.TypeAPIProductName, apiproductName, &apiproduct, getAPIProduct); err != nil {
		return nil, err
	}
	return &apiproduct, nil
}

// Update UPSERTs an apiproduct in database
func (s *APIProductCache) Update(p *types.APIProduct) types.Error {

	s.cache.deleteEntry(types.TypeAPIProductName, p.Name)
	return s.apiproduct.Update(p)
}

// Delete deletes an apiproduct
func (s *APIProductCache) Delete(apiProduct string) types.Error {

	s.cache.deleteEntry(types.TypeAPIProductName, apiProduct)
	return s.apiproduct.Delete(apiProduct)
}
