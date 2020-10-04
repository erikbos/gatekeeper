package main

import (
	"testing"
	"time"

	envoyCluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoyType "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/runtime/protoiface"

	"github.com/erikbos/gatekeeper/pkg/types"
)

func Test_buildEnvoyClusterConfig(t *testing.T) {

	tests := []struct {
		name     string
		cluster  types.Cluster
		expected *envoyCluster.Cluster
	}{
		{
			name: "Build cluster 1",
			cluster: types.Cluster{
				Name: "Example Backend",
				Attributes: types.Attributes{
					{
						Name:  types.AttributeHost,
						Value: "backend42.localhost",
					},
					{
						Name:  types.AttributePort,
						Value: "1975",
					},
				},
			},
			expected: &envoyCluster.Cluster{
				Name: "Example Backend",

				ConnectTimeout: ptypes.DurationProto(types.DefaultClusterConnectTimeout),

				ClusterDiscoveryType: &envoyCluster.Cluster_Type{
					Type: envoyCluster.Cluster_LOGICAL_DNS,
				},

				DnsLookupFamily: envoyCluster.Cluster_AUTO,

				DnsRefreshRate: ptypes.DurationProto(types.DefaultDNSRefreshRate),

				LbPolicy: envoyCluster.Cluster_ROUND_ROBIN,

				CircuitBreakers: &envoyCluster.CircuitBreakers{
					Thresholds: []*envoyCluster.CircuitBreakers_Thresholds{{
						MaxConnections:     protoUint32orNil(0),
						MaxPendingRequests: protoUint32orNil(0),
						MaxRequests:        protoUint32orNil(0),
						MaxRetries:         protoUint32orNil(0),
					}},
				},

				HealthChecks: nil,

				CommonHttpProtocolOptions: &core.HttpProtocolOptions{
					IdleTimeout: ptypes.DurationProto(types.DefaultClusterIdleTimeout),
				},

				TrackClusterStats: &envoyCluster.TrackClusterStats{
					TimeoutBudgets:       true,
					RequestResponseSizes: true,
				},

				LoadAssignment: &endpoint.ClusterLoadAssignment{
					ClusterName: "Example Backend",
					Endpoints: []*endpoint.LocalityLbEndpoints{
						{
							LbEndpoints: []*endpoint.LbEndpoint{
								{
									HostIdentifier: &endpoint.LbEndpoint_Endpoint{
										Endpoint: &endpoint.Endpoint{
											Address: &core.Address{
												Address: &core.Address_SocketAddress{
													SocketAddress: &core.SocketAddress{
														Address: "backend42.localhost",
														PortSpecifier: &core.SocketAddress_PortValue{
															PortValue: 1975,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Build cluster 2",
			cluster: types.Cluster{
				Name: "Example Backend",
				Attributes: types.Attributes{
					{
						Name:  types.AttributeHost,
						Value: "backend42.localhost",
					},
				},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			buildEnvoyClusterConfig(test.cluster), test.name)
	}
}

func Test_clusterConnectTimeout(t *testing.T) {

	tests := []struct {
		name     string
		cluster  types.Cluster
		expected *duration.Duration
	}{
		{
			name: "cluster timeout 131s ttl",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeConnectTimeout,
						Value: "131s",
					},
				},
			},
			expected: ptypes.DurationProto(131 * time.Second),
		},
		{
			name: "cluster timeout no ttl specified",
			cluster: types.Cluster{
				Attributes: types.Attributes{},
			},
			expected: ptypes.DurationProto(types.DefaultClusterConnectTimeout),
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			clusterConnectTimeout(test.cluster), test.name)
	}
}

func Test_clusterLbPolicy(t *testing.T) {

	tests := []struct {
		name     string
		cluster  types.Cluster
		expected envoyCluster.Cluster_LbPolicy
	}{
		{
			name: "lb policy roundrobin",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeLbPolicy,
						Value: types.AttributeValueLBRoundRobin,
					},
				},
			},
			expected: envoyCluster.Cluster_ROUND_ROBIN,
		},
		{
			name: "lb policy least req",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeLbPolicy,
						Value: types.AttributeValueLBLeastRequest,
					},
				},
			},
			expected: envoyCluster.Cluster_LEAST_REQUEST,
		},
		{
			name: "lb policy least ring hash",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeLbPolicy,
						Value: types.AttributeValueLBRingHash,
					},
				},
			},
			expected: envoyCluster.Cluster_RING_HASH,
		},
		{
			name: "lb policy least ring hash",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeLbPolicy,
						Value: types.AttributeValueLBRandom,
					},
				},
			},
			expected: envoyCluster.Cluster_RANDOM,
		},
		{
			name: "lb policy least ring hash",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeLbPolicy,
						Value: types.AttributeValueLBMaglev,
					},
				},
			},
			expected: envoyCluster.Cluster_MAGLEV,
		},
		{
			name: "lb policy not set",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{},
				},
			},
			expected: envoyCluster.Cluster_ROUND_ROBIN,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			clusterLbPolicy(test.cluster), test.name)
	}
}

func Test_clusterLoadAssignment(t *testing.T) {

	tests := []struct {
		name     string
		cluster  types.Cluster
		expected *endpoint.ClusterLoadAssignment
	}{
		{
			name: "Load assignment 1",
			cluster: types.Cluster{
				Name: "backend vips",
				Attributes: types.Attributes{
					{
						Name:  types.AttributeHost,
						Value: "backend",
					},
					{
						Name:  types.AttributePort,
						Value: "987",
					},
				},
			},
			expected: &endpoint.ClusterLoadAssignment{
				ClusterName: "backend vips",
				Endpoints: []*endpoint.LocalityLbEndpoints{
					{
						LbEndpoints: []*endpoint.LbEndpoint{
							{
								HostIdentifier: &endpoint.LbEndpoint_Endpoint{
									Endpoint: &endpoint.Endpoint{
										Address: &core.Address{
											Address: &core.Address_SocketAddress{
												SocketAddress: &core.SocketAddress{
													Address: "backend",
													PortSpecifier: &core.SocketAddress_PortValue{
														PortValue: 987,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Load assignment 2 (port missing)",
			cluster: types.Cluster{
				Name: "backend vips",
				Attributes: types.Attributes{
					{
						Name:  types.AttributeHost,
						Value: "backend52",
					},
				},
			},
			expected: nil,
		},
		{
			name: "Load assignment 3 (host missing)",
			cluster: types.Cluster{
				Name: "backend vips",
				Attributes: types.Attributes{
					{
						Name:  types.AttributePort,
						Value: "987",
					},
				},
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		require.Equalf(t, test.expected,
			clusterLoadAssignment(test.cluster), test.name)
	}
}

func Test_clusterCircuitBreakers(t *testing.T) {

	tests := []struct {
		name     string
		cluster  types.Cluster
		expected *envoyCluster.CircuitBreakers
	}{
		{
			name: "Circuit breakers 1",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeMaxConnections,
						Value: "",
					},
					{
						Name:  types.AttributeMaxPendingRequests,
						Value: "190",
					},
					{
						Name:  types.AttributeMaxRequests,
						Value: "",
					},
					{
						Name:  types.AttributeMaxRetries,
						Value: "736",
					},
				},
			},
			expected: &envoyCluster.CircuitBreakers{
				Thresholds: []*envoyCluster.CircuitBreakers_Thresholds{{
					MaxConnections:     protoUint32orNil(0),
					MaxPendingRequests: protoUint32orNil(190),
					MaxRequests:        protoUint32orNil(0),
					MaxRetries:         protoUint32orNil(736),
				}},
			},
		},
		{
			name: "Circuit breakers 2",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeMaxConnections,
						Value: "5",
					},
					{
						Name:  types.AttributeMaxPendingRequests,
						Value: "",
					},
					{
						Name:  types.AttributeMaxRequests,
						Value: "912",
					},
					{
						Name:  types.AttributeMaxRetries,
						Value: "",
					},
				},
			},
			expected: &envoyCluster.CircuitBreakers{
				Thresholds: []*envoyCluster.CircuitBreakers_Thresholds{{
					MaxConnections:     protoUint32orNil(5),
					MaxPendingRequests: protoUint32orNil(0),
					MaxRequests:        protoUint32orNil(912),
					MaxRetries:         protoUint32orNil(0),
				}},
			},
		},
		{
			name: "Circuit breakers 3, nil",
			cluster: types.Cluster{
				Attributes: types.Attributes{},
			},
			expected: &envoyCluster.CircuitBreakers{
				Thresholds: []*envoyCluster.CircuitBreakers_Thresholds{{}},
			},
		},
	}

	for _, test := range tests {
		require.Equalf(t, test.expected,
			clusterCircuitBreakers(test.cluster), test.name)
	}
}

func Test_clusterHealthChecks(t *testing.T) {

	tests := []struct {
		name     string
		cluster  types.Cluster
		expected []*core.HealthCheck
	}{
		{
			name: "Health check 1",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeHealthCheckProtocol,
						Value: types.AttributeValueHealthCheckProtocolHTTP,
					},
					{
						Name:  types.AttributeHTTPProtocol,
						Value: types.AttributeValueHTTPProtocol2,
					},
					{
						Name:  types.AttributeHealthCheckPath,
						Value: "/checkpoint",
					},
					{
						Name:  types.AttributeHost,
						Value: "www.testsite.com",
					},
					{
						Name:  types.AttributeHealthHostHeader,
						Value: "www.important.com",
					},
					{
						Name:  types.AttributeHealthCheckInterval,
						Value: "33s",
					},
					{
						Name:  types.AttributeHealthCheckTimeout,
						Value: "160ms",
					},
					{
						Name:  types.AttributeHealthCheckUnhealthyThreshold,
						Value: "7",
					},
					{
						Name:  types.AttributeHealthCheckHealthyThreshold,
						Value: "2",
					},
					{
						Name:  types.AttributeHealthCheckLogFile,
						Value: "/tmp/logfile_for_web",
					},
				},
			},
			expected: []*core.HealthCheck{
				{
					HealthChecker: &core.HealthCheck_HttpHealthCheck_{
						HttpHealthCheck: &core.HealthCheck_HttpHealthCheck{
							Host:            "www.important.com",
							Path:            "/checkpoint",
							CodecClientType: envoyType.CodecClientType_HTTP2,
						},
					},
					Interval:           ptypes.DurationProto(33 * time.Second),
					Timeout:            ptypes.DurationProto(160 * time.Millisecond),
					UnhealthyThreshold: protoUint32orNil(7),
					HealthyThreshold:   protoUint32orNil(2),
					EventLogPath:       "/tmp/logfile_for_web",
				},
			},
		},
	}

	for _, test := range tests {
		require.Equalf(t, test.expected,
			clusterHealthChecks(test.cluster), test.name)
	}
}

func Test_clusterHealthCodec(t *testing.T) {

	tests := []struct {
		name     string
		cluster  types.Cluster
		expected envoyType.CodecClientType
	}{
		{
			name: "HTTP11 codec",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeHTTPProtocol,
						Value: types.AttributeValueHTTPProtocol11,
					},
				},
			},
			expected: envoyType.CodecClientType_HTTP1,
		},
		{
			name: "HTTP2 codec",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeHTTPProtocol,
						Value: types.AttributeValueHTTPProtocol2,
					},
				},
			},
			expected: envoyType.CodecClientType_HTTP2,
		},
		{
			name: "HTTP3 codec",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeHTTPProtocol,
						Value: types.AttributeValueHTTPProtocol3,
					},
				},
			},
			expected: envoyType.CodecClientType_HTTP3,
		},
		{
			name: "unknown/default protocol",
			cluster: types.Cluster{
				Attributes: types.Attributes{},
			},
			expected: envoyType.CodecClientType_HTTP1,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			clusterHealthCodec(test.cluster), test.name)
	}
}

func Test_clusterCommonHTTPProtocolOptions(t *testing.T) {

	tests := []struct {
		name     string
		cluster  types.Cluster
		expected *core.HttpProtocolOptions
	}{
		{
			name: "Specific timeout",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeIdleTimeout,
						Value: "42ms",
					},
				},
			},
			expected: &core.HttpProtocolOptions{
				IdleTimeout: ptypes.DurationProto(42 * time.Millisecond),
			},
		},
		{
			name: "Default timeout",
			cluster: types.Cluster{
				Attributes: types.Attributes{},
			},
			expected: &core.HttpProtocolOptions{
				IdleTimeout: ptypes.DurationProto(types.DefaultClusterIdleTimeout),
			},
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			clusterCommonHTTPProtocolOptions(test.cluster), test.name)
	}
}

func Test_clusterHTTP2ProtocolOptions(t *testing.T) {

	tests := []struct {
		name     string
		cluster  types.Cluster
		expected *core.Http2ProtocolOptions
	}{
		{
			name: "HTTP11 protocol",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeHTTPProtocol,
						Value: types.AttributeValueHTTPProtocol11,
					},
				},
			},
			expected: nil,
		},
		{
			name: "HTTP2 protocol",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeHTTPProtocol,
						Value: types.AttributeValueHTTPProtocol2,
					},
				},
			},
			expected: &core.Http2ProtocolOptions{},
		},
		{
			name: "unknown protocol",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeHTTPProtocol,
						Value: "blabla",
					},
				},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			clusterHTTP2ProtocolOptions(test.cluster), test.name)
	}
}

func Test_clusterTransportSocket(t *testing.T) {

	tests := []struct {
		name       string
		cluster    types.Cluster
		tlsContext protoiface.MessageV1
		expected   *core.TransportSocket
	}{
		{
			name: "1",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeTLSMinimumVersion,
						Value: types.AttributeValueTLSVersion12,
					},
					{
						Name:  types.AttributeSNIHostName,
						Value: "www.sni-hostname.com",
					},
				},
			},
			expected: &core.TransportSocket{
				Name: "tls",
				ConfigType: &core.TransportSocket_TypedConfig{
					TypedConfig: mustMarshalAny(&tls.UpstreamTlsContext{
						Sni: "www.sni-hostname.com",
						CommonTlsContext: buildCommonTLSContext("www.example.com",
							types.Attributes{
								{
									Name:  types.AttributeTLSMinimumVersion,
									Value: types.AttributeValueTLSVersion12,
								},
								{
									Name:  types.AttributeSNIHostName,
									Value: "www.example.com",
								},
							}),
					}),
				},
			},
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			clusterTransportSocket(test.cluster), test.name)
	}
}

func Test_clusterSNIHostname(t *testing.T) {

	tests := []struct {
		name     string
		cluster  types.Cluster
		expected string
	}{
		{
			name: "SNI specified",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeSNIHostName,
						Value: "www.example.com",
					},
				},
			},
			expected: "www.example.com",
		},
		{
			name: "No SNI, host specified",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeHost,
						Value: "www.example5.com",
					},
				},
			},
			expected: "www.example5.com",
		},
		{
			name: "SNI specified",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeSNIHostName,
						Value: "www.example.com",
					},
					{
						Name:  types.AttributeHost,
						Value: "www.example5.com",
					},
				},
			},
			expected: "www.example.com",
		},
		{
			name: "No SNI or host specified",
			cluster: types.Cluster{
				Attributes: types.Attributes{},
			},
			expected: "",
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			clusterSNIHostname(test.cluster), test.name)
	}
}

func Test_clusterDNSLookupFamily(t *testing.T) {

	tests := []struct {
		name     string
		cluster  types.Cluster
		expected envoyCluster.Cluster_DnsLookupFamily
	}{
		{
			name: "ipv4",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeDNSLookupFamiliy,
						Value: types.AttributeValueDNSIPV4Only,
					},
				},
			},
			expected: envoyCluster.Cluster_V4_ONLY,
		},
		{
			name: "ipv6",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeDNSLookupFamiliy,
						Value: types.AttributeValueDNSIPV6Only,
					},
				},
			},
			expected: envoyCluster.Cluster_V6_ONLY,
		},
		{
			name: "auto",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeDNSLookupFamiliy,
						Value: types.AttributeValueDNSAUTO,
					},
				},
			},
			expected: envoyCluster.Cluster_AUTO,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			clusterDNSLookupFamily(test.cluster), test.name)
	}
}

func Test_clusterDNSRefreshRate(t *testing.T) {

	tests := []struct {
		name     string
		cluster  types.Cluster
		expected *duration.Duration
	}{
		{
			name: "31s ttl",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeDNSRefreshRate,
						Value: "31s",
					},
				},
			},
			expected: ptypes.DurationProto(31 * time.Second),
		},
		{
			name: "no ttl specified",
			cluster: types.Cluster{
				Attributes: types.Attributes{},
			},
			expected: ptypes.DurationProto(types.DefaultDNSRefreshRate),
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			clusterDNSRefreshRate(test.cluster), test.name)
	}
}

func Test_clusterDNSResolvers(t *testing.T) {

	tests := []struct {
		name     string
		cluster  types.Cluster
		expected []*core.Address
	}{
		{
			name: "one resolver",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeDNSResolvers,
						Value: "8.8.8.8",
					},
				},
			},
			expected: []*core.Address{
				{
					Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Address: "8.8.8.8",
							PortSpecifier: &core.SocketAddress_PortValue{
								PortValue: 53,
							},
						},
					},
				},
			},
		},
		{
			name: "two resolvers",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeDNSResolvers,
						Value: "1.1.1.1,1.2.3.4",
					},
				},
			},
			expected: []*core.Address{
				{
					Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Address: "1.1.1.1",
							PortSpecifier: &core.SocketAddress_PortValue{
								PortValue: 53,
							},
						},
					},
				},
				{
					Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Address: "1.2.3.4",
							PortSpecifier: &core.SocketAddress_PortValue{
								PortValue: 53,
							},
						},
					},
				},
			},
		},
		{
			name: "no resolver",
			cluster: types.Cluster{
				Attributes: types.Attributes{
					{
						Name:  types.AttributeDNSResolvers,
						Value: "",
					},
				},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			clusterDNSResolvers(test.cluster), test.name)
	}
}
