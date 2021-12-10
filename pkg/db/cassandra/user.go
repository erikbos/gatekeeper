package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const ( // List of user columns we use
	userColumns = `name,
display_name,
password,
status,
roles,
created_at,
created_by,
lastmodified_at,
lastmodified_by`

	// Prometheus label for metrics of db interactions
	userMetricLabel = "users"
)

// UserStore holds our database config
type UserStore struct {
	db *Database
}

// NewUserStore creates user instance
func NewUserStore(database *Database) *UserStore {
	return &UserStore{
		db: database,
	}
}

// GetAll retrieves all users
func (s *UserStore) GetAll() (types.Users, types.Error) {

	query := "SELECT " + userColumns + " FROM users"
	users, err := s.runGetUserQuery(query)
	if err != nil {
		s.db.metrics.QueryFailed(userMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	s.db.metrics.QuerySuccessful(userMetricLabel)
	return users, nil
}

// Get retrieves a user from database
func (s *UserStore) Get(userName string) (*types.User, types.Error) {

	query := "SELECT " + userColumns + " FROM users WHERE name = ? LIMIT 1"
	users, err := s.runGetUserQuery(query, userName)
	if err != nil {
		s.db.metrics.QueryFailed(userMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	if len(users) == 0 {
		s.db.metrics.QueryNotFound(userMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("cannot find user '%s'", userName))
	}

	s.db.metrics.QuerySuccessful(userMetricLabel)
	return &users[0], nil
}

// runGetUserQuery executes CQL query and returns resultset
func (s *UserStore) runGetUserQuery(query string, queryParameters ...interface{}) (types.Users, error) {
	var users types.Users

	timer := prometheus.NewTimer(s.db.metrics.queryHistogram)
	defer timer.ObserveDuration()

	iter := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		users = append(users, types.User{
			Name:           columnToString(m, "name"),
			DisplayName:    columnToString(m, "display_name"),
			Password:       columnToString(m, "password"),
			Status:         columnToString(m, "status"),
			Roles:          columnToStringSlice(m, "roles"),
			CreatedAt:      columnToInt64(m, "created_at"),
			CreatedBy:      columnToString(m, "created_by"),
			LastModifiedAt: columnToInt64(m, "lastmodified_at"),
			LastModifiedBy: columnToString(m, "lastmodified_by"),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iter.Close(); err != nil {
		return types.Users{}, err
	}
	return users, nil
}

// Update UPSERTs an user in database
func (s *UserStore) Update(c *types.User) types.Error {

	query := "INSERT INTO users (" + userColumns + ") VALUES(?,?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		c.Name,
		c.DisplayName,
		c.Password,
		c.Status,
		c.Roles,
		c.CreatedAt,
		c.CreatedBy,
		c.LastModifiedAt,
		c.LastModifiedBy).Exec(); err != nil {

		s.db.metrics.QueryFailed(userMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("cannot update user '%s' (%s)", c.Name, err))
	}
	return nil
}

// Delete deletes a user
func (s *UserStore) Delete(userToDelete string) types.Error {

	query := "DELETE FROM users WHERE name = ?"
	if err := s.db.CassandraSession.Query(query, userToDelete).Exec(); err != nil {
		s.db.metrics.QueryFailed(userMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}
