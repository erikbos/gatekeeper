package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/erikbos/apiauth/pkg/shared"
	"github.com/golang/protobuf/ptypes/wrappers"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoymatcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	log "github.com/sirupsen/logrus"
)

const (
	routeRefreshInterval = 2

	attributeDirectResponseStatusCode = "DirectResponseStatusCode"
	attributeDirectResponseBody       = "DirectResponseBody"
	attributePrefixRewrite            = "PrefixRewrite"
	attributeCORSAllowMethods         = "CORSAllowMethods"
	attributeCORSAllowHeaders         = "CORSAllowHeaders"
	attributeCORSExposeHeaders        = "CORSExposeHeaders"
	attributeCORSMaxAge               = "CORSMaxAge"
	attributeCORSAllowCredentials     = "CORSAllowCredentials"
)

var routes []shared.Route

// FIXME this does not detect removed records
// GetRouteConfigFromDatabase continously gets the current configuration
func (s *server) GetRouteConfigFromDatabase() {
	var routesLastUpdate int64
	var routeMutex sync.Mutex

	for {
		var xdsPushNeeded bool

		newRouteList, err := s.db.GetRoutes()
		if err != nil {
			log.Errorf("Could not retrieve routes from database (%s)", err)
		} else {
			for _, s := range newRouteList {
				// Is a cluster updated since last time we stored it?
				if s.LastmodifiedAt > routesLastUpdate {
					routeMutex.Lock()
					routes = newRouteList
					routeMutex.Unlock()

					routesLastUpdate = shared.GetCurrentTimeMilliseconds()
					xdsPushNeeded = true
				}
			}
		}
		if xdsPushNeeded {
			// FIXME this should be notification via channel
			xdsLastUpdate = shared.GetCurrentTimeMilliseconds()
			// Increase xds deployment metric
			s.metricXdsDeployments.WithLabelValues("routes").Inc()
		}
		time.Sleep(routeRefreshInterval * time.Second)
	}
}

// GetRouteCount returns number of routes
func (s *server) GetRouteCount() float64 {
	return float64(len(routes))
}

// getEnvoyRouteConfig returns array of all envoy routes
func getEnvoyRouteConfig() ([]cache.Resource, error) {
	var envoyRoutes []cache.Resource

	RouteSetNames := getRouteSetNames(routes)
	for routeSetName := range RouteSetNames {
		log.Infof("Adding routeset %s", routeSetName)
		envoyRoutes = append(envoyRoutes, buildEnvoyVirtualHostRouteConfig(routeSetName, routes))
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

// buildEnvoyVirtualHostRouteConfig builds vhost and route configuration of one routeset
func buildEnvoyVirtualHostRouteConfig(routeSet string,
	routes []shared.Route) *api.RouteConfiguration {

	return &api.RouteConfiguration{
		Name: routeSet,
		VirtualHosts: []*route.VirtualHost{
			{
				Name:    routeSet,
				Domains: getVirtualHostsOfRouteSet(routeSet),
				Routes:  buildEnvoyRoutes(routeSet, routes),
			},
		},
	}
}

// buildEnvoyRoute returns all Envoy routes belong to one routeset
func buildEnvoyRoutes(routeSet string, routes []shared.Route) []*route.Route {
	var envoyRoutes []*route.Route

	for _, route := range routes {
		if route.RouteSet == routeSet {
			envoyRoutes = append(envoyRoutes, buildEnvoyRoute(route))
		}
	}
	return envoyRoutes
}

// buildEnvoyRoute returns a single Envoy route
func buildEnvoyRoute(routeEntry shared.Route) *route.Route {
	routeMatch := buildRouteMatch(routeEntry)
	if routeMatch == nil {
		log.Warnf("Cannot build route config for route %s", routeEntry.Name)
		return nil
	}

	envoyRoute := &route.Route{
		Name:  routeEntry.Name,
		Match: routeMatch,
	}

	if routeEntry.Cluster != "" {
		envoyRoute.Action = buildRouteActionCluster(routeEntry)
	}

	// In case attribute exist we build direct response config
	_, err := shared.GetAttribute(routeEntry.Attributes, attributeDirectResponseStatusCode)
	if err == nil {
		envoyRoute.Action = buildRouteActionDirectResponse(routeEntry)
	}
	return envoyRoute
}

// buildRouteActionCluster returns route config we match on
func buildRouteMatch(routeEntry shared.Route) *route.RouteMatch {
	switch routeEntry.PathType {
	case "path":
		return buildRouteMatchPath(routeEntry)
	case "prefix":
		return buildRouteMatchPrefix(routeEntry)
	case "regexp":
		return buildRouteMatchRegexp(routeEntry)
	}
	return nil
}

// buildRouteMatchPath returns prefix-based route-match config
func buildRouteMatchPath(routeEntry shared.Route) *route.RouteMatch {
	return &route.RouteMatch{
		PathSpecifier: &route.RouteMatch_Path{
			Path: routeEntry.Path,
		},
	}
}

// buildRouteMatchPrefix returns path-based route-match config
func buildRouteMatchPrefix(routeEntry shared.Route) *route.RouteMatch {
	return &route.RouteMatch{
		PathSpecifier: &route.RouteMatch_Prefix{
			Prefix: routeEntry.Path,
		},
	}
}

// buildRouteMatchPath returns prefix-based route-match config
func buildRouteMatchRegexp(routeEntry shared.Route) *route.RouteMatch {
	return &route.RouteMatch{
		PathSpecifier: &route.RouteMatch_SafeRegex{
			SafeRegex: buildRegexpMatcher(routeEntry.Path),
		},
	}
}

// buildRouteActionCluster return route action in of forwarding to a cluster
func buildRouteActionCluster(routeEntry shared.Route) *route.Route_Route {
	action := &route.Route_Route{
		Route: &route.RouteAction{
			ClusterSpecifier: &route.RouteAction_Cluster{
				Cluster: routeEntry.Cluster,
			},
			Cors: buildCorsPolicy(routeEntry),
		},
	}
	prefixRewrite, err := shared.GetAttribute(routeEntry.Attributes, attributePrefixRewrite)
	if err == nil && prefixRewrite != "" {
		action.Route.PrefixRewrite = prefixRewrite
	}
	return action
}

// buildCorsPolicy return CorsPolicy based upon a route's attribute(s)
func buildCorsPolicy(routeEntry shared.Route) *route.CorsPolicy {
	var corsConfigured bool
	corsPolicy := route.CorsPolicy{}

	corsAllowMethods, err := shared.GetAttribute(routeEntry.Attributes, attributeCORSAllowMethods)
	if err == nil && corsAllowMethods != "" {
		corsPolicy.AllowMethods = corsAllowMethods
		corsConfigured = true
	}

	corsAllowHeaders, err := shared.GetAttribute(routeEntry.Attributes, attributeCORSAllowHeaders)
	if err == nil && corsAllowMethods != "" {
		corsPolicy.AllowHeaders = corsAllowHeaders
		corsConfigured = true
	}

	corsExposeHeaders, err := shared.GetAttribute(routeEntry.Attributes, attributeCORSExposeHeaders)
	if err == nil && corsAllowMethods != "" {
		corsPolicy.ExposeHeaders = corsExposeHeaders
		corsConfigured = true
	}

	corsMaxAge, err := shared.GetAttribute(routeEntry.Attributes, attributeCORSMaxAge)
	if err == nil && corsAllowMethods != "" {
		corsPolicy.MaxAge = corsMaxAge
		corsConfigured = true
	}

	corsAllowCredentials, err := shared.GetAttribute(routeEntry.Attributes, attributeCORSAllowCredentials)
	if err == nil && corsAllowCredentials == attributeValueTrue {
		corsPolicy.AllowCredentials = &wrappers.BoolValue{Value: true}
		corsConfigured = true
	}

	if corsConfigured {
		stringMatch := make([]*envoymatcher.StringMatcher, 0, 1)
		corsPolicy.AllowOriginStringMatch = append(stringMatch, buildStringMatcher("."))
		return &corsPolicy
	}
	return nil
}

func buildStringMatcher(regexp string) *envoymatcher.StringMatcher {
	return &envoymatcher.StringMatcher{
		MatchPattern: &envoymatcher.StringMatcher_SafeRegex{
			SafeRegex: buildRegexpMatcher(regexp),
		},
	}
}

func buildRegexpMatcher(regexp string) *envoymatcher.RegexMatcher {
	return &envoymatcher.RegexMatcher{
		EngineType: &envoymatcher.RegexMatcher_GoogleRe2{
			GoogleRe2: &envoymatcher.RegexMatcher_GoogleRE2{
				MaxProgramSize: &wrappers.UInt32Value{
					Value: 100,
				},
			},
		},
		Regex: regexp,
	}
}

func buildRouteActionDirectResponse(routeEntry shared.Route) *route.Route_DirectResponse {
	directResponseStatusCode, err := shared.GetAttribute(routeEntry.Attributes, attributeDirectResponseStatusCode)
	if err == nil && directResponseStatusCode != "" {
		statusCode, err := strconv.Atoi(directResponseStatusCode)

		if err == nil && statusCode != 0 {
			response := route.Route_DirectResponse{
				DirectResponse: &route.DirectResponseAction{
					Status: uint32(statusCode),
				},
			}
			directResponseStatusBody, err := shared.GetAttribute(routeEntry.Attributes, attributeDirectResponseBody)
			if err == nil && directResponseStatusCode != "" {
				response.DirectResponse.Body = &core.DataSource{
					Specifier: &core.DataSource_InlineString{
						InlineString: directResponseStatusBody,
					},
				}
			}
			return &response
		}
	}
	return nil
}
