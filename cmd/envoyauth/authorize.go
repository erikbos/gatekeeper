package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authservice "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoytype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/gogo/googleapis/google/rpc"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erikbos/gatekeeper/pkg/types"
)

type envoyAuthConfig struct {
	Listen string `yaml:"listen"` // GRPC Address and port to listen for control plane
}

// requestInfo holds all information of a request
type requestInfo struct {
	IP              net.IP
	httpRequest     *authservice.AttributeContext_HttpRequest
	URL             *url.URL
	queryParameters url.Values
	apikey          *string
	oauth2token     *string
	vhost           *types.Listener
	developer       *types.Developer
	developerApp    *types.DeveloperApp
	appCredential   *types.DeveloperAppKey
	APIProduct      *types.APIProduct
}

// startGRPCAuthorizationServer starts extauthz grpc listener
func (a *authorizationServer) StartAuthorizationServer() {

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
func (a *authorizationServer) Check(ctx context.Context,
	authRequest *authservice.CheckRequest) (*authservice.CheckResponse, error) {

	timer := prometheus.NewTimer(a.metrics.authLatencyHistogram)
	defer timer.ObserveDuration()

	request, err := getRequestInfo(authRequest)
	if err != nil {
		a.metrics.connectInfoFailures.Inc()
		return rejectRequest(http.StatusBadRequest, nil, nil, fmt.Sprintf("%s", err))
	}
	a.logRequestDebug(request)

	// FIXME not sure if x-forwarded-proto the way to determine original tcp port used
	request.vhost, err = a.vhosts.Lookup(request.httpRequest.Host, request.httpRequest.Headers["x-forwarded-proto"])
	if err != nil {
		a.increaseCounterRequestRejected(request)
		return rejectRequest(http.StatusNotFound, nil, nil, "unknown vhost")
	}

	vhostPolicyOutcome := &PolicyChainResponse{}
	if request.vhost != nil && request.vhost.Policies != "" {
		vhostPolicyOutcome = (&PolicyChain{
			authServer: a,
			request:    request,
			scope:      policyScopeVhost,
		}).Evaluate()
	}

	APIProductPolicyOutcome := &PolicyChainResponse{}
	if request.APIProduct != nil && request.APIProduct.Policies != "" {
		APIProductPolicyOutcome = (&PolicyChain{
			authServer: a,
			request:    request,
			scope:      policyScopeAPIProduct,
		}).Evaluate()
	}

	log.Debugf("vhostPolicyOutcome: %+v", vhostPolicyOutcome)
	log.Debugf("APIProductPolicyOutcome: %+v", APIProductPolicyOutcome)

	// We reject call in case both vhost & apiproduct policy did not authenticate call
	if (vhostPolicyOutcome != nil && !vhostPolicyOutcome.authenticated) &&
		(APIProductPolicyOutcome != nil && !APIProductPolicyOutcome.authenticated) {

		a.increaseCounterRequestRejected(request)

		return rejectRequest(vhostPolicyOutcome.deniedStatusCode,
			mergeMapsStringString(vhostPolicyOutcome.upstreamHeaders,
				APIProductPolicyOutcome.upstreamHeaders),
			mergeMapsStringString(vhostPolicyOutcome.upstreamDynamicMetadata,
				APIProductPolicyOutcome.upstreamDynamicMetadata),
			vhostPolicyOutcome.deniedMessage)
	}

	a.IncreaseCounterRequestAccept(request)

	return allowRequest(
		mergeMapsStringString(vhostPolicyOutcome.upstreamHeaders,
			APIProductPolicyOutcome.upstreamHeaders),
		mergeMapsStringString(vhostPolicyOutcome.upstreamDynamicMetadata,
			APIProductPolicyOutcome.upstreamDynamicMetadata))
}

// mergeMapsStringString returns merged map[string]string
// it overwriting duplicate keys, you should handle that if there is a need
func mergeMapsStringString(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// allowRequest answers Envoyproxy to authorizates request to go upstream
func allowRequest(headers, metadata map[string]string) (*authservice.CheckResponse, error) {

	response := &authservice.CheckResponse{
		Status: &status.Status{
			Code: int32(rpc.OK),
		},
		HttpResponse: &authservice.CheckResponse_OkResponse{
			OkResponse: &authservice.OkHttpResponse{
				Headers: buildHeadersList(headers),
				// TODO this should be removed with go control plane 0.9.8
				DynamicMetadata: buildDynamicMetadataList(metadata),
			},
		},
		// TODO uncomment in case of go control plane 0.9.8
		// DynamicMetadata: buildDynamicMetadataList(metadata),
	}
	log.Printf("allowRequest: %+v", response)

	return response, nil
}

// rejectCall answers Envoyproxy to reject HTTP request
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
			},
		},
		// TODO uncomment in case of go control plane > 0.9.8
		// DynamicMetadata: buildDynamicMetadataList(metadata),
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
	if ipaddress, ok := newConnection.httpRequest.Headers["x-forwarded-for"]; ok {
		newConnection.IP = net.ParseIP(ipaddress)
	}

	var err error
	if newConnection.URL, err = url.ParseRequestURI(newConnection.httpRequest.Path); err != nil {
		return nil, errors.New("could not parse url")
	}

	if newConnection.queryParameters, err = url.ParseQuery(newConnection.URL.RawQuery); err != nil {
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
