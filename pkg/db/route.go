package db

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// Prometheus label for metrics of db interactions
const routeMetricLabel = "routes"

// GetRoutes retrieves all routes
func (d *Database) GetRoutes() ([]shared.Route, error) {

	query := "SELECT * FROM routes"
	routes, err := d.runGetRouteQuery(query)
	if err != nil {
		return []shared.Route{}, err
	}

	if len(routes) == 0 {
		d.metricsQueryMiss(routeMetricLabel)
		return []shared.Route{}, errors.New("Can not retrieve list of routes")
	}

	d.metricsQueryHit(routeMetricLabel)
	return routes, nil
}

// GetRouteByName retrieves a route from database
func (d *Database) GetRouteByName(routeName string) (shared.Route, error) {

	query := "SELECT * FROM routes WHERE name = ? LIMIT 1"
	routes, err := d.runGetRouteQuery(query, routeName)
	if err != nil {
		return shared.Route{}, err
	}

	if len(routes) == 0 {
		d.metricsQueryMiss(routeMetricLabel)
		return shared.Route{},
			fmt.Errorf("Can not find route (%s)", routeName)
	}

	d.metricsQueryHit(routeMetricLabel)
	return routes[0], nil
}

// runGetRouteQuery executes CQL query and returns resultset
func (d *Database) runGetRouteQuery(query string, queryParameters ...interface{}) ([]shared.Route, error) {
	var routes []shared.Route

	timer := prometheus.NewTimer(d.dbLookupHistogram)
	defer timer.ObserveDuration()

	iter := d.cassandraSession.Query(query, queryParameters...).Iter()
	m := make(map[string]interface{})
	for iter.MapScan(m) {
		newRoute := shared.Route{
			Name:           m["name"].(string),
			DisplayName:    m["display_name"].(string),
			RouteSet:       m["route_set"].(string),
			Path:           m["path"].(string),
			PathType:       m["path_type"].(string),
			Cluster:        m["cluster"].(string),
			CreatedAt:      m["created_at"].(int64),
			CreatedBy:      m["created_by"].(string),
			LastmodifiedAt: m["lastmodified_at"].(int64),
			LastmodifiedBy: m["lastmodified_by"].(string),
		}
		if m["attributes"] != nil {
			newRoute.Attributes = d.unmarshallJSONArrayOfAttributes(m["attributes"].(string))
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

// UpdateRouteByName UPSERTs an route in database
func (d *Database) UpdateRouteByName(route *shared.Route) error {

	route.Attributes = shared.TidyAttributes(route.Attributes)
	route.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()

	if err := d.cassandraSession.Query(`INSERT INTO routes (
name,
display_name,
route_set,
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
		route.RouteSet,
		route.Path,
		route.PathType,
		route.Cluster,
		d.marshallArrayOfAttributesToJSON(route.Attributes),
		route.CreatedAt,
		route.CreatedBy,
		route.LastmodifiedAt,
		route.LastmodifiedBy).Exec(); err != nil {

		return fmt.Errorf("Cannot update route '%s' (%v)", route.Name, err)
	}
	return nil
}

// DeleteRouteByName deletes a route
func (d *Database) DeleteRouteByName(routeToDelete string) error {

	_, err := d.GetRouteByName(routeToDelete)
	if err != nil {
		return err
	}

	query := "DELETE FROM routes WHERE name = ?"
	return d.cassandraSession.Query(query, routeToDelete).Exec()
}
