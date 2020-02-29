package geoip

import (
	"log"
	"testing"
)

func TestGeoLookupIPs(t *testing.T) {

	testData := []struct {
		ipsubnet string
		country  string
	}{
		// xs4all's ip addresses are marked as NL
		{"194.109.0.0", "NL"},
		{"2001:980::42", "NL"},
		{"this_ipaddress_cannot_exist", ""},
	}

	g, err := OpenDatabase("../GeoIP2-City.mmdb")
	if err != nil {
		log.Fatalf("failed to open geoip database: %v", err)
	}

	for _, testSubnet := range testData {
		country, _ := g.GetCountryAndState(testSubnet.ipsubnet)

		if country != testSubnet.country {
			t.Errorf("Lookup of (%s) was incorrect, got: %s, want: %s.", testSubnet.ipsubnet, country, testSubnet.country)
		}
	}
}
