package types

import (
	"fmt"
	"sort"
	"time"
)

// Route holds configuration of a route
type Route struct {
	// Name of route (not changable)
	Name string `json:"name"`

	// Friendly display name of route
	DisplayName string `json:"displayName"`

	// Routegroup this route is part of
	RouteGroup string `json:"RouteGroup"`

	// Path of route
	Path string `json:"path"`

	// Type of pathmatching: path, prefix, regexp
	PathType string `json:"pathType"`

	// Attributes of this route
	Attributes Attributes `json:"attributes"`

	// Created at timestamp in epoch milliseconds
	CreatedAt int64 `json:"createdAt"`

	// Name of user who created this route
	CreatedBy string `json:"createdBy"`

	// Last modified at timestamp in epoch milliseconds
	LastmodifiedAt int64 `json:"lastmodifiedAt"`

	// Name of user who last updated this route
	LastmodifiedBy string `json:"lastmodifiedBy"`
}

// Routes holds one or more routes
type Routes []Route

var (
	// NullRoute is an empty route type
	NullRoute = Route{}

	// NullRoutes is an empty route slice
	NullRoutes = Routes{}
)

// Attributes supported on a route
const (
	// Name of upstream cluster to forward requests to
	AttributeCluster = "Cluster"

	// Weighted list of clusters to load balance requests across
	AttributeWeightedClusters = "WeightedClusters"

	// Enable authentication via extauthz
	AttributeAuthentication = "Authentication"

	// Enable ratelimiting
	AttributeRateLimiting = "RateLimiting"

	// Return an arbitrary HTTP response directly, without proxying
	AttributeDirectResponseStatusCode = "DirectResponseStatusCode"

	// Responsebody to return when direct response is done
	AttributeDirectResponseBody = "DirectResponseBody"

	// Return an HTTP redirect
	AttributeRedirectStatusCode = "RedirectStatusCode"

	// HTTP scheme when generating a redirect
	AttributeRedirectScheme = "RedirectScheme"

	// Hostname when generating a redirect
	AttributeRedirectHostName = "RedirectHostName"

	// Port when generating a redirect
	AttributeRedirectPort = "RedirectPort"

	// Path when generating a redirect
	AttributeRedirectPath = "RedirectPath"

	// Enable removal of query parameters when redirecting
	AttributeRedirectStripQuery = "RedirectStripQuery"

	// Rewrites path when contacting upstream
	AttributePrefixRewrite = "PrefixRewrite"

	// Whether the resource allows credentials
	AttributeCORSAllowCredentials = "CORSAllowCredentials"

	// Value of Access-Control-Allow-Methods header
	AttributeCORSAllowMethods = "CORSAllowMethods"

	// Value of Access-Control-Allow-Headers header
	AttributeCORSAllowHeaders = "CORSAllowHeaders"

	// Value of Access-Control-Expose-Headers header
	AttributeCORSExposeHeaders = "CORSExposeHeaders"

	// Value of Access-Control-Expose-Headers header
	AttributeCORSMaxAge = "CORSMaxAge"

	// Host header to set when forwarding to upstream cluster
	AttributeHostHeader = "HostHeader"

	// Additional header(s) to set before forwarding upstream
	AttributeHeaders = "Headers"

	// Basic authentication header to set before forwarding upstream
	AttributeBasicAuth = "BasicAuth"

	// Conditions under which retry takes place
	AttributeRetryOn = "RetryOn"

	// Upstream timeout per retry attempt
	AttributePerTryTimeout = "PerTryTimeout"

	// Max number of retry attempts
	AttributeNumRetries = "NumRetries"

	// Upstream status codes which are to be retried
	AttributeRetryOnStatusCodes = "RetryOnStatusCodes"

	// Cluster to mirror requests to
	AttributeRequestMirrorCluster = "RequestMirrorCluster"

	// Percentage of traffic traffic to mirror requests to
	AttributeRequestMirrorPercentage = "RequestMirrorPercentage"

	// Tineout for cluster communication
	AttributeTimeout = "Timeout"

	// Default route timeout
	DefaultRouteTimeout = 20 * time.Second
	// Default per retry timeout
	DefaultPerRetryTimeout = 500 * time.Millisecond

	// Default retry-on status codes
	DefaultRetryStatusCodes = "500,503,504"
)

// Sort orders a slice of routes
func (routes Routes) Sort() {
	// Sort routes by routegroup, paths
	sort.SliceStable(routes, func(i, j int) bool {
		if routes[i].RouteGroup == routes[j].RouteGroup {
			return routes[i].Path < routes[j].Path
		}
		return routes[i].RouteGroup < routes[j].RouteGroup
	})
}

// ConfigCheck checks if a route's configuration is correct
func (r *Route) ConfigCheck() error {

	for _, attribute := range r.Attributes {
		if !validRouteAttributes[attribute.Name] {
			return fmt.Errorf("Unknown attribute '%s'", attribute.Name)
		}
	}
	return nil
}

// validRouteAttributes contains all valid attribute names for a route
var validRouteAttributes = map[string]bool{
	AttributeCluster:                  true,
	AttributeWeightedClusters:         true,
	AttributeAuthentication:           true,
	AttributeRateLimiting:             true,
	AttributeDirectResponseStatusCode: true,
	AttributeDirectResponseBody:       true,
	AttributePrefixRewrite:            true,
	AttributeCORSAllowCredentials:     true,
	AttributeCORSAllowMethods:         true,
	AttributeCORSAllowHeaders:         true,
	AttributeCORSExposeHeaders:        true,
	AttributeCORSMaxAge:               true,
	AttributeHostHeader:               true,
	AttributeHeaders:                  true,
	AttributeBasicAuth:                true,
	AttributeRetryOn:                  true,
	AttributePerTryTimeout:            true,
	AttributeNumRetries:               true,
	AttributeRetryOnStatusCodes:       true,
	AttributeRequestMirrorCluster:     true,
	AttributeRequestMirrorPercentage:  true,
	AttributeRedirectStatusCode:       true,
	AttributeRedirectScheme:           true,
	AttributeRedirectHostName:         true,
	AttributeRedirectPort:             true,
	AttributeRedirectPath:             true,
	AttributeRedirectStripQuery:       true,
	AttributeTimeout:                  true,
}
