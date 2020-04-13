package db

import (
	"fmt"

	"github.com/erikbos/apiauth/pkg/types"
)

// Prometheus label for metrics of db interactions
const routeMetricLabel = "routes"

// temp
var routes = []types.Route{
	{
		Name:        "people",
		MatchPrefix: "/people",
		Cluster:     "people",
	},
	{
		Name:        "google",
		MatchPrefix: "/google/",
		Cluster:     "google",
		HostRewrite: "www.google.com",
	},
	{
		Name:                 "default",
		MatchPrefix:          "/",
		DirectResponseBody:   "Nobody home!",
		DirectResponseStatus: 404,
	},
}

// GetRoutes retrieves all routes
func (d *Database) GetRoutes() ([]types.Route, error) {
	return routes, nil

	// query := "SELECT * FROM routes"
	// routes, err := d.runGetRouteQuery(query)
	// if err != nil {
	// 	return []types.Route{}, err
	// }
	// if len(routes) == 0 {
	// 	d.metricsQueryMiss(routeMetricLabel)
	// 	return []types.Route{}, errors.New("Can not retrieve list of routes")
	// }
	// d.metricsQueryHit(routeMetricLabel)
	// return routes, nil
}

// GetRouteByName retrieves a route from database
func (d *Database) GetRouteByName(routeName string) (types.Route, error) {
	for _, value := range routes {
		if value.Name == routeName {
			return value, nil
		}
	}
	return types.Route{}, fmt.Errorf("Can not find route (%s)", routeName)

	// query := "SELECT * FROM routez WHERE key = ? LIMIT 1"
	// routes, err := d.runGetRouteQuery(query, routeName)
	// if err != nil {
	// 	return types.Route{}, err
	// }
	// if len(routes) == 0 {
	// 	d.metricsQueryMiss(routeMetricLabel)
	// 	return types.Route{},
	// 		fmt.Errorf("Can not find route (%s)", routeName)
	// }
	// d.metricsQueryHit(routeMetricLabel)
	// return routes[0], nil
}

// runGetRouteQuery executes CQL query and returns resultset
func (d *Database) runGetRouteQuery(query string, queryParameters ...interface{}) ([]types.Route, error) {
	// 	var routes []types.Route

	// 	timer := prometheus.NewTimer(d.dbLookupHistogram)
	// 	defer timer.ObserveDuration()

	// 	iter := d.cassandraSession.Query(query, queryParameters...).Iter()
	// 	m := make(map[string]interface{})
	// 	for iter.MapScan(m) {
	// 		newRoute := types.Route{
	// 			Name:           m["key"].(string),
	// 			MatchPrefix:       m["host_name"].(string),
	// 			Port:           m["port"].(int),
	// 			Cluster:        m["cluster"].(string),
	// 			PrefixRewrite:  m["PrefixRewrite"].(string),
	// 			CreatedAt:      m["created_at"].(int64),
	// 			CreatedBy:      m["created_by"].(string),
	// 			DisplayName:    m["display_name"].(string),
	// 			LastmodifiedAt: m["lastmodified_at"].(int64),
	// 			LastmodifiedBy: m["lastmodified_by"].(string),
	// 		}
	// 		if m["attributes"] != nil {
	// 			newRoute.Attributes = d.unmarshallJSONArrayOfAttributes(m["attributes"].(string))
	// 		}
	// 		routes = append(routes, newRoute)
	// 		m = map[string]interface{}{}
	// 	}
	// 	// In case query failed we return query error
	// 	if err := iter.Close(); err != nil {
	// 		log.Error(err)
	// 		return []types.Route{}, err
	// 	}
	// 	return routes, nil
	return routes, nil
}

// UpdateRouteByName UPSERTs an route in database
func (d *Database) UpdateRouteByName(updatedRoute *types.Route) error {
	// query := "INSERT INTO routez (key, display_name, " +
	// 	"host_name, port, attributes, " +
	// 	"created_at, created_by, lastmodified_at, lastmodified_by) " +
	// 	"VALUES(?,?,?,?,?,?,?,?,?)"
	// updatedRoute.Attributes = types.TidyAttributes(updatedRoute.Attributes)
	// attributes := d.marshallArrayOfAttributesToJSON(updatedRoute.Attributes)
	// updatedRoute.LastmodifiedAt = types.GetCurrentTimeMilliseconds()
	// if err := d.cassandraSession.Query(query,
	// 	updatedRoute.Name, updatedRoute.DisplayName,
	// 	updatedRoute.HostName, updatedRoute.Port, attributes,
	// 	updatedRoute.CreatedAt, updatedRoute.CreatedBy,
	// 	updatedRoute.LastmodifiedAt,
	// 	updatedRoute.LastmodifiedBy).Exec(); err != nil {
	// 	return fmt.Errorf("Can not update route (%v)", err)
	// }
	return nil
}

// DeleteRouteByName deletes a route
func (d *Database) DeleteRouteByName(routeToDelete string) error {
	// _, err := d.GetRouteByName(routeToDelete)
	// if err != nil {
	// 	return err
	// }
	// query := "DELETE FROM routez WHERE key = ?"
	// return d.cassandraSession.Query(query, routeToDelete).Exec()
	return nil
}
