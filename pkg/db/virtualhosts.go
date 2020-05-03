package db

import (
	"fmt"

	"github.com/erikbos/apiauth/pkg/shared"
)

// Prometheus label for metrics of db interactions
// const routeMetricLabel = "virtualhosts"

// temp
var virtualhosts = []shared.VirtualHost{
	{
		Name:         "nozomi",
		DisplayName:  "Erik home testsetup 0",
		Port:         80,
		VirtualHosts: []string{"nozomi.sievie.nl"},
		RouteSet:     "routes_80",
		Attributes: []shared.AttributeKeyValues{
			{
				Name:  "bla",
				Value: "42",
			},
		},
		LastmodifiedAt: 10000,
	},
	{
		Name:         "nozomi",
		DisplayName:  "Erik home testsetup 1",
		Port:         443,
		VirtualHosts: []string{"nozomi.sievie.be"},
		RouteSet:     "routes_443",
		Attributes: []shared.AttributeKeyValues{
			{
				Name:  "Certificate",
				Value: "-----BEGIN CERTIFICATE-----\nMIIDmzCCAoOgAwIBAgIUbJq6CcBfxqN2Pwuu8l+26Sqa44UwDQYJKoZIhvcNAQEL\nBQAwXTELMAkGA1UEBhMCTkwxCzAJBgNVBAgMAk5IMRIwEAYDVQQHDAlBbXN0ZXJk\nYW0xETAPBgNVBAoMCEVyaWsgSW5jMRowGAYDVQQDDBFub3pvbWkuc2lldmllLmNv\nbTAeFw0xOTExMTUyMDQ0MDdaFw0yMDExMTQyMDQ0MDdaMF0xCzAJBgNVBAYTAk5M\nMQswCQYDVQQIDAJOSDESMBAGA1UEBwwJQW1zdGVyZGFtMREwDwYDVQQKDAhFcmlr\nIEluYzEaMBgGA1UEAwwRbm96b21pLnNpZXZpZS5jb20wggEiMA0GCSqGSIb3DQEB\nAQUAA4IBDwAwggEKAoIBAQDb9JTssv+M1xJbvX6T5TsRXuWzkhOrhevXZAqEHoJk\noo8b3lDLCIN/mF6L7uMJOayVCDHIE10kSBcTbqU6ERI6Iw1lUDfDP6E58UqZNTY4\ngh+3q7pC6/56gftsdHyFezzuRj7xjwMennFQx+RMAXOKkeHrYTYQecwjltlERNez\n7N9ZqSTjDTKkWDGnt1jV69yZ+mj5Eb49XUILitI/JQFSeN5IKF0P1iIy0Ud6On16\nVXCY26rYpgGdAs6kMiAbPSd5F48VbL+k2siPCp2j5fEmz6R4Jqq8U69kekjckiTX\nwCvlkP3p8f1RNIfbYtz/i8Ad0Qnh4DGcvKZV5A3WzxEPAgMBAAGjUzBRMB0GA1Ud\nDgQWBBQI6eYshsVEecejqBkvCr3ZbJy6XDAfBgNVHSMEGDAWgBQI6eYshsVEecej\nqBkvCr3ZbJy6XDAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQC2\nv7548ctujTyVG+rJ1jYUedsuVjSj1vRNZz4DXMB6P73syaloRP20dCVdxeOiNpN/\nv6Qspqxhyb0VicGBOxT/UmERX4/77ZuGxOptDfXH1caQgsB/aaPQmpjIdqJ8AfsM\nIBWfMd97N9DE9yjfT5tf9+vsOOeXvLg9ktc/DzlMrQuXRvtOvrdO/VzBMJFrpfA6\niu3Jg4FgrWA9O0l90yBAKJf6XIkmiUpn7cqPC18arRf+fW+x+Osq/8J8dYVBiOZZ\n8onrWWxdBWBRjn41fe9wmvaLijnSTxnL0x17YpbUp/GrDpF/x1Efdb0psw9LLbne\nOQphHwAS0a+Z48RmDzwA\n-----END CERTIFICATE-----\n",
			},
			{
				Name:  "CertificateKey",
				Value: "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDb9JTssv+M1xJb\nvX6T5TsRXuWzkhOrhevXZAqEHoJkoo8b3lDLCIN/mF6L7uMJOayVCDHIE10kSBcT\nbqU6ERI6Iw1lUDfDP6E58UqZNTY4gh+3q7pC6/56gftsdHyFezzuRj7xjwMennFQ\nx+RMAXOKkeHrYTYQecwjltlERNez7N9ZqSTjDTKkWDGnt1jV69yZ+mj5Eb49XUIL\nitI/JQFSeN5IKF0P1iIy0Ud6On16VXCY26rYpgGdAs6kMiAbPSd5F48VbL+k2siP\nCp2j5fEmz6R4Jqq8U69kekjckiTXwCvlkP3p8f1RNIfbYtz/i8Ad0Qnh4DGcvKZV\n5A3WzxEPAgMBAAECggEBAMrOIf55MM2ghInYF/yvsJ3cnPjMaJyPN5x63oNhSiMW\nC9PLUT1TVUPxrsNheS7JYcpsKtJqoEfSvIwrSedXVDIMnc5bf37kjXjKdVj8Skki\nGbKVgYEw7Yvxi2w9n47Hya99T44UqfCycJLmLCa0c99BkUghcuMQGlx6O0wKGcUH\nnk8FXSGtmAEczdac+JHXbojfVcTm06EKaSBDac7aRxKqIVtKekYWw2ZHRtc/nk+p\nddTJAwDDPh3YS3GJ18C57LaPGnJ3a0xHiXLD96baJhFoKHmNSjj/aa1ZSPAwcLGD\ne+ZprCuhygJGiAnou69UwluD5ocRpHc3t+8qXUTuQekCgYEA85UOvZz2kAteW31a\nIK8G3uirS8I1yp+nW/RaItTaE7rEHCPT8h2ferqRW1TpJLQXmN17JCP7hP54V6WV\nnedEEic5WhonF+n2f8grQK5c9LBkL/uDoPpEi6ZdYYY3hZnntGs6oShrgPV5lGYm\nvNB+Lty+fXmJINOOcr2MbSPvIesCgYEA5ysswen0LTniRfYOgiF678l922ym5lOg\nXBD10HRle59AHlXQHyTcY43PijQpGyCs28kGBtCRDjTyvKtGDGzpwx3aUnGwAnxp\ni4nw8EwmyuOEf2FTNdvlHEqmql4a71q0N6JOdFs2sDIxdBomlbIvG/X4eyKre//+\n+OLMPYDw4G0CgYBi+eCBf7RYl6YBuw/SVAyQqy5fnEzLRtB0dvfhS2hJuAxT+uL2\ncL8K2aCS4g/SUDN+dBDDgLOFOPmhc7E19nEchz+wswvLldAJ4EZjA/bVno83SBYW\nZVtQ+4raQ/VvnjgegavTLF9yiUyb1l5LPtTnKd9lkOr9obkyOn9DIeTbfQKBgFh1\nJv1U/wDHY5SN4WNeWGKlYamzW/JLEdPpEYcg4yx49dol0Cv6uPLHcyFZcFlXGY5I\n0CuPZ9Jd5HzZtUZP7uug4sglhMqOvPyOXko1eaqtgSgVH/g+Gt/GmRwcQoZQ2SFo\n1EimFrk5m77nutgRhQFYECteSux6OyEV+D2Yt5PJAoGAA8zS5w7XMml2nmRZalYy\nScGiebiGNJZHWrfeOMSPPXN1CEs06r1AF454dkJtDsQEvs0+gwCtGWRUWI02T7ub\nploun1+vCNpBRoEUS50xsZAIaXgNLtM2afowfltc1TU1UR2bFg7OS+sW1YLJzm9w\n2E6B7kFT8aKQS6yEnL5+m6M=\n-----END PRIVATE KEY-----\n",
			},
			{
				Name:  "TLSCipherSuites",
				Value: "[ECDHE-RSA-CHACHA20-POLY1305|ECDHE-RSA-AES256-GCM-SHA384|ECDHE-RSA-AES128-GCM-SHA256]",
			},
			{
				Name:  "TLSMinimumVersion",
				Value: "TLSv1_2",
			},
		},
		LastmodifiedAt: 10000,
	},
	{
		Name:         "nozomi",
		DisplayName:  "Erik home testsetup 2",
		Port:         443,
		VirtualHosts: []string{"nozomi.sievie.com"},
		RouteSet:     "routes_443",
		Attributes: []shared.AttributeKeyValues{
			{
				Name:  "Certificate",
				Value: "-----BEGIN CERTIFICATE-----\nMIIDmzCCAoOgAwIBAgIUbJq6CcBfxqN2Pwuu8l+26Sqa44UwDQYJKoZIhvcNAQEL\nBQAwXTELMAkGA1UEBhMCTkwxCzAJBgNVBAgMAk5IMRIwEAYDVQQHDAlBbXN0ZXJk\nYW0xETAPBgNVBAoMCEVyaWsgSW5jMRowGAYDVQQDDBFub3pvbWkuc2lldmllLmNv\nbTAeFw0xOTExMTUyMDQ0MDdaFw0yMDExMTQyMDQ0MDdaMF0xCzAJBgNVBAYTAk5M\nMQswCQYDVQQIDAJOSDESMBAGA1UEBwwJQW1zdGVyZGFtMREwDwYDVQQKDAhFcmlr\nIEluYzEaMBgGA1UEAwwRbm96b21pLnNpZXZpZS5jb20wggEiMA0GCSqGSIb3DQEB\nAQUAA4IBDwAwggEKAoIBAQDb9JTssv+M1xJbvX6T5TsRXuWzkhOrhevXZAqEHoJk\noo8b3lDLCIN/mF6L7uMJOayVCDHIE10kSBcTbqU6ERI6Iw1lUDfDP6E58UqZNTY4\ngh+3q7pC6/56gftsdHyFezzuRj7xjwMennFQx+RMAXOKkeHrYTYQecwjltlERNez\n7N9ZqSTjDTKkWDGnt1jV69yZ+mj5Eb49XUILitI/JQFSeN5IKF0P1iIy0Ud6On16\nVXCY26rYpgGdAs6kMiAbPSd5F48VbL+k2siPCp2j5fEmz6R4Jqq8U69kekjckiTX\nwCvlkP3p8f1RNIfbYtz/i8Ad0Qnh4DGcvKZV5A3WzxEPAgMBAAGjUzBRMB0GA1Ud\nDgQWBBQI6eYshsVEecejqBkvCr3ZbJy6XDAfBgNVHSMEGDAWgBQI6eYshsVEecej\nqBkvCr3ZbJy6XDAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQC2\nv7548ctujTyVG+rJ1jYUedsuVjSj1vRNZz4DXMB6P73syaloRP20dCVdxeOiNpN/\nv6Qspqxhyb0VicGBOxT/UmERX4/77ZuGxOptDfXH1caQgsB/aaPQmpjIdqJ8AfsM\nIBWfMd97N9DE9yjfT5tf9+vsOOeXvLg9ktc/DzlMrQuXRvtOvrdO/VzBMJFrpfA6\niu3Jg4FgrWA9O0l90yBAKJf6XIkmiUpn7cqPC18arRf+fW+x+Osq/8J8dYVBiOZZ\n8onrWWxdBWBRjn41fe9wmvaLijnSTxnL0x17YpbUp/GrDpF/x1Efdb0psw9LLbne\nOQphHwAS0a+Z48RmDzwA\n-----END CERTIFICATE-----\n",
			},
			{
				Name:  "CertificateKey",
				Value: "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDb9JTssv+M1xJb\nvX6T5TsRXuWzkhOrhevXZAqEHoJkoo8b3lDLCIN/mF6L7uMJOayVCDHIE10kSBcT\nbqU6ERI6Iw1lUDfDP6E58UqZNTY4gh+3q7pC6/56gftsdHyFezzuRj7xjwMennFQ\nx+RMAXOKkeHrYTYQecwjltlERNez7N9ZqSTjDTKkWDGnt1jV69yZ+mj5Eb49XUIL\nitI/JQFSeN5IKF0P1iIy0Ud6On16VXCY26rYpgGdAs6kMiAbPSd5F48VbL+k2siP\nCp2j5fEmz6R4Jqq8U69kekjckiTXwCvlkP3p8f1RNIfbYtz/i8Ad0Qnh4DGcvKZV\n5A3WzxEPAgMBAAECggEBAMrOIf55MM2ghInYF/yvsJ3cnPjMaJyPN5x63oNhSiMW\nC9PLUT1TVUPxrsNheS7JYcpsKtJqoEfSvIwrSedXVDIMnc5bf37kjXjKdVj8Skki\nGbKVgYEw7Yvxi2w9n47Hya99T44UqfCycJLmLCa0c99BkUghcuMQGlx6O0wKGcUH\nnk8FXSGtmAEczdac+JHXbojfVcTm06EKaSBDac7aRxKqIVtKekYWw2ZHRtc/nk+p\nddTJAwDDPh3YS3GJ18C57LaPGnJ3a0xHiXLD96baJhFoKHmNSjj/aa1ZSPAwcLGD\ne+ZprCuhygJGiAnou69UwluD5ocRpHc3t+8qXUTuQekCgYEA85UOvZz2kAteW31a\nIK8G3uirS8I1yp+nW/RaItTaE7rEHCPT8h2ferqRW1TpJLQXmN17JCP7hP54V6WV\nnedEEic5WhonF+n2f8grQK5c9LBkL/uDoPpEi6ZdYYY3hZnntGs6oShrgPV5lGYm\nvNB+Lty+fXmJINOOcr2MbSPvIesCgYEA5ysswen0LTniRfYOgiF678l922ym5lOg\nXBD10HRle59AHlXQHyTcY43PijQpGyCs28kGBtCRDjTyvKtGDGzpwx3aUnGwAnxp\ni4nw8EwmyuOEf2FTNdvlHEqmql4a71q0N6JOdFs2sDIxdBomlbIvG/X4eyKre//+\n+OLMPYDw4G0CgYBi+eCBf7RYl6YBuw/SVAyQqy5fnEzLRtB0dvfhS2hJuAxT+uL2\ncL8K2aCS4g/SUDN+dBDDgLOFOPmhc7E19nEchz+wswvLldAJ4EZjA/bVno83SBYW\nZVtQ+4raQ/VvnjgegavTLF9yiUyb1l5LPtTnKd9lkOr9obkyOn9DIeTbfQKBgFh1\nJv1U/wDHY5SN4WNeWGKlYamzW/JLEdPpEYcg4yx49dol0Cv6uPLHcyFZcFlXGY5I\n0CuPZ9Jd5HzZtUZP7uug4sglhMqOvPyOXko1eaqtgSgVH/g+Gt/GmRwcQoZQ2SFo\n1EimFrk5m77nutgRhQFYECteSux6OyEV+D2Yt5PJAoGAA8zS5w7XMml2nmRZalYy\nScGiebiGNJZHWrfeOMSPPXN1CEs06r1AF454dkJtDsQEvs0+gwCtGWRUWI02T7ub\nploun1+vCNpBRoEUS50xsZAIaXgNLtM2afowfltc1TU1UR2bFg7OS+sW1YLJzm9w\n2E6B7kFT8aKQS6yEnL5+m6M=\n-----END PRIVATE KEY-----\n",
			},
			{
				Name:  "TLSCipherSuites",
				Value: "[ECDHE-RSA-CHACHA20-POLY1305|ECDHE-RSA-AES256-GCM-SHA384|ECDHE-RSA-AES128-GCM-SHA256]",
			},
			{
				Name:  "TLSMinimumVersion",
				Value: "TLSv1_2",
			},
		},
		LastmodifiedAt: 10000,
	},
}

// GetVirtualHosts retrieves all virtualhosts
func (d *Database) GetVirtualHosts() ([]shared.VirtualHost, error) {
	return virtualhosts, nil

	// query := "SELECT * FROM virtualhosts"
	// virtualhosts, err := d.runGetVirtualHostQuery(query)
	// if err != nil {
	// 	return []shared.VirtualHost{}, err
	// }
	// if len(virtualhosts) == 0 {
	// 	d.metricsQueryMiss(routeMetricLabel)
	// 	return []shared.VirtualHost{}, errors.New("Can not retrieve list of virtualhosts")
	// }
	// d.metricsQueryHit(routeMetricLabel)
	// return virtualhosts, nil
}

// GetVirtualHostByName retrieves a route from database
func (d *Database) GetVirtualHostByName(virtualHost string) (shared.VirtualHost, error) {
	for _, value := range virtualhosts {
		if value.Name == virtualHost {
			return value, nil
		}
	}
	return shared.VirtualHost{}, fmt.Errorf("Can not find virtualhost (%s)", virtualHost)

	// query := "SELECT * FROM routez WHERE key = ? LIMIT 1"
	// virtualhosts, err := d.runGetVirtualHostQuery(query, routeName)
	// if err != nil {
	// 	return shared.VirtualHost{}, err
	// }
	// if len(virtualhosts) == 0 {
	// 	d.metricsQueryMiss(routeMetricLabel)
	// 	return shared.VirtualHost{},
	// 		fmt.Errorf("Can not find route (%s)", routeName)
	// }
	// d.metricsQueryHit(routeMetricLabel)
	// return virtualhosts[0], nil
}

// runGetVirtualHostQuery executes CQL query and returns resultset
// func (d *Database) runGetVirtualHostQuery(query string, queryParameters ...interface{}) ([]shared.VirtualHost, error) {
// 		var virtualhosts []shared.VirtualHost

// 		timer := prometheus.NewTimer(d.dbLookupHistogram)
// 		defer timer.ObserveDuration()

// 		iter := d.cassandraSession.Query(query, queryParameters...).Iter()
// 		m := make(map[string]interface{})
// 		for iter.MapScan(m) {
// 			newVirtualHost := shared.VirtualHost{
// 				Name:           m["key"].(string),
// 				MatchPrefix:       m["host_name"].(string),
// 				Port:           m["port"].(int),
// 				Cluster:        m["cluster"].(string),
// 				PrefixRewrite:  m["PrefixRewrite"].(string),
// 				CreatedAt:      m["created_at"].(int64),
// 				CreatedBy:      m["created_by"].(string),
// 				DisplayName:    m["display_name"].(string),
// 				LastmodifiedAt: m["lastmodified_at"].(int64),
// 				LastmodifiedBy: m["lastmodified_by"].(string),
// 			}
// 			if m["attributes"] != nil {
// 				newVirtualHost.Attributes = d.unmarshallJSONArrayOfAttributes(m["attributes"].(string))
// 			}
// 			virtualhosts = append(virtualhosts, newVirtualHost)
// 			m = map[string]interface{}{}
// 		}
// 		// In case query failed we return query error
// 		if err := iter.Close(); err != nil {
// 			log.Error(err)
// 			return []shared.VirtualHost{}, err
// 		}
// 		return virtualhosts, nil
// 	return virtualhosts, nil
// }

// UpdateVirtualHostByName UPSERTs an route in database
func (d *Database) UpdateVirtualHostByName(updatedVirtualHost *shared.VirtualHost) error {
	// query := "INSERT INTO routez (key, display_name, " +
	// 	"host_name, port, attributes, " +
	// 	"created_at, created_by, lastmodified_at, lastmodified_by) " +
	// 	"VALUES(?,?,?,?,?,?,?,?,?)"
	// updatedVirtualHost.Attributes = shared.TidyAttributes(updatedVirtualHost.Attributes)
	// attributes := d.marshallArrayOfAttributesToJSON(updatedVirtualHost.Attributes)
	// updatedVirtualHost.LastmodifiedAt = shared.GetCurrentTimeMilliseconds()
	// if err := d.cassandraSession.Query(query,
	// 	updatedVirtualHost.Name, updatedVirtualHost.DisplayName,
	// 	updatedVirtualHost.HostName, updatedVirtualHost.Port, attributes,
	// 	updatedVirtualHost.CreatedAt, updatedVirtualHost.CreatedBy,
	// 	updatedVirtualHost.LastmodifiedAt,
	// 	updatedVirtualHost.LastmodifiedBy).Exec(); err != nil {
	// 	return fmt.Errorf("Can not update route (%v)", err)
	// }
	return nil
}

// DeleteVirtualHostByName deletes a route
func (d *Database) DeleteVirtualHostByName(virtualHostToDelete string) error {
	// _, err := d.GetVirtualHostByName(virtualHostToDelete)
	// if err != nil {
	// 	return err
	// }
	// query := "DELETE FROM virtualhosts WHERE key = ?"
	// return d.cassandraSession.Query(query, virtualHostToDelete).Exec()
	return nil
}
