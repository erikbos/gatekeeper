package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/erikbos/apiauth/pkg/shared"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	log "github.com/sirupsen/logrus"
)

var routes []shared.Route

// var routeLastUpdate int64
var routeMutex sync.Mutex

// // FIXME this should be implemented using channels
// // FIXME this does not detect removed records

// GetRouteConfigFromDatabase continously gets the current configuration
func (s *server) GetRouteConfigFromDatabase() {
	for {
		newRouteList, err := s.db.GetRoutes()
		if err != nil {
			log.Errorf("Could not retrieve routes from database (%s)", err)
		} else {
			for _, s := range newRouteList {
				// Is a cluster updated since last time we stored it?
				if s.LastmodifiedAt > clustersLastUpdate || 1 == 1 {
					// now := shared.GetCurrentTimeMilliseconds()
					routeMutex.Lock()
					routes = newRouteList
					// routeLastUpdate = now
					routeMutex.Unlock()
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
}

// getEnvoyRouteConfig returns array of all envoy routes
func getEnvoyRouteConfig() ([]cache.Resource, error) {
	var envoyRoutes []cache.Resource

	RouteSetNames := getRouteSetNames(routes)
	for routeSetName := range RouteSetNames {
		log.Infof("adding routeset %s", routeSetName)
		envoyRoutes = append(envoyRoutes, buildEnvoyRouteConfig(routeSetName, routes))
	}

	return envoyRoutes, nil
}

// getVirtualHostPorts returns set of unique routeset names
func getRouteSetNames(vhosts []shared.Route) map[string]bool {
	routeSetNames := map[string]bool{}
	for _, routeEntry := range routes {
		routeSetNames[routeEntry.RouteSet] = true
	}
	return routeSetNames
}

// buildEnvoyRouteConfig builds route configuration for routes of one group of vhosts
func buildEnvoyRouteConfig(routeSet string, routes []shared.Route) *api.RouteConfiguration {
	var envoyRoutes []*route.Route
	for _, value := range routes {
		if value.RouteSet == routeSet {
			envoyRoute := &route.Route{
				Name: value.Name,
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: value.MatchPrefix,
					},
				},
			}
			if value.Cluster != "" {
				envoyRoute.Action = &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: value.Cluster,
						},
						PrefixRewrite: "/",
						HostRewriteSpecifier: &route.RouteAction_HostRewrite{
							HostRewrite: value.HostRewrite,
						},
					},
				}
			}
			directResponseStatusCode, err := shared.GetAttribute(value.Attributes, "DirectResponseStatusCode")
			if err == nil && directResponseStatusCode != "" {
				statusCode, err := strconv.Atoi(directResponseStatusCode)

				if err == nil && statusCode != 0 {
					response := route.Route_DirectResponse{
						DirectResponse: &route.DirectResponseAction{
							Status: uint32(statusCode),
						},
					}
					directResponseStatusBody, err := shared.GetAttribute(value.Attributes, "DirectResponseBody")
					if err == nil && directResponseStatusCode != "" {
						response.DirectResponse.Body = &core.DataSource{
							Specifier: &core.DataSource_InlineString{
								InlineString: directResponseStatusBody,
							},
						}
					}
					envoyRoute.Action = &response
				}
			}
			envoyRoutes = append(envoyRoutes, envoyRoute)
		}
	}

	vhosts := getVirtualHostsOfRouteSet(routeSet)
	envoyRouteConfig := &api.RouteConfiguration{
		Name: routeSet,
		VirtualHosts: []*route.VirtualHost{
			{
				Name:    routeSet,
				Domains: vhosts,
				Routes:  envoyRoutes,
			},
		},
	}
	log.Infof("buildEnvoyRouteConfig: %+v", envoyRouteConfig)
	return envoyRouteConfig
}
