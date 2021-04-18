package main

import (
	"testing"
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoyauth "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	envoymatcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/erikbos/gatekeeper/pkg/types"
)

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
			expected: []string{},
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
		route      types.Routes
		expected   []*route.VirtualCluster
	}{
		{
			name:       "routegroup 1 match, 1 mismatch",
			RouteGroup: "bikes",
			route: types.Routes{
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
			route: types.Routes{
				{
					PathType: "unknown",
				},
			}, expected: nil,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			s.buildEnvoyVirtualClusters(test.RouteGroup, test.route), test.name)
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
