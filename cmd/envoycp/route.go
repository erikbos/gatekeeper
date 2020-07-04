package main

import (
	"encoding/base64"
	"strconv"
	"strings"
	"sync"
	"time"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoyauth "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/ext_authz/v2"
	envoymatcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/conversion"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

const (
	routeDataRefreshInterval = 2 * time.Second
)

// FIXME this does not detect removed records
// GetRouteConfigFromDatabase continuously gets the current configuration
func (s *server) GetRouteConfigFromDatabase(n chan xdsNotifyMesssage) {
	var routesLastUpdate int64
	var routeMutex sync.Mutex

	for {
		var xdsPushNeeded bool

		newRouteList, err := s.db.GetRoutes()
		if err != nil {
			log.Errorf("Could not retrieve routes from database (%s)", err)
		} else {
			if routesLastUpdate == 0 {
				log.Info("Initial load of routes")
			}
			for _, route := range newRouteList {
				// Is a cluster updated since last time we stored it?
				if route.LastmodifiedAt > routesLastUpdate {
					routeMutex.Lock()
					s.routes = newRouteList
					routeMutex.Unlock()

					routesLastUpdate = shared.GetCurrentTimeMilliseconds()
					xdsPushNeeded = true

					warnForUnknownRouteAttributes(route)
				}
			}
		}
		if xdsPushNeeded {
			n <- xdsNotifyMesssage{
				resource: "route",
			}
			s.metrics.xdsDeployments.WithLabelValues("routes").Inc()
		}
		time.Sleep(routeDataRefreshInterval)
	}
}

// GetRouteCount returns number of routes
func (s *server) GetRouteCount() float64 {
	return float64(len(s.routes))
}

// getEnvoyRouteConfig returns array of all envoy routes
func (s *server) getEnvoyRouteConfig() ([]cache.Resource, error) {
	var envoyRoutes []cache.Resource

	RouteGroupNames := s.getRouteGroupNames(s.routes)
	for RouteGroupName := range RouteGroupNames {
		log.Infof("Adding routegroup '%s'", RouteGroupName)
		envoyRoutes = append(envoyRoutes, s.buildEnvoyVirtualHostRouteConfig(RouteGroupName, s.routes))
	}

	return envoyRoutes, nil
}

// getVirtualHostPorts returns set of unique RouteGroup names
func (s *server) getRouteGroupNames(vhosts []shared.Route) map[string]bool {
	RouteGroupNames := map[string]bool{}
	for _, routeEntry := range s.routes {
		RouteGroupNames[routeEntry.RouteGroup] = true
	}
	return RouteGroupNames
}

// buildEnvoyVirtualHostRouteConfig builds vhost and route configuration of one RouteGroup
func (s *server) buildEnvoyVirtualHostRouteConfig(RouteGroup string,
	routes []shared.Route) *api.RouteConfiguration {

	return &api.RouteConfiguration{
		Name: RouteGroup,
		VirtualHosts: []*route.VirtualHost{
			{
				Name:    RouteGroup,
				Domains: s.getVirtualHostsOfRouteGroup(RouteGroup),
				Routes:  s.buildEnvoyRoutes(RouteGroup, routes),
			},
		},
	}
}

// buildEnvoyRoute returns all Envoy routes belong to one RouteGroup
func (s *server) buildEnvoyRoutes(RouteGroup string, routes []shared.Route) []*route.Route {
	var envoyRoutes []*route.Route

	for _, route := range routes {
		if route.RouteGroup == RouteGroup {
			envoyRoutes = append(envoyRoutes, s.buildEnvoyRoute(route))
		}
	}
	return envoyRoutes
}

// buildEnvoyRoute returns a single Envoy route
func (s *server) buildEnvoyRoute(routeEntry shared.Route) *route.Route {
	routeMatch := buildRouteMatch(routeEntry)
	if routeMatch == nil {
		log.Warnf("Cannot build route config for route '%s'", routeEntry.Name)
		return nil
	}

	envoyRoute := &route.Route{
		Name:  routeEntry.Name,
		Match: routeMatch,
	}

	if routeEntry.Cluster != "" {
		envoyRoute.Action = buildRouteActionCluster(routeEntry)
	}

	// Add direct response config if required
	_, err := shared.GetAttribute(routeEntry.Attributes, attributeDisableAuthentication)
	if err == nil {
		envoyRoute.PerFilterConfig = buildRoutePerFilterConfig(routeEntry)
		return envoyRoute
	}

	// Add direct response config if required
	_, err = shared.GetAttribute(routeEntry.Attributes, attributeDirectResponseStatusCode)
	if err == nil {
		envoyRoute.Action = buildRouteActionDirectResponse(routeEntry)
		return envoyRoute
	}

	// In case route-level attributes exist we might additional upstream headers
	upstreamHeaders := make(map[string]string)
	handleBasicAuthAttribute(routeEntry, upstreamHeaders)
	envoyRoute.RequestHeadersToAdd = buildHeadersList(upstreamHeaders)

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
			Cors:                 buildCorsPolicy(routeEntry),
			HostRewriteSpecifier: buildHostRewrite(routeEntry),
			RetryPolicy:          buildRetryPolicy(routeEntry),
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

// buildHostRewrite return CorsPolicy based upon a route's attribute(s)
func buildHostRewrite(routeEntry shared.Route) *route.RouteAction_HostRewrite {

	upstreamHostHeader, err := shared.GetAttribute(routeEntry.Attributes, attributeHostHeader)
	if err == nil && upstreamHostHeader != "" {
		return &route.RouteAction_HostRewrite{
			HostRewrite: upstreamHostHeader,
		}
	}

	return nil
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

func handleBasicAuthAttribute(routeEntry shared.Route, headersToAdd map[string]string) {

	usernamePassword, err := shared.GetAttribute(routeEntry.Attributes, attributeBasicAuth)
	if err == nil && usernamePassword != "" {
		authenticationDigest := base64.StdEncoding.EncodeToString([]byte(usernamePassword))

		headersToAdd["Authorization"] = authenticationDigest
	}
}

// buildHeadersList creates map to hold additional upstream headers
func buildHeadersList(headers map[string]string) []*core.HeaderValueOption {
	if len(headers) == 0 {
		return nil
	}

	headerList := make([]*core.HeaderValueOption, 0, len(headers))
	for key, value := range headers {
		headerList = append(headerList, &core.HeaderValueOption{
			Header: &core.HeaderValue{
				Key:   key,
				Value: value,
			},
		})
	}
	return headerList
}

func buildRetryPolicy(routeEntry shared.Route) *route.RetryPolicy {

	RetryOn := shared.GetAttributeAsString(routeEntry.Attributes, attributeRetryOn, "")
	if RetryOn == "" {
		return nil
	}
	perTryTimeout := shared.GetAttributeAsDuration(routeEntry.Attributes, attributePerTryTimeout, perRetryTimeout)
	numRetries := uint32(shared.GetAttributeAsInt(routeEntry.Attributes, attributeNumRetries, 2))
	RetriableStatusCodes := buildStatusCodesSlice(
		shared.GetAttributeAsString(routeEntry.Attributes, attributeRetryOnStatusCodes, "503"))

	return &route.RetryPolicy{
		RetryOn:              RetryOn,
		NumRetries:           &wrappers.UInt32Value{Value: numRetries},
		PerTryTimeout:        ptypes.DurationProto(perTryTimeout),
		RetriableStatusCodes: RetriableStatusCodes,
		RetryHostPredicate: []*route.RetryPolicy_RetryHostPredicate{
			{Name: "envoy.retry_host_predicates.previous_hosts"},
		},
		// HostSelectionRetryMaxAttempts: 5,

	}
}

func buildStatusCodesSlice(statusCodes string) []uint32 {

	var statusCodeSlice []uint32

	for _, statusCode := range strings.Split(statusCodes, ",") {
		// we only add successfully parse integers
		if value, err := strconv.Atoi(statusCode); err == nil {
			statusCodeSlice = append(statusCodeSlice, uint32(value))
		}
	}
	return statusCodeSlice
}

func buildRoutePerFilterConfig(routeEntry shared.Route) map[string]*structpb.Struct {

	perFilterConfig := make(map[string]*structpb.Struct)

	value, err := shared.GetAttribute(routeEntry.Attributes, attributeDisableAuthentication)
	if err == nil && value == attributeValueTrue {

		extAuthzPerRoute, err := conversion.MessageToStruct(
			&envoyauth.ExtAuthzPerRoute{
				Override: &envoyauth.ExtAuthzPerRoute_Disabled{
					Disabled: true,
				},
			})
		if err == nil {
			perFilterConfig["envoy.filters.http.ext_authz"] = extAuthzPerRoute
		}
	}

	return perFilterConfig
}
