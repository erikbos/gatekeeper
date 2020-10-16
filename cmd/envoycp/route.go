package main

import (
	"encoding/base64"
	"strconv"
	"strings"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoyauth "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	envoymatcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	envoytypemetadata "github.com/envoyproxy/go-control-plane/envoy/type/metadata/v3"
	envoytype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// getEnvoyRouteConfig returns array of all envoy routes
func (s *server) getEnvoyRouteConfig() ([]cache.Resource, error) {
	var envoyRoutes []cache.Resource

	RouteGroupNames := s.getRouteGroupNames(s.dbentities.GetRoutes())
	for RouteGroupName := range RouteGroupNames {
		s.logger.Info("Compiling configuration", zap.String("routegroup", RouteGroupName))
		envoyRoutes = append(envoyRoutes,
			s.buildEnvoyListenerRouteConfig(RouteGroupName,
				s.dbentities.GetRoutes()))
	}

	return envoyRoutes, nil
}

// getListenerPorts returns set of unique RouteGroup names
func (s *server) getRouteGroupNames(vhosts types.Routes) map[string]bool {
	RouteGroupNames := map[string]bool{}
	for _, routeEntry := range s.dbentities.GetRoutes() {
		RouteGroupNames[routeEntry.RouteGroup] = true
	}
	return RouteGroupNames
}

// buildEnvoyListenerRouteConfig builds vhost and route configuration of one RouteGroup
func (s *server) buildEnvoyListenerRouteConfig(RouteGroup string,
	routes types.Routes) *route.RouteConfiguration {

	return &route.RouteConfiguration{
		Name: RouteGroup,
		VirtualHosts: []*route.VirtualHost{
			{
				Name:    RouteGroup,
				Domains: s.getVhostsInRouteGroup(RouteGroup),
				Routes:  s.buildEnvoyRoutes(RouteGroup, routes),
			},
		},
	}
}

// buildEnvoyRoute returns all Envoy routes belong to one RouteGroup
func (s *server) buildEnvoyRoutes(RouteGroup string, routes types.Routes) []*route.Route {
	var envoyRoutes []*route.Route

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
func (s *server) buildEnvoyRoute(routeEntry types.Route) *route.Route {
	routeMatch := buildRouteMatch(routeEntry)
	if routeMatch == nil {
		s.logger.Warn("Cannot build route match config", zap.String("route", routeEntry.Name))
		return nil
	}

	envoyRoute := &route.Route{
		Name:  routeEntry.Name,
		Match: routeMatch,
	}

	// Set all route specific filter options
	envoyRoute.TypedPerFilterConfig = buildPerRouteFilterConfig(routeEntry)

	// Add direct response if configured: in this case Envoy itself will answer
	if _, err := routeEntry.Attributes.Get(types.AttributeDirectResponseStatusCode); err == nil {
		envoyRoute.Action = buildRouteActionDirectResponse(routeEntry)
		return envoyRoute
	}

	// Add redirect response if configured: in this case Envoy will generate HTTP redirect
	if _, err := routeEntry.Attributes.Get(types.AttributeRedirectStatusCode); err == nil {
		envoyRoute.Action = s.buildRouteActionRedirectResponse(routeEntry)
		return envoyRoute
	}

	// Add cluster(s) to forward this route to
	_, err1 := routeEntry.Attributes.Get(types.AttributeCluster)
	_, err2 := routeEntry.Attributes.Get(types.AttributeWeightedClusters)
	if err1 != nil || err2 != nil {
		envoyRoute.Action = s.buildRouteActionCluster(routeEntry)
	}

	// Stop in case we do not have any route action defined
	// WARNING: We cannot compare against just "nil" as .Action is an interface
	if envoyRoute.Action == (*route.Route_Route)(nil) {
		s.logger.Warn("Route without destination cluster", zap.String("route", routeEntry.Name))
		return nil
	}

	envoyRoute.RequestHeadersToAdd = buildUpstreamHeadersToAdd(routeEntry)
	envoyRoute.RequestHeadersToRemove = buildUpstreamHeadersToRemove(routeEntry)

	return envoyRoute
}

// buildRouteMatch returns route config we match on
func buildRouteMatch(routeEntry types.Route) *route.RouteMatch {

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
func (s *server) buildRouteActionCluster(routeEntry types.Route) *route.Route_Route {

	action := &route.Route_Route{
		Route: &route.RouteAction{
			Cors:                 buildCorsPolicy(routeEntry),
			HostRewriteSpecifier: buildHostRewriteSpecifier(routeEntry),
			RetryPolicy:          buildRetryPolicy(routeEntry),
			Timeout: ptypes.DurationProto(routeEntry.Attributes.GetAsDuration(types.AttributeTimeout,
				types.DefaultRouteTimeout)),
		},
	}

	prefixRewrite, err := routeEntry.Attributes.Get(types.AttributePrefixRewrite)
	if err == nil && prefixRewrite != "" {
		action.Route.PrefixRewrite = prefixRewrite
	}

	if cluster, err := routeEntry.Attributes.Get(types.AttributeCluster); err == nil {
		action.Route.ClusterSpecifier = &route.RouteAction_Cluster{
			Cluster: cluster,
		}
	}

	if _, err := routeEntry.Attributes.Get(types.AttributeWeightedClusters); err == nil {
		action.Route.ClusterSpecifier = s.buildWeightedClusters(routeEntry)
	}

	// Stop in case we do not have any route cluster destination
	if action.Route.ClusterSpecifier == nil {
		return nil
	}

	action.Route.RateLimits = buildRateLimits(routeEntry)
	action.Route.RequestMirrorPolicies = s.buildRequestMirrorPolicies(routeEntry)

	return action
}

func (s *server) buildWeightedClusters(routeEntry types.Route) *route.RouteAction_WeightedClusters {

	weightedClusterSpec, err := routeEntry.Attributes.Get(types.AttributeWeightedClusters)
	if err != nil || weightedClusterSpec == "" {
		return nil
	}

	weightedClusters := make([]*route.WeightedCluster_ClusterWeight, 0)
	var clusterWeight, totalWeight int

	for _, cluster := range strings.Split(weightedClusterSpec, ",") {

		clusterConfig := strings.Split(cluster, ":")

		if len(clusterConfig) == 2 {
			clusterWeight, _ = strconv.Atoi(clusterConfig[1])
		} else {
			s.logger.Warn("Weighted destination cluster does not have weight value",
				zap.String("route", routeEntry.Name), zap.String("cluster", cluster))

			clusterWeight = 1
		}
		totalWeight += clusterWeight

		clusterToAdd := &route.WeightedCluster_ClusterWeight{
			Name:   strings.TrimSpace(clusterConfig[0]),
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

func (s *server) buildRequestMirrorPolicies(routeEntry types.Route) []*route.RouteAction_RequestMirrorPolicy {

	mirrorCluster := routeEntry.Attributes.GetAsString(types.AttributeRequestMirrorCluster, "")
	mirrorPercentage := routeEntry.Attributes.GetAsString(types.AttributeRequestMirrorPercentage, "")

	if mirrorCluster == "" || mirrorPercentage == "" {
		return nil
	}

	percentage, _ := strconv.Atoi(mirrorPercentage)
	if percentage < 0 || percentage > 100 {
		s.logger.Warn("Route has incorrect request mirror ratio",
			zap.String("route", routeEntry.Name), zap.String("cluster", mirrorCluster),
			zap.String("percentage", mirrorPercentage))
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
func buildCorsPolicy(routeEntry types.Route) *route.CorsPolicy {

	var corsConfigured bool
	corsPolicy := route.CorsPolicy{}

	corsAllowMethods, err := routeEntry.Attributes.Get(types.AttributeCORSAllowMethods)
	if err == nil && corsAllowMethods != "" {
		corsPolicy.AllowMethods = corsAllowMethods
		corsConfigured = true
	}

	corsAllowHeaders, err := routeEntry.Attributes.Get(types.AttributeCORSAllowHeaders)
	if err == nil && corsAllowMethods != "" {
		corsPolicy.AllowHeaders = corsAllowHeaders
		corsConfigured = true
	}

	corsExposeHeaders, err := routeEntry.Attributes.Get(types.AttributeCORSExposeHeaders)
	if err == nil && corsAllowMethods != "" {
		corsPolicy.ExposeHeaders = corsExposeHeaders
		corsConfigured = true
	}

	corsMaxAge, err := routeEntry.Attributes.Get(types.AttributeCORSMaxAge)
	if err == nil && corsAllowMethods != "" {
		corsPolicy.MaxAge = corsMaxAge
		corsConfigured = true
	}

	corsAllowCredentials, err := routeEntry.Attributes.Get(types.AttributeCORSAllowCredentials)
	if err == nil && corsAllowCredentials == types.AttributeValueTrue {
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
func buildHostRewriteSpecifier(routeEntry types.Route) *route.RouteAction_HostRewriteLiteral {

	upstreamHostHeader, err := routeEntry.Attributes.Get(types.AttributeHostHeader)
	if err == nil && upstreamHostHeader != "" {
		return &route.RouteAction_HostRewriteLiteral{
			HostRewriteLiteral: upstreamHostHeader,
		}
	}

	return nil
}

// buildRouteActionDirectResponse builds route config to have Envoy itself answer with status code and body
func buildRouteActionDirectResponse(routeEntry types.Route) *route.Route_DirectResponse {

	directResponseStatusCode, err := routeEntry.Attributes.Get(types.AttributeDirectResponseStatusCode)
	if err == nil && directResponseStatusCode != "" {
		statusCode, err := strconv.Atoi(directResponseStatusCode)

		if err == nil && statusCode != 0 {
			response := route.Route_DirectResponse{
				DirectResponse: &route.DirectResponseAction{
					Status: uint32(statusCode),
				},
			}
			directResponseStatusBody, err := routeEntry.Attributes.Get(types.AttributeDirectResponseBody)
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
func (s *server) buildRouteActionRedirectResponse(routeEntry types.Route) *route.Route_Redirect {

	redirectStatusCode := routeEntry.Attributes.GetAsString(types.AttributeRedirectStatusCode, "")
	redirectScheme := routeEntry.Attributes.GetAsString(types.AttributeRedirectScheme, "")
	RedirectHostName := routeEntry.Attributes.GetAsString(types.AttributeRedirectHostName, "")
	redirectPort := routeEntry.Attributes.GetAsString(types.AttributeRedirectPort, "")
	redirectPath := routeEntry.Attributes.GetAsString(types.AttributeRedirectPath, "")
	redirectStripQuery := routeEntry.Attributes.GetAsString(types.AttributeRedirectStripQuery, "")

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
		s.logger.Warn("Unsupported redirect status code value",
			zap.String("route", routeEntry.Name),
			zap.String("attribute", types.AttributeRedirectStatusCode))
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
	if redirectStripQuery == types.AttributeValueTrue {
		response.Redirect.StripQuery = true
	}

	return &response
}

// buildUpstreamHeadersToAdd consolidates all possible source for route upstream headers
func buildUpstreamHeadersToAdd(routeEntry types.Route) []*core.HeaderValueOption {

	// In case route-level attributes exist we have additional upstream headers
	headersToAdd := make(map[string]string)

	buildBasicAuth(routeEntry, headersToAdd)
	buildOptionalUpstreamHeader(routeEntry, headersToAdd, types.AttributeHeaderToAdd1)
	buildOptionalUpstreamHeader(routeEntry, headersToAdd, types.AttributeHeaderToAdd2)
	buildOptionalUpstreamHeader(routeEntry, headersToAdd, types.AttributeHeaderToAdd3)
	buildOptionalUpstreamHeader(routeEntry, headersToAdd, types.AttributeHeaderToAdd4)
	buildOptionalUpstreamHeader(routeEntry, headersToAdd, types.AttributeHeaderToAdd5)

	if len(headersToAdd) != 0 {
		return buildHeadersList(headersToAdd)
	}
	return nil
}

// buildBasicAuth sets Basic Authentication for upstream requests
func buildBasicAuth(routeEntry types.Route, headersToAdd map[string]string) {

	usernamePassword, err := routeEntry.Attributes.Get(types.AttributeBasicAuth)
	if err == nil && usernamePassword != "" {
		authenticationDigest := base64.StdEncoding.EncodeToString([]byte(usernamePassword))

		headersToAdd["Authorization"] = authenticationDigest
	}
}

func buildOptionalUpstreamHeader(routeEntry types.Route,
	headersToAdd map[string]string, attributeName string) {

	if headerToSet, err := routeEntry.Attributes.Get(attributeName); err == nil {
		if headerValue := strings.Split(headerToSet, "="); len(headerValue) == 2 {
			headersToAdd[headerValue[0]] = headerValue[1]
		}
	}
}

// buildUpstreamHeadersToRemove compiles list of headers we need to remove
func buildUpstreamHeadersToRemove(routeEntry types.Route) []string {

	headersToRemove, err := routeEntry.Attributes.Get(types.AttributeHeadersToRemove)
	if err != nil || headersToRemove == "" {
		return nil
	}
	h := make([]string, 0, 10)
	h = append(h, strings.Split(headersToRemove, ",")...)
	for _, value := range strings.Split(headersToRemove, ",") {
		h = append(h, strings.TrimSpace(value))
	}
	if len(h) != 0 {
		return h
	}
	return nil
}

// buildHeadersList creates map to hold headers to add or remove
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

// buildRateLimits returns ratelimit configuration for a route
func buildRateLimits(routeEntry types.Route) []*route.RateLimit {

	// In case attribute RateLimiting does not exist or is not set to true
	// no ratelimiting enabled
	value, err := routeEntry.Attributes.Get(types.AttributeRateLimiting)
	if err != nil || value != types.AttributeValueTrue {
		return nil
	}

	// Enable ratelimiting
	// TOD (fix configuration)
	return []*route.RateLimit{{
		Actions: []*route.RateLimit_Action{{
			ActionSpecifier: &route.RateLimit_Action_DynamicMetadata{
				DynamicMetadata: &route.RateLimit_Action_DynamicMetaData{
					DescriptorKey: "1",
					MetadataKey: &envoytypemetadata.MetadataKey{
						Key: wellknown.HTTPExternalAuthorization,
						Path: []*envoytypemetadata.MetadataKey_PathSegment{{
							Segment: &envoytypemetadata.MetadataKey_PathSegment_Key{
								Key: "42",
							},
						}},
					},
					DefaultValue: "5",
				},
			},
		}},
		Stage: protoUint32(0),
		Limit: &route.RateLimit_Override{
			OverrideSpecifier: &route.RateLimit_Override_DynamicMetadata_{
				DynamicMetadata: &route.RateLimit_Override_DynamicMetadata{
					MetadataKey: &envoytypemetadata.MetadataKey{
						Key: wellknown.HTTPExternalAuthorization,
						Path: []*envoytypemetadata.MetadataKey_PathSegment{{
							Segment: &envoytypemetadata.MetadataKey_PathSegment_Key{
								Key: "rl",
							},
						}},
					},
				},
			},
		},
	}}
}

func buildRetryPolicy(routeEntry types.Route) *route.RetryPolicy {

	RetryOn := routeEntry.Attributes.GetAsString(types.AttributeRetryOn, "")
	if RetryOn == "" {
		return nil
	}
	perTryTimeout := routeEntry.Attributes.GetAsDuration(types.AttributePerTryTimeout,
		types.DefaultPerRetryTimeout)
	numRetries := uint32(routeEntry.Attributes.GetAsUInt32(types.AttributeNumRetries, 2))
	RetriableStatusCodes := buildStatusCodesSlice(
		routeEntry.Attributes.GetAsString(types.AttributeRetryOnStatusCodes,
			types.DefaultRetryStatusCodes))

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

func buildPerRouteFilterConfig(routeEntry types.Route) map[string]*any.Any {

	perRouteFilterConfigMap := make(map[string]*any.Any)

	if authzFilterConfig := perRouteAuthzFilterConfig(routeEntry); authzFilterConfig != nil {
		perRouteFilterConfigMap[wellknown.HTTPExternalAuthorization] = authzFilterConfig
	}

	if len(perRouteFilterConfigMap) != 0 {
		return perRouteFilterConfigMap
	}
	return nil
}

// perRouteAuthzFilterConfig sets all the authz specific filter options of a route
func perRouteAuthzFilterConfig(routeEntry types.Route) *anypb.Any {

	// filter http.ext_authz is always inline in the default filter chain configuration
	//

	// In case attribute AuthzAuthentication=true is set for this route
	// we enable authz on this route by not configuring filter settings for this route
	value, err := routeEntry.Attributes.Get(types.AttributeAuthentication)
	if err == nil && value == types.AttributeValueTrue {
		return nil
	}

	// Our default HTTP forwarding behaviour is to not authenticate,
	// hence we need to disable this filter per route
	perFilterExtAuthzConfig := &envoyauth.ExtAuthzPerRoute{
		Override: &envoyauth.ExtAuthzPerRoute_Disabled{
			Disabled: true,
		},
	}
	ratelimitTypedConf, e := ptypes.MarshalAny(perFilterExtAuthzConfig)
	if e != nil {
		return nil
	}
	return ratelimitTypedConf
}
