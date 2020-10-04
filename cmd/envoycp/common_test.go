package main

import (
	"testing"
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/erikbos/gatekeeper/pkg/types"
)

func Test_buildAddress(t *testing.T) {

	tests := []struct {
		name     string
		hostname string
		port     uint32
		expected *core.Address
	}{
		{
			name:     "1",
			hostname: "localhost",
			port:     80,
			expected: &core.Address{
				Address: &core.Address_SocketAddress{
					SocketAddress: &core.SocketAddress{
						Address:  "localhost",
						Protocol: core.SocketAddress_TCP,
						PortSpecifier: &core.SocketAddress_PortValue{
							PortValue: 80,
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			buildAddress(test.hostname, test.port), test.name)
	}
}

func Test_buildGRPCService(t *testing.T) {

	tests := []struct {
		name        string
		clusterName string
		timeout     time.Duration
		expected    *core.GrpcService
	}{
		{
			name:        "1",
			clusterName: "testCluster",
			timeout:     82 * time.Second,
			expected: &core.GrpcService{
				Timeout: ptypes.DurationProto(82 * time.Second),
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
						ClusterName: "testCluster",
					},
				},
			},
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			buildGRPCService(test.clusterName, test.timeout), test.name)
	}
}

func Test_buildConfigSource(t *testing.T) {

	tests := []struct {
		name        string
		clusterName string
		timeout     time.Duration
		expected    *core.ConfigSource
	}{
		{
			name:        "1",
			clusterName: "testbackend5",
			timeout:     82 * time.Second,
			expected: &core.ConfigSource{
				ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
					ApiConfigSource: &core.ApiConfigSource{
						ApiType: core.ApiConfigSource_GRPC,
						GrpcServices: []*core.GrpcService{
							{
								Timeout: ptypes.DurationProto(82 * time.Second),
								TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
									EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
										ClusterName: "testbackend5",
									},
								},
							},
						},
						TransportApiVersion: core.ApiVersion_V3,
					},
				},
				ResourceApiVersion: core.ApiVersion_V3,
			},
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			buildConfigSource(test.clusterName, test.timeout), test.name)
	}
}

func Test_buildTransportSocket(t *testing.T) {

	tests := []struct {
		name         string
		resourceName string
		tlsContext   protoiface.MessageV1
		expected     *core.TransportSocket
	}{
		{
			name:         "1",
			resourceName: "cluster",
			tlsContext: &tls.UpstreamTlsContext{
				Sni: "cluster5",
				CommonTlsContext: buildCommonTLSContext("www.example.com",
					types.Attributes{
						{
							Name:  types.AttributeTLSMinimumVersion,
							Value: types.AttributeValueTLSVersion12,
						},
						{
							Name:  types.AttributeTLSMaximumVersion,
							Value: types.AttributeValueTLSVersion13,
						},
					}),
			},
			expected: &core.TransportSocket{
				Name: "tls",
				ConfigType: &core.TransportSocket_TypedConfig{
					TypedConfig: mustMarshalAny(&tls.UpstreamTlsContext{
						Sni: "cluster5",
						CommonTlsContext: buildCommonTLSContext("www.example.com",
							types.Attributes{
								{
									Name:  types.AttributeTLSMinimumVersion,
									Value: types.AttributeValueTLSVersion12,
								},
								{
									Name:  types.AttributeTLSMaximumVersion,
									Value: types.AttributeValueTLSVersion13,
								},
							}),
					}),
				},
			},
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			buildTransportSocket(test.resourceName, test.tlsContext), test.name)
	}
}

func Test_buildCommonTLSContext(t *testing.T) {

	tests := []struct {
		name         string
		resourceName string
		attributes   types.Attributes
		expected     *tls.CommonTlsContext
	}{
		{
			name:         "1",
			resourceName: "example.com",
			attributes: types.Attributes{
				{
					Name:  types.AttributeTLSMaximumVersion,
					Value: types.AttributeValueTLSVersion11,
				},
			},
			expected: &tls.CommonTlsContext{
				AlpnProtocols: buildALPNProtocols("example.com", types.Attributes{
					{
						Name:  types.AttributeTLSMaximumVersion,
						Value: types.AttributeValueTLSVersion11,
					},
				}),
				TlsParams: buildTLSParameters(types.Attributes{
					{
						Name:  types.AttributeTLSMaximumVersion,
						Value: types.AttributeValueTLSVersion11},
				}),
				TlsCertificates: buildTLSCertificates(types.Attributes{
					{
						Name:  types.AttributeTLSMaximumVersion,
						Value: types.AttributeValueTLSVersion11},
				}),
			},
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			buildCommonTLSContext(test.resourceName, test.attributes), test.name)
	}
}

func Test_buildTLSCipherSuites(t *testing.T) {

	tests := []struct {
		name       string
		attributes types.Attributes
		expected   []string
	}{
		{
			name: "One ciphers",
			attributes: types.Attributes{
				{
					Name:  types.AttributeTLSCipherSuites,
					Value: "ECDHE-RSA-AES128-GCM-SHA256",
				},
			},
			expected: []string{"ECDHE-RSA-AES128-GCM-SHA256"},
		},
		{
			name: "Two ciphers",
			attributes: types.Attributes{
				{Name: types.AttributeTLSCipherSuites, Value: "a,b"},
			},
			expected: []string{"a", "b"},
		},
		{
			name: "Empty ciphers",
			attributes: types.Attributes{
				{Name: types.AttributeTLSCipherSuites, Value: ""},
			},
			expected: nil,
		},
		{
			name: "No ciphers",
			attributes: types.Attributes{
				{Name: "", Value: ""},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			buildTLSCipherSuites(test.attributes), test.name)
	}
}

func Test_buildTLSParameters(t *testing.T) {

	tests := []struct {
		name       string
		attributes types.Attributes
		expected   *tls.TlsParameters
	}{
		{
			name: "One cipher + TLS1.0",
			attributes: types.Attributes{
				{
					Name:  types.AttributeTLSCipherSuites,
					Value: "ECDHE-RSA-AES128-GCM-SHA256"},
				{
					Name:  types.AttributeTLSMinimumVersion,
					Value: types.AttributeValueTLSVersion10,
				},
			},
			expected: &tls.TlsParameters{
				CipherSuites:              []string{"ECDHE-RSA-AES128-GCM-SHA256"},
				TlsMinimumProtocolVersion: tls.TlsParameters_TLSv1_0,
			},
		},
		{
			name: "Max TLS1.3",
			attributes: types.Attributes{
				{
					Name:  types.AttributeTLSCipherSuites,
					Value: "ECDHE-RSA-CHACHA20-POLY1305",
				},
				{
					Name:  types.AttributeTLSMaximumVersion,
					Value: types.AttributeValueTLSVersion13,
				},
			},
			expected: &tls.TlsParameters{
				CipherSuites:              []string{"ECDHE-RSA-CHACHA20-POLY1305"},
				TlsMaximumProtocolVersion: tls.TlsParameters_TLSv1_3,
			},
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			buildTLSParameters(test.attributes), test.name)
	}
}

func Test_buildTLSVersion(t *testing.T) {

	tests := []struct {
		name     string
		version  string
		expected tls.TlsParameters_TlsProtocol
	}{
		{
			name:     "10",
			version:  types.AttributeValueTLSVersion10,
			expected: tls.TlsParameters_TLSv1_0,
		},
		{
			name:     "11",
			version:  types.AttributeValueTLSVersion11,
			expected: tls.TlsParameters_TLSv1_1,
		},
		{
			name:     "12",
			version:  types.AttributeValueTLSVersion12,
			expected: tls.TlsParameters_TLSv1_2,
		},
		{
			name:     "13",
			version:  types.AttributeValueTLSVersion11,
			expected: tls.TlsParameters_TLSv1_1,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			buildTLSVersion(test.version), test.name)
	}
}

func Test_buildALPNProtocols(t *testing.T) {

	tests := []struct {
		name       string
		entity     string
		attributes types.Attributes
		expected   []string
	}{
		{
			name: "ALPN HTTP/1.1",
			attributes: types.Attributes{
				{
					Name:  types.AttributeHTTPProtocol,
					Value: types.AttributeValueHTTPProtocol11,
				},
			},
			expected: []string{alpnProtocolHTTP11},
		},
		{
			name: "ALPN HTTP/2",
			attributes: types.Attributes{
				{
					Name:  types.AttributeHTTPProtocol,
					Value: types.AttributeValueHTTPProtocol2,
				},
			},
			expected: []string{alpnProtocolHTTP2, alpnProtocolHTTP11},
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			buildALPNProtocols(test.entity, test.attributes), test.name)
	}
}

func Test_buildTLSCertificates(t *testing.T) {

	certificate := "-----BEGIN CERTIFICATE-----\nMIIDmzCCAoOgAwIBAgIUbJq6CcBfxqN2Pwuu8l+26Sqa44UwDQYJKoZIhvcNAQEL\nBQAwXTELMAkGA1UEBhMCTkwxCzAJBgNVBAgMAk5IMRIwEAYDVQQHDAlBbXN0ZXJk\nYW0xETAPBgNVBAoMCEVyaWsgSW5jMRowGAYDVQQDDBFub3pvbWkuc2lldmllLmNv\nbTAeFw0xOTExMTUyMDQ0MDdaFw0yMDExMTQyMDQ0MDdaMF0xCzAJBgNVBAYTAk5M\nMQswCQYDVQQIDAJOSDESMBAGA1UEBwwJQW1zdGVyZGFtMREwDwYDVQQKDAhFcmlr\nIEluYzEaMBgGA1UEAwwRbm96b21pLnNpZXZpZS5jb20wggEiMA0GCSqGSIb3DQEB\nAQUAA4IBDwAwggEKAoIBAQDb9JTssv+M1xJbvX6T5TsRXuWzkhOrhevXZAqEHoJk\noo8b3lDLCIN/mF6L7uMJOayVCDHIE10kSBcTbqU6ERI6Iw1lUDfDP6E58UqZNTY4\ngh+3q7pC6/56gftsdHyFezzuRj7xjwMennFQx+RMAXOKkeHrYTYQecwjltlERNez\n7N9ZqSTjDTKkWDGnt1jV69yZ+mj5Eb49XUILitI/JQFSeN5IKF0P1iIy0Ud6On16\nVXCY26rYpgGdAs6kMiAbPSd5F48VbL+k2siPCp2j5fEmz6R4Jqq8U69kekjckiTX\nwCvlkP3p8f1RNIfbYtz/i8Ad0Qnh4DGcvKZV5A3WzxEPAgMBAAGjUzBRMB0GA1Ud\nDgQWBBQI6eYshsVEecejqBkvCr3ZbJy6XDAfBgNVHSMEGDAWgBQI6eYshsVEecej\nqBkvCr3ZbJy6XDAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQC2\nv7548ctujTyVG+rJ1jYUedsuVjSj1vRNZz4DXMB6P73syaloRP20dCVdxeOiNpN/\nv6Qspqxhyb0VicGBOxT/UmERX4/77ZuGxOptDfXH1caQgsB/aaPQmpjIdqJ8AfsM\nIBWfMd97N9DE9yjfT5tf9+vsOOeXvLg9ktc/DzlMrQuXRvtOvrdO/VzBMJFrpfA6\niu3Jg4FgrWA9O0l90yBAKJf6XIkmiUpn7cqPC18arRf+fW+x+Osq/8J8dYVBiOZZ\n8onrWWxdBWBRjn41fe9wmvaLijnSTxnL0x17YpbUp/GrDpF/x1Efdb0psw9LLbne\nOQphHwAS0a+Z48RmDzwA\n-----END CERTIFICATE-----"
	certificateKey := "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDb9JTssv+M1xJb\nvX6T5TsRXuWzkhOrhevXZAqEHoJkoo8b3lDLCIN/mF6L7uMJOayVCDHIE10kSBcT\nbqU6ERI6Iw1lUDfDP6E58UqZNTY4gh+3q7pC6/56gftsdHyFezzuRj7xjwMennFQ\nx+RMAXOKkeHrYTYQecwjltlERNez7N9ZqSTjDTKkWDGnt1jV69yZ+mj5Eb49XUIL\nitI/JQFSeN5IKF0P1iIy0Ud6On16VXCY26rYpgGdAs6kMiAbPSd5F48VbL+k2siP\nCp2j5fEmz6R4Jqq8U69kekjckiTXwCvlkP3p8f1RNIfbYtz/i8Ad0Qnh4DGcvKZV\n5A3WzxEPAgMBAAECggEBAMrOIf55MM2ghInYF/yvsJ3cnPjMaJyPN5x63oNhSiMW\nC9PLUT1TVUPxrsNheS7JYcpsKtJqoEfSvIwrSedXVDIMnc5bf37kjXjKdVj8Skki\nGbKVgYEw7Yvxi2w9n47Hya99T44UqfCycJLmLCa0c99BkUghcuMQGlx6O0wKGcUH\nnk8FXSGtmAEczdac+JHXbojfVcTm06EKaSBDac7aRxKqIVtKekYWw2ZHRtc/nk+p\nddTJAwDDPh3YS3GJ18C57LaPGnJ3a0xHiXLD96baJhFoKHmNSjj/aa1ZSPAwcLGD\ne+ZprCuhygJGiAnou69UwluD5ocRpHc3t+8qXUTuQekCgYEA85UOvZz2kAteW31a\nIK8G3uirS8I1yp+nW/RaItTaE7rEHCPT8h2ferqRW1TpJLQXmN17JCP7hP54V6WV\nnedEEic5WhonF+n2f8grQK5c9LBkL/uDoPpEi6ZdYYY3hZnntGs6oShrgPV5lGYm\nvNB+Lty+fXmJINOOcr2MbSPvIesCgYEA5ysswen0LTniRfYOgiF678l922ym5lOg\nXBD10HRle59AHlXQHyTcY43PijQpGyCs28kGBtCRDjTyvKtGDGzpwx3aUnGwAnxp\ni4nw8EwmyuOEf2FTNdvlHEqmql4a71q0N6JOdFs2sDIxdBomlbIvG/X4eyKre//+\n+OLMPYDw4G0CgYBi+eCBf7RYl6YBuw/SVAyQqy5fnEzLRtB0dvfhS2hJuAxT+uL2\ncL8K2aCS4g/SUDN+dBDDgLOFOPmhc7E19nEchz+wswvLldAJ4EZjA/bVno83SBYW\nZVtQ+4raQ/VvnjgegavTLF9yiUyb1l5LPtTnKd9lkOr9obkyOn9DIeTbfQKBgFh1\nJv1U/wDHY5SN4WNeWGKlYamzW/JLEdPpEYcg4yx49dol0Cv6uPLHcyFZcFlXGY5I\n0CuPZ9Jd5HzZtUZP7uug4sglhMqOvPyOXko1eaqtgSgVH/g+Gt/GmRwcQoZQ2SFo\n1EimFrk5m77nutgRhQFYECteSux6OyEV+D2Yt5PJAoGAA8zS5w7XMml2nmRZalYy\nScGiebiGNJZHWrfeOMSPPXN1CEs06r1AF454dkJtDsQEvs0+gwCtGWRUWI02T7ub\nploun1+vCNpBRoEUS50xsZAIaXgNLtM2afowfltc1TU1UR2bFg7OS+sW1YLJzm9w\n2E6B7kFT8aKQS6yEnL5+m6M=\n-----END PRIVATE KEY-----"

	tests := []struct {
		name       string
		entity     string
		attributes types.Attributes
		expected   []*tls.TlsCertificate
	}{
		{
			name: "certificate",
			attributes: types.Attributes{
				{
					Name:  types.AttributeTLSCertificate,
					Value: certificate,
				},
				{
					Name:  types.AttributeTLSCertificateKey,
					Value: certificateKey,
				},
			},
			expected: []*tls.TlsCertificate{
				{
					CertificateChain: &core.DataSource{
						Specifier: &core.DataSource_InlineString{
							InlineString: certificate,
						},
					},
					PrivateKey: &core.DataSource{
						Specifier: &core.DataSource_InlineString{
							InlineString: certificateKey,
						},
					},
				},
			},
		},
		{
			name: "certificate, no key",
			attributes: types.Attributes{
				{
					Name:  types.AttributeTLSCertificate,
					Value: certificate,
				},
			},
			expected: nil,
		},
		{
			name: "certificate key only",
			attributes: types.Attributes{
				{
					Name:  types.AttributeTLSCertificateKey,
					Value: certificateKey,
				},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		require.Equalf(t, test.expected,
			buildTLSCertificates(test.attributes), test.name)
	}
}

func Test_protoBool(t *testing.T) {

	require.Equal(t, &wrappers.BoolValue{Value: true},
		protoBool(true))
	require.Equal(t, &wrappers.BoolValue{Value: false},
		protoBool(false))

}

func Test_protoUint32(t *testing.T) {

	require.Equal(t, &wrappers.UInt32Value{Value: 42},
		protoUint32(42))
}

func Test_protoUint32orNil(t *testing.T) {

	require.Equal(t, &wrappers.UInt32Value{Value: 75},
		protoUint32orNil(75))

	require.Equal(t, (*wrapperspb.UInt32Value)(nil),
		protoUint32orNil(0))
}

func mustMarshalAny(pb protoiface.MessageV1) *any.Any {

	a, err := ptypes.MarshalAny(pb)
	if err != nil {
		panic(err.Error())
	}
	return a
}
