package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/erikbos/apiauth/pkg/shared"

	"github.com/bmatcuk/doublestar"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type"
	"github.com/gogo/googleapis/google/rpc"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
)

var apiProxyBasePaths = map[string][]string{
	"Petstore": {"/pets/2", "/pets/3"},
	"Cycling":  {"/search/1"},
	"People":   {"/people"},
}

// CheckAllowedPath check if apicall is allowed:
// - retrieve apikey details
// - 	check if key is approved
// - iterate over products in apikey
// - 	check if product is approved
// -    check if product is deployed in same org as apikey
// - 	iterate over resource path(s) of each product:
// - 		if requestor path matches (proxyBaseBath + resource_path)
// -			- return 200
// - if not 403
//
// FIXME this should be replaced with:
// loading all known paths in a radix tree
// radix tree entry contains path property struct { whichproduct, enabled, etc}
//
func (a *authorizationServer) CheckAllowedPath(requestURIPath, organization, apiKey string) (int, shared.AppCredential, shared.DeveloperApp, shared.APIProduct, error) {
	appcredential, err := a.c.GetAppCredentialCached(a.db, organization, apiKey)
	if err != nil {
		return 403, shared.AppCredential{}, shared.DeveloperApp{}, shared.APIProduct{}, errors.New("Could not find apikey")
	}

	// we immediately lookup developer app as we always needs its information
	// (e.g. logging with as much customer detail as possible)
	developerApp, err := a.c.GetDeveloperAppCached(a.db, appcredential.OrganizationAppID)
	if err != nil {
		return 403, appcredential, shared.DeveloperApp{}, shared.APIProduct{}, errors.New("Could not find developer app of apikey")
	}
	if appcredential.Status != "approved" {
		return 403, appcredential, developerApp, shared.APIProduct{}, errors.New("Unapproved apikey")
	}
	if appcredential.ExpiresAt != -1 {
		currentTime := time.Now().UnixNano() / 1000000
		// fmt.Printf("Current time: %d\n", currentTime)
		// fmt.Printf("Expires time: %d\n", appcredential.expires_at)
		if currentTime > appcredential.ExpiresAt {
			return 403, appcredential, developerApp, shared.APIProduct{}, errors.New("Expired apikey")
		}
	}
	// iterate over this key's apiproducts entitlement
	for productIndex := range appcredential.APIProducts {
		// Check if product has been approved
		if appcredential.APIProducts[productIndex].Status == "approved" {
			// retrieve each api product
			apiproduct, err := a.c.GetAPIProductCached(a.db, organization, appcredential.APIProducts[productIndex].Apiproduct)
			log.Debugf("CheckPath() evaluation apiproduct: %+v\n", appcredential.APIProducts[productIndex].Apiproduct)

			if err != nil {
				// FIXME should we continue in case a single product is not retrievable?
				return 503, appcredential, developerApp, shared.APIProduct{}, errors.New("Cannot retrieve product(s) of apikey")
			}
			// FIX ME/TBC should we skip product if it deployed in different org?
			// (very) unlikely scenario?
			if appcredential.OrganizationName != apiproduct.OrganizationName {
				log.Debugf("CheckPath() organization mismatch: %s != %s",
					appcredential.OrganizationName, apiproduct.OrganizationName)
				continue
			}
			// Get basePaths of this proxy as we need it as prefix for the whitelisted
			// resource_path(s)
			basePaths, ok := apiProxyBasePaths[apiproduct.Name]
			if ok {
				// Try if we have match for each of the basePath defined
				for basePathIndex := range basePaths {
					// Iterate over resource_paths of apiproduct to check if URI of
					// requestor matches one of these paths
					log.Debugf("CheckPath() evaluation basepaths: %s <= %+v", basePaths[basePathIndex], apiproduct.APIResources)
					for index := range apiproduct.APIResources {
						proxyURIPath := basePaths[basePathIndex] + apiproduct.APIResources[index]
						log.Debugf("CheckPath() Testing proxypath of %s/%s in (%s)", apiproduct.Name, proxyURIPath, requestURIPath)
						// handle resource_path equals /
						if apiproduct.APIResources[index] == "/" && proxyURIPath == requestURIPath {
							// log.Debugf("CheckAllowedPath: match on resource root")
							return 200, appcredential, developerApp, apiproduct, nil
						}
						// handle resource_path equals /* or /**
						ok, _ := doublestar.Match(proxyURIPath, requestURIPath)
						if ok {
							// log.Debugf("CheckAllowedPath: match on resource wildcard")
							return 200, appcredential, developerApp, apiproduct, nil
						}
					}
					// in case none of the resource paths matches, but the basepath is the prefix of the uri..
					// (could also exact match on base path!)
					if strings.HasPrefix(requestURIPath, basePaths[basePathIndex]) {
						return 404, appcredential, developerApp, shared.APIProduct{}, errors.New("access to product endpoint denied")
					}
				}
			} else {
				return 503, appcredential, developerApp, shared.APIProduct{}, errors.New("Cannot find basepath of product(s) of apikey")
			}
		}
		// Else, skipping unapproved product included in apikey
	}
	return 404, appcredential, developerApp, shared.APIProduct{}, errors.New("Cannot find product")
}

//lookUpAttribute find one named attribute in array of attributes (developer or developerapp)

func lookUpAttribute(attributes []shared.AttributeKeyValues, requestedAttributeName string) string {
	for attributeIndex := range attributes {
		if attributes[attributeIndex].Name == requestedAttributeName {
			return attributes[attributeIndex].Value
		}
	}
	return ""
}

// getQPSQuotaKeyAndLimit returns QPS quotakey to be used by Lyft ratelimiter
// QPS set as developer app attribute has priority over quota set as product attribute
//
func getQPSQuotaKeyAndLimit(apiKey string, apiproduct shared.APIProduct, developerapp shared.DeveloperApp) (string, string) {
	quotaAttributeName := apiproduct.Name + "_quotaPerSecond"
	// QPS set as developer app attribute has priority over quota set as product attribute

	value := lookUpAttribute(developerapp.Attributes, quotaAttributeName)
	if value != "" {
		return apiKey + "_a_" + quotaAttributeName, value
	}
	value = lookUpAttribute(apiproduct.Attributes, quotaAttributeName)
	if value != "" {
		return apiKey + "_p_" + quotaAttributeName, value
	}
	return "", ""
}

// allowCall answers gRPC to allow HTTP request
//
// message: body to return by auth module to HTTP requestor

func startGRPCAuthenticationServer(a authorizationServer) {
	a.authLatencyHistogram = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name:       "apiauth_request_latency",
			Help:       "Authentication latency in seconds.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		})
	prometheus.MustRegister(a.authLatencyHistogram)

	lis, err := net.Listen("tcp", a.config.AuthGRPCListen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("GRPC listening on %s", lis.Addr())

	grpcServer := grpc.NewServer()
	auth.RegisterAuthorizationServer(grpcServer, &a)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

//Check gets called by Envoy to authenticate & authorize a request
//
func (a *authorizationServer) Check(ctx context.Context, req *auth.CheckRequest) (*auth.CheckResponse, error) {

	timer := prometheus.NewTimer(a.authLatencyHistogram)
	defer timer.ObserveDuration()

	httpRequest := req.Attributes.Request.Http
	//	ttmethod, ok := httpRequest.Headers["ttmethod"]

	URL, _ := url.ParseRequestURI(httpRequest.Path)
	queryParameters, _ := url.ParseQuery(URL.RawQuery)
	apiKey := queryParameters["apikey"][0]

	log.Debugf("Check() rx path: %s", httpRequest.Path)
	log.Debugf("Check() rx uri: %s", URL.Path)
	log.Debugf("Check() rx apikey: %s", apiKey)
	for k, v := range httpRequest.Headers {
		log.Debugf("Check() rx header [%s] = %s", k, v)
	}

	if value, ok := httpRequest.Headers["x-forwarded-for"]; ok {
		country, state := a.g.GetCountryAndState(value)
		log.Debugf("Check() rx ip country: %s, state: %s", country, state)
	}

	log.Printf("Query String: %v", req.GetAttributes().GetRequest().GetHttp().GetQuery())

	// We staticly pass "petstore" as organization (should be derived from vhost -> org mapping)
	// Lookup apikey, developApp, and apiproduct and try to path match URL with allowed paths
	statusCode, appCrendentials, developerApp, APIProduct, err := a.CheckAllowedPath("petstore", URL.Path, apiKey)

	log.Debugf("Check() statuscode: %d", statusCode)
	log.Debugf("Check() appCrendentials: %+v\n", appCrendentials)
	log.Debugf("Check() developerApp: %+v\n", developerApp)
	log.Debugf("Check() APIProduct: %+v\n", APIProduct)

	if statusCode != 200 {
		log.Debugf("Check() CheckAllowedPath rejected path: %s", err.Error())

		return rejectCall(envoy_type.StatusCode_Forbidden, err.Error())
	}

	// We are going to allow this request, let's build the quotakey
	quotaKey, quotaLimit := getQPSQuotaKeyAndLimit(apiKey, APIProduct, developerApp)
	log.Debugf("Check() Quotakey: %s, %s", quotaKey, quotaLimit)
	log.Debugf("Check() CheckAllowedPath allowed path %s", URL.Path)
	return allowCall("")
}

func allowCall(message string) (*auth.CheckResponse, error) {
	return &auth.CheckResponse{
		Status: &status.Status{
			Code: int32(rpc.OK),
		},
		HttpResponse: &auth.CheckResponse_OkResponse{
			OkResponse: &auth.OkHttpResponse{
				Headers: []*core.HeaderValueOption{
					{
						Header: &core.HeaderValue{
							Key:   "developerApp",
							Value: "42",
						},
					},
				},
			},
		},
	}, nil
}

// rejectCall answers gRPC to deny HTTP request
//
// statusCode:	401 = envoy_type.StatusCode_Unauthorized
//				403 = envoy_type.StatusCode_Forbidden
//				503 = envoy_type.StatusCode_ServiceUnavailable
// (https://github.com/envoyproxy/envoy/blob/master/source/common/http/codes.cc)
//
// message: body to return by auth module to HTTP requestor

func rejectCall(statusCode envoy_type.StatusCode, message string) (*auth.CheckResponse, error) {
	errorPayload := fmt.Sprintf("{ %s }", message)

	return &auth.CheckResponse{
		Status: &status.Status{
			Code: int32(rpc.UNAUTHENTICATED),
		},
		HttpResponse: &auth.CheckResponse_DeniedResponse{
			DeniedResponse: &auth.DeniedHttpResponse{
				Status: &envoy_type.HttpStatus{
					Code: statusCode,
				},
				Headers: []*core.HeaderValueOption{
					{
						Header: &core.HeaderValue{
							Key:   "content-type",
							Value: "json",
						},
					},
				},
				Body: errorPayload,
			},
		},
	}, nil
}
