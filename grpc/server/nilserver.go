package server

import "github.com/zalgonoise/zlog/log/event"

type nilLogServer struct{}

// Serve is the implementation of the `Serve()` method, from the LogServer interface
func (nilLogServer) Serve() {}

// Stop is the implementation of the `Stop()` method, from the LogServer interface
func (nilLogServer) Stop() {}

// Channels is the implementation of the `Channels()` method, from the LogServer interface
func (nilLogServer) Channels() (logCh, logSvCh chan *event.Event, errCh chan error) {
	return nil, nil, nil
}

// NilServer func creates a gRPC Log Server that doesn't do anything,
// mostly for tests using the LogServer interface
func NilServer() LogServer {
	return &nilLogServer{}
}
