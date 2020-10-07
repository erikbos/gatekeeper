package main

import (
	"context"
	"sync"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"go.uber.org/zap"
)

type callback struct {
	mutex       sync.Mutex
	signal      chan newNode
	connections map[int64]*core.Node
	logger      *zap.Logger
	metrics     *metrics
}

type newNode struct {
	nodeID string
}

func newCallback(s *server) *callback {

	return &callback{
		signal:      make(chan newNode),
		connections: make(map[int64]*core.Node),
		logger:      s.logger,
		metrics:     s.metrics,
	}
}

// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb *callback) OnStreamOpen(ctx context.Context, id int64, typ string) error {

	cb.logger.Info("OnStreamOpen", zap.Int64("stream", id), zap.String("type", typ))
	cb.metrics.IncXDSMessageCount("OnStreamOpen")

	return nil
}

// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (cb *callback) OnStreamClosed(id int64) {

	// Lock as we might receive multiple connections of new Envoys simultaneously
	cb.mutex.Lock()
	// Remove so we do not update this connection's snapshot anymore when configuration changes
	delete(cb.connections, id)
	cb.mutex.Unlock()

	cb.logger.Info("OnStreamClosed", zap.Int64("stream", id))
	cb.metrics.IncXDSMessageCount("OnStreamClosed")
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

	cb.logger.Info("OnStreamRequest",
		zap.Int64("stream", id),
		zap.String("useragent", request.Node.UserAgentName),
		zap.String("cluster", request.Node.Cluster),
		zap.String("id", request.Node.Id))
	cb.metrics.IncXDSMessageCount("OnStreamRequest")

	return nil
}

// OnStreamResponse is called immediately prior to sending a response on a stream.
func (cb *callback) OnStreamResponse(id int64, request *discovery.DiscoveryRequest, response *discovery.DiscoveryResponse) {

	cb.logger.Info("OnStreamResponse", zap.Int64("stream", id), zap.String("type", response.TypeUrl))
	cb.metrics.IncXDSMessageCount("OnStreamResponse")
}

// OnFetchRequest is called for each Fetch request. Returning an error will end processing of the
// request and respond with an error.
func (cb *callback) OnFetchRequest(ctx context.Context, request *discovery.DiscoveryRequest) error {

	cb.logger.Info("OnFetchRequest")
	cb.metrics.IncXDSMessageCount("OnFetchRequest")

	return nil
}

// OnFetchResponse is called immediately prior to sending a response.
func (cb *callback) OnFetchResponse(*discovery.DiscoveryRequest, *discovery.DiscoveryResponse) {

	cb.logger.Info("OnFetchResponse")
	cb.metrics.IncXDSMessageCount("OnFetchResponse")
}
