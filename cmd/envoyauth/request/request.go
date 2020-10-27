package request

import (
	"net"
	"net/url"

	authservice "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"

	"github.com/erikbos/gatekeeper/pkg/types"
)

// State holds all details of a received request
type State struct {
	// Timestamp contains current time
	Timestamp       int64
	HTTPRequest     *authservice.AttributeContext_HttpRequest
	IP              net.IP
	URL             *url.URL
	QueryParameters url.Values
	ConsumerKey     *string
	OauthToken      *string
	Listener        *types.Listener
	Developer       *types.Developer
	DeveloperApp    *types.DeveloperApp
	DeveloperAppKey *types.DeveloperAppKey
	APIProduct      *types.APIProduct
}
