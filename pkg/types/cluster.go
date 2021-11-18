package types

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-playground/validator/v10"
)

// Cluster holds configuration of an upstream cluster
//
// Field validation (binding) is done using https://godoc.org/github.com/go-playground/validator
type (
	Cluster struct {
		// Name of cluster (not changable)
		Name string `validate:"required,min=1"`

		// Friendly display name of cluster
		DisplayName string

		// Attributes of this cluster
		Attributes Attributes

		// Created at timestamp in epoch milliseconds
		CreatedAt int64

		// Name of user who created this cluster
		CreatedBy string

		// Last modified at timestamp in epoch milliseconds
		LastModifiedAt int64

		// Name of user who last updated this cluster
		LastModifiedBy string
	}

	// Clusters holds one or more clusters
	Clusters []Cluster
)

var (
	// NullCluster is an empty cluster type
	NullCluster = Cluster{}

	// NullClusters is an empty cluster slice
	NullClusters = Clusters{}
)

// Cluster specific attributes
const (
	// Hostname of cluster
	AttributeHost = "Host"

	// Port of cluster
	AttributePort = "Port"

	// Timeout for new network connections to cluster
	AttributeConnectTimeout = "ConnectTimeout"

	// Tdle timeout for connections, for the period in which there are no active requests
	AttributeIdleTimeout = "IdleTimeout"

	// Determines whether to enable TLS or not, HTTP/2 always uses TLS
	AttributeTLS = "TLS"

	// Holds hostname to send during TLS handshake (if not set a cluster's hostname will be used)
	AttributeSNIHostName = "SNIHostName"

	// Sets network protocol to use for health check
	AttributeHealthCheckProtocol = "HealthCheckProtocol"

	// Determines host header to use for health check
	AttributeHealthHostHeader = "HealthCheckHostHeader"

	// Determines http path of health check probe
	AttributeHealthCheckPath = "HealthCheckPath"

	// Health check interval for probes
	AttributeHealthCheckInterval = "HealthCheckInterval"

	// Health check timeout
	AttributeHealthCheckTimeout = "HealthCheckTimeout"

	// Threshold of attempts before declaring cluster unhealthly
	AttributeHealthCheckUnhealthyThreshold = "HealthCheckUnhealthyThreshold"

	// Threshold of attempts before declaring cluster healthly
	AttributeHealthCheckHealthyThreshold = "HealthCheckHealthyThreshold"

	// Logfile name for healthcheck probes
	AttributeHealthCheckLogFile = "HealthCheckLogFile"

	// Maximum number of connects to cluster
	AttributeMaxConnections = "MaxConnections"

	// Maximum number of pending cluster requests
	AttributeMaxPendingRequests = "MaxPendingRequests"

	// Maximum number of parallel requests to cluster
	AttributeMaxRequests = "MaxRequests"

	// Maximum number of retries to cluster
	AttributeMaxRetries = "MaxRetries"

	// IP network address family to use for contacting cluster
	AttributeDNSLookupFamily = "DNSLookupFamily"

	// dns resolving using v4 only
	AttributeValueDNSIPV4Only = "V4_ONLY"

	// dns resolving using v6 only
	AttributeValueDNSIPV6Only = "V6_ONLY"

	// dns resolving via both v4 & v6
	AttributeValueDNSAUTO = "AUTO"

	// Refresh rate for resolving cluster hostname
	AttributeDNSRefreshRate = "DNSRefreshRate"

	// Resolver ip address(es) to use for dns resolution (multiple can be comma separated)
	AttributeDNSResolvers = "DNSResolvers"

	AttributeLbPolicy            = "LbPolicy"
	AttributeValueLBRoundRobin   = "ROUND_ROBIN"
	AttributeValueLBLeastRequest = "LEAST_REQUEST"
	AttributeValueLBRingHash     = "RING_HASH"
	AttributeValueLBRandom       = "RANDOM"
	AttributeValueLBMaglev       = "MAGLEV"

	// Default connection timeout
	DefaultClusterConnectTimeout = 5 * time.Second

	// Default cluster connect idle timeout
	DefaultClusterIdleTimeout = 15 * time.Minute

	// Default health check interval
	DefaultHealthCheckInterval = 5 * time.Second

	// Default health check timeout
	DefaultHealthCheckTimeout = 10 * time.Second

	// Default unhealthy threshold
	DefaultHealthCheckUnhealthyThreshold = 2

	// Default healthy threshold
	DefaultHealthCheckHealthyThreshold = 2

	// Default dns resolution interval of cluster hostname
	DefaultDNSRefreshRate = 5 * time.Second
)

// Sort orders a slice of clusters
func (clusters Clusters) Sort() {
	// Sort clusters by name
	sort.SliceStable(clusters, func(i, j int) bool {
		return clusters[i].Name < clusters[j].Name
	})
}

// Validate checks if a cluster's configuration is correct
func (c *Cluster) Validate() error {

	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return err
	}
	for _, attribute := range c.Attributes {
		if !validClusterAttributes[attribute.Name] {
			return fmt.Errorf("unknown attribute '%s'", attribute.Name)
		}
	}
	return nil
}

// validClusterAttributes contains all valid attribute names for a cluster
var validClusterAttributes = map[string]bool{
	AttributeHost:                          true,
	AttributePort:                          true,
	AttributeConnectTimeout:                true,
	AttributeIdleTimeout:                   true,
	AttributeDNSLookupFamily:               true,
	AttributeDNSRefreshRate:                true,
	AttributeDNSResolvers:                  true,
	AttributeTLS:                           true,
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
