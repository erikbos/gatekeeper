package types

import (
	"fmt"
	"sort"
)

// Listener contains everything about downstream configuration of listener and http virtual hosts
type Listener struct {
	Name             string      `json:"name"`             // Name of listener (not changable)
	DisplayName      string      `json:"displayName"`      // Friendly display name of listener
	VirtualHosts     StringSlice `json:"virtualHosts"`     // Virtual hosts of this listener
	Port             int         `json:"port"`             // tcp port to listen on
	RouteGroup       string      `json:"routeGroup"`       // Routegroup to forward traffic to
	Policies         string      `json:"policies"`         // Comma separated list of policynames, to apply to requests
	Attributes       Attributes  `json:"attributes"`       // Attributes of this listener
	OrganizationName string      `json:"organizationName"` // Organization this listener belongs to (not used)
	CreatedAt        int64       `json:"createdAt"`        // Created at timestamp in epoch milliseconds
	CreatedBy        string      `json:"createdBy"`        // Name of user who created this listener
	LastmodifiedAt   int64       `json:"lastmodifiedAt"`   // Last modified at timestamp in epoch milliseconds
	LastmodifiedBy   string      `json:"lastmodifiedBy"`   // Name of user who last updated this listener
}

// Listeners holds one or more Listeners
type Listeners []Listener

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
