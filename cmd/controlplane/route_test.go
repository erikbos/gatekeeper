package main

import (
	"sort"
	"strings"
	"testing"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_filter_authz "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	envoy_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/erikbos/gatekeeper/pkg/types"
)

func Test_buildEnvoyRoute(t *testing.T) {

	s := newServerForTesting()

	// s.dbentities = db.NewEntityCache()

	tests := []struct {
		name       string
		RouteGroup string
		routes     types.Routes
		expected   []*envoy_route.Route
	}{
		{
			name:       "1 upstream cluster + properties",
			RouteGroup: "default",

			routes: types.Routes{
				{
					Name:       "default",
					Path:       "/default",
					PathType:   types.AttributeValuePathTypePrefix,
					RouteGroup: "default",
					Attributes: types.Attributes{
						{
							Name:  types.AttributeCluster,
							Value: "upstream",
						},
						{
							Name:  types.AttributePrefixRewrite,
							Value: "/seconddefault",
						},
						{
							Name:  types.AttributeHostHeader,
							Value: "www.example.com",
						},
						{
							Name:  types.AttributeTimeout,
							Value: "2s",
						},
					},
				},
			},
			expected: []*envoy_route.Route{
				{
					Name: "default",
					Match: &envoy_route.RouteMatch{
						PathSpecifier: &envoy_route.RouteMatch_Prefix{
							Prefix: "/default",
						},
					},
					Action: &envoy_route.Route_Route{
						Route: &envoy_route.RouteAction{
							ClusterSpecifier: &envoy_route.RouteAction_Cluster{
								Cluster: "upstream",
							},
							PrefixRewrite: "/seconddefault",
							HostRewriteSpecifier: &envoy_route.RouteAction_HostRewriteLiteral{
								HostRewriteLiteral: "www.example.com",
							},
							Timeout: durationpb.New(2 * time.Second),
						},
					},
					TypedPerFilterConfig: map[string]*anypb.Any{
						wellknown.HTTPExternalAuthorization: mustMarshalAny(&envoy_filter_authz.ExtAuthzPerRoute{
							Override: &envoy_filter_authz.ExtAuthzPerRoute_Disabled{
								Disabled: true,
							},
						}),
					},
				},
			},
		},
		{
			name:       "2 weighted upstream clusters + non-matching routegroup",
			RouteGroup: "default",

			routes: types.Routes{
				{
					Name:       "default",
					Path:       "/",
					PathType:   types.AttributeValuePathTypePrefix,
					RouteGroup: "default",
					Attributes: types.Attributes{
						{
							Name:  types.AttributeWeightedClusters,
							Value: "upstream1:10,upstream2:25",
						},
					},
				},
				{
					Name:       "backup",
					Path:       "/backup",
					PathType:   types.AttributeValuePathTypePrefix,
					RouteGroup: "non-http",
					Attributes: types.Attributes{
						{
							Name:  types.AttributeCluster,
							Value: "upstream",
						},
					},
				},
			},
			expected: []*envoy_route.Route{
				{
					Name: "default",
					Match: &envoy_route.RouteMatch{
						PathSpecifier: &envoy_route.RouteMatch_Prefix{
							Prefix: "/",
						},
					},
					Action: &envoy_route.Route_Route{
						Route: &envoy_route.RouteAction{
							ClusterSpecifier: &envoy_route.RouteAction_WeightedClusters{
								WeightedClusters: &envoy_route.WeightedCluster{
									Clusters: []*envoy_route.WeightedCluster_ClusterWeight{
										{
											Name: "upstream1",
											Weight: &wrapperspb.UInt32Value{
												Value: 10,
											},
										},
										{
											Name: "upstream2",
											Weight: &wrapperspb.UInt32Value{
												Value: 25,
											},
										},
									},
									TotalWeight: &wrapperspb.UInt32Value{
										Value: 35,
									},
								},
							},
						},
					},
					TypedPerFilterConfig: map[string]*anypb.Any{
						wellknown.HTTPExternalAuthorization: mustMarshalAny(&envoy_filter_authz.ExtAuthzPerRoute{
							Override: &envoy_filter_authz.ExtAuthzPerRoute_Disabled{
								Disabled: true,
							},
						}),
					},
				},
			},
		},
		{
			name:       "1 direct response",
			RouteGroup: "default",

			routes: types.Routes{
				{
					Name:       "default",
					Path:       "/dsr",
					PathType:   types.AttributeValuePathTypePath,
					RouteGroup: "default",
					Attributes: types.Attributes{
						{
							Name:  types.AttributeDirectResponseStatusCode,
							Value: "200",
						},
						{
							Name:  types.AttributeDirectResponseBody,
							Value: "Hello World",
						},
					},
				},
			},
			expected: []*envoy_route.Route{
				{
					Name: "default",
					Match: &envoy_route.RouteMatch{
						PathSpecifier: &envoy_route.RouteMatch_Path{
							Path: "/dsr",
						},
					},
					Action: &envoy_route.Route_DirectResponse{
						DirectResponse: &envoy_route.DirectResponseAction{
							Status: 200,
							Body: &envoy_core.DataSource{
								Specifier: &envoy_core.DataSource_InlineString{
									InlineString: "Hello World",
								},
							},
						},
					},
					TypedPerFilterConfig: map[string]*anypb.Any{
						wellknown.HTTPExternalAuthorization: mustMarshalAny(&envoy_filter_authz.ExtAuthzPerRoute{
							Override: &envoy_filter_authz.ExtAuthzPerRoute_Disabled{
								Disabled: true,
							},
						}),
					},
				},
			},
		},
		{
			name:       "no upstream cluster",
			RouteGroup: "default",

			routes: types.Routes{
				{
					Name:       "default",
					Path:       "/",
					PathType:   types.AttributeValuePathTypePrefix,
					RouteGroup: "default",
					Attributes: types.Attributes{
						{
							Name:  types.AttributePrefixRewrite,
							Value: "/seconddefault",
						},
						{
							Name:  types.AttributeHostHeader,
							Value: "www.example.com",
						},
						{
							Name:  types.AttributeTimeout,
							Value: "2s",
						},
					},
				},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		equalf(t, test.expected,
			s.buildEnvoyRoutes(test.RouteGroup, test.routes), test.name)
	}
}

func Test_buildRouteMatch(t *testing.T) {

	tests := []struct {
		name     string
		route    types.Route
		expected *envoy_route.RouteMatch
	}{
		{
			name: "path match",
			route: types.Route{
				PathType: types.AttributeValuePathTypePath,
				Path:     "/users",
			},
			expected: &envoy_route.RouteMatch{
				PathSpecifier: &envoy_route.RouteMatch_Path{
					Path: "/users",
				},
			},
		},
		{
			name: "prefix match",
			route: types.Route{
				PathType: types.AttributeValuePathTypePrefix,
				Path:     "/developers",
			},
			expected: &envoy_route.RouteMatch{
				PathSpecifier: &envoy_route.RouteMatch_Prefix{
					Prefix: "/developers",
				},
			},
		},
		{
			name: "regexp match",
			route: types.Route{
				PathType: types.AttributeValuePathTypeRegexp,
				Path:     "/products",
			},
			expected: &envoy_route.RouteMatch{
				PathSpecifier: &envoy_route.RouteMatch_SafeRegex{
					SafeRegex: buildRegexpMatcher("/products"),
				},
			},
		},
		{
			name: "no weighted loadbalancing",
			route: types.Route{
				PathType: "unknown",
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		equalf(t, test.expected, buildRouteMatch(test.route), test.name)
	}
}

func Test_buildWeightedClusters(t *testing.T) {

	s := newServerForTesting()

	tests := []struct {
		name     string
		route    types.Route
		expected *envoy_route.RouteAction_WeightedClusters
	}{
		{
			name: "1 weighted cluster",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeWeightedClusters,
						Value: "cluster1:25",
					},
				},
			},
			expected: &envoy_route.RouteAction_WeightedClusters{
				WeightedClusters: &envoy_route.WeightedCluster{
					Clusters: []*envoy_route.WeightedCluster_ClusterWeight{
						{
							Name:   strings.TrimSpace("cluster1"),
							Weight: protoUint32(uint32(25)),
						},
					},
					TotalWeight: protoUint32(uint32(25)),
				},
			},
		},
		{
			name: "2 clusters, one with weight",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeWeightedClusters,
						Value: "cluster1:50,cluster2,cluster3:75",
					},
				},
			},
			expected: &envoy_route.RouteAction_WeightedClusters{
				WeightedClusters: &envoy_route.WeightedCluster{
					Clusters: []*envoy_route.WeightedCluster_ClusterWeight{
						{
							Name:   strings.TrimSpace("cluster1"),
							Weight: protoUint32(uint32(50)),
						},
						{
							Name:   strings.TrimSpace("cluster3"),
							Weight: protoUint32(uint32(75)),
						},
					},
					TotalWeight: protoUint32(uint32(125)),
				},
			},
		},
		{
			name: "3 weighted clusters",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeWeightedClusters,
						Value: "cluster1:25,cluster2:50,cluster3:75",
					},
				},
			},
			expected: &envoy_route.RouteAction_WeightedClusters{
				WeightedClusters: &envoy_route.WeightedCluster{
					Clusters: []*envoy_route.WeightedCluster_ClusterWeight{
						{
							Name:   strings.TrimSpace("cluster1"),
							Weight: protoUint32(uint32(25)),
						},
						{
							Name:   strings.TrimSpace("cluster2"),
							Weight: protoUint32(uint32(50)),
						},
						{
							Name:   strings.TrimSpace("cluster3"),
							Weight: protoUint32(uint32(75)),
						},
					},
					TotalWeight: protoUint32(uint32(150)),
				},
			},
		},
		{
			name: "no weighted loadbalancing",
			route: types.Route{
				Attributes: types.Attributes{
					{},
				},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		equalf(t, test.expected,
			s.buildWeightedClusters(test.route), test.name)
	}
}

func Test_buildRequestMirrorPolicies(t *testing.T) {

	s := newServerForTesting()

	tests := []struct {
		name     string
		route    types.Route
		expected []*envoy_route.RouteAction_RequestMirrorPolicy
	}{
		{
			name: "mirror configuration ok",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRequestMirrorCluster,
						Value: "second_cluster",
					},
					{
						Name:  types.AttributeRequestMirrorPercentage,
						Value: "55",
					},
				},
			},
			expected: []*envoy_route.RouteAction_RequestMirrorPolicy{
				{
					Cluster: "second_cluster",
					RuntimeFraction: &envoy_core.RuntimeFractionalPercent{
						DefaultValue: &envoy_type.FractionalPercent{
							Numerator:   uint32(55),
							Denominator: envoy_type.FractionalPercent_HUNDRED,
						},
					},
				},
			},
		},
		{
			name: "cluster set, mirror percentage wrong 1",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRequestMirrorCluster,
						Value: "second_cluster",
					},
					{
						Name:  types.AttributeRequestMirrorPercentage,
						Value: "101",
					},
				},
			},
			expected: nil,
		},
		{
			name: "cluster set, mirror percentage wrong 2",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRequestMirrorCluster,
						Value: "second_cluster",
					},
					{
						Name:  types.AttributeRequestMirrorPercentage,
						Value: "-1",
					},
				},
			},
			expected: nil,
		},
		{
			name: "mirror cluster only, no percentage",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRequestMirrorCluster,
						Value: "second_cluster",
					},
				},
			},
			expected: nil,
		},
		{
			name: "mirror percentage only, no cluster",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRequestMirrorPercentage,
						Value: "25",
					},
				},
			},
			expected: nil,
		},
		{
			name: "no mirror config",
			route: types.Route{
				Attributes: types.Attributes{
					{},
				},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		equalf(t, test.expected,
			s.buildRequestMirrorPolicies(test.route), test.name)
	}
}

func Test_buildCorsPolicy(t *testing.T) {

	stringMatch := make([]*envoy_matcher.StringMatcher, 0, 1)
	defaultStringMatcher := append(stringMatch, buildStringMatcher("."))

	tests := []struct {
		name     string
		route    types.Route
		expected *envoy_route.CorsPolicy
	}{
		{
			name: "CORS methods",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeCORSAllowMethods,
						Value: "GET",
					},
				},
			},
			expected: &envoy_route.CorsPolicy{
				AllowMethods:           "GET",
				AllowOriginStringMatch: defaultStringMatcher,
			},
		},
		{
			name: "CORS allowheaders",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeCORSAllowHeaders,
						Value: "supercors",
					},
				},
			},
			expected: &envoy_route.CorsPolicy{
				AllowHeaders:           "supercors",
				AllowOriginStringMatch: defaultStringMatcher,
			},
		},
		{
			name: "COFS exposeheaders",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeCORSExposeHeaders,
						Value: "accept",
					},
				},
			},
			expected: &envoy_route.CorsPolicy{
				ExposeHeaders:          "accept",
				AllowOriginStringMatch: defaultStringMatcher,
			},
		},
		{
			name: "CORS maxage",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeCORSMaxAge,
						Value: "3600",
					},
				},
			},
			expected: &envoy_route.CorsPolicy{
				MaxAge:                 "3600",
				AllowOriginStringMatch: defaultStringMatcher,
			},
		},
		{
			name: "CORS allow credentials",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeCORSAllowCredentials,
						Value: types.AttributeValueTrue,
					},
				},
			},
			expected: &envoy_route.CorsPolicy{
				AllowCredentials:       protoBool(true),
				AllowOriginStringMatch: defaultStringMatcher,
			},
		},
		{
			name: "no cors",
			route: types.Route{
				Attributes: types.Attributes{},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		equalf(t, test.expected,
			buildCorsPolicy(test.route), test.name)
	}
}

func Test_buildRouteActionDirectResponse(t *testing.T) {

	tests := []struct {
		name     string
		route    types.Route
		expected *envoy_route.Route_DirectResponse
	}{
		{
			name: "200 direct",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeDirectResponseStatusCode,
						Value: "200",
					},
				},
			},
			expected: &envoy_route.Route_DirectResponse{
				DirectResponse: &envoy_route.DirectResponseAction{
					Status: uint32(200),
				},
			},
		},
		{
			name: "255 + body",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeDirectResponseStatusCode,
						Value: "255",
					},
					{
						Name:  types.AttributeDirectResponseBody,
						Value: "Hello World",
					},
				},
			},
			expected: &envoy_route.Route_DirectResponse{
				DirectResponse: &envoy_route.DirectResponseAction{
					Status: uint32(255),
					Body: &envoy_core.DataSource{
						Specifier: &envoy_core.DataSource_InlineString{
							InlineString: "Hello World",
						},
					},
				},
			},
		},
		{
			name: "999",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeDirectResponseStatusCode,
						Value: "999",
					},
				},
			},
			expected: nil,
		},
		{
			name: "no direct response",
			route: types.Route{
				Attributes: types.Attributes{},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		equalf(t, test.expected,
			buildRouteActionDirectResponse(test.route), test.name)
	}
}

func Test_buildRouteActionRedirectResponse(t *testing.T) {

	s := newServerForTesting()

	tests := []struct {
		name     string
		route    types.Route
		expected *envoy_route.Route_Redirect
	}{
		{
			name: "301 redirect status code + scheme",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRedirectStatusCode,
						Value: "301",
					},
					{
						Name:  types.AttributeRedirectScheme,
						Value: "https",
					},
				},
			},
			expected: &envoy_route.Route_Redirect{
				Redirect: &envoy_route.RedirectAction{
					ResponseCode: envoy_route.RedirectAction_MOVED_PERMANENTLY,
					SchemeRewriteSpecifier: &envoy_route.RedirectAction_SchemeRedirect{
						SchemeRedirect: "https",
					},
				},
			},
		},
		{
			name: "302 redirect status code + hostname",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRedirectStatusCode,
						Value: "302",
					},
					{
						Name:  types.AttributeRedirectHostName,
						Value: "www.example.com",
					},
				},
			},
			expected: &envoy_route.Route_Redirect{
				Redirect: &envoy_route.RedirectAction{
					ResponseCode: envoy_route.RedirectAction_FOUND,
					HostRedirect: "www.example.com",
				},
			},
		},
		{
			name: "303 redirect status code + port",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRedirectStatusCode,
						Value: "303",
					},
					{
						Name:  types.AttributeRedirectPort,
						Value: "8080",
					},
				},
			},
			expected: &envoy_route.Route_Redirect{
				Redirect: &envoy_route.RedirectAction{
					ResponseCode: envoy_route.RedirectAction_SEE_OTHER,
					PortRedirect: 8080,
				},
			},
		},
		{
			name: "307 redirect status code + path",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRedirectStatusCode,
						Value: "307",
					},
					{
						Name:  types.AttributeRedirectPath,
						Value: "/newlocation",
					},
				},
			},
			expected: &envoy_route.Route_Redirect{
				Redirect: &envoy_route.RedirectAction{
					ResponseCode: envoy_route.RedirectAction_TEMPORARY_REDIRECT,
					PathRewriteSpecifier: &envoy_route.RedirectAction_PathRedirect{
						PathRedirect: "/newlocation",
					},
				},
			},
		},
		{
			name: "308 redirect status code + path",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRedirectStatusCode,
						Value: "308",
					},
					{
						Name:  types.AttributeRedirectStripQuery,
						Value: types.AttributeValueTrue,
					},
				},
			},
			expected: &envoy_route.Route_Redirect{
				Redirect: &envoy_route.RedirectAction{
					ResponseCode: envoy_route.RedirectAction_PERMANENT_REDIRECT,
					StripQuery:   true,
				},
			},
		},
		{
			name: "Empty redirect status code",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRedirectStatusCode,
						Value: "",
					},
				},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		equalf(t, test.expected,
			s.buildRouteActionRedirectResponse(test.route), test.name)
	}
}

func Test_buildUpstreamHeadersToAdd(t *testing.T) {

	tests := []struct {
		name     string
		route    types.Route
		expected []*envoy_core.HeaderValueOption
	}{
		{
			name: "1 upstream header",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRequestHeaderToAdd1,
						Value: "extra=900",
					},
				},
			},
			expected: []*envoy_core.HeaderValueOption{
				{
					Header: &envoy_core.HeaderValue{
						Key:   "extra",
						Value: "900",
					},
				},
			},
		},
		{
			name: "basic auth",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeBasicAuth,
						Value: "test:123",
					},
					{
						Name:  types.AttributeRequestHeaderToAdd1,
						Value: "name=api",
					},
				},
			},
			expected: []*envoy_core.HeaderValueOption{
				{
					Header: &envoy_core.HeaderValue{
						Key:   "Authorization",
						Value: "Basic dGVzdDoxMjM=",
					},
				},
				{
					Header: &envoy_core.HeaderValue{
						Key:   "name",
						Value: "api",
					},
				},
			},
		},
		{
			name: "no headers to add",
			route: types.Route{
				Attributes: types.Attributes{},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		received := buildUpstreamHeadersToAdd(test.route)

		// Sort both maps by Key so compare does not fail
		sort.Slice(test.expected, func(i, j int) bool {
			return test.expected[i].Header.Key < test.expected[j].Header.Key
		})
		sort.Slice(received, func(i, j int) bool {
			return received[i].Header.Key < received[j].Header.Key
		})

		equalf(t, test.expected, received, test.name)
	}
}

func Test_buildUpstreamHeadersToRemove(t *testing.T) {

	tests := []struct {
		name     string
		route    types.Route
		expected []string
	}{
		{
			name: "1 header to remove",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRequestHeadersToRemove,
						Value: "user-agent",
					},
				},
			},
			expected: []string{"user-agent"},
		},
		{
			name: "2 headers to remove, No Authorization",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRequestHeadersToRemove,
						Value: "accept",
					},
					{
						Name:  types.AttributeRouteExtAuthz,
						Value: types.AttributeValueTrue,
					},
				},
			},
			expected: []string{"Authorization", "accept"},
		},
		{
			name: "default header to remove",
			route: types.Route{
				Attributes: types.Attributes{},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		equalf(t, test.expected,
			buildUpstreamHeadersToRemove(test.route), test.name)
	}
}

func Test_buildHeadersList(t *testing.T) {

	tests := []struct {
		name     string
		headers  map[string]string
		expected []*envoy_core.HeaderValueOption
	}{
		{
			name: "1 header",
			headers: map[string]string{
				"user-agent": "netscape",
			},
			expected: []*envoy_core.HeaderValueOption{
				{
					Header: &envoy_core.HeaderValue{
						Key:   "user-agent",
						Value: "netscape",
					},
				},
			},
		},
		{
			name:     "no headers",
			headers:  map[string]string{},
			expected: nil,
		},
	}
	for _, test := range tests {
		equalf(t, test.expected,
			buildHeadersList(test.headers), test.name)
	}
}

func Test_buildRateLimits(t *testing.T) {

}

func Test_buildRetryPolicy(t *testing.T) {

	tests := []struct {
		name     string
		route    types.Route
		expected *envoy_route.RetryPolicy
	}{
		{
			name: "retry policy 5xx",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRetryOn,
						Value: "5xx",
					},
				},
			},
			expected: &envoy_route.RetryPolicy{
				RetryOn:              "5xx",
				NumRetries:           protoUint32(types.DefaultNumRetries),
				PerTryTimeout:        durationpb.New(types.DefaultPerRetryTimeout),
				RetriableStatusCodes: buildStatusCodesSlice(types.DefaultRetryStatusCodes),
				RetryHostPredicate: []*envoy_route.RetryPolicy_RetryHostPredicate{
					{Name: "envoy.retry_host_predicates.previous_hosts"},
				},
			},
		},
		{
			name: "retry policy 5xx",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRetryOn,
						Value: "connect-failure,reset",
					},
					{
						Name:  types.AttributeNumRetries,
						Value: "7",
					},
					{
						Name:  types.AttributePerTryTimeout,
						Value: "21ms",
					},
					{
						Name:  types.AttributeRetryOnStatusCodes,
						Value: "506,507",
					},
				},
			},
			expected: &envoy_route.RetryPolicy{
				RetryOn:              "connect-failure,reset",
				NumRetries:           protoUint32(7),
				PerTryTimeout:        durationpb.New(21 * time.Millisecond),
				RetriableStatusCodes: buildStatusCodesSlice("506,507"),
				RetryHostPredicate: []*envoy_route.RetryPolicy_RetryHostPredicate{
					{Name: "envoy.retry_host_predicates.previous_hosts"},
				},
			},
		},
		{
			name: "retry policy empty",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRetryOn,
						Value: "",
					},
				},
			},
			expected: nil,
		},
		{
			name: "no retry policy",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeHost,
						Value: "www.example.com",
					},
				},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		equalf(t, test.expected,
			buildRetryPolicy(test.route), test.name)
	}
}

func Test_buildStatusCodesSlice(t *testing.T) {

	tests := []struct {
		name        string
		statusCodes string
		expected    []uint32
	}{
		{
			name:        "1 code",
			statusCodes: "404",
			expected:    []uint32{404},
		},
		{
			name:        "3 codes, 1 non-int",
			statusCodes: "200,201,burp,206",
			expected:    []uint32{200, 201, 206},
		},
		{
			name:        "no codes",
			statusCodes: "bla",
			expected:    nil,
		},
	}
	for _, test := range tests {
		equalf(t, test.expected,
			buildStatusCodesSlice(test.statusCodes), test.name)
	}
}

func Test_buildPerRouteFilterConfig(t *testing.T) {

	tests := []struct {
		name     string
		route    types.Route
		expected map[string]*anypb.Any
	}{
		{
			name: "route auth explicit disabled",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRouteExtAuthz,
						Value: types.AttributeValueFalse,
					},
				},
			},
			expected: map[string]*anypb.Any{
				wellknown.HTTPExternalAuthorization: mustMarshalAny(&envoy_filter_authz.ExtAuthzPerRoute{
					Override: &envoy_filter_authz.ExtAuthzPerRoute_Disabled{
						Disabled: true,
					},
				}),
			},
		},
		{
			name: "route auth explicit enabled",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRouteExtAuthz,
						Value: types.AttributeValueTrue,
					},
				},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		equalf(t, test.expected,
			buildPerRouteFilterConfig(test.route), test.name)
	}
}

func Test_perRouteAuthzFilterConfig(t *testing.T) {

	tests := []struct {
		name     string
		route    types.Route
		expected *anypb.Any
	}{
		{
			name: "route auth explicit disabled",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRouteExtAuthz,
						Value: types.AttributeValueFalse,
					},
				},
			},
			expected: mustMarshalAny(&envoy_filter_authz.ExtAuthzPerRoute{
				Override: &envoy_filter_authz.ExtAuthzPerRoute_Disabled{
					Disabled: true,
				},
			}),
		},
		{
			name: "route auth explicit enabled",
			route: types.Route{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRouteExtAuthz,
						Value: types.AttributeValueTrue,
					},
				},
			},
			expected: nil,
		},
		{
			name: "route auth not configured",
			route: types.Route{
				Attributes: types.Attributes{},
			},
			expected: mustMarshalAny(&envoy_filter_authz.ExtAuthzPerRoute{
				Override: &envoy_filter_authz.ExtAuthzPerRoute_Disabled{
					Disabled: true,
				},
			}),
		},
	}
	for _, test := range tests {
		equalf(t, test.expected,
			perRouteAuthzFilterConfig(test.route), test.name)
	}
}

func Test_buildEnvoyVirtualClusters(t *testing.T) {

	s := newServerForTesting()

	tests := []struct {
		name       string
		RouteGroup string
		routes     types.Routes
		expected   []*envoy_route.VirtualCluster
	}{
		{
			name:       "routegroup 1 match, 1 mismatch",
			RouteGroup: "bikes",
			routes: types.Routes{
				{

					RouteGroup: "bikes",
					PathType:   "path",
					Path:       "/bikes",
				},
				{

					RouteGroup: "cars",
					PathType:   "path",
					Path:       "/cars",
				},
			},
			expected: []*envoy_route.VirtualCluster{
				{
					Headers: []*envoy_route.HeaderMatcher{
						{
							Name: ":path",
							HeaderMatchSpecifier: &envoy_route.HeaderMatcher_ExactMatch{
								ExactMatch: "/bikes",
							},
						},
					},
				},
			},
		},
		{
			name: "unknown match",
			routes: types.Routes{
				{
					PathType: "unknown",
				},
			}, expected: nil,
		},
	}
	for _, test := range tests {
		equalf(t, test.expected,
			s.buildEnvoyVirtualClusters(test.RouteGroup, test.routes), test.name)
	}
}

func Test_buildEnvoyVirtualClusterPathMatch(t *testing.T) {

	tests := []struct {
		name     string
		route    types.Route
		expected *envoy_route.HeaderMatcher
	}{
		{
			name: "path match",
			route: types.Route{
				PathType: "path",
				Path:     "/customers",
			},
			expected: &envoy_route.HeaderMatcher{
				Name: ":path",
				HeaderMatchSpecifier: &envoy_route.HeaderMatcher_ExactMatch{
					ExactMatch: "/customers",
				},
			},
		},
		{
			name: "prefix match",
			route: types.Route{
				PathType: "prefix",
				Path:     "/users",
			},
			expected: &envoy_route.HeaderMatcher{
				Name: ":path",
				HeaderMatchSpecifier: &envoy_route.HeaderMatcher_PrefixMatch{
					PrefixMatch: "/users",
				},
			},
		},
		{
			name: "regexp match",
			route: types.Route{
				PathType: "regexp",
				Path:     "/*bills*",
			},
			expected: &envoy_route.HeaderMatcher{
				Name: ":path",
				HeaderMatchSpecifier: &envoy_route.HeaderMatcher_SafeRegexMatch{
					SafeRegexMatch: &envoy_matcher.RegexMatcher{
						Regex: "/*bills*",
					},
				},
			},
		},
		{
			name: "unknown match",
			route: types.Route{
				PathType: "unknown",
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		equalf(t, test.expected,
			buildEnvoyVirtualClusterPathMatch(test.route), test.name)
	}
}
