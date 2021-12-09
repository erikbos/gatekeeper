package policy

// Camnot run test without geoip db

// func TestGeoLookupIPs(t *testing.T) {
// 	assert := assert.New(t)

// 	tests := []struct {
// 		ipsubnet string
// 		country  string
// 	}{
// 		// xs4all's ip addresses are marked as NL
// 		{"194.109.0.0", "NL"},
// 		{"2001:980::42", "NL"},
// 		{"this_ipaddress_cannot_exist", ""},
// 	}

// 	g, err := OpenGeoipDatabase("../../GeoIP2-City.mmdb")
// 	if err != nil {
// 		log.Fatalf("failed to open geoip database: %v", err)
// 	}
// 	for _, test := range tests {
// 		ipaddress := net.ParseIP(test.ipsubnet)
// 		country, _ := g.GetCountryAndState(ipaddress)

// 		assert.Equalf(test.country, country, "GeoIP lookup mismatch")
// 	}
// }
