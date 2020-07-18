package main

import (
	"encoding/base64"
	"strconv"
	"strings"
	"sync"
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoyauth "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	envoymatcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	envoytype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	log "github.com/sirupsen/logrus"

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
	routes []shared.Route) *route.RouteConfiguration {

	return &route.RouteConfiguration{
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
		log.Warnf("Cannot build route match config for route '%s'", routeEntry.Name)
		return nil
	}

	envoyRoute := &route.Route{
		Name:  routeEntry.Name,
		Match: routeMatch,
	}

	// disable extauth if requested.
	// extauth is first configed filter, so this needs to be done before anything else
	_, err := shared.GetAttribute(routeEntry.Attributes, attributeDisableAuthentication)
	if err == nil {
		envoyRoute.TypedPerFilterConfig = buildRoutePerFilterConfig(routeEntry)
	}

	// Add direct response if configured: in this case Envoy itself will answer
	_, err = shared.GetAttribute(routeEntry.Attributes, attributeDirectResponseStatusCode)
	if err == nil {
		envoyRoute.Action = buildRouteActionDirectResponse(routeEntry)
		return envoyRoute
	}

	// Add redirect response if configured: in this case Envoy will generate HTTP redirect
	_, err = shared.GetAttribute(routeEntry.Attributes, attributeRedirectStatusCode)
	if err == nil {
		envoyRoute.Action = buildRouteActionRedirectResponse(routeEntry)
		return envoyRoute
	}

	// Add cluster(s) to forward to
	if routeEntry.Cluster != "" {
		envoyRoute.Action = buildRouteActionCluster(routeEntry)
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
		return &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Path{
				Path: routeEntry.Path,
			},
		}

	case "prefix":
		return &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: routeEntry.Path,
			},
		}

	case "regexp":
		return &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_SafeRegex{
				SafeRegex: buildRegexpMatcher(routeEntry.Path),
			},
		}
	}
	return nil
}

// buildRouteActionCluster return route action in of forwarding to a cluster
func buildRouteActionCluster(routeEntry shared.Route) *route.Route_Route {

	action := &route.Route_Route{
		Route: &route.RouteAction{
			Cors:                 buildCorsPolicy(routeEntry),
			HostRewriteSpecifier: buildHostRewrite(routeEntry),
			RetryPolicy:          buildRetryPolicy(routeEntry),
		},
	}

	prefixRewrite, err := shared.GetAttribute(routeEntry.Attributes, attributePrefixRewrite)
	if err == nil && prefixRewrite != "" {
		action.Route.PrefixRewrite = prefixRewrite
	}

	if routeEntry.Cluster != "" {
		if strings.Contains(routeEntry.Cluster, ",") {
			action.Route.ClusterSpecifier = buildWeightedClusters(routeEntry)
		} else {
			action.Route.ClusterSpecifier = &route.RouteAction_Cluster{
				Cluster: routeEntry.Cluster,
			}
		}
	}

	action.Route.RequestMirrorPolicies = buildRequestMirrorPolicy(routeEntry)

	return action
}

func buildWeightedClusters(routeEntry shared.Route) *route.RouteAction_WeightedClusters {

	weightedClusters := make([]*route.WeightedCluster_ClusterWeight, 0)
	var clusterWeight, totalWeight int

	for _, cluster := range strings.Split(routeEntry.Cluster, ",") {

		clusterConfig := strings.Split(cluster, ":")

		if len(clusterConfig) == 2 {
			clusterWeight, _ = strconv.Atoi(clusterConfig[1])
		} else {
			log.Warningf("Route '%s' cluster '%s' subcluster '%s' does not have weight value",
				routeEntry.Name, routeEntry.Cluster, cluster)

			clusterWeight = 1
		}
		totalWeight += clusterWeight

		clusterToAdd := &route.WeightedCluster_ClusterWeight{
			Name:   clusterConfig[0],
			Weight: protoUint32(uint32(clusterWeight)),
		}
		weightedClusters = append(weightedClusters, clusterToAdd)
	}

	return &route.RouteAction_WeightedClusters{
		WeightedClusters: &route.WeightedCluster{
			Clusters:    weightedClusters,
			TotalWeight: protoUint32(uint32(totalWeight)),
		},
	}
}

func buildRequestMirrorPolicy(routeEntry shared.Route) []*route.RouteAction_RequestMirrorPolicy {

	mirrorCluster := shared.GetAttributeAsString(routeEntry.Attributes, attributeRequestMirrorClusterName, "")
	mirrorPercentage := shared.GetAttributeAsString(routeEntry.Attributes, attributeRequestMirrorPercentage, "")

	if mirrorCluster == "" || mirrorPercentage == "" {
		return nil
	}

	percentage, _ := strconv.Atoi(mirrorPercentage)
	if percentage < 0 || percentage > 100 {
		log.Warningf("Route '%s' cluster '%s' incorrect request mirror percentage '%s'",
			routeEntry.Name, routeEntry.Cluster, mirrorPercentage)
		return nil
	}

	return []*route.RouteAction_RequestMirrorPolicy{{
		Cluster: mirrorCluster,
		RuntimeFraction: &core.RuntimeFractionalPercent{
			DefaultValue: &envoytype.FractionalPercent{
				Numerator:   uint32(percentage),
				Denominator: envoytype.FractionalPercent_HUNDRED,
			},
		},
	}}
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
		corsPolicy.AllowCredentials = protoBool(true)
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
			GoogleRe2: &envoymatcher.RegexMatcher_GoogleRE2{},
		},
		Regex: regexp,
	}
}

// buildHostRewrite returns HostRewrite config based upon route attribute(s)
func buildHostRewrite(routeEntry shared.Route) *route.RouteAction_HostRewriteLiteral {

	upstreamHostHeader, err := shared.GetAttribute(routeEntry.Attributes, attributeHostHeader)
	if err == nil && upstreamHostHeader != "" {
		return &route.RouteAction_HostRewriteLiteral{
			HostRewriteLiteral: upstreamHostHeader,
		}
	}

	return nil
}

// buildRouteActionDirectResponse builds response to have Envoy itself answer with status code and body
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

// buildRouteActionRedirectResponse builds response to have Envoy redirect to different path
func buildRouteActionRedirectResponse(routeEntry shared.Route) *route.Route_Redirect {

	redirectStatusCode := shared.GetAttributeAsString(routeEntry.Attributes, attributeRedirectStatusCode, "")
	redirectScheme := shared.GetAttributeAsString(routeEntry.Attributes, attributeRedirectScheme, "")
	RedirectHostName := shared.GetAttributeAsString(routeEntry.Attributes, attributeRedirectHostName, "")
	redirectPort := shared.GetAttributeAsString(routeEntry.Attributes, attributeRedirectPort, "")
	redirectPath := shared.GetAttributeAsString(routeEntry.Attributes, attributeRedirectPath, "")
	redirectStripQuery := shared.GetAttributeAsString(routeEntry.Attributes, attributeRedirectStripQuery, "")

	// Status code must be set other wise we not build redirect configuration
	if redirectStatusCode == "" {
		return nil
	}

	response := route.Route_Redirect{
		Redirect: &route.RedirectAction{},
	}

	// Do we have to set a particular statuscode?
	statusCode, _ := strconv.Atoi(redirectStatusCode)
	switch statusCode {
	case 301:
		response.Redirect.ResponseCode = route.RedirectAction_MOVED_PERMANENTLY
	case 302:
		response.Redirect.ResponseCode = route.RedirectAction_FOUND
	case 303:
		response.Redirect.ResponseCode = route.RedirectAction_SEE_OTHER
	case 307:
		response.Redirect.ResponseCode = route.RedirectAction_TEMPORARY_REDIRECT
	case 308:
		response.Redirect.ResponseCode = route.RedirectAction_PERMANENT_REDIRECT
	default:
		log.Warningf("Route '%s' attribute '%s' has unsupported status '%s'",
			routeEntry.Name, attributeRedirectStatusCode, redirectStatusCode)
		return nil
	}
	// Do we have to update scheme to http or https?
	if redirectScheme != "" {
		response.Redirect.SchemeRewriteSpecifier = &route.RedirectAction_SchemeRedirect{
			SchemeRedirect: redirectScheme,
		}
	}
	// Do we have to redirect to a specific hostname?
	if RedirectHostName != "" {
		response.Redirect.HostRedirect = RedirectHostName
	}
	// Do we have to redirect to a specific port?
	if redirectPort != "" {
		port, _ := strconv.Atoi(redirectPort)
		response.Redirect.PortRedirect = uint32(port)
	}
	// Do we have to redirect to a specific path?
	if redirectPath != "" {
		response.Redirect.PathRewriteSpecifier = &route.RedirectAction_PathRedirect{
			PathRedirect: redirectPath,
		}
	}
	// Do we have to enable stripping of query string parameters?
	if redirectStripQuery == attributeValueTrue {
		response.Redirect.StripQuery = true
	}

	return &response
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
		NumRetries:           protoUint32(numRetries),
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

func buildRoutePerFilterConfig(routeEntry shared.Route) map[string]*any.Any {

	perFilterConfigMap := make(map[string]*any.Any)

	value, err := shared.GetAttribute(routeEntry.Attributes, attributeDisableAuthentication)
	if err == nil && value == attributeValueTrue {

		perFilterExtAuthzConfig := envoyauth.ExtAuthzPerRoute{
			Override: &envoyauth.ExtAuthzPerRoute_Disabled{
				Disabled: true,
			},
		}

		b := proto.NewBuffer(nil)
		b.SetDeterministic(true)
		_ = b.Marshal(&perFilterExtAuthzConfig)

		filter := &any.Any{
			TypeUrl: "type.googleapis.com/" + proto.MessageName(&perFilterExtAuthzConfig),
			Value:   b.Bytes(),
		}

		perFilterConfigMap[wellknown.HTTPExternalAuthorization] = filter
	}

	return perFilterConfigMap
}
