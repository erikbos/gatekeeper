package cassandra

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (

	// Prometheus label for metrics of db interactions
	routeMetricLabel = "routes"

	// List of route columns we use
	routeColumns = `name,
display_name,
route_group,
path,
path_type,
attributes,
created_at,
created_by,
lastmodified_at,
lastmodified_by`
)

// RouteStore holds our route config
type RouteStore struct {
	db *Database
}

// NewRouteStore creates route instance
func NewRouteStore(database *Database) *RouteStore {
	return &RouteStore{
		db: database,
	}
}

// GetAll retrieves all routes
func (s *RouteStore) GetAll() (types.Routes, types.Error) {

	query := "SELECT * FROM routes"
	routes, err := s.runGetRouteQuery(query)
	if err != nil {
		s.db.metrics.QueryFailed(routeMetricLabel)
		return types.NullRoutes, types.NewDatabaseError(err)
	}

	s.db.metrics.QueryHit(routeMetricLabel)
	return routes, nil
}

// Get retrieves a route from database
func (s *RouteStore) Get(routeName string) (*types.Route, types.Error) {

	query := "SELECT * FROM routes WHERE name = ? LIMIT 1"
	routes, err := s.runGetRouteQuery(query, routeName)
	if err != nil {
		s.db.metrics.QueryFailed(routeMetricLabel)
		return nil, types.NewDatabaseError(err)
	}

	if len(routes) == 0 {
		s.db.metrics.QueryMiss(routeMetricLabel)
		return nil, types.NewItemNotFoundError(
			fmt.Errorf("can not find route '%s'", routeName))
	}

	s.db.metrics.QueryHit(routeMetricLabel)
	return &routes[0], nil
}

// runGetRouteQuery executes CQL query and returns resultset
func (s *RouteStore) runGetRouteQuery(query string, queryParameters ...interface{}) (types.Routes, error) {
	var routes types.Routes

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	iter := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		routes = append(routes, types.Route{
			Attributes:     AttributesUnmarshal(columnValueString(m, "attributes")),
			CreatedAt:      columnValueInt64(m, "created_at"),
			CreatedBy:      columnValueString(m, "created_by"),
			DisplayName:    columnValueString(m, "display_name"),
			LastModifiedAt: columnValueInt64(m, "lastmodified_at"),
			LastModifiedBy: columnValueString(m, "lastmodified_by"),
			Name:           columnValueString(m, "name"),
			Path:           columnValueString(m, "path"),
			PathType:       columnValueString(m, "path_type"),
			RouteGroup:     columnValueString(m, "route_group"),
		})
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iter.Close(); err != nil {
		return types.Routes{}, err
	}
	return routes, nil
}

// Update UPSERTs an route
func (s *RouteStore) Update(r *types.Route) types.Error {

	query := "INSERT INTO routes (" + routeColumns + ") VALUES(?,?,?,?,?,?,?,?,?,?)"
	if err := s.db.CassandraSession.Query(query,
		r.Name,
		r.DisplayName,
		r.RouteGroup,
		r.Path,
		r.PathType,
		AttributesMarshal(r.Attributes),
		r.CreatedAt,
		r.CreatedBy,
		r.LastModifiedAt,
		r.LastModifiedBy).Exec(); err != nil {

		s.db.metrics.QueryFailed(routeMetricLabel)
		return types.NewDatabaseError(
			fmt.Errorf("cannot update route '%s' (%s)", r.Name, err))
	}
	return nil
}

// Delete deletes a route
func (s *RouteStore) Delete(routeToDelete string) types.Error {

	query := "DELETE FROM routes WHERE name = ?"
	if err := s.db.CassandraSession.Query(query, routeToDelete).Exec(); err != nil {
		s.db.metrics.QueryFailed(routeMetricLabel)
		return types.NewDatabaseError(err)
	}
	return nil
}
