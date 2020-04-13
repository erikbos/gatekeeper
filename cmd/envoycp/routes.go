package main

import (
	"github.com/erikbos/apiauth/pkg/types"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	log "github.com/sirupsen/logrus"
)

var routes []types.Cluster

// var clustersLastUpdate int64
// var mux sync.Mutex

// // FIXME this should be implemented using channels
// // FIXME this does not detect removed records

// // getClusterConfigFromDatabase continously gets the current configuration
// func (s *server) GetClusterConfigFromDatabase() {
// 	for {
// 		newClusterList, err := s.db.GetClusters()
// 		if err != nil {
// 			log.Errorf("Could not retrieve clusters from database (%s)", err)
// 		} else {
// 			for _, s := range newClusterList {
// 				// Is a cluster updated since last time we stored it?
// 				if s.LastmodifiedAt > clustersLastUpdate {
// 					now := types.GetCurrentTimeMilliseconds()
// 					mux.Lock()
// 					clusters = newClusterList
// 					clustersLastUpdate = now
// 					mux.Unlock()
// 				}
// 			}
// 		}
// 		time.Sleep(2 * time.Second)
// 	}
// }

// getClusterConfig returns array of all envoyclusters
func getEnvoyRouteConfig() ([]cache.Resource, error) {
	envoyRoutes := []cache.Resource{}
	// for _, s := range clusters {
	envoyRoutes = append(envoyRoutes, buildEnvoyRouteConfig())
	// }
	return envoyRoutes, nil
}

// buildEnvoyRouteConfig buils one envoy route configuration
func buildEnvoyRouteConfig() *api.RouteConfiguration {
	routes := []*route.Route{
		{
			Name: "google",
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
					Prefix: "/google/",
				},
			},
			Action: &route.Route_Route{
				Route: &route.RouteAction{
					ClusterSpecifier: &route.RouteAction_Cluster{
						Cluster: "ttapi",
					},
					PrefixRewrite: "/",
					HostRewriteSpecifier: &route.RouteAction_HostRewrite{
						HostRewrite: "www.nu.nl",
					},
				},
			},
		},
		{
			Name: "people",
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
					Prefix: "/people/",
				},
			},
			Action: &route.Route_Route{
				Route: &route.RouteAction{
					ClusterSpecifier: &route.RouteAction_Cluster{
						Cluster: "people",
					},
					// PrefixRewrite: "/",
					// HostRewriteSpecifier: &route.RouteAction_HostRewrite{
					// 	HostRewrite: "www.nu.nl",
					// },
				},
			},
		},
		{
			Name: "health-check",
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Path{
					Path: "/hc",
				},
			},
			Action: &route.Route_DirectResponse{
				DirectResponse: &route.DirectResponseAction{
					Status: 200,
				},
			},
		},
		{
			Name: "default",
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
					Prefix: "/",
				},
			},
			Action: &route.Route_DirectResponse{
				DirectResponse: &route.DirectResponseAction{
					Status: 406,
				},
			},
		},
	}

	route := &api.RouteConfiguration{
		Name: "generic_route",
		VirtualHosts: []*route.VirtualHost{
			{
				Name:    "virtual_host",
				Domains: []string{"*"},
				Routes:  routes,
			},
		},
	}

	log.Debugf("buildEnvoyRouteConfig: %+v", route)
	return route
}

// buildEnvoyRouteConfig2 buils one envoy route configuration
func buildEnvoyRouteConfig2() *api.RouteConfiguration {
	var clusterPrefixMap = map[string]string{
		// "bla1": "bla2",
		"/people*": "ttapi",
	}

	routes := make([]*route.Route, len(clusterPrefixMap))

	index := 0
	for prefix, clusterName := range clusterPrefixMap {
		routes[index] = &route.Route{
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
					Prefix: prefix,
				},
			},
			Action: &route.Route_Route{
				Route: &route.RouteAction{
					PrefixRewrite: "/",
					ClusterSpecifier: &route.RouteAction_Cluster{
						Cluster: clusterName,
					},
				},
			},
		}
		index++
	}

	envoyRoute := &api.RouteConfiguration{
		Name: "generic_route",
		VirtualHosts: []*route.VirtualHost{
			{
				Name:    "virtual_host",
				Domains: []string{"*"},
				Routes:  routes,
			},
		},
	}

	log.Debugf("buildEnvoyRouteConfig: %+v", envoyRoute)
	return envoyRoute
}
