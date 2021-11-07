package types

import (
	"fmt"
	"sort"
)

// Listener contains everything about downstream configuration of listener and http virtual hosts
//
// Field validation (binding) is done using https://godoc.org/github.com/go-playground/validator
type (
	Listener struct {
		// Name of listener (not changable)
		Name string `binding:"required,min=4"`

		// Friendly display name of listener
		DisplayName string

		// Virtual hosts of this listener (at least one, each value must be a fqdn)
		VirtualHosts []string `binding:"required,min=1,dive,fqdn"`

		// tcp port to listen on
		Port int `binding:"required,min=1,max=65535"`

		// Routegroup to forward traffic to
		RouteGroup string `binding:"required"`

		// Comma separated list of policynames, to apply to requests
		Policies string

		// Attributes of this listener
		Attributes Attributes

		// Created at timestamp in epoch milliseconds
		CreatedAt int64

		// Name of user who created this listener
		CreatedBy string

		// Last modified at timestamp in epoch milliseconds
		LastModifiedAt int64

		// Name of user who last updated this listener
		LastModifiedBy string
	}

	// Listeners holds one or more Listeners
	Listeners []Listener
)

var (
	// NullListener is an empty listener type
	NullListener = Listener{}

	// NullListeners is an empty listener slice
	NullListeners = Listeners{}
)

// listener specific attributes
const (
	AttributeListenerFilters = "Filters"

	// File for storing access logs
	AttributeAccessLogFile = "AccessLogFile"

	// Field configuration for access logging to file
	AttributeAccessLogFileFields = "AccessLogFileFields"

	// Cluster to send access logs to
	AttributeAccessLogCluster = "AccessLogCluster"

	// In memory buffer size for access logs
	AttributeAccessLogClusterBufferSize = "AccessLogClusterBufferSize"

	// Server name to respond with
	AttributeServerName = "ServerName"

	// HTTP/2 max concurrent streams per connection
	AttributeMaxConcurrentStreams = "MaxConcurrentStreams"

	// HTTP/2 initial connection window size
	AttributeInitialConnectionWindowSize = "InitialConnectionWindowSize"

	// HTTP/2 initial window size
	AttributeInitialStreamWindowSize = "InitialStreamWindowSize"

	// Name of extzauth cluster
	AttributeExtAuthzCluster = "ExtAuthzCluster"

	// Extauthz cluster request timeout
	AttributeExtAuthzTimeout = "ExtAuthzTimeout"

	// Are requests allowed in case authentication times out
	AttributeExtAuthzFailureModeAllow = "ExtAuthzFailureModeAllow"

	// Number of bytes of POST request to include in authentication request
	AttributeExtAuthzRequestBodySize = "ExtAuthzRequestBodySize"

	// Ratelimiting
	AttributeRateLimitingCluster = "RateLimitingCluster"

	//
	AttributeRateLimitingTimeout = "RateLimitingTimeout"

	//
	AttributeRateLimitingDomain = "RateLimitingDomain"

	//
	AttributeRateLimitingFailureModeAllow = "RateLimitingFailureModeAllow"
)

// Attributes which are shared amongst listener, route and cluster
const (
	// AttributeTLSCertificate holds pem encoded certicate
	AttributeTLSCertificate = "TLSCertificate"

	// AttributeTLSCertificateKey holds certicate key
	AttributeTLSCertificateKey = "TLSCertificateKey"

	// AttributeTLSCertificateFile holds filename of pem encoded certicate
	AttributeTLSCertificateFile = "TLSCertificateFile"

	// AttributeTLSCertificateKeyFile holds filename of certicate key
	AttributeTLSCertificateKeyFile = "TLSCertificateKeyFile"

	// AttributeTLSMinimumVersion determines minimum TLS version accepted
	AttributeTLSMinimumVersion = "TLSMinimumVersion"

	// AttributeTLSMaximumVersion determines maximum TLS version accepted
	AttributeTLSMaximumVersion = "TLSMaximumVersion"

	// AttributeTLSCipherSuites determines set of allowed TLS ciphers
	AttributeTLSCipherSuites = "TLSCipherSuites"

	// AttributeTLSCipherSuites sets HTTP protocol to accept
	AttributeHTTPProtocol = "HTTPProtocol"

	AttributeValueTrue                    = "true"
	AttributeValueFalse                   = "false"
	AttributeValueTLSVersion10            = "TLS1.0"
	AttributeValueTLSVersion11            = "TLS1.1"
	AttributeValueTLSVersion12            = "TLS1.2"
	AttributeValueTLSVersion13            = "TLS1.3"
	AttributeValueHTTPProtocol11          = "HTTP/1.1"
	AttributeValueHTTPProtocol2           = "HTTP/2"
	AttributeValueHTTPProtocol3           = "HTTP/3"
	AttributeValueHealthCheckProtocolHTTP = "HTTP"
)

// Sort a slice of listeners
func (listeners Listeners) Sort() {
	// Sort listeners by routegroup, paths
	sort.SliceStable(listeners, func(i, j int) bool {
		if listeners[i].Port == listeners[j].Port {
			return listeners[i].Name < listeners[j].Name
		}
		return listeners[i].Port < listeners[j].Port
	})
}

// ConfigCheck checks if a listener's configuration is correct
func (l *Listener) ConfigCheck() error {

	for _, attribute := range l.Attributes {
		if !validListenerAttributes[attribute.Name] {
			return fmt.Errorf("unknown attribute '%s'", attribute.Name)
		}
	}
	return nil
}

// validListenerAttributes contains all valid attribute names for a listener
var validListenerAttributes = map[string]bool{
	AttributeListenerFilters:             true,
	AttributeAccessLogFile:               true,
	AttributeAccessLogCluster:            true,
	AttributeAccessLogClusterBufferSize:  true,
	AttributeHTTPProtocol:                true,
	AttributeTLS:                         true,
	AttributeTLSMinimumVersion:           true,
	AttributeTLSMaximumVersion:           true,
	AttributeTLSCertificate:              true,
	AttributeTLSCertificateFile:          true,
	AttributeTLSCertificateKey:           true,
	AttributeTLSCertificateKeyFile:       true,
	AttributeTLSCipherSuites:             true,
	AttributeServerName:                  true,
	AttributeMaxConcurrentStreams:        true,
	AttributeInitialConnectionWindowSize: true,
	AttributeInitialStreamWindowSize:     true,
}
