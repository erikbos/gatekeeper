package types

import (
	"fmt"
	"sort"
	"time"
)

// Cluster holds configuration of an upstream cluster
type Cluster struct {
	Name           string     `json:"name"`           // Name of cluster (not changable)
	DisplayName    string     `json:"displayName"`    // Friendly display name of cluster
	HostName       string     `json:"hostName"`       // Hostname of cluster
	Port           int        `json:"port"`           // tcp port of cluster
	Attributes     Attributes `json:"attributes"`     // Attributes of this cluster
	CreatedAt      int64      `json:"createdAt"`      // Created at timestamp in epoch milliseconds
	CreatedBy      string     `json:"createdBy"`      // Name of user who created this cluster
	LastmodifiedAt int64      `json:"lastmodifiedAt"` // Last modified at timestamp in epoch milliseconds
	LastmodifiedBy string     `json:"lastmodifiedBy"` // Name of user who last updated this cluster
}

// Clusters holds one or more clusters
type Clusters []Cluster

const (
	// AttributeConnectTimeout is timeout for new network connections to cluster
	AttributeConnectTimeout = "ConnectTimeout"

	// AttributeIdleTimeout is idle timeout for connections, for the period in which there are no active requests
	AttributeIdleTimeout = "IdleTimeout"

	// AttributeTLSEnable determines whether to enable TLS or not, HTTP/2 always uses TLS
	AttributeTLSEnable = "TLSEnable"

	// AttributeSNIHostName holds hostname to send during TLS handshake (if not set a cluster's hostname will be used)
	AttributeSNIHostName = "SNIHostName"

	// AttributeHealthCheckProtocol determines network protocol to use for health check
	AttributeHealthCheckProtocol = "HealthCheckProtocol"

	// AttributeHealthHostHeader determines host header to use for health check
	AttributeHealthHostHeader = "HealthCheckHostHeader"

	// AttributeHealthCheckPath determines http path of health check probe
	AttributeHealthCheckPath = "HealthCheckPath"

	// AttributeHealthCheckInterval determines health check interval
	AttributeHealthCheckInterval = "HealthCheckInterval"

	// AttributeHealthCheckTimeout determines health check timeout
	AttributeHealthCheckTimeout = "HealthCheckTimeout"

	// AttributeHealthCheckUnhealthyThreshold sets threshold of attempts before declaring cluster unhealthly
	AttributeHealthCheckUnhealthyThreshold = "HealthCheckUnhealthyThreshold"

	// AttributeHealthCheckHealthyThreshold sets threshold of attempts before declaring cluster healthly
	AttributeHealthCheckHealthyThreshold = "HealthCheckHealthyThreshold"

	// AttributeHealthCheckLogFile determines logfile name for healthcheck probes
	AttributeHealthCheckLogFile = "HealthCheckLogFile"

	// AttributeMaxConnections determines maximum number of connects to cluster
	AttributeMaxConnections = "MaxConnections"

	// AttributeMaxPendingRequests sets maximum number of pending cluster requests
	AttributeMaxPendingRequests = "MaxPendingRequests"

	// AttributeMaxRequests sets maximum number of parallel requests to cluster
	AttributeMaxRequests = "MaxRequests"

	// AttributeMaxRetries sets maximum number of retries to cluster
	AttributeMaxRetries = "MaxRetries"

	// AttributeDNSLookupFamiliy sets IP network address family to use for contacting cluster
	AttributeDNSLookupFamiliy = "DNSLookupFamily"

	// AttributeValueDNSIPV4Only set IPv4 dns resolving
	AttributeValueDNSIPV4Only = "V4_ONLY"

	// AttributeValueDNSIPV6Only set IPv6 dns resolving
	AttributeValueDNSIPV6Only = "V6_ONLY"

	// AttributeValueDNSAUTO set IPv4 or IPv6 dns resolving
	AttributeValueDNSAUTO = "AUTO"

	// AttributeDNSRefreshRate sets refresh rate for resolving cluster hostname
	AttributeDNSRefreshRate = "DNSRefreshRate"

	// AttributeDNSResolvers sets resolver ip address to resolve cluster hostname (multiple can be comma separated)
	AttributeDNSResolvers = "DNSResolvers"

	AttributeLbPolicy            = "LbPolicy"
	AttributeValueLBRoundRobin   = "ROUND_ROBIN"
	AttributeValueLBLeastRequest = "LEAST_REQUEST"
	AttributeValueLBRingHash     = "RING_HASH"
	AttributeValueLBRandom       = "RANDOM"
	AttributeValueLBMaglev       = "MAGLEV"

	// DefaultClusterConnectTimeout holds default connection timeout
	DefaultClusterConnectTimeout = 5 * time.Second

	// DefaultClusterIdleTimeout holds default cluster connect idle timeout
	DefaultClusterIdleTimeout = 15 * time.Minute

	// DefaultHealthCheckInterval holds default health check interval
	DefaultHealthCheckInterval = 5 * time.Second

	// DefaultHealthCheckTimeout holds default health check timeout
	DefaultHealthCheckTimeout = 10 * time.Second

	// DefaultHealthCheckUnhealthyThreshold holds default unhealthy threshold
	DefaultHealthCheckUnhealthyThreshold = 2

	// DefaultHealthCheckHealthyThreshold holds default healthy threshold
	DefaultHealthCheckHealthyThreshold = 2

	// DefaultDNSRefreshRate holds default dns resolution interval of cluster hostname
	DefaultDNSRefreshRate = 5 * time.Second
)

// Sort orders a slice of clusters
func (clusters Clusters) Sort() {
	// Sort clusters by name
	sort.SliceStable(clusters, func(i, j int) bool {
		return clusters[i].Name < clusters[j].Name
	})
}

// ConfigCheck checks if a cluster's configuration is correct
func (c *Cluster) ConfigCheck() error {

	for _, attribute := range c.Attributes {
		if !validClusterAttributes[attribute.Name] {
			return fmt.Errorf("Unknown attribute '%s'", attribute.Name)
		}
	}
	return nil
}

// validClusterAttributes contains all valid attribute names for a cluster
var validClusterAttributes = map[string]bool{
	AttributeConnectTimeout:                true,
	AttributeIdleTimeout:                   true,
	AttributeDNSLookupFamiliy:              true,
	AttributeDNSRefreshRate:                true,
	AttributeDNSResolvers:                  true,
	AttributeTLSEnable:                     true,
	AttributeTLSMinimumVersion:             true,
	AttributeTLSMaximumVersion:             true,
	AttributeTLSCipherSuites:               true,
	AttributeHTTPProtocol:                  true,
	AttributeSNIHostName:                   true,
	AttributeLbPolicy:                      true,
	AttributeHealthCheckProtocol:           true,
	AttributeHealthCheckPath:               true,
	AttributeHealthCheckInterval:           true,
	AttributeHealthCheckTimeout:            true,
	AttributeHealthCheckUnhealthyThreshold: true,
	AttributeHealthCheckHealthyThreshold:   true,
	AttributeHealthCheckLogFile:            true,
	AttributeMaxConnections:                true,
	AttributeMaxPendingRequests:            true,
	AttributeMaxRequests:                   true,
	AttributeMaxRetries:                    true,
}
