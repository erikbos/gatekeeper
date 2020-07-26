package cassandra

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Prometheus label for metrics of db interactions
const routeMetricLabel = "routes"

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
func (s *RouteStore) GetAll() ([]shared.Route, error) {

	query := "SELECT * FROM routes"
	routes, err := s.runGetRouteQuery(query)
	if err != nil {
		return []shared.Route{}, err
	}

	if len(routes) == 0 {
		s.db.metrics.QueryMiss(routeMetricLabel)
		return []shared.Route{}, errors.New("Can not retrieve list of routes")
	}

	s.db.metrics.QueryHit(routeMetricLabel)
	return routes, nil
}

// GetRouteByName retrieves a route from database
func (s *RouteStore) GetRouteByName(routeName string) (*shared.Route, error) {

	query := "SELECT * FROM routes WHERE name = ? LIMIT 1"
	routes, err := s.runGetRouteQuery(query, routeName)
	if err != nil {
		return nil, err
	}

	if len(routes) == 0 {
		s.db.metrics.QueryMiss(routeMetricLabel)
		return nil, fmt.Errorf("Can not find route (%s)", routeName)
	}

	s.db.metrics.QueryHit(routeMetricLabel)
	return &routes[0], nil
}

// runGetRouteQuery executes CQL query and returns resultset
func (s *RouteStore) runGetRouteQuery(query string, queryParameters ...interface{}) ([]shared.Route, error) {
	var routes []shared.Route

	timer := prometheus.NewTimer(s.db.metrics.LookupHistogram)
	defer timer.ObserveDuration()

	iter := s.db.CassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		newRoute := shared.Route{
			Name:           m["name"].(string),
			DisplayName:    m["display_name"].(string),
			RouteGroup:     m["route_group"].(string),
			Path:           m["path"].(string),
			PathType:       m["path_type"].(string),
			Cluster:        m["cluster"].(string),
			CreatedAt:      m["created_at"].(int64),
			CreatedBy:      m["created_by"].(string),
			LastmodifiedAt: m["lastmodified_at"].(int64),
			LastmodifiedBy: m["lastmodified_by"].(string),
		}
		if m["attributes"] != nil {
			newRoute.Attributes = s.db.UnmarshallJSONArrayOfAttributes(m["attributes"].(string))
		}
		routes = append(routes, newRoute)
		m = map[string]interface{}{}
	}
	// In case query failed we return query error
	if err := iter.Close(); err != nil {
		log.Error(err)
		return []shared.Route{}, err
	}
	return routes, nil
}

// UpdateRouteByName UPSERTs an route
func (s *RouteStore) UpdateRouteByName(route *shared.Route) error {

	route.Attributes = shared.TidyAttributes(route.Attributes)
	route.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	if err := s.db.CassandraSession.Query(`INSERT INTO routes (
name,
display_name,
route_group,
path,
path_type,
cluster,
attributes,
created_at,
created_by,
lastmodified_at,
lastmodified_by) VALUES(?,?,?,?,?,?,?,?,?,?,?)`,

		route.Name,
		route.DisplayName,
		route.RouteGroup,
		route.Path,
		route.PathType,
		route.Cluster,
		s.db.MarshallArrayOfAttributesToJSON(route.Attributes),
		route.CreatedAt,
		route.CreatedBy,
		route.LastmodifiedAt,
		route.LastmodifiedBy).Exec(); err != nil {

		return fmt.Errorf("Cannot update route '%s' (%v)", route.Name, err)
	}
	return nil
}

// DeleteRouteByName deletes a route
func (s *RouteStore) DeleteRouteByName(routeToDelete string) error {

	_, err := s.GetRouteByName(routeToDelete)
	if err != nil {
		return err
	}

	query := "DELETE FROM routes WHERE name = ?"
	return s.db.CassandraSession.Query(query, routeToDelete).Exec()
}
