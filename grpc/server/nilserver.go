package server

type nilLogServer struct{}

func (nilLogServer) Serve() {}
func (nilLogServer) Stop()  {}

// NilServer func creates a gRPC Log Server that doesn't do anything,
// mostly for tests using the LogServer interface
func NilServer() LogServer {
	return &nilLogServer{}
}
