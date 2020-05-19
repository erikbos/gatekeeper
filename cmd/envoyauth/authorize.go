package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

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
	vhost           *shared.VirtualHost
	developer       shared.Developer
	developerApp    shared.DeveloperApp
	appCredential   shared.AppCredential
	APIProduct      shared.APIProduct
}

// startGRPCAuthorizationServer starts extauthz grpc listener
func (a *authorizationServer) startGRPCAuthorizationServer() {

	lis, err := net.Listen("tcp", a.config.AuthGRPCListen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("GRPC listening on %s", a.config.AuthGRPCListen)

	grpcServer := grpc.NewServer()
	auth.RegisterAuthorizationServer(grpcServer, a)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Check (called by Envoy) to authenticate & authorize a HTTP request
func (a *authorizationServer) Check(ctx context.Context, authRequest *auth.CheckRequest) (*auth.CheckResponse, error) {

	timer := prometheus.NewTimer(a.metrics.authLatencyHistogram)
	defer timer.ObserveDuration()

	request, err := getRequestInfo(authRequest)
	if err != nil {
		a.metrics.connectInfoFailures.Inc()
		return rejectRequest(http.StatusBadRequest, nil, fmt.Sprintf("%s", err))
	}

	a.logConnectionDebug(&request)
	log.Printf("Incoming host: %s", request.httpRequest.Host)

	upstreamHeaders := make(map[string]string)

	// FIXME remove
	for _, vhostEntry := range a.virtualhosts {
		for _, vhost := range vhostEntry.VirtualHosts {
			if vhost == request.httpRequest.Host {
				if vhostEntry.Policies != "" {
					log.Infof("found host: %s, policies: %s", vhost, vhostEntry.Policies)
					_, err := a.handlePolicies(&request, vhostEntry.Policies, a.handleVhostPolicy, upstreamHeaders)
					// In case a policy wants us to stop we reject call
					if err != nil {
						a.increaseRequestRejectCounter(&request)
						return rejectRequest(envoy_type.StatusCode_Forbidden, nil, err.Error())
					}
				}

			}
		}
	}

	// Invoke any product policy that we need to invoke
	if request.APIProduct.Scopes != "" {
		_, err := a.handlePolicies(&request, request.APIProduct.Scopes, a.handlePolicy, upstreamHeaders)
		// In case a policy wants us to stop we reject call
		if err != nil {
			a.increaseRequestRejectCounter(&request)
			return rejectRequest(envoy_type.StatusCode_Forbidden, nil, err.Error())
		}
	}

	a.IncreaseRequestAcceptCounter(&request)
	return allowRequest(upstreamHeaders)
}

type function func(policy string, request *requestInfo) (map[string]string, error)

// handlePolicies invokes all policy functions to set additional upstream headers
func (a *authorizationServer) handlePolicies(request *requestInfo, policies string,
	policyHandler function, newUpstreamHeaders map[string]string) (int, error) {

	for _, policy := range strings.Split(policies, ",") {
		headersToAdd, err := policyHandler(policy, request)

		// Stop and return error in case policy indicates we should stop
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

// JSONErrorMessage is the format for our error messages
const JSONErrorMessage = `{
 "message": "%s"
}
`

// returns a well structured JSON-formatted message
func buildJSONErrorMessage(message string) string {
	return fmt.Sprintf(JSONErrorMessage, message)
}
