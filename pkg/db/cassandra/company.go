package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// List of company columns we use
	companyColumns = `name,
organization_name,
display_name,
attributes,
status,
created_at,
created_by,
lastmodified_at,
lastmodified_by`
	// Prometheus label for metrics of db interactions
	companyMetricLabel = "companies"
)

// CompanyStore holds our database config
type CompanyStore struct {
	db *Database
}

// NewCompanyStore creates company instance
func NewCompanyStore(database *Database) *CompanyStore {
	return &CompanyStore{
		db: database,
	}
}

// GetAll retrieves all companys
func (s *CompanyStore) GetAll(organizationName string) (types.Companies, types.Error) {

	query := "SELECT " + companyColumns + " FROM companys WHERE organization_name = ?"
	companys, err := s.runGetCompanyQuery(query, organizationName)
	if err != nil {
		s.db.metrics.QueryFailed(companyMetricLabel)
		return types.NullCompanies, types.NewDatabaseError(err)
	}

	s.db.metrics.QuerySuccessful(companyMetricLabel)
	return companys, nil
}

// Get retrieves a company from database
func (s *CompanyStore) Get(organizationName, companyName string) (*types.Company, types.Error) {

	query := "SELECT " + companyColumns + " FROM companys WHERE WHERE organization_name = ? AND name = ? LIMIT 1"
	companys, err := s.runGetCompanyQuery(query, organizationName, companyName)
	if err != nil {
		s.db.metrics.QueryFailed(companyMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	if len(companys) == 0 {
		s.db.metrics.QueryNotFound(companyMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("cannot find company '%s'", companyName))
	}

	s.db.metrics.QuerySuccessful(companyMetricLabel)
	return &companys[0], nil
}

// runGetCompanyQuery executes CQL query and returns resultset
func (s *CompanyStore) runGetCompanyQuery(query string, queryParameters ...interface{}) (types.Companies, error) {
	var companys types.Companies

	timer := prometheus.NewTimer(s.db.metrics.queryHistogram)
	defer timer.ObserveDuration()

	iter := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		companys = append(companys, types.Company{
			Name:           columnToString(m, "name"),
			DisplayName:    columnToString(m, "display_name"),
			Attributes:     columnToAttributes(m, "attributes"),
			Status:         columnToString(m, "status"),
			CreatedAt:      columnToInt64(m, "created_at"),
			CreatedBy:      columnToString(m, "created_by"),
			LastModifiedAt: columnToInt64(m, "lastmodified_at"),
			LastModifiedBy: columnToString(m, "lastmodified_by"),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iter.Close(); err != nil {
		return types.Companies{}, err
	}
	return companys, nil
}

// Update UPSERTs an company in database
func (s *CompanyStore) Update(organizationName string, c *types.Company) types.Error {

	query := "INSERT INTO companys (" + companyColumns + ") VALUES(?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		c.Name,
		organizationName,
		c.DisplayName,
		attributesToColumn(c.Attributes),
		c.Status,
		c.CreatedAt,
		c.CreatedBy,
		c.LastModifiedAt,
		c.LastModifiedBy).Exec(); err != nil {

		s.db.metrics.QueryFailed(companyMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("cannot update company '%s' (%s)", c.Name, err))
	}
	return nil
}

// Delete deletes a company
func (s *CompanyStore) Delete(organizationName, companyToDelete string) types.Error {

	query := "DELETE FROM companys WHERE organizationName = ? and name = ?"
	if err := s.db.CassandraSession.Query(query, organizationName, companyToDelete).Exec(); err != nil {
		s.db.metrics.QueryFailed(companyMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}
