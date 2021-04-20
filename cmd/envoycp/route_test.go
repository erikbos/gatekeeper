package main

import (
	"strings"
	"testing"
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoyauth "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	envoymatcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	envoytype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/require"
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
		expected   []*route.Route
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
			expected: []*route.Route{
				{
					Name: "default",
					Match: &route.RouteMatch{
						PathSpecifier: &route.RouteMatch_Prefix{
							Prefix: "/default",
						},
					},
					Action: &route.Route_Route{
						Route: &route.RouteAction{
							ClusterSpecifier: &route.RouteAction_Cluster{
								Cluster: "upstream",
							},
							PrefixRewrite: "/seconddefault",
							HostRewriteSpecifier: &route.RouteAction_HostRewriteLiteral{
								HostRewriteLiteral: "www.example.com",
							},
							Timeout: durationpb.New(2 * time.Second),
						},
					},
					TypedPerFilterConfig: map[string]*any.Any{
						wellknown.HTTPExternalAuthorization: mustMarshalAny(&envoyauth.ExtAuthzPerRoute{
							Override: &envoyauth.ExtAuthzPerRoute_Disabled{
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
			expected: []*route.Route{
				{
					Name: "default",
					Match: &route.RouteMatch{
						PathSpecifier: &route.RouteMatch_Prefix{
							Prefix: "/",
						},
					},
					Action: &route.Route_Route{
						Route: &route.RouteAction{
							ClusterSpecifier: &route.RouteAction_WeightedClusters{
								WeightedClusters: &route.WeightedCluster{
									Clusters: []*route.WeightedCluster_ClusterWeight{
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
					TypedPerFilterConfig: map[string]*any.Any{
						wellknown.HTTPExternalAuthorization: mustMarshalAny(&envoyauth.ExtAuthzPerRoute{
							Override: &envoyauth.ExtAuthzPerRoute_Disabled{
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
			expected: []*route.Route{
				{
					Name: "default",
					Match: &route.RouteMatch{
						PathSpecifier: &route.RouteMatch_Path{
							Path: "/dsr",
						},
					},
					Action: &route.Route_DirectResponse{
						DirectResponse: &route.DirectResponseAction{
							Status: 200,
							Body: &core.DataSource{
								Specifier: &core.DataSource_InlineString{
									InlineString: "Hello World",
								},
							},
						},
					},
					TypedPerFilterConfig: map[string]*any.Any{
						wellknown.HTTPExternalAuthorization: mustMarshalAny(&envoyauth.ExtAuthzPerRoute{
							Override: &envoyauth.ExtAuthzPerRoute_Disabled{
								Disabled: true,
							},
						}),
					},
				},
			},
		},
		{
			name:       "1 upstream cluster",
			RouteGroup: "default",

			routes: types.Routes{
				{
					Name:       "default",
					Path:       "/",
					PathType:   types.AttributeValuePathTypePrefix,
					RouteGroup: "default",
					Attributes: types.Attributes{
						{
							Name:  types.AttributeCluster,
							Value: "upstream",
						},
					},
				},
			},
			expected: []*route.Route{
				{
					Name: "default",
					Match: &route.RouteMatch{
						PathSpecifier: &route.RouteMatch_Prefix{
							Prefix: "/",
						},
					},
					Action: &route.Route_Route{
						Route: &route.RouteAction{
							ClusterSpecifier: &route.RouteAction_Cluster{
								Cluster: "upstream",
							},
						},
					},
					TypedPerFilterConfig: map[string]*any.Any{
						wellknown.HTTPExternalAuthorization: mustMarshalAny(&envoyauth.ExtAuthzPerRoute{
							Override: &envoyauth.ExtAuthzPerRoute_Disabled{
								Disabled: true,
							},
						}),
					},
				},
			},
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			s.buildEnvoyRoutes(test.RouteGroup, test.routes), test.name)
	}
}

func Test_buildRouteMatch(t *testing.T) {

	tests := []struct {
		name     string
		route    types.Route
		expected *route.RouteMatch
	}{
		{
			name: "path match",
			route: types.Route{
				PathType: types.AttributeValuePathTypePath,
				Path:     "/users",
			},
			expected: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Path{
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
			expected: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
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
			expected: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_SafeRegex{
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
		require.Equalf(t, test.expected,
			buildRouteMatch(test.route), test.name)
	}
}

func Test_buildWeightedClusters(t *testing.T) {

	s := newServerForTesting()

	tests := []struct {
		name     string
		route    types.Route
		expected *route.RouteAction_WeightedClusters
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
			expected: &route.RouteAction_WeightedClusters{
				WeightedClusters: &route.WeightedCluster{
					Clusters: []*route.WeightedCluster_ClusterWeight{
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
						Value: "cluster1,cluster2:50,cluster3:75",
					},
				},
			},
			expected: &route.RouteAction_WeightedClusters{
				WeightedClusters: &route.WeightedCluster{
					Clusters: []*route.WeightedCluster_ClusterWeight{
						{
							Name:   strings.TrimSpace("cluster1"),
							Weight: protoUint32(uint32(1)),
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
					TotalWeight: protoUint32(uint32(126)),
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
			expected: &route.RouteAction_WeightedClusters{
				WeightedClusters: &route.WeightedCluster{
					Clusters: []*route.WeightedCluster_ClusterWeight{
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
		require.Equalf(t, test.expected,
			s.buildWeightedClusters(test.route), test.name)
	}
}

func Test_buildRequestMirrorPolicies(t *testing.T) {

	s := newServerForTesting()

	tests := []struct {
		name     string
		route    types.Route
		expected []*route.RouteAction_RequestMirrorPolicy
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
			expected: []*route.RouteAction_RequestMirrorPolicy{
				{
					Cluster: "second_cluster",
					RuntimeFraction: &core.RuntimeFractionalPercent{
						DefaultValue: &envoytype.FractionalPercent{
							Numerator:   uint32(55),
							Denominator: envoytype.FractionalPercent_HUNDRED,
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
		require.Equalf(t, test.expected,
			s.buildRequestMirrorPolicies(test.route), test.name)
	}
}

func Test_buildCorsPolicy(t *testing.T) {

	stringMatch := make([]*envoymatcher.StringMatcher, 0, 1)
	defaultStringMatcher := append(stringMatch, buildStringMatcher("."))

	tests := []struct {
		name     string
		route    types.Route
		expected *route.CorsPolicy
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
			expected: &route.CorsPolicy{
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
			expected: &route.CorsPolicy{
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
			expected: &route.CorsPolicy{
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
			expected: &route.CorsPolicy{
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
			expected: &route.CorsPolicy{
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
		require.Equalf(t, test.expected,
			buildCorsPolicy(test.route), test.name)
	}
}

func Test_buildRouteActionDirectResponse(t *testing.T) {

	tests := []struct {
		name     string
		route    types.Route
		expected *route.Route_DirectResponse
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
			expected: &route.Route_DirectResponse{
				DirectResponse: &route.DirectResponseAction{
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
			expected: &route.Route_DirectResponse{
				DirectResponse: &route.DirectResponseAction{
					Status: uint32(255),
					Body: &core.DataSource{
						Specifier: &core.DataSource_InlineString{
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
		require.Equalf(t, test.expected,
			buildRouteActionDirectResponse(test.route), test.name)
	}
}

func Test_buildRouteActionRedirectResponse(t *testing.T) {

	s := newServerForTesting()

	tests := []struct {
		name     string
		route    types.Route
		expected *route.Route_Redirect
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
			expected: &route.Route_Redirect{
				Redirect: &route.RedirectAction{
					ResponseCode: route.RedirectAction_MOVED_PERMANENTLY,
					SchemeRewriteSpecifier: &route.RedirectAction_SchemeRedirect{
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
			expected: &route.Route_Redirect{
				Redirect: &route.RedirectAction{
					ResponseCode: route.RedirectAction_FOUND,
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
			expected: &route.Route_Redirect{
				Redirect: &route.RedirectAction{
					ResponseCode: route.RedirectAction_SEE_OTHER,
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
			expected: &route.Route_Redirect{
				Redirect: &route.RedirectAction{
					ResponseCode: route.RedirectAction_TEMPORARY_REDIRECT,
					PathRewriteSpecifier: &route.RedirectAction_PathRedirect{
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
			expected: &route.Route_Redirect{
				Redirect: &route.RedirectAction{
					ResponseCode: route.RedirectAction_PERMANENT_REDIRECT,
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
		require.Equalf(t, test.expected,
			s.buildRouteActionRedirectResponse(test.route), test.name)
	}
}

func Test_buildUpstreamHeadersToAdd(t *testing.T) {

	tests := []struct {
		name     string
		route    types.Route
		expected []*core.HeaderValueOption
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
			expected: []*core.HeaderValueOption{
				{
					Header: &core.HeaderValue{
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
			expected: []*core.HeaderValueOption{
				{
					Header: &core.HeaderValue{
						Key:   "Authorization",
						Value: "Basic dGVzdDoxMjM=",
					},
				},
				{
					Header: &core.HeaderValue{
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
		require.Equalf(t, test.expected,
			buildUpstreamHeadersToAdd(test.route), test.name)
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
		require.Equalf(t, test.expected,
			buildUpstreamHeadersToRemove(test.route), test.name)
	}
}

func Test_buildHeadersList(t *testing.T) {

	tests := []struct {
		name     string
		headers  map[string]string
		expected []*core.HeaderValueOption
	}{
		{
			name: "1 header",
			headers: map[string]string{
				"user-agent": "netscape",
			},
			expected: []*core.HeaderValueOption{
				{
					Header: &core.HeaderValue{
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
		require.Equalf(t, test.expected,
			buildHeadersList(test.headers), test.name)
	}
}

func Test_buildRateLimits(t *testing.T) {

}

func Test_buildRetryPolicy(t *testing.T) {

	tests := []struct {
		name     string
		route    types.Route
		expected *route.RetryPolicy
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
			expected: &route.RetryPolicy{
				RetryOn:              "5xx",
				NumRetries:           protoUint32(types.DefaultNumRetries),
				PerTryTimeout:        durationpb.New(types.DefaultPerRetryTimeout),
				RetriableStatusCodes: buildStatusCodesSlice(types.DefaultRetryStatusCodes),
				RetryHostPredicate: []*route.RetryPolicy_RetryHostPredicate{
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
			expected: &route.RetryPolicy{
				RetryOn:              "connect-failure,reset",
				NumRetries:           protoUint32(7),
				PerTryTimeout:        durationpb.New(21 * time.Millisecond),
				RetriableStatusCodes: buildStatusCodesSlice("506,507"),
				RetryHostPredicate: []*route.RetryPolicy_RetryHostPredicate{
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
		require.Equalf(t, test.expected,
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
		require.Equalf(t, test.expected,
			buildStatusCodesSlice(test.statusCodes), test.name)
	}
}

func Test_buildPerRouteFilterConfig(t *testing.T) {

	tests := []struct {
		name     string
		route    types.Route
		expected map[string]*any.Any
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
			expected: map[string]*any.Any{
				wellknown.HTTPExternalAuthorization: mustMarshalAny(&envoyauth.ExtAuthzPerRoute{
					Override: &envoyauth.ExtAuthzPerRoute_Disabled{
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
		require.Equalf(t, test.expected,
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
			expected: mustMarshalAny(&envoyauth.ExtAuthzPerRoute{
				Override: &envoyauth.ExtAuthzPerRoute_Disabled{
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
			expected: mustMarshalAny(&envoyauth.ExtAuthzPerRoute{
				Override: &envoyauth.ExtAuthzPerRoute_Disabled{
					Disabled: true,
				},
			}),
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			perRouteAuthzFilterConfig(test.route), test.name)
	}
}

func Test_buildEnvoyVirtualClusters(t *testing.T) {

	s := newServerForTesting()

	tests := []struct {
		name       string
		RouteGroup string
		routes     types.Routes
		expected   []*route.VirtualCluster
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
			expected: []*route.VirtualCluster{
				{
					Headers: []*route.HeaderMatcher{
						{
							Name: ":path",
							HeaderMatchSpecifier: &route.HeaderMatcher_ExactMatch{
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
		require.Equalf(t, test.expected,
			s.buildEnvoyVirtualClusters(test.RouteGroup, test.routes), test.name)
	}
}

func Test_buildEnvoyVirtualClusterPathMatch(t *testing.T) {

	tests := []struct {
		name     string
		route    types.Route
		expected *route.HeaderMatcher
	}{
		{
			name: "path match",
			route: types.Route{
				PathType: "path",
				Path:     "/customers",
			},
			expected: &route.HeaderMatcher{
				Name: ":path",
				HeaderMatchSpecifier: &route.HeaderMatcher_ExactMatch{
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
			expected: &route.HeaderMatcher{
				Name: ":path",
				HeaderMatchSpecifier: &route.HeaderMatcher_PrefixMatch{
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
			expected: &route.HeaderMatcher{
				Name: ":path",
				HeaderMatchSpecifier: &route.HeaderMatcher_SafeRegexMatch{
					SafeRegexMatch: &envoymatcher.RegexMatcher{
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
		require.Equalf(t, test.expected,
			buildEnvoyVirtualClusterPathMatch(test.route), test.name)
	}
}
