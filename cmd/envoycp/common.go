package main

import (
	"strings"
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/runtime/protoiface"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

// buildAddress builds an Envoy address to connect to
func buildAddress(hostname string, port int) *core.Address {

	return &core.Address{
		Address: &core.Address_SocketAddress{
			SocketAddress: &core.SocketAddress{
				Address:  hostname,
				Protocol: core.SocketAddress_TCP,
				PortSpecifier: &core.SocketAddress_PortValue{
					PortValue: uint32(port),
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

func buildConfigSource(cluster string, timeout time.Duration) *core.ConfigSource {

	grpcService := []*core.GrpcService{
		buildGRPCService(cluster, timeout),
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
		log.Warnf("Cannot encode resource '%s' as transportsocket", resourceName)
		return nil
	}
	return &core.TransportSocket{
		Name: "tls",
		ConfigType: &core.TransportSocket_TypedConfig{
			TypedConfig: tlsContextProtoBuf,
		},
	}
}

func buildCommonTLSContext(resourceName string, attributes shared.Attributes) *tls.CommonTlsContext {

	return &tls.CommonTlsContext{
		AlpnProtocols:   buildALPNProtocols(resourceName, attributes),
		TlsParams:       buildTLSParameters(attributes),
		TlsCertificates: buildTLSCertificates(attributes),
	}
}

func buildTLSCipherSuites(attributes shared.Attributes) []string {

	value, err := attributes.Get(attributeTLSCipherSuites)
	if err == nil {
		var ciphers []string

		for _, cipher := range strings.Split(value, ",") {
			ciphers = append(ciphers, strings.TrimSpace(cipher))
		}
		return ciphers
	}
	return nil
}

func buildTLSParameters(attributes shared.Attributes) *tls.TlsParameters {

	tlsParameters := &tls.TlsParameters{
		CipherSuites: buildTLSCipherSuites(attributes),
	}

	if minVersion, err := attributes.Get(attributeTLSMinimumVersion); err == nil {
		tlsParameters.TlsMinimumProtocolVersion = buildTLSVersion(minVersion)
	}

	if maxVersion, err := attributes.Get(attributeTLSMaximumVersion); err == nil {
		tlsParameters.TlsMaximumProtocolVersion = buildTLSVersion(maxVersion)
	}
	return tlsParameters
}

func buildTLSVersion(version string) tls.TlsParameters_TlsProtocol {

	switch version {
	case attributeValueTLSVersion10:
		return tls.TlsParameters_TLSv1_0
	case attributeValueTLSVersion11:
		return tls.TlsParameters_TLSv1_1
	case attributeValueTLSVersion12:
		return tls.TlsParameters_TLSv1_2
	case attributeValueTLSVersion13:
		return tls.TlsParameters_TLSv1_3
	}
	return tls.TlsParameters_TLS_AUTO
}

// buildALPNOptions return supported ALPN protocols
func buildALPNProtocols(resourceName string, attributes shared.Attributes) []string {

	value, err := attributes.Get(attributeHTTPProtocol)
	if err == nil {
		switch value {
		case attributeValueHTTPProtocol11:
			return []string{"http/1.1"}

		case attributeValueHTTPProtocol2:
			return []string{"h2", "http/1.1"}

		default:
			log.Warnf("Resource '%s' has attribute '%s' with unsupported value '%s'",
				resourceName, attributeHTTPProtocol, value)
		}
	}
	return []string{"http/1.1"}
}

func buildTLSCertificates(attributes shared.Attributes) []*tls.TlsCertificate {

	certificate, certificateError := attributes.Get(attributeTLSCertificate)
	certificateKey, certificateKeyError := attributes.Get(attributeTLSCertificateKey)

	if certificateError != nil && certificateKeyError != nil {
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
