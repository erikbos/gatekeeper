package main

import (
	"sync"
	"time"

	"github.com/erikbos/apiauth/pkg/shared"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	log "github.com/sirupsen/logrus"
)

var routes []shared.Route

var routeLastUpdate int64
var routeMutex sync.Mutex

// // FIXME this should be implemented using channels
// // FIXME this does not detect removed records

// GetRouteConfigFromDatabase continously gets the current configuration
func (s *server) GetRouteConfigFromDatabase() {
	log.Printf("bla")
	for {
		newRouteList, err := s.db.GetRoutes()
		if err != nil {
			log.Errorf("Could not retrieve routes from database (%s)", err)
		} else {
			for _, s := range newRouteList {
				// Is a cluster updated since last time we stored it?
				if s.LastmodifiedAt > clustersLastUpdate || 1 == 1 {
					now := shared.GetCurrentTimeMilliseconds()
					routeMutex.Lock()
					routes = newRouteList
					routeLastUpdate = now
					routeMutex.Unlock()
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
}

// getEnvoyRouteConfig returns array of all envoyclusters
func getEnvoyRouteConfig() ([]cache.Resource, error) {
	envoyRoutes := []cache.Resource{}

	envoyRoutes = append(envoyRoutes, buildEnvoyRouteConfig())

	return envoyRoutes, nil
}

// buildEnvoyRouteConfig buils one envoy route configuration
func buildEnvoyRouteConfig() *api.RouteConfiguration {
	envoyRoutes := make([]*route.Route, len(routes))

	for index, value := range routes {
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
		if value.DirectResponseStatus != 0 {
			envoyRoute.Action = &route.Route_DirectResponse{
				DirectResponse: &route.DirectResponseAction{
					Status: uint32(value.DirectResponseStatus),
				},
			}
		}
		envoyRoutes[index] = envoyRoute
	}
	envoyRouteConfig := &api.RouteConfiguration{
		Name: "generic_route",
		VirtualHosts: []*route.VirtualHost{
			{
				Name:    "virtual_host",
				Domains: []string{"*"},
				Routes:  envoyRoutes,
			},
		},
	}

	log.Infof("buildEnvoyRouteConfig: %+v", envoyRouteConfig)
	return envoyRouteConfig
}
