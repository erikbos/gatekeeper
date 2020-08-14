package cassandra

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Prometheus label for metrics of db interactions
const apiProductsMetricLabel = "apiproducts"

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
func (s *APIProductStore) GetAll() ([]shared.APIProduct, error) {
	query := "SELECT * FROM api_products"

	apiproducts, err := s.runGetAPIProductQuery(query)
	if err != nil {
		return []shared.APIProduct{}, err
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
func (s *APIProductStore) GetByOrganization(organizationName string) ([]shared.APIProduct, error) {
	query := "SELECT * FROM api_products WHERE organization_name = ? ALLOW FILTERING"

	apiproducts, err := s.runGetAPIProductQuery(query, organizationName)
	if err != nil {
		return []shared.APIProduct{}, err
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
func (s *APIProductStore) GetByName(organizationName, apiproductName string) (*shared.APIProduct, error) {

	query := "SELECT * FROM api_products WHERE organization_name = ? AND name = ? LIMIT 1"

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
func (s *APIProductStore) runGetAPIProductQuery(query string, queryParameters ...interface{}) ([]shared.APIProduct, error) {
	var apiproducts []shared.APIProduct

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
		return []shared.APIProduct{}, nil
	}

	m := make(map[string]interface{})
	for iter.MapScan(m) {
		apiproduct := shared.APIProduct{
			Name:             m["name"].(string),
			DisplayName:      m["display_name"].(string),
			Description:      m["description"].(string),
			RouteGroup:       m["route_group"].(string),
			OrganizationName: m["organization_name"].(string),
			Policies:         m["policies"].(string),
			CreatedAt:        m["created_at"].(int64),
			CreatedBy:        m["created_by"].(string),
			LastmodifiedAt:   m["lastmodified_at"].(int64),
			LastmodifiedBy:   m["lastmodified_by"].(string),
		}
		apiproduct.Paths = s.db.UnmarshallJSONArrayOfStrings(m["paths"].(string))
		apiproduct.Attributes = s.db.UnmarshallJSONArrayOfAttributes(m["attributes"].(string))
		apiproducts = append(apiproducts, apiproduct)
		m = map[string]interface{}{}
	}

	if err := iter.Close(); err != nil {
		log.Error(err)
		return []shared.APIProduct{}, err
	}

	return apiproducts, nil
}

// UpdateByName UPSERTs an apiproduct in database
func (s *APIProductStore) UpdateByName(p *shared.APIProduct) error {

	p.Attributes = shared.TidyAttributes(p.Attributes)
	p.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	if err := s.db.CassandraSession.Query(`INSERT INTO api_products (
name,
display_name,
attributes,
route_group,
paths,
policies,
created_at,
created_by,
lastmodified_at,
lastmodified_by,
organization_name) VALUES(?,?,?,?,?,?,?,?,?,?,?)`,

		p.Name,
		p.DisplayName,
		s.db.MarshallArrayOfAttributesToJSON(p.Attributes),
		p.RouteGroup,
		s.db.MarshallArrayOfStringsToJSON(p.Paths),
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
