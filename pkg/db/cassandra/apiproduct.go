package cassandra

import (
	"encoding/json"
	"fmt"

	"github.com/gocql/gocql"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// List of apiproduct columns we use
	apiProductsColumns = `key,
name,
description,
display_name,
organization_name,
approval_type,
api_resources,
route_group,
scopes,
policies,
attributes,
created_at,
created_by,
lastmodified_at,
lastmodified_by`

	// Prometheus label for metrics of db interactions
	apiProductsMetricLabel = "apiproducts"
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

	query := "SELECT " + apiProductsColumns + " FROM api_products WHERE organization_name = ?"
	apiproducts, err := s.runGetAPIProductQuery(query, organizationName)
	if err != nil {
		s.db.metrics.QueryFailed(apiProductsMetricLabel)
		return types.NullAPIProducts, types.NewDatabaseError(err)
	}

	s.db.metrics.QuerySuccessful(apiProductsMetricLabel)
	return apiproducts, nil
}

// Get returns an apiproduct
func (s *APIProductStore) Get(organizationName, apiproductName string) (*types.APIProduct, types.Error) {

	query := "SELECT " + apiProductsColumns + " FROM api_products WHERE key = ? LIMIT 1"
	apiproducts, err := s.runGetAPIProductQuery(query, s.generatePrimaryKey(organizationName, apiproductName))
	if err != nil {
		s.db.metrics.QueryFailed(apiProductsMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	if len(apiproducts) == 0 {
		s.db.metrics.QueryNotFound(apiProductsMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("cannot find apiproduct '%s'", apiproductName))
	}

	s.db.metrics.QuerySuccessful(apiProductsMetricLabel)
	return &apiproducts[0], nil
}

// runGetAPIProductQuery executes CQL query and returns resultset
func (s *APIProductStore) runGetAPIProductQuery(query string, queryParameters ...interface{}) (types.APIProducts, error) {
	var apiproducts types.APIProducts

	timer := prometheus.NewTimer(s.db.metrics.queryHistogram)
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
			Name:           columnToString(m, "name"),
			Description:    columnToString(m, "description"),
			DisplayName:    columnToString(m, "display_name"),
			ApprovalType:   columnToString(m, "approval_type"),
			APIResources:   columnToStringSlice(m, "api_resources"),
			RouteGroup:     columnToString(m, "route_group"),
			Scopes:         columnToStringSlice(m, "scopes"),
			Policies:       columnToString(m, "policies"),
			Attributes:     columnToAttributes(m, "attributes"),
			CreatedAt:      columnToInt64(m, "created_at"),
			CreatedBy:      columnToString(m, "created_by"),
			LastModifiedAt: columnToInt64(m, "lastmodified_at"),
			LastModifiedBy: columnToString(m, "lastmodified_by"),
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

	query := "INSERT INTO api_products (" + apiProductsColumns + ") VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		s.generatePrimaryKey(organizationName, p.Name),
		p.Name,
		p.Description,
		p.DisplayName,
		organizationName,
		p.ApprovalType,
		p.APIResources,
		p.RouteGroup,
		p.Scopes,
		p.Policies,
		attributesToColumn(p.Attributes),
		p.CreatedAt,
		p.CreatedBy,
		p.LastModifiedAt,
		p.LastModifiedBy).Exec(); err != nil {

		s.db.metrics.QueryFailed(apiProductsMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("cannot update apiproduct '%s' (%s)", p.Name, err))
	}
	return nil
}

// Delete deletes an apiproduct
func (s *APIProductStore) Delete(organizationName, apiProduct string) types.Error {

	query := "DELETE FROM api_products WHERE key = ?"
	key := s.generatePrimaryKey(organizationName, apiProduct)
	if err := s.db.CassandraSession.Query(query, key).Exec(); err != nil {
		s.db.metrics.QueryFailed(apiProductsMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}

// generatePrimaryKey returns unique primary key based upon organization & apiproduct
func (s *APIProductStore) generatePrimaryKey(organization, apiProduct string) string {
	// Combine organization and apiproduct to make globally unique key
	return fmt.Sprintf("%s@@@%s", organization, apiProduct)
}

// Unmarshal unpacks a key's product statuses
// Example input: [{"name":"S","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
func KeyAPIProductStatusesUnmarshal(jsonProductStatuses string) types.KeyAPIProductStatuses {

	if jsonProductStatuses != "" {
		var productStatus = make([]types.KeyAPIProductStatus, 0)
		if err := json.Unmarshal([]byte(jsonProductStatuses), &productStatus); err == nil {
			return productStatus
		}
	}
	return types.KeyAPIProductStatuses{}
}

// Marshal packs a key's product statuses into JSON
// Example input: [{"name":"DisplayName","value":"erikbos teleporter"},{"name":"ErikbosTeleporterExtraAttribute","value":"42"}]
func KeyAPIProductStatusesMarshal(ps types.KeyAPIProductStatuses) string {

	if len(ps) > 0 {
		ArrayOfAttributesInJSON, err := json.Marshal(ps)
		if err == nil {
			return string(ArrayOfAttributesInJSON)
		}
	}
	return "[]"
}
