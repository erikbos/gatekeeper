package main

import (
	"context"
	"fmt"
	"sync"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	log "github.com/sirupsen/logrus"
)

type callbacks struct {
	signal   chan struct{}
	fetches  int
	requests int
	mu       sync.Mutex
}

func (cb *callbacks) Report() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	fields := log.Fields{
		"fetches":  cb.fetches,
		"requests": cb.requests,
	}
	log.WithFields(fields).Info("OnstreamReport")
}

// OnStreamOpen is called once an xDS stream is open with a stream ID and the type URL (or "" for ADS).
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb *callbacks) OnStreamOpen(ctx context.Context, id int64, typ string) error {
	fields := log.Fields{
		"stream": id,
		"type":   typ,
	}
	log.WithFields(fields).Info("OnStreamOpen")
	return nil
}

// OnStreamClosed is called immediately prior to closing an xDS stream with a stream ID.
func (cb *callbacks) OnStreamClosed(id int64) {
	fields := log.Fields{
		"stream": id,
	}
	log.WithFields(fields).Info("OnStreamClosed")
}

// OnStreamRequest is called once a request is received on a stream.
// Returning an error will end processing and close the stream. OnStreamClosed will still be called.
func (cb *callbacks) OnStreamRequest(id int64, request *v2.DiscoveryRequest) error {
	fields := log.Fields{
		"stream":    id,
		"useragent": request.Node.UserAgentName,
		"cluster":   request.Node.Cluster,
		"id":        request.Node.Id,
		"version": fmt.Sprintf("%d.%d.%d", request.Node.GetUserAgentBuildVersion().Version.MajorNumber,
			request.Node.GetUserAgentBuildVersion().Version.MinorNumber,
			request.Node.GetUserAgentBuildVersion().Version.Patch),
	}
	log.WithFields(fields).Info("OnStreamRequest")

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
func (cb *callbacks) OnStreamResponse(id int64, request *v2.DiscoveryRequest, response *v2.DiscoveryResponse) {
	fields := log.Fields{
		"stream": id,
		"type":   response.TypeUrl,
	}
	log.WithFields(fields).Info("OnStreamResponse")
	cb.Report()
}

// OnFetchRequest is called for each Fetch request. Returning an error will end processing of the
// request and respond with an error.
func (cb *callbacks) OnFetchRequest(ctx context.Context, request *v2.DiscoveryRequest) error {
	log.WithFields(log.Fields{}).Info("OnFetchRequest")

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
func (cb *callbacks) OnFetchResponse(*v2.DiscoveryRequest, *v2.DiscoveryResponse) {
	log.Infof("OnFetchResponse")
}
