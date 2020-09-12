package main

import (
	"testing"

	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
)

func (s *server) TestbuildFilter(t *testing.T) {

	tests := []struct {
		s    *server
		want []*hcm.HttpFilter
	}{
		// test 1
		{
			&server{
				config: &EnvoyCPConfig{
					Envoyproxy: envoyProxyConfig{
						ExtAuthz: extAuthzConfig{
							Enable: false,
						},
						RateLimiter: rateLimiterConfig{
							Enable: false,
						},
					},
				},
			},
			[]*hcm.HttpFilter{
				{
					ConfigType: &hcm.HttpFilter_TypedConfig{},
				},
			},
		},
	}

	// assert := assert.New(t)

	for _, test := range tests {
		test.s.buildFilter()
		// assert.EqualValues((test.country, country, "GeoIP lookup mismatch")

	}
}
