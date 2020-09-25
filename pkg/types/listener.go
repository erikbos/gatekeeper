package types

import (
	"fmt"
	"sort"
)

// Listener contains everything about downstream configuration of listener and http virtual hosts
type Listener struct {
	// Name of listener (not changable)
	Name string `json:"name"`

	// Friendly display name of listener
	DisplayName string `json:"displayName"`

	// Virtual hosts of this listener
	VirtualHosts StringSlice `json:"virtualHosts"`

	// tcp port to listen on
	Port int `json:"port"`

	// Routegroup to forward traffic to
	RouteGroup string `json:"routeGroup"`

	// Comma separated list of policynames, to apply to requests
	Policies string `json:"policies"`

	// Attributes of this listener
	Attributes Attributes `json:"attributes"`

	// Organization this listener belongs to (not used)
	OrganizationName string `json:"organizationName"`

	// Created at timestamp in epoch milliseconds
	CreatedAt int64 `json:"createdAt"`
	// Name of user who created this listener

	CreatedBy string `json:"createdBy"`
	// Last modified at timestamp in epoch milliseconds
	LastmodifiedAt int64 `json:"lastmodifiedAt"`

	// Name of user who last updated this listener
	LastmodifiedBy string `json:"lastmodifiedBy"`
}

// Listeners holds one or more Listeners
type Listeners []Listener

var (
	// NullListener is an empty listener type
	NullListener = Listener{}

	// NullListeners is an empty listener slice
	NullListeners = Listeners{}
)

// listener specific attributes
const (
	// File for storing access logs
	AttributeAccessLogFile = "AccessLogFile"

	// Cluster to send access logs to
	AttributeAccessLogCluster = "AccessLogCluster"
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
			return fmt.Errorf("Unknown attribute '%s'", attribute.Name)
		}
	}
	return nil
}

// validListenerAttributes contains all valid attribute names for a listener
var validListenerAttributes = map[string]bool{
	AttributeAccessLogFile:     true,
	AttributeAccessLogCluster:  true,
	AttributeHTTPProtocol:      true,
	AttributeTLSEnable:         true,
	AttributeTLSMinimumVersion: true,
	AttributeTLSMaximumVersion: true,
	AttributeTLSCertificate:    true,
	AttributeTLSCertificateKey: true,
	AttributeTLSCipherSuites:   true,
}
