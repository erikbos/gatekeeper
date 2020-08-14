package shared

import (
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// GetCurrentTimeMilliseconds returns current epoch time in milliseconds
func GetCurrentTimeMilliseconds() int64 {
	return time.Now().UTC().UnixNano() / int64(time.Millisecond)
}

// TimeMillisecondsToString return timestamp as string in UTC
func TimeMillisecondsToString(timestamp int64) string {
	return time.Unix(0, timestamp*int64(time.Millisecond)).UTC().Format(time.RFC3339)
}

// TimeMillisecondsToInt64 returns time.Time as int64
func TimeMillisecondsToInt64(timestamp time.Time) int64 {
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
		} else {
			log.Debugf("FIXME increase unparsable ip ACL counter")
		}
	}
	// source ip did not match any of the ACL subnets, request rejected
	return false
}
