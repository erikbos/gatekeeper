package shared

import (
	"net"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// GetCurrentTimeMilliseconds returns current epoch time in milliseconds
func GetCurrentTimeMilliseconds() int64 {
	return time.Now().UTC().UnixNano() / int64(time.Millisecond)
}

// TimeMillisecondsToString return timestamp as string in UTC
func TimeMillisecondsToString(timestamp int64) string {
	return time.Unix(0, timestamp*int64(time.Millisecond)).UTC().Format(time.RFC3339)
}

// TimeMillisecondsToInt64 returns time.Time as int64, if not set -1 will be returned
func TimeMillisecondsToInt64(timestamp time.Time) int64 {
	if timestamp.IsZero() {
		return -1
	}
	return timestamp.UTC().UnixNano() / int64(time.Millisecond)
}

// CheckIPinAccessList checks if ip addresses is in one of the subnets of IP ACL
func CheckIPinAccessList(ip net.IP, ipAccessList string) bool {

	if ipAccessList == "" {
		return false
	}
	for _, subnet := range strings.Split(ipAccessList, ",") {
		if _, network, err := net.ParseCIDR(strings.TrimSpace(subnet)); err == nil {
			if network.Contains(ip) {
				// OK, we have a match
				return true
			}
		}
		// We should not parse ACL, let's continue
	}
	// source ip did not match any of the ACL subnets, request rejected
	return false
}

// LoadYAMLConfiguration loads a configuration file and parses it as YAML
func LoadYAMLConfiguration(filename *string, config interface{}) (interface{}, error) {

	file, err := os.Open(*filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	yamlDecoder := yaml.NewDecoder(file)
	yamlDecoder.SetStrict(true)
	if err := yamlDecoder.Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}
