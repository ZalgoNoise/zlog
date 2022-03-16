package server

import (
	"fmt"
	"net"

	"github.com/zalgonoise/zlog/log"
	pb "github.com/zalgonoise/zlog/proto/message"
	"google.golang.org/grpc"
)

type GRPCLogServer struct {
	Addr   string
	opts   []grpc.ServerOption
	Logger log.Logger
	ErrCh  chan error
	LogSv  *pb.LogServer
	Server *grpc.Server
}

func New(opts ...LogServerConfig) *GRPCLogServer {
	server := &GRPCLogServer{
		ErrCh: make(chan error),
		LogSv: pb.NewLogServer(),
	}

	for _, opt := range opts {
		opt.Apply(server)
	}

	if server.Addr == "" {
		WithAddr("").Apply(server)
	}

	if server.Logger != nil {
		WithLogger().Apply(server)
	}

	if server.opts == nil {
		WithGRPCOpts().Apply(server)
	}

	return server

}

func (s GRPCLogServer) listen() net.Listener {
	lis, err := net.Listen("tcp", s.Addr)

	if err != nil {
		s.ErrCh <- err
		return nil
	}

	return lis
}

func (s GRPCLogServer) handleMessages() {
	for {
		select {
		case msg := <-s.LogSv.MsgCh:
			logmsg := log.NewMessage().FromProto(msg).Build()
			s.Logger.Log(logmsg)

		case <-s.LogSv.Done:
			return
		}
	}
}

func (s GRPCLogServer) Serve() {
	lis := s.listen()
	go s.handleMessages()

	s.Server = grpc.NewServer(s.opts...)
	pb.RegisterLogServiceServer(s.Server, s.LogSv)

	s.Logger.Log(log.NewMessage().Message(fmt.Sprintf("gRPC server running on port %s", s.Addr)).Build())

	if err := s.Server.Serve(lis); err != nil {
		s.ErrCh <- err
		return
	}

}

func (s GRPCLogServer) Stop() {
	s.LogSv.Done <- struct{}{}
	s.Server.Stop()
}
