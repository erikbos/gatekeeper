package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authservice "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoytype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/gogo/googleapis/google/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erikbos/gatekeeper/pkg/shared"
	"github.com/erikbos/gatekeeper/pkg/types"
)

type envoyAuthConfig struct {

	// GRPC Address and port to listen for control plane
	Listen string `yaml:"listen"`
}

// requestDetails holds all information of a request
type requestDetails struct {
	// Timestamp contains current time
	timestamp       int64
	httpRequest     *authservice.AttributeContext_HttpRequest
	IP              net.IP
	URL             *url.URL
	queryParameters url.Values
	consumerKey     *string
	oauthToken      *string
	listener        *types.Listener
	developer       *types.Developer
	developerApp    *types.DeveloperApp
	developerAppKey *types.DeveloperAppKey
	APIProduct      *types.APIProduct
}

// startGRPCAuthorizationServer starts extauthz grpc listener
func (s *server) StartAuthorizationServer() {

	lis, err := net.Listen("tcp", s.config.EnvoyAuth.Listen)
	if err != nil {
		s.logger.Fatal("failed to listen", zap.Error(err))
	}
	s.logger.Info("GRPC listening on " + s.config.EnvoyAuth.Listen)

	grpcServer := grpc.NewServer()
	authservice.RegisterAuthorizationServer(grpcServer, s)

	if err := grpcServer.Serve(lis); err != nil {
		s.logger.Fatal("Failed to start server", zap.Error(err))
	}
}

// Check (called by Envoy) to authenticate & authorize a HTTP request
func (s *server) Check(ctx context.Context,
	extauthzRequest *authservice.CheckRequest) (*authservice.CheckResponse, error) {

	timer := prometheus.NewTimer(s.metrics.authLatency)
	defer timer.ObserveDuration()

	request, err := getRequestDetails(extauthzRequest)
	if err != nil {
		s.metrics.IncConnectionInfoFailure()
		return s.rejectRequest(http.StatusBadRequest, nil, nil, fmt.Sprintf("%s", err))
	}
	request.timestamp = shared.GetCurrentTimeMilliseconds()

	s.logger.Debug("extauthz",
		zap.String("path", request.httpRequest.Path),
		zap.Any("headers", request.httpRequest.Headers))

	// FIXME not sure if x-forwarded-proto the way to determine original tcp port used
	request.listener, err = s.vhosts.Lookup(request.httpRequest.Host, request.httpRequest.Headers["x-forwarded-proto"])
	if err != nil {
		s.metrics.IncAuthenticationRejected(request)
		return s.rejectRequest(http.StatusNotFound, nil, nil, "U1nknown vhost")
	}

	vhostPolicyOutcome := &PolicyChainResponse{}
	if request.listener != nil && request.listener.Policies != "" {
		vhostPolicyOutcome = (&PolicyChain{
			authServer: s,
			request:    request,
			scope:      policyScopeVhost,
		}).Evaluate()
	}

	APIProductPolicyOutcome := &PolicyChainResponse{}
	if request.APIProduct != nil && request.APIProduct.Policies != "" {
		APIProductPolicyOutcome = (&PolicyChain{
			authServer: s,
			request:    request,
			scope:      policyScopeAPIProduct,
		}).Evaluate()
	}

	s.logger.Debug("vhostPolicyOutcome", zap.Reflect("debug", vhostPolicyOutcome))
	s.logger.Debug("APIProductPolicyOutcome", zap.Reflect("debug", APIProductPolicyOutcome))

	// We reject call in case both vhost & apiproduct policy did not authenticate call
	if (vhostPolicyOutcome != nil && !vhostPolicyOutcome.authenticated) &&
		(APIProductPolicyOutcome != nil && !APIProductPolicyOutcome.authenticated) {

		s.metrics.IncAuthenticationRejected(request)

		return s.rejectRequest(vhostPolicyOutcome.deniedStatusCode,
			mergeMapsStringString(vhostPolicyOutcome.upstreamHeaders,
				APIProductPolicyOutcome.upstreamHeaders),
			mergeMapsStringString(vhostPolicyOutcome.upstreamDynamicMetadata,
				APIProductPolicyOutcome.upstreamDynamicMetadata),
			vhostPolicyOutcome.deniedMessage)
	}

	s.metrics.IncAuthenticationAccepted(request)

	return s.allowRequest(
		mergeMapsStringString(vhostPolicyOutcome.upstreamHeaders,
			APIProductPolicyOutcome.upstreamHeaders),
		mergeMapsStringString(vhostPolicyOutcome.upstreamDynamicMetadata,
			APIProductPolicyOutcome.upstreamDynamicMetadata))
}

// allowRequest answers Envoyproxy to authorizates request to go upstream
func (s *server) allowRequest(headers, metadata map[string]string) (
	*authservice.CheckResponse, error) {

	dynamicMetadata := buildDynamicMetadataList(metadata)

	response := &authservice.CheckResponse{
		Status: &status.Status{
			Code: int32(rpc.OK),
		},
		HttpResponse: &authservice.CheckResponse_OkResponse{
			OkResponse: &authservice.OkHttpResponse{
				Headers: buildHeadersList(headers),
				// Required for < Envoy 0.17
				DynamicMetadata: dynamicMetadata,
			},
		},
		// Required for > Envoy 0.16
		// DynamicMetadata: dynamicMetadata,
	}

	s.logger.Debug("allowRequest", zap.Reflect("response", response))
	return response, nil
}

// rejectRequest answers Envoyproxy to reject HTTP request
func (s *server) rejectRequest(statusCode int, headers,
	metadata map[string]string, message string) (*authservice.CheckResponse, error) {

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
		DynamicMetadata: buildDynamicMetadataList(metadata),
	}

	s.logger.Debug("rejectRequest", zap.Reflect("response", response))
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

// buildDynamicMetadataList creates struct for all additional metadata to be returned when accepting a request.
//
// Potential use cases:
// 1) insert metadata into upstream headers using %DYNAMIC_METADATA%
// 2) have accesslog log metadata which are not susposed to go upstream as HTTP headers
// 3) to provide configuration to ratelimiter filter
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
	// Convert rate limiter values into ratelimiter configuration
	// a route's ratelimiteraction will check for this metadata key
	if rateLimitOverride := buildRateLimiterOveride(metadata); rateLimitOverride != nil {
		metadataStruct.Fields["rl.override"] = rateLimitOverride
	}
	return &metadataStruct
}

// buildRateLimiterOveride builds route RateLimiterOverride configuration based upon
// metadata keys "rl.requests_per_unit" & "rl.unit"
func buildRateLimiterOveride(metadata map[string]string) *structpb.Value {

	var requestsPerUnit float64
	if value, found := metadata["rl.requests_per_unit"]; found {
		var err error
		if requestsPerUnit, err = strconv.ParseFloat(value, 64); err != nil {
			return nil
		}
	}
	var unit string
	unit, found := metadata["rl.unit"]
	if !found {
		return nil
	}
	return &structpb.Value{
		Kind: &structpb.Value_StructValue{
			StructValue: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"requests_per_unit": {
						Kind: &structpb.Value_NumberValue{
							NumberValue: requestsPerUnit,
						},
					},
					"unit": {
						Kind: &structpb.Value_StringValue{
							StringValue: unit,
						},
					},
				},
			},
		},
	}
}

// getRequestDetails returns HTTP data of a request
func getRequestDetails(req *authservice.CheckRequest) (*requestDetails, error) {

	r := requestDetails{
		httpRequest: req.Attributes.Request.Http,
	}
	if ipaddress, ok := r.httpRequest.Headers["x-forwarded-for"]; ok {
		r.IP = net.ParseIP(ipaddress)
	}

	var err error
	if r.URL, err = url.ParseRequestURI(r.httpRequest.Path); err != nil {
		return nil, errors.New("Cannot parse url")
	}

	if r.queryParameters, err = url.ParseQuery(r.URL.RawQuery); err != nil {
		return nil, errors.New("Cannot parse query parameters")
	}

	return &r, nil
}

// returns a well structured JSON-formatted message
func buildJSONErrorMessage(message *string) string {

	const JSONErrorMessage = `{
	"message": "%s"
   }
   `

	return fmt.Sprintf(JSONErrorMessage, *message)
}

// mergeMapsStringString merges multiple map[string]string into one
// be aware: it does overwriting duplicate keys
func mergeMapsStringString(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
