package cassandra

import (
	"encoding/json"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// List of role columns we use
	roleColumns = `name,
display_name,
permissions,
created_at,
created_by,
lastmodified_at,
lastmodified_by`

	// Prometheus label for metrics of db interactions
	roleMetricLabel = "roles"
)

// RoleStore holds our database config
type RoleStore struct {
	db *Database
}

// NewRoleStore creates role instance
func NewRoleStore(database *Database) *RoleStore {
	return &RoleStore{
		db: database,
	}
}

// GetAll retrieves all roles
func (s *RoleStore) GetAll() (types.Roles, types.Error) {

	query := "SELECT " + roleColumns + " FROM roles"
	roles, err := s.runGetRoleQuery(query)
	if err != nil {
		s.db.metrics.QueryFailed(roleMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	s.db.metrics.QuerySuccessful(roleMetricLabel)
	return roles, nil
}

// Get retrieves a role from database
func (s *RoleStore) Get(roleName string) (*types.Role, types.Error) {

	query := "SELECT " + roleColumns + " FROM roles WHERE name = ? LIMIT 1"
	roles, err := s.runGetRoleQuery(query, roleName)
	if err != nil {
		s.db.metrics.QueryFailed(roleMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	if len(roles) == 0 {
		s.db.metrics.QueryNotFound(roleMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("cannot find role '%s'", roleName))
	}

	s.db.metrics.QuerySuccessful(roleMetricLabel)
	return &roles[0], nil
}

// runGetRoleQuery executes CQL query and returns resultset
func (s *RoleStore) runGetRoleQuery(query string, queryParameters ...interface{}) (types.Roles, error) {
	var roles types.Roles

	timer := prometheus.NewTimer(s.db.metrics.queryHistogram)
	defer timer.ObserveDuration()

	iter := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		roles = append(roles, types.Role{
			Name:           columnToString(m, "name"),
			DisplayName:    columnToString(m, "display_name"),
			Permissions:    PermissionsUnmarshal(columnToString(m, "permissions")),
			CreatedAt:      columnToInt64(m, "created_at"),
			CreatedBy:      columnToString(m, "created_by"),
			LastModifiedAt: columnToInt64(m, "lastmodified_at"),
			LastModifiedBy: columnToString(m, "lastmodified_by"),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iter.Close(); err != nil {
		return types.Roles{}, err
	}
	return roles, nil
}

// Update UPSERTs an role in database
func (s *RoleStore) Update(c *types.Role) types.Error {

	query := "INSERT INTO roles (" + roleColumns + ") VALUES(?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		c.Name,
		c.DisplayName,
		PermissionsMarshal(c.Permissions),
		c.CreatedAt,
		c.CreatedBy,
		c.LastModifiedAt,
		c.LastModifiedBy).Exec(); err != nil {

		s.db.metrics.QueryFailed(roleMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("cannot update role '%s' (%s)", c.Name, err))
	}
	return nil
}

// Delete deletes a role
func (s *RoleStore) Delete(roleToDelete string) types.Error {

	query := "DELETE FROM roles WHERE name = ?"
	if err := s.db.CassandraSession.Query(query, roleToDelete).Exec(); err != nil {
		s.db.metrics.QueryFailed(roleMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}

// PermissionsUnmarshal unpacks JSON-encoded role permissions into Permissions
func PermissionsUnmarshal(rolePermissionsAsJSON string) types.Permissions {

	if rolePermissionsAsJSON != "" {
		var permissions types.Permissions
		if err := json.Unmarshal([]byte(rolePermissionsAsJSON), &permissions); err == nil {
			return permissions
		}
	}
	return types.NullPermissions
}

// PermissionsMarshal packs role Permissions into JSON
func PermissionsMarshal(a types.Permissions) string {

	if json, err := json.Marshal(a); err == nil {
		return string(json)
	}
	return "[]"
}
