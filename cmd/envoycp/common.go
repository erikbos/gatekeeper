package main

import (
	"strings"
	"testing"
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// ALPN protocols
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/transport_sockets/tls/v3/tls.proto.html?highlight=alpn
	alpnProtocolHTTP11 = "http/1.1"
	alpnProtocolHTTP2  = "h2"
)

// buildAddress builds an Envoy address to connect to
func buildAddress(hostname string, port uint32) *core.Address {

	return &core.Address{
		Address: &core.Address_SocketAddress{
			SocketAddress: &core.SocketAddress{
				Address:  hostname,
				Protocol: core.SocketAddress_TCP,
				PortSpecifier: &core.SocketAddress_PortValue{
					PortValue: port,
				},
			},
		},
	}
}

func buildGRPCService(clusterName string, timeout time.Duration) *core.GrpcService {

	return &core.GrpcService{
		Timeout: ptypes.DurationProto(timeout),
		TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
			EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
				ClusterName: clusterName,
			},
		},
	}
}

func buildConfigSource(clusterName string, timeout time.Duration) *core.ConfigSource {

	grpcService := []*core.GrpcService{
		buildGRPCService(clusterName, timeout),
	}
	return &core.ConfigSource{
		ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
			ApiConfigSource: &core.ApiConfigSource{
				ApiType:             core.ApiConfigSource_GRPC,
				GrpcServices:        grpcService,
				TransportApiVersion: core.ApiVersion_V3,
			},
		},
		ResourceApiVersion: core.ApiVersion_V3,
	}
}

func buildTransportSocket(resourceName string, tlsContext protoiface.MessageV1) *core.TransportSocket {

	tlsContextProtoBuf, err := ptypes.MarshalAny(tlsContext)
	if err != nil {
		return nil
	}
	return &core.TransportSocket{
		Name: "tls",
		ConfigType: &core.TransportSocket_TypedConfig{
			TypedConfig: tlsContextProtoBuf,
		},
	}
}

func buildCommonTLSContext(resourceName string, attributes types.Attributes) *tls.CommonTlsContext {

	return &tls.CommonTlsContext{
		AlpnProtocols:   buildALPNProtocols(resourceName, attributes),
		TlsParams:       buildTLSParameters(attributes),
		TlsCertificates: buildTLSCertificates(attributes),
	}
}

func buildTLSCipherSuites(attributes types.Attributes) []string {

	value, err := attributes.Get(types.AttributeTLSCipherSuites)
	if err == nil && value != "" {
		var ciphers []string

		for _, cipher := range strings.Split(value, ",") {
			ciphers = append(ciphers, strings.TrimSpace(cipher))
		}
		return ciphers
	}
	return nil
}

func buildTLSParameters(attributes types.Attributes) *tls.TlsParameters {

	tlsParameters := &tls.TlsParameters{
		CipherSuites: buildTLSCipherSuites(attributes),
	}

	if minVersion, err := attributes.Get(types.AttributeTLSMinimumVersion); err == nil {
		tlsParameters.TlsMinimumProtocolVersion = buildTLSVersion(minVersion)
	}

	if maxVersion, err := attributes.Get(types.AttributeTLSMaximumVersion); err == nil {
		tlsParameters.TlsMaximumProtocolVersion = buildTLSVersion(maxVersion)
	}
	return tlsParameters
}

func buildTLSVersion(version string) tls.TlsParameters_TlsProtocol {

	switch version {
	case types.AttributeValueTLSVersion10:
		return tls.TlsParameters_TLSv1_0
	case types.AttributeValueTLSVersion11:
		return tls.TlsParameters_TLSv1_1
	case types.AttributeValueTLSVersion12:
		return tls.TlsParameters_TLSv1_2
	case types.AttributeValueTLSVersion13:
		return tls.TlsParameters_TLSv1_3
	}
	return tls.TlsParameters_TLS_AUTO
}

// buildALPNOptions return supported ALPN protocols
func buildALPNProtocols(entity string, attributes types.Attributes) []string {

	value, err := attributes.Get(types.AttributeHTTPProtocol)
	if err == nil {
		switch value {
		case types.AttributeValueHTTPProtocol11:
			return []string{alpnProtocolHTTP11}

		case types.AttributeValueHTTPProtocol2:
			return []string{alpnProtocolHTTP2, alpnProtocolHTTP11}
		}
	}
	return []string{alpnProtocolHTTP11}
}

func buildTLSCertificates(attributes types.Attributes) []*tls.TlsCertificate {

	certificate, certificateError := attributes.Get(types.AttributeTLSCertificate)
	certificateKey, certificateKeyError := attributes.Get(types.AttributeTLSCertificateKey)

	if certificateError != nil || certificateKeyError != nil {
		return nil
	}

	return []*tls.TlsCertificate{
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
	}
}

func protoBool(b bool) *wrappers.BoolValue {

	if b {
		return &wrappers.BoolValue{Value: true}
	}
	return &wrappers.BoolValue{Value: false}
}

func protoUint32(i uint32) *wrappers.UInt32Value {

	return &wrappers.UInt32Value{Value: i}
}

func protoUint32orNil(val uint32) *wrappers.UInt32Value {

	if val == 0 {
		return nil
	}
	return protoUint32(val)
}

// RequireEqual will test that want == got for protobufs, call t.fatal if it does not,
// This mimics the behavior of the testify `require` functions.
func RequireEqual(t *testing.T, want, got interface{}) {
	t.Helper()

	diff := cmp.Diff(want, got, protocmp.Transform())
	if diff != "" {
		t.Fatal(diff)
	}
}
