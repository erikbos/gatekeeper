package request

import (
	"errors"
	"net"
	"net/url"

	envoy_service_auth_v3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// Request holds all details of a received Envoy auth request and looked up database entities
type Request struct {
	// Timestamp contains current time
	Timestamp       int64
	HTTPRequest     *envoy_service_auth_v3.AttributeContext_HttpRequest
	IP              net.IP
	Port            int
	URL             *url.URL
	QueryParameters url.Values
	ConsumerKey     *string
	OauthToken      *string
	Listener        *types.Listener
	Organization    *types.Organization
	Developer       *types.Developer
	DeveloperApp    *types.DeveloperApp
	Key             *types.Key
	APIProduct      *types.APIProduct
}

// DecodeAuthRequest returns details of an Envoy auth request
func DecodeAuthRequest(req *envoy_service_auth_v3.CheckRequest) (*Request, error) {

	r := &Request{
		HTTPRequest: req.Attributes.Request.Http,
	}
	if ipaddress, ok := r.HTTPRequest.Headers["x-forwarded-for"]; ok {
		r.IP = net.ParseIP(ipaddress)
	}

	var err error
	if r.URL, err = url.ParseRequestURI(r.HTTPRequest.Path); err != nil {
		return nil, errors.New("cannot parse url")
	}

	if r.QueryParameters, err = url.ParseQuery(r.URL.RawQuery); err != nil {
		return nil, errors.New("cannot parse query parameters")
	}

	// TODO/FIXME not sure if x-forwarded-proto the way to determine original tcp port used
	if proto, ok := r.HTTPRequest.Headers["x-forwarded-proto"]; ok {
		switch proto {
		case "http":
			r.Port = 80
		case "https":
			r.Port = 443
		default:
			return nil, errors.New("cannot determine port")
		}
	} else {
		return nil, errors.New("cannot determine port, missing header")
	}

	return r, nil
}
