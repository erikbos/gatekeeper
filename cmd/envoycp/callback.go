package main

import (
	"context"
	"sync"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	log "github.com/sirupsen/logrus"
)

type callbacks struct {
	signal   chan struct{}
	fetches  int
	requests int
	mu       sync.Mutex
	srv      *server
}

func (cb *callbacks) Report() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	fields := log.Fields{
		"fetches":  cb.fetches,
		"requests": cb.requests,
	}
	log.WithFields(fields).Info("OnstreamReport")

	cb.srv.increaseCounterXDSMessage("Report")
}

// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb *callbacks) OnStreamOpen(ctx context.Context, id int64, typ string) error {
	fields := log.Fields{
		"stream": id,
		"type":   typ,
	}
	log.WithFields(fields).Info("OnStreamOpen")

	cb.srv.increaseCounterXDSMessage("OnStreamOpen")

	return nil
}

// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (cb *callbacks) OnStreamClosed(id int64) {
	fields := log.Fields{
		"stream": id,
	}
	log.WithFields(fields).Info("OnStreamClosed")

	cb.srv.increaseCounterXDSMessage("OnStreamClosed")
}

// OnStreamRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb *callbacks) OnStreamRequest(id int64, request *discovery.DiscoveryRequest) error {
	fields := log.Fields{
		"stream":    id,
		"useragent": request.Node.UserAgentName,
		"cluster":   request.Node.Cluster,
		"id":        request.Node.Id,
		// "version": fmt.Sprintf("%d.%d.%d", request.Node.GetUserAgentBuildVersion().Version.MajorNumber,
		// 	request.Node.GetUserAgentBuildVersion().Version.MinorNumber,
		// 	request.Node.GetUserAgentBuildVersion().Version.Patch),
	}
	log.WithFields(fields).Info("OnStreamRequest")

	cb.srv.increaseCounterXDSMessage("OnStreamRequest")

	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.requests++
	if cb.signal != nil {
		close(cb.signal)
		cb.signal = nil
	}
	return nil
}

// OnStreamResponse is called immediately prior to sending a response on a stream.
func (cb *callbacks) OnStreamResponse(id int64, request *discovery.DiscoveryRequest, response *discovery.DiscoveryResponse) {
	fields := log.Fields{
		"stream": id,
		"type":   response.TypeUrl,
	}
	log.WithFields(fields).Info("OnStreamResponse")

	cb.srv.increaseCounterXDSMessage("OnStreamResponse")

	cb.Report()
}

// OnFetchRequest is called for each Fetch request. Returning an error will end processing of the
// request and respond with an error.
func (cb *callbacks) OnFetchRequest(ctx context.Context, request *discovery.DiscoveryRequest) error {
	log.WithFields(log.Fields{}).Info("OnFetchRequest")

	cb.srv.increaseCounterXDSMessage("OnFetchRequest")

	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.fetches++
	if cb.signal != nil {
		close(cb.signal)
		cb.signal = nil
	}
	return nil
}

// OnFetchResponse is called immediately prior to sending a response.
func (cb *callbacks) OnFetchResponse(*discovery.DiscoveryRequest, *discovery.DiscoveryResponse) {
	log.Infof("OnFetchResponse")

	cb.srv.increaseCounterXDSMessage("OnFetchResponse")
}
