package main

import (
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_filter_authz "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	envoy_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/erikbos/gatekeeper/pkg/types"
	"github.com/golang/protobuf/ptypes/any"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

// getEnvoyRouteConfig returns array of all envoy routes
func (s *server) getEnvoyRouteConfig() ([]cache.Resource, error) {
	var envoyRoutes []cache.Resource

	RouteGroupNames := s.getRouteGroupNames(s.dbentities.GetRoutes())
	for RouteGroupName := range RouteGroupNames {
		s.logger.Info("Compiling configuration", zap.String("routegroup", RouteGroupName))
		envoyRoutes = append(envoyRoutes,
			s.buildEnvoyListenerRouteConfig(RouteGroupName, s.dbentities.GetRoutes()))
	}

	return envoyRoutes, nil
}

// getListenerPorts returns set of unique RouteGroup names
func (s *server) getRouteGroupNames(vhosts types.Routes) map[string]bool {
	RouteGroupNames := map[string]bool{}
	for _, route := range s.dbentities.GetRoutes() {
		RouteGroupNames[route.RouteGroup] = true
	}
	return RouteGroupNames
}

// buildEnvoyListenerRouteConfig builds vhost and route configuration of one RouteGroup
func (s *server) buildEnvoyListenerRouteConfig(RouteGroup string,
	routes types.Routes) *envoy_route.RouteConfiguration {

	return &envoy_route.RouteConfiguration{
		Name: RouteGroup,
		VirtualHosts: []*envoy_route.VirtualHost{
			{
				Name:            RouteGroup,
				Domains:         s.getVhostsInRouteGroup(RouteGroup),
				Routes:          s.buildEnvoyRoutes(RouteGroup, routes),
				VirtualClusters: s.buildEnvoyVirtualClusters(RouteGroup, routes),
			},
		},
	}
}

// buildEnvoyRoutes returns all Envoy routes belonging to a RouteGroup
func (s *server) buildEnvoyRoutes(RouteGroup string, routes types.Routes) []*envoy_route.Route {
	var envoyRoutes []*envoy_route.Route

	for _, route := range routes {
		if err := route.ConfigCheck(); err != nil {
			s.logger.Warn("Unsupported configuration", zap.String("route", route.Name), zap.Error(err))
		}

		if route.RouteGroup == RouteGroup {
			if routeToAdd := s.buildEnvoyRoute(route); routeToAdd != nil {
				envoyRoutes = append(envoyRoutes, routeToAdd)
			}
		}
	}
	return envoyRoutes
}

// buildEnvoyRoute returns a single Envoy route
func (s *server) buildEnvoyRoute(route types.Route) *envoy_route.Route {
	routeMatch := buildRouteMatch(route)
	if routeMatch == nil {
		s.logger.Warn("Cannot build route match config", zap.String("route", route.Name))
		return nil
	}

	envoyRoute := &envoy_route.Route{
		Name:  route.Name,
		Match: routeMatch,
	}

	// Set all route specific filter options
	envoyRoute.TypedPerFilterConfig = buildPerRouteFilterConfig(route)

	// Add direct response if configured: in this case Envoy itself will answer
	if _, err := route.Attributes.Get(types.AttributeDirectResponseStatusCode); err == nil {
		envoyRoute.Action = buildRouteActionDirectResponse(route)
		return envoyRoute
	}

	// Add redirect response if configured: in this case Envoy will generate HTTP redirect
	if _, err := route.Attributes.Get(types.AttributeRedirectStatusCode); err == nil {
		envoyRoute.Action = s.buildRouteActionRedirectResponse(route)
		return envoyRoute
	}

	// Add cluster(s) to forward this route to
	cluster, _ := route.Attributes.Get(types.AttributeCluster)
	weightedClusters, _ := route.Attributes.Get(types.AttributeWeightedClusters)
	if cluster == "" && weightedClusters == "" {
		// Stop in case we do not have any route cluster destination as
		// this route would be invalid..
		s.logger.Warn("Route without destination cluster", zap.String("route", route.Name))
		return nil
	}

	envoyRoute.Action = s.buildRouteActionCluster(route)
	envoyRoute.RequestHeadersToAdd = buildUpstreamHeadersToAdd(route)
	envoyRoute.RequestHeadersToRemove = buildUpstreamHeadersToRemove(route)

	return envoyRoute
}

// buildRouteMatch returns route config to match on
func buildRouteMatch(route types.Route) *envoy_route.RouteMatch {

	switch route.PathType {
	case types.AttributeValuePathTypePath:
		return &envoy_route.RouteMatch{
			PathSpecifier: &envoy_route.RouteMatch_Path{
				Path: route.Path,
			},
		}

	case types.AttributeValuePathTypePrefix:
		return &envoy_route.RouteMatch{
			PathSpecifier: &envoy_route.RouteMatch_Prefix{
				Prefix: route.Path,
			},
		}

	case types.AttributeValuePathTypeRegexp:
		return &envoy_route.RouteMatch{
			PathSpecifier: &envoy_route.RouteMatch_SafeRegex{
				SafeRegex: buildRegexpMatcher(route.Path),
			},
		}
	}
	return nil
}

// buildRouteActionCluster return route action in of forwarding to a cluster
func (s *server) buildRouteActionCluster(route types.Route) *envoy_route.Route_Route {

	action := &envoy_route.Route_Route{
		Route: &envoy_route.RouteAction{
			Cors:        buildCorsPolicy(route),
			RetryPolicy: buildRetryPolicy(route),
		},
	}

	upstreamTimeout, err := route.Attributes.Get(types.AttributeTimeout)
	if err == nil && upstreamTimeout != "" {
		if upstreamTimeoutDuration, err := time.ParseDuration(upstreamTimeout); err == nil {
			action.Route.Timeout = durationpb.New(upstreamTimeoutDuration)
		}
	}

	upstreamHostHeader, err := route.Attributes.Get(types.AttributeHostHeader)
	if err == nil && upstreamHostHeader != "" {
		action.Route.HostRewriteSpecifier = &envoy_route.RouteAction_HostRewriteLiteral{
			HostRewriteLiteral: upstreamHostHeader,
		}
	}

	prefixRewrite, err := route.Attributes.Get(types.AttributePrefixRewrite)
	if err == nil && prefixRewrite != "" {
		action.Route.PrefixRewrite = prefixRewrite
	}

	cluster, err := route.Attributes.Get(types.AttributeCluster)
	if err == nil && cluster != "" {
		action.Route.ClusterSpecifier = &envoy_route.RouteAction_Cluster{
			Cluster: cluster,
		}
	}

	weightedClusters, err := route.Attributes.Get(types.AttributeWeightedClusters)
	if err == nil && weightedClusters != "" {
		action.Route.ClusterSpecifier = s.buildWeightedClusters(route)
	}

	action.Route.RateLimits = buildRateLimits(route)
	action.Route.RequestMirrorPolicies = s.buildRequestMirrorPolicies(route)

	return action
}

func (s *server) buildWeightedClusters(route types.Route) *envoy_route.RouteAction_WeightedClusters {

	weightedClusterSpec, err := route.Attributes.Get(types.AttributeWeightedClusters)
	if err != nil || weightedClusterSpec == "" {
		return nil
	}

	weightedClusters := make([]*envoy_route.WeightedCluster_ClusterWeight, 0)

	var totalWeight int
	for _, cluster := range strings.Split(weightedClusterSpec, ",") {
		// format = clustername : weight
		clusterConfig := strings.Split(cluster, ":")
		if len(clusterConfig) == 2 {

			var clusterName, clusterWeight string = clusterConfig[0], clusterConfig[1]

			if weight, err := strconv.Atoi(clusterWeight); err == nil {
				if weight >= 1 && weight <= 10000000 {
					weightedClusters = append(weightedClusters,
						&envoy_route.WeightedCluster_ClusterWeight{
							Name:   strings.TrimSpace(clusterName),
							Weight: protoUint32(uint32(weight)),
						})
					totalWeight += weight
				}
			} else {
				s.logger.Warn("Weighted cluster unparsable weight value",
					zap.String("route", route.Name), zap.String("cluster", cluster))
			}
		} else {
			s.logger.Warn("Weighted cluster does not have weight value",
				zap.String("route", route.Name), zap.String("cluster", cluster))
		}
	}

	return &envoy_route.RouteAction_WeightedClusters{
		WeightedClusters: &envoy_route.WeightedCluster{
			Clusters:    weightedClusters,
			TotalWeight: protoUint32(uint32(totalWeight)),
		},
	}
}

func (s *server) buildRequestMirrorPolicies(route types.Route) []*envoy_route.RouteAction_RequestMirrorPolicy {

	mirrorCluster := route.Attributes.GetAsString(types.AttributeRequestMirrorCluster, "")
	mirrorPercentage := route.Attributes.GetAsString(types.AttributeRequestMirrorPercentage, "")

	if mirrorCluster == "" || mirrorPercentage == "" {
		return nil
	}

	percentage, _ := strconv.Atoi(mirrorPercentage)
	if percentage < 0 || percentage > 100 {
		s.logger.Warn("Route has incorrect request mirror ratio",
			zap.String("route", route.Name), zap.String("cluster", mirrorCluster),
			zap.String("percentage", mirrorPercentage))
		return nil
	}

	return []*envoy_route.RouteAction_RequestMirrorPolicy{
		{
			Cluster: mirrorCluster,
			RuntimeFraction: &envoy_core.RuntimeFractionalPercent{
				DefaultValue: &envoy_type.FractionalPercent{
					Numerator:   uint32(percentage),
					Denominator: envoy_type.FractionalPercent_HUNDRED,
				},
			},
		},
	}
}

// buildCorsPolicy return CorsPolicy based upon a route's attribute(s)
func buildCorsPolicy(route types.Route) *envoy_route.CorsPolicy {

	var corsConfigured bool
	corsPolicy := envoy_route.CorsPolicy{}

	corsAllowMethods, err := route.Attributes.Get(types.AttributeCORSAllowMethods)
	if err == nil && corsAllowMethods != "" {
		corsPolicy.AllowMethods = corsAllowMethods
		corsConfigured = true
	}

	corsAllowHeaders, err := route.Attributes.Get(types.AttributeCORSAllowHeaders)
	if err == nil && corsAllowHeaders != "" {
		corsPolicy.AllowHeaders = corsAllowHeaders
		corsConfigured = true
	}

	corsExposeHeaders, err := route.Attributes.Get(types.AttributeCORSExposeHeaders)
	if err == nil && corsExposeHeaders != "" {
		corsPolicy.ExposeHeaders = corsExposeHeaders
		corsConfigured = true
	}

	corsMaxAge, err := route.Attributes.Get(types.AttributeCORSMaxAge)
	if err == nil && corsMaxAge != "" {
		corsPolicy.MaxAge = corsMaxAge
		corsConfigured = true
	}

	corsAllowCredentials, err := route.Attributes.Get(types.AttributeCORSAllowCredentials)
	if err == nil && corsAllowCredentials == types.AttributeValueTrue {
		corsPolicy.AllowCredentials = protoBool(true)
		corsConfigured = true
	}

	if corsConfigured {
		stringMatch := make([]*envoy_matcher.StringMatcher, 0, 1)
		corsPolicy.AllowOriginStringMatch = append(stringMatch, buildStringMatcher("."))
		return &corsPolicy
	}
	return nil
}

func buildStringMatcher(regexp string) *envoy_matcher.StringMatcher {

	return &envoy_matcher.StringMatcher{
		MatchPattern: &envoy_matcher.StringMatcher_SafeRegex{
			SafeRegex: buildRegexpMatcher(regexp),
		},
	}
}

func buildRegexpMatcher(regexp string) *envoy_matcher.RegexMatcher {

	return &envoy_matcher.RegexMatcher{
		EngineType: &envoy_matcher.RegexMatcher_GoogleRe2{
			GoogleRe2: &envoy_matcher.RegexMatcher_GoogleRE2{},
		},
		Regex: regexp,
	}
}

// buildRouteActionDirectResponse builds route config to have Envoy itself answer with status code and body
func buildRouteActionDirectResponse(route types.Route) *envoy_route.Route_DirectResponse {

	directResponseStatusCode, err := route.Attributes.Get(types.AttributeDirectResponseStatusCode)
	if err == nil && directResponseStatusCode != "" {
		statusCode, err := strconv.Atoi(directResponseStatusCode)

		if statusCode < 100 || statusCode > 500 {
			return nil
		}

		if err == nil && statusCode != 0 {
			response := envoy_route.Route_DirectResponse{
				DirectResponse: &envoy_route.DirectResponseAction{
					Status: uint32(statusCode),
				},
			}
			directResponseStatusBody, err := route.Attributes.Get(types.AttributeDirectResponseBody)
			if err == nil && directResponseStatusCode != "" {
				response.DirectResponse.Body = &envoy_core.DataSource{
					Specifier: &envoy_core.DataSource_InlineString{
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
func (s *server) buildRouteActionRedirectResponse(route types.Route) *envoy_route.Route_Redirect {

	redirectStatusCode := route.Attributes.GetAsString(types.AttributeRedirectStatusCode, "")
	redirectScheme := route.Attributes.GetAsString(types.AttributeRedirectScheme, "")
	RedirectHostName := route.Attributes.GetAsString(types.AttributeRedirectHostName, "")
	redirectPort := route.Attributes.GetAsString(types.AttributeRedirectPort, "")
	redirectPath := route.Attributes.GetAsString(types.AttributeRedirectPath, "")
	redirectStripQuery := route.Attributes.GetAsString(types.AttributeRedirectStripQuery, "")

	// Status code must be set other wise we not build redirect configuration
	if redirectStatusCode == "" {
		return nil
	}

	response := envoy_route.Route_Redirect{
		Redirect: &envoy_route.RedirectAction{},
	}

	// Do we have to set a particular statuscode?
	statusCode, _ := strconv.Atoi(redirectStatusCode)
	switch statusCode {
	case 301:
		response.Redirect.ResponseCode = envoy_route.RedirectAction_MOVED_PERMANENTLY
	case 302:
		response.Redirect.ResponseCode = envoy_route.RedirectAction_FOUND
	case 303:
		response.Redirect.ResponseCode = envoy_route.RedirectAction_SEE_OTHER
	case 307:
		response.Redirect.ResponseCode = envoy_route.RedirectAction_TEMPORARY_REDIRECT
	case 308:
		response.Redirect.ResponseCode = envoy_route.RedirectAction_PERMANENT_REDIRECT
	default:
		s.logger.Warn("Unsupported redirect status code value",
			zap.String("route", route.Name),
			zap.String("attribute", types.AttributeRedirectStatusCode))
		return nil
	}
	// Do we have to update scheme to http or https?
	if redirectScheme != "" {
		response.Redirect.SchemeRewriteSpecifier = &envoy_route.RedirectAction_SchemeRedirect{
			SchemeRedirect: redirectScheme,
		}
	}
	// Do we have to redirect to a specific hostname?
	if RedirectHostName != "" {
		response.Redirect.HostRedirect = RedirectHostName
	}
	// Do we have to redirect to a specific port?
	if redirectPort != "" {
		if port, err := strconv.Atoi(redirectPort); err == nil {
			if port >= 0 && port <= 65535 {
				response.Redirect.PortRedirect = uint32(port)
			}
		}
	}
	// Do we have to redirect to a specific path?
	if redirectPath != "" {
		response.Redirect.PathRewriteSpecifier = &envoy_route.RedirectAction_PathRedirect{
			PathRedirect: redirectPath,
		}
	}
	// Do we have to enable stripping of query string parameters?
	if redirectStripQuery == types.AttributeValueTrue {
		response.Redirect.StripQuery = true
	}

	return &response
}

// buildUpstreamHeadersToAdd consolidates all possible source for route upstream headers
func buildUpstreamHeadersToAdd(route types.Route) []*envoy_core.HeaderValueOption {

	// In case route-level attributes exist we have additional upstream headers
	headersToAdd := make(map[string]string)

	buildBasicAuth(route, headersToAdd)
	buildOptionalUpstreamHeader(route, headersToAdd, types.AttributeRequestHeaderToAdd1)
	buildOptionalUpstreamHeader(route, headersToAdd, types.AttributeRequestHeaderToAdd2)
	buildOptionalUpstreamHeader(route, headersToAdd, types.AttributeRequestHeaderToAdd3)
	buildOptionalUpstreamHeader(route, headersToAdd, types.AttributeRequestHeaderToAdd4)
	buildOptionalUpstreamHeader(route, headersToAdd, types.AttributeRequestHeaderToAdd5)

	if len(headersToAdd) != 0 {
		return buildHeadersList(headersToAdd)
	}
	return nil
}

// buildBasicAuth sets Basic Authentication for upstream requests
func buildBasicAuth(route types.Route, headersToAdd map[string]string) {

	usernamePassword, err := route.Attributes.Get(types.AttributeBasicAuth)
	if err == nil && usernamePassword != "" {
		authenticationDigest := base64.StdEncoding.EncodeToString([]byte(usernamePassword))

		headersToAdd["Authorization"] = "Basic " + authenticationDigest
	}
}

func buildOptionalUpstreamHeader(route types.Route,
	headersToAdd map[string]string, attributeName string) {

	if headerToSet, err := route.Attributes.Get(attributeName); err == nil {
		if headerValue := strings.Split(headerToSet, "="); len(headerValue) == 2 {
			headersToAdd[headerValue[0]] = headerValue[1]
		}
	}
}

// buildUpstreamHeadersToRemove compiles list of headers we need to remove
func buildUpstreamHeadersToRemove(route types.Route) []string {

	h := make([]string, 0, 10)

	// In case attribute AuthzAuthentication=true is set for this route we assume
	// extauthz will authenticate each request by evaluation the Authorization header.
	// Given this we should not send the original Authorization upstream as it very unlikely
	/// upstream will do authentication
	if value, err := route.Attributes.Get(types.AttributeRouteExtAuthz); err == nil &&
		value == types.AttributeValueTrue {
		h = append(h, "Authorization")
	}

	headersToRemove, err := route.Attributes.Get(types.AttributeRequestHeadersToRemove)
	if err == nil || headersToRemove != "" {
		for _, value := range strings.Split(headersToRemove, ",") {
			h = append(h, strings.TrimSpace(value))
		}
	}
	if len(h) == 0 {
		return nil
	}
	return h
}

// buildHeadersList creates map to hold headers to add or remove
func buildHeadersList(headers map[string]string) []*envoy_core.HeaderValueOption {

	if len(headers) == 0 {
		return nil
	}

	headerList := make([]*envoy_core.HeaderValueOption, 0, len(headers))
	for key, value := range headers {
		headerList = append(headerList, &envoy_core.HeaderValueOption{
			Header: &envoy_core.HeaderValue{
				Key:   key,
				Value: value,
			},
		})
	}
	return headerList
}

// buildRateLimits returns ratelimit configuration for a route
func buildRateLimits(route types.Route) []*envoy_route.RateLimit {

	// In case attribute RateLimiting does not exist or is not set to true
	// no ratelimiting will be configured
	value, err := route.Attributes.Get(types.AttributeRouteRateLimiting)
	if err != nil || value != types.AttributeValueTrue {
		return nil
	}

	return nil

	// TODO

	// return []*route.RateLimit{{
	// 	Actions: []*route.RateLimit_Action{{
	// 		ActionSpecifier: &route.RateLimit_Action_DynamicMetadata{
	// 			DynamicMetadata: &route.RateLimit_Action_DynamicMetaData{
	// 				DescriptorKey: "key",
	// 				DefaultValue:  "default_descriptor",
	// 				MetadataKey: &envoytypemetadata.MetadataKey{
	// 					Key: wellknown.HTTPExternalAuthorization,
	// 					Path: []*envoytypemetadata.MetadataKey_PathSegment{{
	// 						Segment: &envoytypemetadata.MetadataKey_PathSegment_Key{
	// 							Key: "rl.descriptor",
	// 						},
	// 					}},
	// 				},
	// 			},
	// 		},
	// 	}},
	// 	Stage: protoUint32(0),
	// 	Limit: &route.RateLimit_Override{
	// 		OverrideSpecifier: &route.RateLimit_Override_DynamicMetadata_{
	// 			DynamicMetadata: &route.RateLimit_Override_DynamicMetadata{
	// 				MetadataKey: &envoytypemetadata.MetadataKey{
	// 					Key: wellknown.HTTPExternalAuthorization,
	// 					Path: []*envoytypemetadata.MetadataKey_PathSegment{{
	// 						Segment: &envoytypemetadata.MetadataKey_PathSegment_Key{
	// 							Key: "rl.override",
	// 						},
	// 					}},
	// 				},
	// 			},
	// 		},
	// 	},
	// }}
}

func buildRetryPolicy(route types.Route) *envoy_route.RetryPolicy {

	RetryOn := route.Attributes.GetAsString(types.AttributeRetryOn, "")
	if RetryOn == "" {
		return nil
	}
	perTryTimeout := route.Attributes.GetAsDuration(types.AttributePerTryTimeout,
		types.DefaultPerRetryTimeout)
	numRetries := uint32(route.Attributes.GetAsUInt32(types.AttributeNumRetries,
		types.DefaultNumRetries))
	RetriableStatusCodes := buildStatusCodesSlice(
		route.Attributes.GetAsString(types.AttributeRetryOnStatusCodes,
			types.DefaultRetryStatusCodes))

	return &envoy_route.RetryPolicy{
		RetryOn:              RetryOn,
		NumRetries:           protoUint32(numRetries),
		PerTryTimeout:        durationpb.New(perTryTimeout),
		RetriableStatusCodes: RetriableStatusCodes,
		RetryHostPredicate: []*envoy_route.RetryPolicy_RetryHostPredicate{
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
			if value >= 100 && value < 600 {
				statusCodeSlice = append(statusCodeSlice, uint32(value))
			}
		}
	}
	return statusCodeSlice
}

func buildPerRouteFilterConfig(route types.Route) map[string]*any.Any {

	perRouteFilterConfigMap := make(map[string]*any.Any)

	if authzFilterConfig := perRouteAuthzFilterConfig(route); authzFilterConfig != nil {
		perRouteFilterConfigMap[wellknown.HTTPExternalAuthorization] = authzFilterConfig
	}

	if len(perRouteFilterConfigMap) != 0 {
		return perRouteFilterConfigMap
	}
	return nil
}

// perRouteAuthzFilterConfig sets all the authz specific filter options of a route
func perRouteAuthzFilterConfig(route types.Route) *anypb.Any {

	// filter http.ext_authz is always inline in the default filter chain configuration
	//

	// In case attribute AuthzAuthentication=true is set for this route
	// we enable authz on this route by not configuring filter settings for this route
	value, err := route.Attributes.Get(types.AttributeRouteExtAuthz)
	if err == nil && value == types.AttributeValueTrue {
		return nil
	}

	// Our default HTTP forwarding behaviour is to not authenticate,
	// hence we need to disable this filter per route
	perFilterExtAuthzConfig := &envoy_filter_authz.ExtAuthzPerRoute{
		Override: &envoy_filter_authz.ExtAuthzPerRoute_Disabled{
			Disabled: true,
		},
	}
	extAuthzTypedConf, e := anypb.New(perFilterExtAuthzConfig)
	if e != nil {
		return nil
	}
	return extAuthzTypedConf
}

// buildEnvoyVirtualClusters returns a VirtualCluster configuration for each route
func (s *server) buildEnvoyVirtualClusters(RouteGroup string, routes types.Routes) []*envoy_route.VirtualCluster {

	var envoyVirtualClusters []*envoy_route.VirtualCluster

	for _, route := range routes {
		if route.RouteGroup == RouteGroup {
			pathMatch := buildEnvoyVirtualClusterPathMatch(route)
			if pathMatch == nil {
				s.logger.Warn("Cannot build virtualcluster header match config",
					zap.String("route", route.Name), zap.String("routegroup", RouteGroup))
				return nil
			}
			envoyVirtualClusters = append(envoyVirtualClusters, &envoy_route.VirtualCluster{
				Name: route.Name,
				Headers: []*envoy_route.HeaderMatcher{
					pathMatch,
				},
			})
		}
	}
	return envoyVirtualClusters
}

// buildEnvoyVirtualClusterPathMatch returns a matcher based upon the path type of a route
func buildEnvoyVirtualClusterPathMatch(route types.Route) *envoy_route.HeaderMatcher {

	matcher := &envoy_route.HeaderMatcher{
		Name: ":path",
	}
	switch route.PathType {
	case types.AttributeValuePathTypePath:
		matcher.HeaderMatchSpecifier = &envoy_route.HeaderMatcher_ExactMatch{
			ExactMatch: route.Path,
		}
		return matcher
	case types.AttributeValuePathTypePrefix:
		matcher.HeaderMatchSpecifier = &envoy_route.HeaderMatcher_PrefixMatch{
			PrefixMatch: route.Path,
		}
		return matcher
	case types.AttributeValuePathTypeRegexp:
		matcher.HeaderMatchSpecifier = &envoy_route.HeaderMatcher_SafeRegexMatch{
			SafeRegexMatch: &envoy_matcher.RegexMatcher{
				Regex: route.Path,
			},
		}
		return matcher
	}
	return nil
}
