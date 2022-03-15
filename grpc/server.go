package main

import (
	"fmt"
	"net"

	"github.com/zalgonoise/zlog/log"
	"google.golang.org/grpc"
)

func setup() chan error {
	return make(chan error)
}

func listen(port string, errCh chan error) net.Listener {

	lis, err := net.Listen("tcp", port)

	if err != nil {
		errCh <- err
		return nil
	}

	return lis
}

func grpcServer(port string, errCh chan error) {
	lis := listen(port, errCh)

	// push log server handler
	s := &log.LogServer{}

	grpcServer := grpc.NewServer()

	log.RegisterLogServiceServer(grpcServer, s)

	if err := grpcServer.Serve(lis); err != nil {
		errCh <- err
		return
	}

}

func main() {
	port := ":9000"
	errCh := setup()
	go grpcServer(port, errCh)
	for {
		err := <-errCh
		panic(fmt.Errorf("errored out: %s", err))
	}

}
