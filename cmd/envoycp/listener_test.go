package main

import (
	"testing"
	"time"

	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	ratelimitconf "github.com/envoyproxy/go-control-plane/envoy/config/ratelimit/v3"
	fileaccesslog "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	grpcaccesslog "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
	extauthz "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	ratelimit "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ratelimit/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erikbos/gatekeeper/pkg/types"
)

func Test_buildConnectionManager(t *testing.T) {

	s := server{}
	listener1 := types.Listener{
		Attributes: types.Attributes{
			{
				Name:  types.AttributeServerName,
				Value: "QWERTY",
			},
		},
	}
	expected1 := &hcm.HttpConnectionManager{
		CodecType:                 hcm.HttpConnectionManager_AUTO,
		StatPrefix:                "ingress_http",
		UseRemoteAddress:          protoBool(true),
		HttpFilters:               s.buildFilter(listener1),
		RouteSpecifier:            s.buildRouteSpecifierRDS(listener1.RouteGroup),
		AccessLog:                 s.buildAccessLog(listener1),
		CommonHttpProtocolOptions: listenerCommonHTTPProtocolOptions(listener1),
		Http2ProtocolOptions:      buildHTTP2ProtocolOptions(listener1),
		ServerName:                "QWERTY",
	}

	require.Equalf(t, expected1,
		s.buildConnectionManager(listener1), "test1")
}

func Test_buildFilter(t *testing.T) {

	tests := []struct {
		name     string
		s        server
		listener types.Listener
		expected []*hcm.HttpFilter
	}{
		{
			name: "BuildAuthz 1 (authz enabled, no CORS)",
			listener: types.Listener{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeAuthentication,
						Value: types.AttributeValueTrue,
					},
					{
						Name:  types.AttributeAuthenticationRequestBodySize,
						Value: "3000",
					},
					{
						Name:  types.AttributeAuthenticationFailureModeAllow,
						Value: "true",
					},
					{
						Name:  types.AttributeAuthenticationCluster,
						Value: "authz_cluster2",
					},
					{
						Name:  types.AttributeAuthenticationTimeout,
						Value: "48s",
					},
				},
			},
			s: server{},
			expected: []*hcm.HttpFilter{
				{
					Name: wellknown.HTTPExternalAuthorization,
					ConfigType: &hcm.HttpFilter_TypedConfig{
						TypedConfig: mustMarshalAny(&extauthz.ExtAuthz{
							FailureModeAllow: true,
							Services: &extauthz.ExtAuthz_GrpcService{
								GrpcService: buildGRPCService("authz_cluster2",
									48*time.Second),
							},
							WithRequestBody: &extauthz.BufferSettings{
								MaxRequestBytes:     uint32(3000),
								AllowPartialMessage: false,
							},
							TransportApiVersion: core.ApiVersion_V3,
						}),
					},
				},
				{
					Name: wellknown.Router,
				},
			},
		},
		{
			name: "BuildAuthz 2 (authz & CORS enabled)",
			listener: types.Listener{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeAuthentication,
						Value: types.AttributeValueTrue,
					},
					{
						Name:  types.AttributeAuthenticationRequestBodySize,
						Value: "3000",
					},
					{
						Name:  types.AttributeAuthenticationFailureModeAllow,
						Value: "true",
					},
					{
						Name:  types.AttributeAuthenticationCluster,
						Value: "authz_cluster2",
					},
					{
						Name:  types.AttributeAuthenticationTimeout,
						Value: "48s",
					},
					{
						Name:  types.AttributeCORS,
						Value: types.AttributeValueTrue,
					},
				},
			},
			s: server{},
			expected: []*hcm.HttpFilter{
				{
					Name: wellknown.HTTPExternalAuthorization,
					ConfigType: &hcm.HttpFilter_TypedConfig{
						TypedConfig: mustMarshalAny(&extauthz.ExtAuthz{
							FailureModeAllow: true,
							Services: &extauthz.ExtAuthz_GrpcService{
								GrpcService: buildGRPCService("authz_cluster2",
									48*time.Second),
							},
							WithRequestBody: &extauthz.BufferSettings{
								MaxRequestBytes:     uint32(3000),
								AllowPartialMessage: false,
							},
							TransportApiVersion: core.ApiVersion_V3,
						}),
					},
				},
				{
					Name: wellknown.CORS,
				},
				{
					Name: wellknown.Router,
				},
			},
		},
		{
			name: "BuildAuthz 3 (only CORS enabled)",
			listener: types.Listener{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeCORS,
						Value: types.AttributeValueTrue,
					},
				},
			},
			s: server{},
			expected: []*hcm.HttpFilter{
				{
					Name: wellknown.CORS,
				},
				{
					Name: wellknown.Router,
				},
			},
		},
		{
			name: "BuildAuthz 4 (ratelimiter enabled)",
			listener: types.Listener{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRateLimiting,
						Value: types.AttributeValueTrue,
					},
					{
						Name:  types.AttributeRateLimitingDomain,
						Value: "ebnl",
					},
					{
						Name:  types.AttributeRateLimitingFailureModeAllow,
						Value: "true",
					},
					{
						Name:  types.AttributeRateLimitingCluster,
						Value: "ratelimiter_node",
					},
					{
						Name:  types.AttributeRateLimitingTimeout,
						Value: "64ms",
					},
				},
			},
			s: server{},
			expected: []*hcm.HttpFilter{
				{
					Name: wellknown.HTTPRateLimit,
					ConfigType: &hcm.HttpFilter_TypedConfig{
						TypedConfig: mustMarshalAny(&ratelimit.RateLimit{
							Domain:          "ebnl",
							Stage:           0,
							FailureModeDeny: true,
							Timeout:         durationpb.New(64 * time.Millisecond),
							RateLimitService: &ratelimitconf.RateLimitServiceConfig{
								GrpcService:         buildGRPCService("ratelimiter_node", 64*time.Millisecond),
								TransportApiVersion: core.ApiVersion_V3,
							},
						}),
					},
				},
				{
					Name: wellknown.Router,
				},
			},
		},
		{
			name: "BuildAuthz 5 (ratelimiter enabled, default timeout)",
			listener: types.Listener{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeRateLimiting,
						Value: types.AttributeValueTrue,
					},
					{
						Name:  types.AttributeRateLimitingDomain,
						Value: "ebnl",
					},
					{
						Name:  types.AttributeRateLimitingFailureModeAllow,
						Value: "true",
					},
					{
						Name:  types.AttributeRateLimitingCluster,
						Value: "ratelimiter_node",
					},
				},
			},
			s: server{},
			expected: []*hcm.HttpFilter{
				{
					Name: wellknown.HTTPRateLimit,
					ConfigType: &hcm.HttpFilter_TypedConfig{
						TypedConfig: mustMarshalAny(&ratelimit.RateLimit{
							Domain:          "ebnl",
							Stage:           0,
							FailureModeDeny: true,
							Timeout:         durationpb.New(defaultRateLimitingTimeout),
							RateLimitService: &ratelimitconf.RateLimitServiceConfig{
								GrpcService:         buildGRPCService("ratelimiter_node", defaultRateLimitingTimeout),
								TransportApiVersion: core.ApiVersion_V3,
							},
						}),
					},
				},
				{
					Name: wellknown.Router,
				},
			},
		},
		{
			name:     "BuildAuthz 9 (no specific filters)",
			listener: types.Listener{},
			s:        server{},
			expected: []*hcm.HttpFilter{
				{
					Name: wellknown.Router,
				},
			},
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			test.s.buildFilter(test.listener), test.name)
	}
}

func Test_buildHTTPFilterExtAuthzConfig(t *testing.T) {

	tests := []struct {
		name     string
		s        server
		listener types.Listener
		expected *anypb.Any
	}{
		{
			name: "BuildAuthz 1",
			listener: types.Listener{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeAuthentication,
						Value: types.AttributeValueTrue,
					},
					{
						Name:  types.AttributeAuthenticationRequestBodySize,
						Value: "6000",
					},
					{
						Name:  types.AttributeAuthenticationFailureModeAllow,
						Value: "true",
					},
					{
						Name:  types.AttributeAuthenticationCluster,
						Value: "authz_cluster",
					},
					{
						Name:  types.AttributeAuthenticationTimeout,
						Value: "24s",
					},
				},
			},
			expected: mustMarshalAny(&extauthz.ExtAuthz{
				FailureModeAllow: true,
				Services: &extauthz.ExtAuthz_GrpcService{
					GrpcService: buildGRPCService("authz_cluster",
						24*time.Second),
				},
				WithRequestBody: &extauthz.BufferSettings{
					MaxRequestBytes:     uint32(6000),
					AllowPartialMessage: false,
				},
				TransportApiVersion: core.ApiVersion_V3,
			}),
		},
		{
			name: "BuildAuthz 2 (not enabled)",
			listener: types.Listener{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeAuthenticationCluster,
						Value: "authz_cluster",
					},
					{
						Name:  types.AttributeAuthenticationTimeout,
						Value: "25s",
					},
				},
			},
			expected: nil,
		},
		{
			name: "BuildAuthz 3 (no cluster)",
			listener: types.Listener{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeAuthentication,
						Value: types.AttributeValueTrue,
					},
					{
						Name:  types.AttributeAuthenticationTimeout,
						Value: "25s",
					},
				},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			test.s.buildHTTPFilterExtAuthzConfig(test.listener), test.name)
	}
}

func Test_buildRouteSpecifierRDS(t *testing.T) {

	tests := []struct {
		name       string
		routeGroup string
		s          server
		expected   *hcm.HttpConnectionManager_Rds
	}{
		{
			name:       "RouteSpecificer RDS 1",
			routeGroup: "routes_747",
			s: server{
				config: &EnvoyCPConfig{
					XDS: xdsConfig{
						Cluster: "rds_cluster",
						Timeout: 12 * time.Second,
					},
				},
			},
			expected: &hcm.HttpConnectionManager_Rds{
				Rds: &hcm.Rds{
					RouteConfigName: "routes_747",
					ConfigSource: buildConfigSource("rds_cluster",
						12*time.Second),
				},
			},
		},
		{
			name:       "RouteSpecificer RDS 2 (no cluster)",
			routeGroup: "routes_747",
			s: server{
				config: &EnvoyCPConfig{
					XDS: xdsConfig{
						Cluster: "",
						Timeout: 12 * time.Second,
					},
				},
			},
			expected: nil,
		},
		{
			name:       "RouteSpecificer RDS 3 (no route group)",
			routeGroup: "",
			s: server{
				config: &EnvoyCPConfig{
					XDS: xdsConfig{
						Cluster: "rds_source",
						Timeout: 12 * time.Second,
					},
				},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			test.s.buildRouteSpecifierRDS(test.routeGroup), test.name)
	}
}

func Test_buildAccessLog(t *testing.T) {

	s := server{}

	tests := []struct {
		name     string
		listener types.Listener
		expected []*accesslog.AccessLog
	}{
		{
			name: "accesslog 1",
			listener: types.Listener{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeAccessLogFile,
						Value: "/dev/log/myaccesslog3",
					},
					{
						Name:  types.AttributeAccessLogFileFields,
						Value: "request_id=%REQ(REQUEST-ID)%,key2=value2, key3 =VALUE3",
					},
				},
			},
			expected: []*accesslog.AccessLog{
				s.buildFileAccessLog("/dev/log/myaccesslog3",
					"request_id=%REQ(REQUEST-ID)%,key2=value2, key3 =VALUE3"),
			},
		},
		{
			name: "accesslog 2",
			listener: types.Listener{
				Name: "www.example.com",
				Attributes: types.Attributes{
					{
						Name:  types.AttributeAccessLogCluster,
						Value: "als_c",
					},
					{
						Name:  types.AttributeAccessLogClusterBufferSize,
						Value: "1024000",
					},
				},
			},
			expected: []*accesslog.AccessLog{
				s.buildGRPCAccessLog("als_c", "www.example.com",
					types.DefaultClusterConnectTimeout, 1024000),
			},
		},
	}
	for _, test := range tests {
		RequireEqual(t, test.expected,
			s.buildAccessLog(test.listener))
		// require.Equalf(t, test.expected,
		// 	buildAccessLog(test.listener), test.name)
	}
}

func Test_buildFileAccessLog(t *testing.T) {

	s := server{}

	tests := []struct {
		name     string
		path     string
		fields   string
		expected *accesslog.AccessLog
	}{
		{
			name:   "file accesslog 1",
			path:   "/var/log/access2.log",
			fields: "request_id=%REQ(REQUEST-ID)%,key2=value2, key3 =VALUE3",

			expected: &accesslog.AccessLog{
				Name: wellknown.FileAccessLog,
				ConfigType: &accesslog.AccessLog_TypedConfig{
					TypedConfig: mustMarshalAny(&fileaccesslog.FileAccessLog{
						Path: "/var/log/access2.log",
						AccessLogFormat: &fileaccesslog.FileAccessLog_LogFormat{
							LogFormat: &core.SubstitutionFormatString{
								Format: &core.SubstitutionFormatString_JsonFormat{
									JsonFormat: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"request_id": {
												Kind: &structpb.Value_StringValue{
													StringValue: "%REQ(REQUEST-ID)%",
												},
											},
											"key2": {
												Kind: &structpb.Value_StringValue{
													StringValue: "value2",
												},
											},
											"key3": {
												Kind: &structpb.Value_StringValue{
													StringValue: "VALUE3",
												},
											},
										},
									},
								},
							},
						},
					}),
				},
			},
		},
	}
	for _, test := range tests {
		RequireEqual(t, test.expected,
			s.buildFileAccessLog(test.path, test.fields))
	}
}

func Test_buildGRPCAccessLog(t *testing.T) {

	s := server{}

	tests := []struct {
		name        string
		clusterName string
		logName     string
		timeout     time.Duration
		bufferSize  uint32
		expected    *accesslog.AccessLog
	}{
		{
			name:        "Specific timeout",
			clusterName: "accesslogserver",
			logName:     "www",
			timeout:     types.DefaultClusterConnectTimeout,
			bufferSize:  10240,
			expected: &accesslog.AccessLog{
				Name: wellknown.HTTPGRPCAccessLog,
				ConfigType: &accesslog.AccessLog_TypedConfig{
					TypedConfig: mustMarshalAny(&grpcaccesslog.HttpGrpcAccessLogConfig{
						CommonConfig: &grpcaccesslog.CommonGrpcAccessLogConfig{
							LogName: "www",
							GrpcService: buildGRPCService(
								"accesslogserver", types.DefaultClusterConnectTimeout),
							TransportApiVersion: core.ApiVersion_V3,
							BufferSizeBytes:     protoUint32orNil(10240),
						},
					},
					)},
			},
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			s.buildGRPCAccessLog(test.clusterName, test.logName,
				test.timeout, test.bufferSize), test.name)
	}
}

func Test_buildCommonHTTPProtocolOptions(t *testing.T) {

	tests := []struct {
		name     string
		listener types.Listener
		expected *core.HttpProtocolOptions
	}{
		{
			name: "Specific timeout",
			listener: types.Listener{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeIdleTimeout,
						Value: "266ms",
					},
				},
			},
			expected: &core.HttpProtocolOptions{
				IdleTimeout: durationpb.New(266 * time.Millisecond),
			},
		},
		{
			name: "Default timeout",
			listener: types.Listener{
				Attributes: types.Attributes{},
			},
			expected: &core.HttpProtocolOptions{
				IdleTimeout: durationpb.New(listenerIdleTimeout),
			},
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			listenerCommonHTTPProtocolOptions(test.listener), test.name)
	}
}

func Test_buildHTTP2ProtocolOptions(t *testing.T) {

	tests := []struct {
		name     string
		listener types.Listener
		expected *core.Http2ProtocolOptions
	}{
		{
			name: "HTTP2 options 1",
			listener: types.Listener{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeMaxConcurrentStreams,
						Value: "100",
					},
					{
						Name:  types.AttributeInitialConnectionWindowSize,
						Value: "190",
					},
					{
						Name:  types.AttributeInitialStreamWindowSize,
						Value: "65536",
					},
				},
			},
			expected: &core.Http2ProtocolOptions{
				MaxConcurrentStreams:        protoUint32orNil(100),
				InitialConnectionWindowSize: protoUint32orNil(190),
				InitialStreamWindowSize:     protoUint32orNil(65536),
			},
		},
		{
			name: "HTTP2 options 2, nil",
			listener: types.Listener{
				Attributes: types.Attributes{},
			},
			expected: &core.Http2ProtocolOptions{
				MaxConcurrentStreams:        protoUint32orNil(0),
				InitialConnectionWindowSize: protoUint32orNil(0),
				InitialStreamWindowSize:     protoUint32orNil(0),
			},
		},
	}

	for _, test := range tests {
		require.Equalf(t, test.expected,
			buildHTTP2ProtocolOptions(test.listener), test.name)
	}
}
