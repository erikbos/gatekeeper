package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// List of company columns we use
	companyColumns = `key,
name,
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

// GetAll retrieves all companies
func (s *CompanyStore) GetAll(organizationName string) (types.Companies, types.Error) {

	query := "SELECT " + companyColumns + " FROM companies WHERE organization_name = ?"
	companies, err := s.runGetCompanyQuery(query, organizationName)
	if err != nil {
		s.db.metrics.QueryFailed(companyMetricLabel)
		return types.NullCompanies, types.NewDatabaseError(err)
	}

	s.db.metrics.QuerySuccessful(companyMetricLabel)
	return companies, nil
}

// Get retrieves a company from database
func (s *CompanyStore) Get(organizationName, companyName string) (*types.Company, types.Error) {

	query := "SELECT " + companyColumns + " FROM companies WHERE key = ? LIMIT 1"
	companies, err := s.runGetCompanyQuery(query, s.generatePrimaryKey(organizationName, companyName))
	if err != nil {
		s.db.metrics.QueryFailed(companyMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	if len(companies) == 0 {
		s.db.metrics.QueryNotFound(companyMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("cannot find company '%s'", companyName))
	}

	s.db.metrics.QuerySuccessful(companyMetricLabel)
	return &companies[0], nil
}

// runGetCompanyQuery executes CQL query and returns resultset
func (s *CompanyStore) runGetCompanyQuery(query string, queryParameters ...interface{}) (types.Companies, error) {
	var companies types.Companies

	timer := prometheus.NewTimer(s.db.metrics.queryHistogram)
	defer timer.ObserveDuration()

	iter := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		companies = append(companies, types.Company{
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
	return companies, nil
}

// Update UPSERTs an company in database
func (s *CompanyStore) Update(organizationName string, c *types.Company) types.Error {

	query := "INSERT INTO companies (" + companyColumns + ") VALUES(?,?,?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		s.generatePrimaryKey(organizationName, c.Name),
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

	query := "DELETE FROM companies WHERE key = ?"
	key := s.generatePrimaryKey(organizationName, companyToDelete)
	if err := s.db.CassandraSession.Query(query, key).Exec(); err != nil {
		s.db.metrics.QueryFailed(companyMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}

// generatePrimaryKey returns unique primary key based upon organization & companyName
func (s *CompanyStore) generatePrimaryKey(organization, companyName string) string {
	// Combine organization and companyname to make globally unique key
	return fmt.Sprintf("%s@@@%s", organization, companyName)
}
