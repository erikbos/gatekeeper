package cassandra

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// Prometheus label for metrics of db interactions
	apiProductsMetricLabel = "apiproducts"

	// List of apiproduct columns we use
	apiProductsColumns = `name,
display_name,
description,
attributes,
route_group,
paths,
policies,
created_at,
created_by,
lastmodified_at,
lastmodified_by,
organization_name`
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
func (s *APIProductStore) GetAll() (types.APIProducts, error) {

	query := "SELECT " + apiProductsColumns + " FROM api_products"

	apiproducts, err := s.runGetAPIProductQuery(query)
	if err != nil {
		return types.APIProducts{}, err
	}

	if len(apiproducts) == 0 {
		//db.metrics.QueryMiss
		s.db.metrics.QueryMiss(apiProductsMetricLabel)
		return apiproducts, fmt.Errorf("Can not find apiproducts")
	}

	s.db.metrics.QueryHit(apiProductsMetricLabel)
	return apiproducts, nil
}

// GetByOrganization retrieves all api products belonging to an organization
func (s *APIProductStore) GetByOrganization(organizationName string) (types.APIProducts, error) {
	query := "SELECT " + apiProductsColumns + " FROM api_products WHERE organization_name = ? ALLOW FILTERING"

	apiproducts, err := s.runGetAPIProductQuery(query, organizationName)
	if err != nil {
		return types.APIProducts{}, err
	}

	if len(apiproducts) == 0 {
		//db.metrics.QueryMiss
		s.db.metrics.QueryMiss(apiProductsMetricLabel)
		return apiproducts,
			fmt.Errorf("Can not find apiproducts in organization %s", organizationName)
	}

	s.db.metrics.QueryHit(apiProductsMetricLabel)
	return apiproducts, nil
}

// GetByName returns an apiproduct
func (s *APIProductStore) GetByName(organizationName, apiproductName string) (*types.APIProduct, error) {

	query := "SELECT " + apiProductsColumns + " FROM api_products WHERE organization_name = ? AND name = ? LIMIT 1"

	apiproducts, err := s.runGetAPIProductQuery(query, organizationName, apiproductName)
	if err != nil {
		return nil, err
	}

	if len(apiproducts) == 0 {
		s.db.metrics.QueryMiss(apiProductsMetricLabel)
		return nil, fmt.Errorf("Could not find apiproduct (%s)", apiproductName)
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
		return types.APIProducts{}, nil
	}

	m := make(map[string]interface{})
	for iter.MapScan(m) {
		apiproducts = append(apiproducts, types.APIProduct{
			Name:             columnValueString(m, "name"),
			DisplayName:      columnValueString(m, "display_name"),
			Description:      m["description"].(string),
			Attributes:       types.APIProduct{}.Attributes.Unmarshal(columnValueString(m, "attributes")),
			RouteGroup:       m["route_group"].(string),
			Paths:            types.APIProduct{}.Paths.Unmarshal(columnValueString(m, "paths")),
			Policies:         m["policies"].(string),
			CreatedAt:        columnValueInt64(m, "created_at"),
			CreatedBy:        columnValueString(m, "created_by"),
			LastmodifiedAt:   columnValueInt64(m, "lastmodified_at"),
			LastmodifiedBy:   columnValueString(m, "lastmodified_by"),
			OrganizationName: m["organization_name"].(string),
		})
		m = map[string]interface{}{}
	}

	if err := iter.Close(); err != nil {
		log.Error(err)
		return types.APIProducts{}, err
	}

	return apiproducts, nil
}

// UpdateByName UPSERTs an apiproduct in database
func (s *APIProductStore) UpdateByName(p *types.APIProduct) error {

	p.Attributes.Tidy()
	p.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	query := "INSERT INTO api_products (" + apiProductsColumns + ") VALUES(?,?,?,?,?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		p.Name,
		p.DisplayName,
		p.Description,
		p.Attributes.Marshal(),
		p.RouteGroup,
		p.Paths.Marshal(),
		p.Policies,
		p.CreatedAt,
		p.CreatedBy,
		p.LastmodifiedAt,
		p.LastmodifiedBy,
		p.OrganizationName).Exec(); err != nil {

		return fmt.Errorf("Can not update apiproduct '%s' (%v)", p.Name, err)
	}
	return nil
}

// DeleteByName deletes an apiproduct
func (s *APIProductStore) DeleteByName(organizationName, apiProduct string) error {

	apiproduct, err := s.GetByName(organizationName, apiProduct)
	if err != nil {
		return err
	}

	query := "DELETE FROM api_products WHERE name = ?"
	return s.db.CassandraSession.Query(query, apiproduct.Name).Exec()
}
