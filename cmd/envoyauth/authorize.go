package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/bmatcuk/doublestar"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type"
	"github.com/gogo/googleapis/google/rpc"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"

	"github.com/erikbos/apiauth/pkg/shared"
)

// requestInfo holds all information of a request
type requestInfo struct {
	IP              net.IP
	httpRequest     *auth.AttributeContext_HttpRequest
	URL             *url.URL
	queryParameters url.Values
	apikey          string
	developer       shared.Developer
	developerApp    shared.DeveloperApp
	appCredential   shared.AppCredential
	APIProduct      shared.APIProduct
}

// startGRPCAuthorizationServer starts extauthz grpc listener
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
func (a *authorizationServer) Check(ctx context.Context, authRequest *auth.CheckRequest) (*auth.CheckResponse, error) {

	timer := prometheus.NewTimer(a.metrics.authLatencyHistogram)
	defer timer.ObserveDuration()

	upstreamHeaders := make(map[string]string)

	request, err := getRequestInfo(authRequest)
	if err != nil {
		a.metrics.connectInfoFailures.Inc()
		return rejectRequest(http.StatusBadRequest, nil, fmt.Sprintf("%s", err))
	}
	a.getCountryAndStateOfRequestorIP(&request, upstreamHeaders)
	a.logConnectionDebug(&request)

	err = a.CheckProductEntitlement("petstore", &request)
	if err != nil {
		log.Debugf("Check() not allowed '%s' (%s)", request.URL.Path, err.Error())
		a.increaseCounterApikeyNotfound(&request)
		return rejectRequest(envoy_type.StatusCode_Forbidden, nil, err.Error())
	}

	// Invoke any policy that we have to apply
	if request.APIProduct.Scopes != "" {
		_, err := handlePolicies(&request, upstreamHeaders)
		// In case a policy wants us to stop we reject call
		if err != nil {
			a.increaseRequestRejectCounter(&request)
			return rejectRequest(envoy_type.StatusCode_Forbidden, nil, err.Error())
		}
	}

	a.IncreaseRequestAcceptCounter(&request)
	return allowRequest(upstreamHeaders)
}

// increaseCounterPerCountry counts hits per country
func (a *authorizationServer) increaseCounterPerCountry(country string) {

	a.metrics.requestsPerCountry.WithLabelValues(country).Inc()
}

// increaseCounterApikeyNotfound requests with unknown apikey
func (a *authorizationServer) increaseCounterApikeyNotfound(r *requestInfo) {

	a.metrics.requestsApikeyNotFound.WithLabelValues(
		r.httpRequest.Host,
		r.httpRequest.Protocol,
		r.httpRequest.Method).Inc()
}

// increaseRequestRejectCounter counts requests that are rejected
func (a *authorizationServer) increaseRequestRejectCounter(r *requestInfo) {

	a.metrics.requestsRejected.WithLabelValues(
		r.httpRequest.Host,
		r.httpRequest.Protocol,
		r.httpRequest.Method,
		r.APIProduct.Name).Inc()
}

// IncreaseRequestAcceptCounter counts requests that are accpeted
func (a *authorizationServer) IncreaseRequestAcceptCounter(r *requestInfo) {

	a.metrics.requestsAccepted.WithLabelValues(
		r.httpRequest.Host,
		r.httpRequest.Protocol,
		r.httpRequest.Method,
		r.APIProduct.Name).Inc()
}

// allowRequest authorizates customer request to go upstreadm
func allowRequest(headers map[string]string) (*auth.CheckResponse, error) {
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

func rejectRequest(statusCode envoy_type.StatusCode,
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

// getRequestInfo returns HTTP data of a request
func getRequestInfo(req *auth.CheckRequest) (requestInfo, error) {

	newConnection := requestInfo{
		httpRequest: req.Attributes.Request.Http,
	}
	if ipaddress, ok := newConnection.httpRequest.Headers["x-forwarded-for"]; ok {
		newConnection.IP = net.ParseIP(ipaddress)
	}
	var err error
	newConnection.URL, err = url.ParseRequestURI(newConnection.httpRequest.Path)
	if err != nil {
		return requestInfo{}, errors.New("could not parse url")
	}
	newConnection.queryParameters, _ = url.ParseQuery(newConnection.URL.RawQuery)
	if err != nil {
		return requestInfo{}, errors.New("could not parse query parameters")
	}

	newConnection.apikey = newConnection.queryParameters["apikey"][0]

	return newConnection, nil
}

func (a *authorizationServer) logConnectionDebug(request *requestInfo) {
	log.Debugf("Check() rx path: %s", request.httpRequest.Path)
	log.Debugf("Check() rx uri: %s", request.URL.Path)
	log.Debugf("Check() rx qp: %s", request.queryParameters)
	log.Debugf("Check() rx apikey: %s", request.apikey)

	for key, value := range request.httpRequest.Headers {
		log.Debugf("Check() rx header [%s] = %s", key, value)
	}

	// log.Printf("Query String: %v", authRequest.GetAttributes().GetRequest().GetHttp().GetQuery())
}

// CheckProductEntitlement verifies whether the requested path is allowed to be called
func (a *authorizationServer) CheckProductEntitlement(organization string, request *requestInfo) error {

	err := a.getEntitlementDetails(organization, request)
	if err != nil {
		return err
	}
	if err = checkAppCredentialValidity(request.appCredential); err != nil {
		return err
	}
	request.APIProduct, err = a.IsRequestPathAllowed(organization, request.URL.Path, request.appCredential)
	if err != nil {
		return errors.New("No product match")
	}
	return nil
}

// getEntitlementDetails populates apikey, developer and developerapp details
func (a *authorizationServer) getEntitlementDetails(organization string, request *requestInfo) error {
	var err error

	request.appCredential, err = a.db.GetAppCredentialByKey(organization, request.apikey)
	if err != nil {
		// FIX ME increase unknown apikey counter (not an error state)
		return errors.New("Could not find apikey")
	}

	request.developerApp, err = a.db.GetDeveloperAppByID(organization, request.appCredential.OrganizationAppID)
	if err != nil {
		// FIX ME increase counter as every apikey should link to dev app (error state)
		return errors.New("Could not find developer app of this apikey")
	}

	request.developer, err = a.db.GetDeveloperByID(request.developerApp.ParentID)
	if err != nil {
		// FIX ME increase counter as every devapp should link to developer (error state)
		return errors.New("Could not find developer of this apikey")
	}

	return nil
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
					// log.Debugf("IsRequestPathAllowed() Matching path %s in %s", requestPath, productPath)
					if ok, _ := doublestar.Match(productPath, requestPath); ok {
						// log.Debugf("IsRequestPathAllowed: match!")
						return apiproduct, nil
					}
				}
			}
		}
	}
	return shared.APIProduct{}, errors.New("No access")
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
