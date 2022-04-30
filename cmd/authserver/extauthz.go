package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_auth_v3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erikbos/gatekeeper/cmd/authserver/policy"
	"github.com/erikbos/gatekeeper/cmd/authserver/request"
	"github.com/erikbos/gatekeeper/pkg/shared"
)

type envoyAuthConfig struct {
	// GRPC Address and port to listen for extauthz requests
	Listen string
	// default Organization to use for authentication of incoming requests
	defaultOrganization string
}

// startGRPCAuthorizationServer starts extauthz grpc listener
func (s *server) StartAuthorizationServer() {

	lis, err := net.Listen("tcp", s.config.EnvoyAuth.Listen)
	if err != nil {
		s.logger.Fatal("failed to listen", zap.Error(err))
	}
	s.logger.Info("GRPC listening on " + s.config.EnvoyAuth.Listen)

	grpcServer := grpc.NewServer()
	envoy_service_auth_v3.RegisterAuthorizationServer(grpcServer, s)
	reflection.Register(grpcServer)
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())

	if err := grpcServer.Serve(lis); err != nil {
		s.logger.Fatal("Failed to start server", zap.Error(err))
	}
}

// Check (called by Envoy) to authenticate & authorize a HTTP request
func (s *server) Check(ctx context.Context,
	extauthzRequest *envoy_service_auth_v3.CheckRequest) (*envoy_service_auth_v3.CheckResponse, error) {

	timer := s.metrics.NewTimerAuthLatency()
	defer timer.ObserveDuration()

	request, err := request.DecodeAuthRequest(extauthzRequest)
	if err != nil {
		s.metrics.IncConnectionInfoFailure()
		return s.rejectRequest(http.StatusServiceUnavailable, nil, nil, err.Error())
	}
	request.Timestamp = shared.GetCurrentTimeMilliseconds()

	s.logger.Debug("extauthz",
		zap.String("path", request.HTTPRequest.Path),
		zap.Any("headers", request.HTTPRequest.Headers))

	// Lookup listener & organization assigned to this vhost
	request.Listener, request.Organization, err = s.vhosts.Lookup(request.HTTPRequest.Host, request.Port)
	if err != nil {
		s.metrics.IncAuthenticationRejected(request)
		return s.rejectRequest(http.StatusServiceUnavailable, nil, nil, "Unknown vhost/port")
	}

	policyConfig := policy.NewChainConfig(s.db, s.oauth, s.geoip, s.metrics, s.logger)

	// Evaluate policies, if any, assigned to listener
	vhostPolicyOut := &policy.ChainOutcome{}
	if request.Listener != nil && request.Listener.Policies != "" {
		vhostPolicyOut = policy.NewChain(request,
			policy.PolicyScopeVhost, policyConfig).Evaluate()

		s.logger.Debug("vhostPolicyOutcome", zap.Reflect("debug", vhostPolicyOut))
	}

	// Evaluate policies assigned, if any, that are assigned to requested apiproduct
	APIProductPolicyOut := &policy.ChainOutcome{}
	if request.APIProduct != nil && request.APIProduct.Policies != "" {
		APIProductPolicyOut = policy.NewChain(request,
			policy.PolicyScopeAPIProduct, policyConfig).Evaluate()

		s.logger.Debug("APIProductPolicyOutcome", zap.Reflect("debug", APIProductPolicyOut))
	}

	// We reject request in case both vhost & apiproduct policy did not authenticate request
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

	// Allow request
	s.metrics.IncAuthenticationAccepted(request)

	return s.allowRequest(
		mergeMapsStringString(vhostPolicyOut.UpstreamHeaders,
			APIProductPolicyOut.UpstreamHeaders),
		mergeMapsStringString(vhostPolicyOut.UpstreamDynamicMetadata,
			APIProductPolicyOut.UpstreamDynamicMetadata))
}

// allowRequest answers Envoyproxy to authorizates request to go upstream
func (s *server) allowRequest(headers, metadata map[string]string) (
	*envoy_service_auth_v3.CheckResponse, error) {

	response := &envoy_service_auth_v3.CheckResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		HttpResponse: &envoy_service_auth_v3.CheckResponse_OkResponse{
			OkResponse: &envoy_service_auth_v3.OkHttpResponse{
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
	metadata map[string]string, message string) (*envoy_service_auth_v3.CheckResponse, error) {

	var envoyStatusCode envoy_type_v3.StatusCode

	switch statusCode {
	case http.StatusUnauthorized:
		envoyStatusCode = envoy_type_v3.StatusCode_Unauthorized
	case http.StatusForbidden:
		envoyStatusCode = envoy_type_v3.StatusCode_Forbidden
	case http.StatusServiceUnavailable:
		envoyStatusCode = envoy_type_v3.StatusCode_ServiceUnavailable
	default:
		envoyStatusCode = envoy_type_v3.StatusCode_Forbidden
	}

	response := &envoy_service_auth_v3.CheckResponse{
		Status: &status.Status{
			Code: int32(code.Code_UNAUTHENTICATED),
		},
		HttpResponse: &envoy_service_auth_v3.CheckResponse_DeniedResponse{
			DeniedResponse: &envoy_service_auth_v3.DeniedHttpResponse{
				Status: &envoy_type_v3.HttpStatus{
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
func buildHeadersList(headers map[string]string) []*envoy_config_core_v3.HeaderValueOption {

	if len(headers) == 0 {
		return nil
	}

	headerList := make([]*envoy_config_core_v3.HeaderValueOption, 0, len(headers))
	for key, value := range headers {
		headerList = append(headerList, &envoy_config_core_v3.HeaderValueOption{
			Header: &envoy_config_core_v3.HeaderValue{
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
