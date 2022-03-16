package main

import (
	"fmt"
	"net"

	"github.com/zalgonoise/zlog/log"
	pb "github.com/zalgonoise/zlog/proto/message"
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

func grpcServer(port string, errCh chan error, logger log.LoggerI) {
	lis := listen(port, errCh)

	msgCh := make(chan *pb.MessageRequest)

	// push log server handler
	s := pb.NewLogServer(msgCh)
	go handleMessages(msgCh, logger)

	grpcServer := grpc.NewServer()

	pb.RegisterLogServiceServer(grpcServer, s)

	if err := grpcServer.Serve(lis); err != nil {
		errCh <- err
		return
	}

}

func handleMessages(msgCh chan *pb.MessageRequest, logger log.LoggerI) {
	for {
		msg := <-msgCh

		logmsg := log.NewMessage().FromProto(msg).Build()

		logger.Log(logmsg)
	}
}

func main() {
	port := ":9000"
	errCh := setup()
	logger := log.New()
	go grpcServer(port, errCh, logger)
	for {
		err := <-errCh
		panic(fmt.Errorf("errored out: %s", err))
	}

}
