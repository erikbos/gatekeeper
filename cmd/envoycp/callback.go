package main

import (
	"context"
	"sync"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	log "github.com/sirupsen/logrus"
)

type callback struct {
	srv         *server
	mutex       sync.Mutex
	signal      chan newNode
	connections map[int64]*core.Node
}

type newNode struct {
	nodeID string
}

func newCallback(s *server) *callback {

	return &callback{
		srv:         s,
		signal:      make(chan newNode),
		connections: make(map[int64]*core.Node),
	}
}

// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb *callback) OnStreamOpen(ctx context.Context, id int64, typ string) error {

	fields := log.Fields{
		"stream": id,
		"type":   typ,
	}
	log.WithFields(fields).Info("OnStreamOpen")
	cb.srv.increaseCounterXDSMessage("OnStreamOpen")

	return nil
}

// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (cb *callback) OnStreamClosed(id int64) {

	// Lock as we might receive multiple connections of new Envoys simultaneously
	cb.mutex.Lock()
	// Remove so we do not update this connection's snapshot anymore when configuration changes
	delete(cb.connections, id)
	cb.mutex.Unlock()

	fields := log.Fields{
		"stream": id,
	}
	log.WithFields(fields).Info("OnStreamClosed")
	cb.srv.increaseCounterXDSMessage("OnStreamClosed")
}

// OnStreamRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb *callback) OnStreamRequest(id int64, request *discovery.DiscoveryRequest) error {

	// Check if we have a connected Envoy on this connection id
	// If not remember connection id & node info and signal that cache should be updated for this new Envoy
	if request.Node != nil && request.Node.Id != "" {
		if _, ok := cb.connections[id]; !ok {

			// Lock as we might receive multiple connections of new Envoys simultaneously
			cb.mutex.Lock()
			// Add so we do update this connection's snapshot when configuration changes
			cb.connections[id] = request.Node
			cb.mutex.Unlock()

			// Notify to have populate cache for this new Envoy
			cb.signal <- newNode{
				nodeID: request.Node.Id,
			}
		}
	}

	fields := log.Fields{
		"stream":    id,
		"useragent": request.Node.UserAgentName,
		"cluster":   request.Node.Cluster,
		"id":        request.Node.Id,
	}
	log.WithFields(fields).Info("OnStreamRequest")
	cb.srv.increaseCounterXDSMessage("OnStreamRequest")

	return nil
}

// OnStreamResponse is called immediately prior to sending a response on a stream.
func (cb *callback) OnStreamResponse(id int64, request *discovery.DiscoveryRequest, response *discovery.DiscoveryResponse) {

	fields := log.Fields{
		"stream": id,
		"type":   response.TypeUrl,
	}
	log.WithFields(fields).Info("OnStreamResponse")
	cb.srv.increaseCounterXDSMessage("OnStreamResponse")
}

// OnFetchRequest is called for each Fetch request. Returning an error will end processing of the
// request and respond with an error.
func (cb *callback) OnFetchRequest(ctx context.Context, request *discovery.DiscoveryRequest) error {

	log.WithFields(log.Fields{}).Info("OnFetchRequest")
	cb.srv.increaseCounterXDSMessage("OnFetchRequest")

	return nil
}

// OnFetchResponse is called immediately prior to sending a response.
func (cb *callback) OnFetchResponse(*discovery.DiscoveryRequest, *discovery.DiscoveryResponse) {

	log.Infof("OnFetchResponse")
	cb.srv.increaseCounterXDSMessage("OnFetchResponse")
}
