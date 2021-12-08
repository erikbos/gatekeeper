package main

import (
	"strings"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/erikbos/gatekeeper/pkg/types"
)

const (
	// ALPN protocols
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/transport_sockets/tls/v3/tls.proto.html?highlight=alpn
	alpnProtocolHTTP11 = "http/1.1"
	alpnProtocolHTTP2  = "h2"
)

// buildAddress builds an Envoy address to connect to
func buildAddress(hostname string, port uint32) *envoy_core.Address {

	return &envoy_core.Address{
		Address: &envoy_core.Address_SocketAddress{
			SocketAddress: &envoy_core.SocketAddress{
				Address:  hostname,
				Protocol: envoy_core.SocketAddress_TCP,
				PortSpecifier: &envoy_core.SocketAddress_PortValue{
					PortValue: port,
				},
			},
		},
	}
}

func buildGRPCService(clusterName string, timeout time.Duration) *envoy_core.GrpcService {

	return &envoy_core.GrpcService{
		Timeout: durationpb.New(timeout),
		TargetSpecifier: &envoy_core.GrpcService_EnvoyGrpc_{
			EnvoyGrpc: &envoy_core.GrpcService_EnvoyGrpc{
				ClusterName: clusterName,
			},
		},
	}
}

func buildConfigSource(clusterName string, timeout time.Duration) *envoy_core.ConfigSource {

	grpcService := []*envoy_core.GrpcService{
		buildGRPCService(clusterName, timeout),
	}
	return &envoy_core.ConfigSource{
		ConfigSourceSpecifier: &envoy_core.ConfigSource_ApiConfigSource{
			ApiConfigSource: &envoy_core.ApiConfigSource{
				ApiType:             envoy_core.ApiConfigSource_GRPC,
				GrpcServices:        grpcService,
				TransportApiVersion: envoy_core.ApiVersion_V3,
			},
		},
		ResourceApiVersion: envoy_core.ApiVersion_V3,
	}
}

func buildTransportSocket(resourceName string, tlsContext protoreflect.ProtoMessage) *envoy_core.TransportSocket {

	tlsContextProtoBuf, err := anypb.New(tlsContext)
	if err != nil {
		return nil
	}
	return &envoy_core.TransportSocket{
		Name: "tls",
		ConfigType: &envoy_core.TransportSocket_TypedConfig{
			TypedConfig: tlsContextProtoBuf,
		},
	}
}

func buildCommonTLSContext(resourceName string, attributes types.Attributes) *envoy_tls.CommonTlsContext {

	return &envoy_tls.CommonTlsContext{
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

func buildTLSParameters(attributes types.Attributes) *envoy_tls.TlsParameters {

	tlsParameters := &envoy_tls.TlsParameters{
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

func buildTLSVersion(version string) envoy_tls.TlsParameters_TlsProtocol {

	switch version {
	case types.AttributeValueTLSVersion10:
		return envoy_tls.TlsParameters_TLSv1_0
	case types.AttributeValueTLSVersion11:
		return envoy_tls.TlsParameters_TLSv1_1
	case types.AttributeValueTLSVersion12:
		return envoy_tls.TlsParameters_TLSv1_2
	case types.AttributeValueTLSVersion13:
		return envoy_tls.TlsParameters_TLSv1_3
	}
	return envoy_tls.TlsParameters_TLS_AUTO
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

func buildTLSCertificates(attributes types.Attributes) []*envoy_tls.TlsCertificate {

	var certificateChain, privateKey *envoy_core.DataSource

	certificateString, certificateError := attributes.Get(types.AttributeTLSCertificate)
	certificateKeyString, certificateKeyError := attributes.Get(types.AttributeTLSCertificateKey)
	if certificateError == nil && certificateKeyError == nil {
		certificateChain = &envoy_core.DataSource{
			Specifier: &envoy_core.DataSource_InlineString{
				InlineString: certificateString,
			},
		}
		privateKey = &envoy_core.DataSource{
			Specifier: &envoy_core.DataSource_InlineString{
				InlineString: certificateKeyString,
			},
		}
	}

	certificateFile, certificateError := attributes.Get(types.AttributeTLSCertificateFile)
	certificateKeyFile, certificateKeyError := attributes.Get(types.AttributeTLSCertificateKeyFile)
	if certificateError == nil && certificateKeyError == nil {
		certificateChain = &envoy_core.DataSource{
			Specifier: &envoy_core.DataSource_InlineString{
				InlineString: certificateFile,
			},
		}
		privateKey = &envoy_core.DataSource{
			Specifier: &envoy_core.DataSource_InlineString{
				InlineString: certificateKeyFile,
			},
		}
	}

	if certificateChain == nil || privateKey == nil {
		return nil
	}

	return []*envoy_tls.TlsCertificate{
		{
			CertificateChain: certificateChain,
			PrivateKey:       privateKey,
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
