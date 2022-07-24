package server

import "github.com/zalgonoise/zlog/log/event"

type nilLogServer struct{}

func (nilLogServer) Serve() {}
func (nilLogServer) Stop()  {}
func (nilLogServer) Channels() (logCh, logSvCh chan *event.Event, errCh chan error) {
	return nil, nil, nil
}

// NilServer func creates a gRPC Log Server that doesn't do anything,
// mostly for tests using the LogServer interface
func NilServer() LogServer {
	return &nilLogServer{}
}
