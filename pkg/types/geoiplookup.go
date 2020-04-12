package types

import (
	"fmt"
	"net"

	"github.com/oschwald/maxminddb-golang"
)

//Geoip bla
type Geoip struct {
	mdb *maxminddb.Reader
}

// GetCountryAndState returns country and state of the location of an ip address
//
func (g *Geoip) GetCountryAndState(ipaddress string) (string, string) {
	if g.mdb != nil {
		ip := net.ParseIP(ipaddress)
		if ip != nil {
			var record struct {
				Country struct {
					ISOCode string `maxminddb:"iso_code"`
				} `maxminddb:"country"`
				Subdivisions []struct {
					ISOCode string `maxminddb:"iso_code"`
				} `maxminddb:"subdivisions"`
			}
			err := g.mdb.Lookup(ip, &record)
			if err == nil {
				// Do we have geoip state information?
				if len(record.Subdivisions) != 0 {
					return record.Country.ISOCode, record.Subdivisions[0].ISOCode
				}
				return record.Country.ISOCode, ""
			}
		}
	}
	return "", ""
}

// OpenDatabase opens a Maxmind geoip database
//
func OpenDatabase(filename string) (*Geoip, error) {
	var err error
	g := Geoip{}
	g.mdb, err = maxminddb.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Could not open %s", filename)
	}
	return &g, nil
}
