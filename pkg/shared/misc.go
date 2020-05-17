package shared

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// returnJSONMessage returns an error message
func returnJSONMessage(c *gin.Context, statusCode int, errorMessage error) {
	c.IndentedJSON(statusCode, gin.H{"message": fmt.Sprintf("%s", errorMessage)})
}

// AbortIfContentTypeNotJSON checks for json content-type and abort request
func AbortIfContentTypeNotJSON(c *gin.Context) {
	if c.Request.Header.Get("content-type") != "application/json" {
		returnJSONMessage(c, http.StatusUnsupportedMediaType,
			errors.New("Content-type application/json required when submitting data"))
		// do not continue request handling
		c.Abort()
	}
}

// GetCurrentTimeMilliseconds returns current epoch time in milliseconds
func GetCurrentTimeMilliseconds() int64 {
	return time.Now().UTC().UnixNano() / 1000000
}

// TimeMillisecondsToString return time as string
func TimeMillisecondsToString(timestamp int64) string {
	return time.Unix(0, timestamp*int64(time.Millisecond)).String()
}

// CheckIPinAccessList checks if ip addresses is in one of the subnets of IP ACL
func CheckIPinAccessList(ip net.IP, ipAccessList string) bool {

	if ipAccessList == "" {
		return false
	}
	for _, subnet := range strings.Split(strings.TrimSpace(ipAccessList), ",") {
		if _, network, err := net.ParseCIDR(subnet); err == nil {
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
