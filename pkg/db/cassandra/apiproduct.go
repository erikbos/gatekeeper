package cassandra

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// Prometheus label for metrics of db interactions
	apiProductsMetricLabel = "apiproducts"

	// List of apiproduct columns we use
	apiProductsColumns = `approval_type,
api_resources,
attributes,
created_at,
created_by,
description,
display_name,
lastmodified_at,
lastmodified_by,
name,
route_group,
policies`
)

// APIProductStore holds our database config
type APIProductStore struct {
	db *Database
}

// NewAPIProductStore creates api product instance
func NewAPIProductStore(database *Database) *APIProductStore {
	return &APIProductStore{
		db: database,
	}
}

// GetAll retrieves all api products
func (s *APIProductStore) GetAll(organizationName string) (types.APIProducts, types.Error) {

	query := "SELECT " + apiProductsColumns + " FROM api_products"

	apiproducts, err := s.runGetAPIProductQuery(query)
	if err != nil {
		s.db.metrics.QueryFailed(apiProductsMetricLabel)
		return types.NullAPIProducts, types.NewDatabaseError(err)
	}

	s.db.metrics.QueryHit(apiProductsMetricLabel)
	return apiproducts, nil
}

// Get returns an apiproduct
func (s *APIProductStore) Get(organizationName, apiproductName string) (*types.APIProduct, types.Error) {

	query := "SELECT " + apiProductsColumns + " FROM api_products WHERE name = ? LIMIT 1"

	apiproducts, err := s.runGetAPIProductQuery(query, apiproductName)
	if err != nil {
		s.db.metrics.QueryFailed(apiProductsMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	if len(apiproducts) == 0 {
		s.db.metrics.QueryMiss(apiProductsMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("cannot find apiproduct '%s'", apiproductName))
	}

	s.db.metrics.QueryHit(apiProductsMetricLabel)
	return &apiproducts[0], nil
}

// runGetAPIProductQuery executes CQL query and returns resultset
func (s *APIProductStore) runGetAPIProductQuery(query string, queryParameters ...interface{}) (types.APIProducts, error) {
	var apiproducts types.APIProducts

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	var iter *gocql.Iter
	if queryParameters == nil {
		iter = s.db.CassandraSession.Query(query).Iter()
	} else {
		iter = s.db.CassandraSession.Query(query, queryParameters...).Iter()
	}
	if iter.NumRows() == 0 {
		_ = iter.Close()
		return types.NullAPIProducts, nil
	}

	m := make(map[string]interface{})
	for iter.MapScan(m) {
		apiproducts = append(apiproducts, types.APIProduct{
			ApprovalType:   columnValueString(m, "approval_type"),
			Attributes:     AttributesUnmarshal(columnValueString(m, "attributes")),
			CreatedAt:      columnValueInt64(m, "created_at"),
			CreatedBy:      columnValueString(m, "created_by"),
			Description:    columnValueString(m, "description"),
			DisplayName:    columnValueString(m, "display_name"),
			LastModifiedAt: columnValueInt64(m, "lastmodified_at"),
			LastModifiedBy: columnValueString(m, "lastmodified_by"),
			Name:           columnValueString(m, "name"),
			APIResources:   stringSliceUnmarshal(columnValueString(m, "api_resources")),
			Policies:       columnValueString(m, "policies"),
			RouteGroup:     columnValueString(m, "route_group"),
		})
		m = map[string]interface{}{}
	}

	if err := iter.Close(); err != nil {
		return types.NullAPIProducts, err
	}

	return apiproducts, nil
}

// Update UPSERTs an apiproduct in database
func (s *APIProductStore) Update(organizationName string, p *types.APIProduct) types.Error {

	query := "INSERT INTO api_products (" + apiProductsColumns + ") VALUES(?,?,?,?,?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		p.ApprovalType,
		stringSliceMarshal(p.APIResources),
		AttributesMarshal(p.Attributes),
		p.CreatedAt,
		p.CreatedBy,
		p.Description,
		p.DisplayName,
		p.LastModifiedAt,
		p.LastModifiedBy,
		p.Name,
		p.Policies,
		p.RouteGroup).Exec(); err != nil {

		s.db.metrics.QueryFailed(apiProductsMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("cannot update apiproduct '%s'", p.Name))
	}
	return nil
}

// Delete deletes an apiproduct
func (s *APIProductStore) Delete(organizationName, apiProduct string) types.Error {

	// apiproduct, err := s.Get(apiProduct)
	// if err != nil {
	// 	s.db.metrics.QueryFailed(apiProductsMetricLabel)
	// 	return err
	// }

	query := "DELETE FROM api_products WHERE name = ?"
	if err := s.db.CassandraSession.Query(query, apiProduct).Exec(); err != nil {
		s.db.metrics.QueryFailed(apiProductsMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}
