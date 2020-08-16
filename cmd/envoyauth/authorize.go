package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authservice "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoytype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/gogo/googleapis/google/rpc"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erikbos/gatekeeper/pkg/shared"
)

type envoyAuthConfig struct {
	Listen string `yaml:"listen"`
}

// requestInfo holds all information of a request
type requestInfo struct {
	IP              net.IP
	httpRequest     *authservice.AttributeContext_HttpRequest
	URL             *url.URL
	queryParameters url.Values
	apikey          *string
	vhost           *shared.VirtualHost
	developer       *shared.Developer
	developerApp    *shared.DeveloperApp
	appCredential   *shared.DeveloperAppKey
	APIProduct      *shared.APIProduct
}

// startGRPCAuthorizationServer starts extauthz grpc listener
func (a *authorizationServer) startGRPCAuthorizationServer() {

	lis, err := net.Listen("tcp", a.config.EnvoyAuth.Listen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("GRPC listening on %s", a.config.EnvoyAuth.Listen)

	grpcServer := grpc.NewServer()
	authservice.RegisterAuthorizationServer(grpcServer, a)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Check (called by Envoy) to authenticate & authorize a HTTP request
func (a *authorizationServer) Check(ctx context.Context, authRequest *authservice.CheckRequest) (*authservice.CheckResponse, error) {

	timer := prometheus.NewTimer(a.metrics.authLatencyHistogram)
	defer timer.ObserveDuration()

	request, err := getRequestInfo(authRequest)
	if err != nil {
		a.metrics.connectInfoFailures.Inc()
		return rejectRequest(http.StatusBadRequest, nil, nil, fmt.Sprintf("%s", err))
	}
	a.logRequestDebug(request)

	// FIXME not sure if x-forwarded-proto the way to determine original tcp port used
	request.vhost, err = a.lookupVhost(request.httpRequest.Host, request.httpRequest.Headers["x-forwarded-proto"])
	if err != nil {
		a.increaseCounterRequestRejected(request)
		return rejectRequest(http.StatusNotFound, nil, nil, "unknown vhost")
	}

	// upstreamHeaders will receive all new headers to add before request goes upstream
	upstreamHeaders := make(map[string]string)

	// dynamicMetadata will receive all metadata to be used by next envoy filters
	// examples are: ratelimiter and extra fields to be logged
	dynamicMetadata := make(map[string]string)

	if request.vhost.Policies != "" {
		errorStatusCode, err := a.handlePolicies(request, &request.vhost.Policies, a.handleVhostPolicy, upstreamHeaders)
		// In case a policy wants us to stop we reject call
		if err != nil {
			a.increaseCounterRequestRejected(request)
			return rejectRequest(errorStatusCode, nil, dynamicMetadata, err.Error())
		}
	}

	// In case HTTP method is OPTIONS we skip parsing product/upstream policy
	// as we're not sending anything upstream, Envoy should answer all OPTIONS requests
	if request.httpRequest.Method != "OPTIONS" {
		if request.APIProduct.Policies != "" {
			errorStatusCode, err := a.handlePolicies(request, &request.APIProduct.Policies, a.handlePolicy, upstreamHeaders)
			// In case a policy wants us to stop we reject call
			if err != nil {
				compileMetadata(request, dynamicMetadata)

				a.increaseCounterRequestRejected(request)
				return rejectRequest(errorStatusCode, nil, dynamicMetadata, err.Error())
			}
		}
	}

	for k, v := range upstreamHeaders {
		log.Debugf("upstream header: %s = %s", k, v)
	}

	compileMetadata(request, dynamicMetadata)

	a.IncreaseCounterRequestAccept(request)
	return allowRequest(upstreamHeaders, dynamicMetadata)
}

// FIXME this should be lookup in map instead of for loops
// FIXXE map should have a key vhost:port
func (a *authorizationServer) lookupVhost(hostname, protocol string) (*shared.VirtualHost, error) {

	for _, vhostEntry := range a.virtualhosts {
		for _, vhost := range vhostEntry.VirtualHosts {
			if vhost == hostname {
				if (vhostEntry.Port == 80 && protocol == "http") ||
					(vhostEntry.Port == 443 && protocol == "https") {
					return &vhostEntry, nil
				}
			}
		}
	}

	return nil, errors.New("vhost not found")
}

// handlePolicies invokes all policy functions to set additional upstream headers
func (a *authorizationServer) handlePolicies(request *requestInfo, policies *string,
	policyHandler function, newUpstreamHeaders map[string]string) (int, error) {

	for _, policy := range strings.Split(*policies, ",") {
		trimmedPolicy := strings.TrimSpace(policy)
		headersToAdd, err := policyHandler(&trimmedPolicy, request)

		// Stop and return error in case policy indicates we must stop
		if err != nil {
			return http.StatusForbidden, err
		}

		// Add policy generated headers for upstream
		for key, value := range headersToAdd {
			newUpstreamHeaders[key] = value
		}
	}
	return http.StatusOK, nil
}

func compileMetadata(request *requestInfo, metadata map[string]string) {

	if request.developer != nil {
		if request.developer.Email != "" {
			metadata["developerEmail"] = request.developer.Email
		}
		if request.developer.DeveloperID != "" {
			metadata["developerID"] = request.developer.DeveloperID
		}
	}
	if request.developerApp != nil {
		if request.developerApp.AppID != "" {
			metadata["developerAppID"] = request.developerApp.AppID
		}
		if request.developerApp.Name != "" {
			metadata["developerAppName"] = request.developerApp.Name
		}
	}
	if request.APIProduct != nil && request.APIProduct.Name != "" {
		metadata["APIProduct"] = request.APIProduct.Name
	}
	if request.apikey != nil {
		metadata["apikey"] = *request.apikey
	}
}

// allowRequest authorizates customer request to go upstream
func allowRequest(headers, metadata map[string]string) (*authservice.CheckResponse, error) {

	response := &authservice.CheckResponse{
		Status: &status.Status{
			Code: int32(rpc.OK),
		},
		HttpResponse: &authservice.CheckResponse_OkResponse{
			OkResponse: &authservice.OkHttpResponse{
				Headers:         buildHeadersList(headers),
				DynamicMetadata: buildDynamicMetadataList(metadata),
			},
		},
	}
	log.Printf("allowRequest: %+v", response)

	return response, nil
}

// rejectCall answers to Envoy to reject HTTP request
func rejectRequest(statusCode int, headers, metadata map[string]string,
	message string) (*authservice.CheckResponse, error) {

	var envoyStatusCode envoytype.StatusCode

	switch statusCode {
	case http.StatusUnauthorized:
		envoyStatusCode = envoytype.StatusCode_Unauthorized
	case http.StatusForbidden:
		envoyStatusCode = envoytype.StatusCode_Forbidden
	case http.StatusServiceUnavailable:
		envoyStatusCode = envoytype.StatusCode_ServiceUnavailable
	default:
		envoyStatusCode = envoytype.StatusCode_Forbidden
	}

	response := &authservice.CheckResponse{
		Status: &status.Status{
			Code: int32(rpc.UNAUTHENTICATED),
		},
		HttpResponse: &authservice.CheckResponse_DeniedResponse{
			DeniedResponse: &authservice.DeniedHttpResponse{
				Status: &envoytype.HttpStatus{
					Code: envoyStatusCode,
				},
				Headers: buildHeadersList(headers),
				Body:    buildJSONErrorMessage(&message),
				// FIXME this is not yet supported by envoy
				//				DynamicMetadata: buildDynamicMetadataList(metadata),
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

// buildDynamicMetadataList creates struct for all additional metadata to be returned when accepting a request
// purpose is:
// 1) have accesslog log additional fields which are not susposed to go upstream as HTTP headers
// 2) to provide configuration to ratelimiter filter
func buildDynamicMetadataList(metadata map[string]string) *structpb.Struct {

	if len(metadata) == 0 {
		return nil
	}

	metadataStruct := structpb.Struct{
		Fields: map[string]*structpb.Value{},
	}
	for key, value := range metadata {
		metadataStruct.Fields[key] = &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: value,
			},
		}
	}
	return &metadataStruct
}

// getRequestInfo returns HTTP data of a request
func getRequestInfo(req *authservice.CheckRequest) (*requestInfo, error) {

	newConnection := requestInfo{
		httpRequest: req.Attributes.Request.Http,
	}
	ipaddress, ok := newConnection.httpRequest.Headers["x-forwarded-for"]
	if ok {
		newConnection.IP = net.ParseIP(ipaddress)
	}

	var err error
	newConnection.URL, err = url.ParseRequestURI(newConnection.httpRequest.Path)
	if err != nil {
		return nil, errors.New("could not parse url")
	}

	newConnection.queryParameters, _ = url.ParseQuery(newConnection.URL.RawQuery)
	if err != nil {
		return nil, errors.New("could not parse query parameters")
	}

	return &newConnection, nil
}

func (a *authorizationServer) logRequestDebug(request *requestInfo) {
	log.Debugf("Check() rx path: %s", request.httpRequest.Path)

	for key, value := range request.httpRequest.Headers {
		log.Debugf("Check() rx header [%s] = %s", key, value)
	}
}

// JSONErrorMessage is the format for our error messages
const JSONErrorMessage = `{
 "message": "%s"
}
`

// returns a well structured JSON-formatted message
func buildJSONErrorMessage(message *string) string {

	return fmt.Sprintf(JSONErrorMessage, *message)
}
