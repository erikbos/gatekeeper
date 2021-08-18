package service

import (
	"testing"

)

func Test_clearSensitiveFields(t *testing.T) {

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
