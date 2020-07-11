package main

import (
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
)

// buildAddress builds an Envoy address to connect to
func buildAddress(hostname string, port int) *core.Address {

	return &core.Address{Address: &core.Address_SocketAddress{
		SocketAddress: &core.SocketAddress{
			Address:  hostname,
			Protocol: core.SocketAddress_TCP,
			PortSpecifier: &core.SocketAddress_PortValue{
				PortValue: uint32(port),
			},
		},
	}}
}

func buildGRPCService(clusterName string, d time.Duration) *core.GrpcService {

	return &core.GrpcService{
		Timeout: ptypes.DurationProto(d),
		TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
			EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
				ClusterName: clusterName,
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

	return &wrappers.UInt32Value{
		Value: i,
	}
}

func protoUint32orNil(val int) *wrappers.UInt32Value {

	if val == 0 {
		return nil
	}
	return protoUint32(uint32(val))
}

// FIXME probably more of the transportsocket related stuff can move here as
// both listener and cluster use the same types
