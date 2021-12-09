package policy

import (
	"net"

	"github.com/oschwald/maxminddb-golang"
)

// Geoip hold our configuration
type Geoip struct {
	Database string `yaml:"database"`
	mdb      *maxminddb.Reader
}

// GetCountryAndState returns country and state of the location of an ip address
func (g *Geoip) GetCountryAndState(ipaddress net.IP) (string, string) {

	if g.mdb == nil || ipaddress == nil {
		return "", ""
	}

	var record struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
		Subdivisions []struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"subdivisions"`
	}

	err := g.mdb.Lookup(ipaddress, &record)
	if err != nil {
		return "", ""
	}
	// Do we have geoip state information?
	if len(record.Subdivisions) != 0 {
		return record.Country.ISOCode, record.Subdivisions[0].ISOCode
	}
	return record.Country.ISOCode, ""
}

// OpenGeoipDatabase opens a Maxmind geoip database
func OpenGeoipDatabase(filename string) (*Geoip, error) {

	var err error
	g := Geoip{Database: filename}

	g.mdb, err = maxminddb.Open(filename)
	if err != nil {
		return nil, err
	}
	return &g, nil
}
