package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/bmatcuk/doublestar"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type"
	"github.com/erikbos/apiauth/pkg/shared"
	"github.com/gogo/googleapis/google/rpc"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
)

type sessionState struct {
	apikey        string
	appCredential shared.AppCredential
	developerApp  shared.DeveloperApp
	developer     shared.Developer
	APIProduct    shared.APIProduct
}

func startGRPCAuthorizationServer(a authorizationServer) {
	lis, err := net.Listen("tcp", a.config.AuthGRPCListen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("GRPC listening on %s", a.config.AuthGRPCListen)

	grpcServer := grpc.NewServer()
	auth.RegisterAuthorizationServer(grpcServer, &a)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Check (called by Envoy) to authenticate & authorize a HTTP request
func (a *authorizationServer) Check(ctx context.Context, req *auth.CheckRequest) (*auth.CheckResponse, error) {

	timer := prometheus.NewTimer(a.metrics.authLatencyHistogram)
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

	headers := make(map[string]string)
	if value, ok := httpRequest.Headers["x-forwarded-for"]; ok {
		country, state := a.g.GetCountryAndState(value)
		headers["geoip-country"] = country
		headers["geoip-state"] = state
		log.Debugf("Check() rx ip country: %s, state: %s", country, state)
	}

	log.Printf("Query String: %v", req.GetAttributes().GetRequest().GetHttp().GetQuery())

	entitlement, err := a.CheckProductEntitlement("petstore", URL.Path, apiKey)
	if err != nil {
		log.Debugf("Check() not allowed '%s' (%s)", URL.Path, err.Error())
		return checkRejectCall(envoy_type.StatusCode_Forbidden, nil, err.Error())
	}
	if entitlement.APIProduct.Scopes != "" {
		handlePolicies(entitlement, headers)
	}
	return checkAllowCall(headers)
}

func checkAllowCall(headers map[string]string) (*auth.CheckResponse, error) {
	response := &auth.CheckResponse{
		Status: &status.Status{
			Code: int32(rpc.OK),
		},
		HttpResponse: &auth.CheckResponse_OkResponse{
			OkResponse: &auth.OkHttpResponse{
				Headers: buildHeadersList(headers),
			},
		},
	}
	log.Printf("allowCall: %v", response)
	return response, nil
}

// rejectCall answers Envoy to deny HTTP request
//
// statusCode:	401 = envoy_type.StatusCode_Unauthorized
//				403 = envoy_type.StatusCode_Forbidden
//				503 = envoy_type.StatusCode_ServiceUnavailable
// (https://github.com/envoyproxy/envoy/blob/master/source/common/http/codes.cc)

func checkRejectCall(statusCode envoy_type.StatusCode,
	headers map[string]string, message string) (*auth.CheckResponse, error) {

	response := &auth.CheckResponse{
		Status: &status.Status{
			Code: int32(rpc.UNAUTHENTICATED),
		},
		HttpResponse: &auth.CheckResponse_DeniedResponse{
			DeniedResponse: &auth.DeniedHttpResponse{
				Status: &envoy_type.HttpStatus{
					Code: statusCode,
				},
				Headers: buildHeadersList(headers),
				Body:    buildJSONErrorMessage(message),
			},
		},
	}
	log.Printf("rejectCall: %v", response)
	return response, nil
}

// buildHeadersList creates map to hold additional upstream headers
func buildHeadersList(headers map[string]string) []*core.HeaderValueOption {
	if len(headers) == 0 {
		return nil
	}

	headerList := make([]*core.HeaderValueOption, 0, len(headers))
	for key, value := range headers {
		headerList = append(headerList, &core.HeaderValueOption{
			Header: &core.HeaderValue{
				Key:   key,
				Value: value,
			},
		})
	}
	return headerList
}

// JSONErrorMessage is the format for our error messages
const JSONErrorMessage = `{
 "message": "%s"
}
`

// returns a well structured JSON-formatted message
func buildJSONErrorMessage(message string) string {
	return fmt.Sprintf(JSONErrorMessage, message)
}

// CheckProductEntitlement
func (a *authorizationServer) CheckProductEntitlement(organization, requestPath,
	apiKey string) (sessionState, error) {

	session, err := a.getEntitlementDetails(organization, requestPath, apiKey)
	if err != nil {
		return session, err
	}
	if err = checkAppCredentialValidity(session.appCredential); err != nil {
		return session, err
	}
	session.APIProduct, err = a.IsRequestPathAllowed(organization, requestPath, session.appCredential)
	if err != nil {
		return session, errors.New("No product match")
	}
	return session, nil
}

// getEntitlementDetails returns full apikey and dev app details
func (a *authorizationServer) getEntitlementDetails(organization, requestPath, apiKey string) (sessionState, error) {
	session := sessionState{}

	session.apikey = apiKey

	var err error
	session.appCredential, err = a.db.GetAppCredentialByKey(organization, apiKey)
	if err != nil {
		// FIX ME increase unknown apikey counter (not an error state)
		return session, errors.New("Could not find apikey")
	}
	session.developerApp, err = a.db.GetDeveloperAppByID(organization, session.appCredential.OrganizationAppID)
	if err != nil {
		// FIX ME increase counter as every apikey should link to dev app (error state)
		return session, errors.New("Could not find developer app of this apikey")
	}
	session.developer, err = a.db.GetDeveloperByID(session.developerApp.ParentID)
	if err != nil {
		// FIX ME increase counter as every devapp should link to developer (error state)
		return session, errors.New("Could not find developer of this apikey")
	}
	return session, nil
}

// checkAppCredentialValidity checks devapp approval and expiry status
func checkAppCredentialValidity(appcredential shared.AppCredential) error {
	if appcredential.Status != "approved" {
		// FIXME increase unapproved counter (not an error state)
		return errors.New("Unapproved apikey")
	}
	if appcredential.ExpiresAt != -1 {
		if shared.GetCurrentTimeMilliseconds() > appcredential.ExpiresAt {
			// FIXME increase expired credentials counter (not an error state))
			return errors.New("Expired apikey")
		}
	}
	return nil
}

// IsRequestPathAllowed
// - iterate over products in apikey
// - 	iterate over resource path(s) of each product:
// - 		if requestor path matches apiresource_path(s)
// -			- return 200
// - if not 403

func (a *authorizationServer) IsRequestPathAllowed(organization, requestPath string,
	appcredential shared.AppCredential) (shared.APIProduct, error) {

	// Iterate over this key's apiproducts
	for _, apiproduct := range appcredential.APIProducts {
		if apiproduct.Status == "approved" {
			// Retrieve details of each api product embedded in key
			apiproduct, err := a.db.GetAPIProductByName(organization, apiproduct.Apiproduct)
			if err != nil {
				// FIXME increase unknown product in apikey counter (not an error state)
			} else {
				// Iterate over apiresource(paths) of apiproduct
				for _, productPath := range apiproduct.APIResources {
					log.Debugf("CheckPath() Matching path %s in %s", requestPath, productPath)
					if ok, _ := doublestar.Match(productPath, requestPath); ok {
						log.Debugf("CheckAllowedPath: match!")
						return apiproduct, nil
					}
				}
			}
		}
	}
	return shared.APIProduct{}, errors.New("No access")
}
