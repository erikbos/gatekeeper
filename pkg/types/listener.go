package types

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Listener contains everything about downstream configuration of listener and http virtual hosts
type (
	Listener struct {
		// Name of listener (not changable)
		Name string `validate:"required,min=1"`

		// Friendly display name of listener
		DisplayName string

		// Virtual hosts of this listener (at least one, each value must be a fqdn)
		VirtualHosts []string `validate:"required,min=1,dive,fqdn"`

		// tcp port to listen on
		Port int `validate:"required,min=1,max=65535"`

		// Routegroup to forward traffic to
		RouteGroup string `validate:"required"`

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

	// Organization to be used for lookups by envoyauth when authentication requests
	AttributeOrganization = "Organization"

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
	for _, l := range listeners {
		l.Attributes.Sort()
	}
}

// Validate checks if a listener's configuration is correct
func (l *Listener) Validate() error {

	validate := validator.New()
	if err := validate.Struct(l); err != nil {
		return err
	}
	for _, attribute := range l.Attributes {
		if !validListenerAttributes[attribute.Name] {
			return fmt.Errorf("unknown attribute '%s'", attribute.Name)
		}
	}
	// scan for duplicate vhosts
	hostsSeen := make(map[string]bool, len(l.VirtualHosts))
	for _, host := range l.VirtualHosts {
		hostLower := strings.ToLower(host)
		if found := hostsSeen[hostLower]; !found {
			hostsSeen[hostLower] = true
		} else {
			return fmt.Errorf("no duplicate virtual hosts allowed (%s)", host)
		}
	}
	return nil
}

// validListenerAttributes contains all valid attribute names for a listener
var validListenerAttributes = map[string]bool{
	AttributeAccessLogCluster:             true,
	AttributeAccessLogClusterBufferSize:   true,
	AttributeAccessLogFile:                true,
	AttributeAccessLogFileFields:          true,
	AttributeExtAuthzCluster:              true,
	AttributeExtAuthzFailureModeAllow:     true,
	AttributeExtAuthzRequestBodySize:      true,
	AttributeExtAuthzTimeout:              true,
	AttributeHTTPProtocol:                 true,
	AttributeIdleTimeout:                  true,
	AttributeInitialConnectionWindowSize:  true,
	AttributeInitialStreamWindowSize:      true,
	AttributeListenerFilters:              true,
	AttributeMaxConcurrentStreams:         true,
	AttributeOrganization:                 true,
	AttributeRateLimitingCluster:          true,
	AttributeRateLimitingDomain:           true,
	AttributeRateLimitingFailureModeAllow: true,
	AttributeRateLimitingTimeout:          true,
	AttributeServerName:                   true,
	AttributeTLS:                          true,
	AttributeTLSCertificate:               true,
	AttributeTLSCertificateFile:           true,
	AttributeTLSCertificateKey:            true,
	AttributeTLSCertificateKeyFile:        true,
	AttributeTLSCipherSuites:              true,
	AttributeTLSMaximumVersion:            true,
	AttributeTLSMinimumVersion:            true,
}
