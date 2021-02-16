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
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erikbos/gatekeeper/cmd/envoyauth/policy"
	"github.com/erikbos/gatekeeper/cmd/envoyauth/request"
	"github.com/erikbos/gatekeeper/pkg/shared"
)

type envoyAuthConfig struct {

	// GRPC Address and port to listen for control plane
	Listen string `yaml:"listen"`
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

	timer := s.metrics.NewTimerAuthLatency()
	defer timer.ObserveDuration()

	request, err := getrequestDetails(extauthzRequest)
	if err != nil {
		s.metrics.IncConnectionInfoFailure()
		return s.rejectRequest(http.StatusBadRequest, nil, nil, fmt.Sprintf("%s", err))
	}
	request.Timestamp = shared.GetCurrentTimeMilliseconds()

	s.logger.Debug("extauthz",
		zap.String("path", request.HTTPRequest.Path),
		zap.Any("headers", request.HTTPRequest.Headers))

	// FIXME not sure if x-forwarded-proto the way to determine original tcp port used
	request.Listener, err = s.vhosts.Lookup(request.HTTPRequest.Host,
		request.HTTPRequest.Headers["x-forwarded-proto"])
	if err != nil {
		s.metrics.IncAuthenticationRejected(request)
		return s.rejectRequest(http.StatusNotFound, nil, nil, "U1nknown vhost")
	}

	policyConfig := policy.NewChainConfig(s.db, s.oauth, s.geoip, s.metrics, s.logger)

	vhostPolicyOut := &policy.ChainOutcome{}
	if request.Listener != nil && request.Listener.Policies != "" {
		vhostPolicyOut = policy.NewChain(request,
			policy.PolicyScopeVhost, policyConfig).Evaluate()

		s.logger.Debug("vhostPolicyOutcome", zap.Reflect("debug", vhostPolicyOut))
	}

	APIProductPolicyOut := &policy.ChainOutcome{}
	if request.APIProduct != nil && request.APIProduct.Policies != "" {
		APIProductPolicyOut = policy.NewChain(request,
			policy.PolicyScopeAPIProduct, policyConfig).Evaluate()

		s.logger.Debug("APIProductPolicyOutcome", zap.Reflect("debug", APIProductPolicyOut))
	}

	// We reject call in case both vhost & apiproduct policy did not authenticate call
	if (vhostPolicyOut != nil && !vhostPolicyOut.Authenticated) &&
		(APIProductPolicyOut != nil && !APIProductPolicyOut.Authenticated) {

		s.metrics.IncAuthenticationRejected(request)

		return s.rejectRequest(vhostPolicyOut.DeniedStatusCode,
			mergeMapsStringString(vhostPolicyOut.UpstreamHeaders,
				APIProductPolicyOut.UpstreamHeaders),
			mergeMapsStringString(vhostPolicyOut.UpstreamDynamicMetadata,
				APIProductPolicyOut.UpstreamDynamicMetadata),
			vhostPolicyOut.DeniedMessage)
	}

	s.metrics.IncAuthenticationAccepted(request)

	return s.allowRequest(
		mergeMapsStringString(vhostPolicyOut.UpstreamHeaders,
			APIProductPolicyOut.UpstreamHeaders),
		mergeMapsStringString(vhostPolicyOut.UpstreamDynamicMetadata,
			APIProductPolicyOut.UpstreamDynamicMetadata))
}

// allowRequest answers Envoyproxy to authorizates request to go upstream
func (s *server) allowRequest(headers, metadata map[string]string) (
	*authservice.CheckResponse, error) {

	response := &authservice.CheckResponse{
		Status: &status.Status{
			Code: int32(rpc.OK),
		},
		HttpResponse: &authservice.CheckResponse_OkResponse{
			OkResponse: &authservice.OkHttpResponse{
				Headers: buildHeadersList(headers),
			},
		},
		DynamicMetadata: buildDynamicMetadataList(metadata),
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

// getrequestDetails returns details of an incoming request
func getrequestDetails(req *authservice.CheckRequest) (*request.State, error) {

	r := &request.State{
		HTTPRequest: req.Attributes.Request.Http,
	}
	if ipaddress, ok := r.HTTPRequest.Headers["x-forwarded-for"]; ok {
		r.IP = net.ParseIP(ipaddress)
	}

	var err error
	if r.URL, err = url.ParseRequestURI(r.HTTPRequest.Path); err != nil {
		return nil, errors.New("Cannot parse url")
	}

	if r.QueryParameters, err = url.ParseQuery(r.URL.RawQuery); err != nil {
		return nil, errors.New("Cannot parse query parameters")
	}

	return r, nil
}

// returns a well structured JSON-formatted message
func buildJSONErrorMessage(message *string) string {

	const JSONErrorMessage = `{ "message": "%s" }`

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
