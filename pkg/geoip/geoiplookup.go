package geoip

import (
	"fmt"
	"net"

	"github.com/oschwald/maxminddb-golang"
)

//Lookup bla
type Lookup struct {
	mdb *maxminddb.Reader
}

// GetCountryAndState returns country and state of the location of an ip address
//
func (l *Lookup) GetCountryAndState(ipaddress string) (string, string) {
	if l.mdb != nil {
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
			err := l.mdb.Lookup(ip, &record)
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
func OpenDatabase(filename string) (*Lookup, error) {
	var err error
	l := Lookup{}
	l.mdb, err = maxminddb.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Could not open %s", filename)
	}
	return &l, nil
}
