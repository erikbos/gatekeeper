package main

import (
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
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

// Uint32orNil returns value in *wrapperspb.UInt32Value
func Uint32orNil(val int) *wrappers.UInt32Value {

	if val == 0 {
		return nil
	}
	return &wrappers.UInt32Value{
		Value: uint32(val),
	}
}

// FIXME probably more of the transportsocket related stuff can move here as
// both listener and cluster use the same types
