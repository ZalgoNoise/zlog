package message

import (
	"context"
	"fmt"
	"io"
)

type LogServer struct {
	MsgCh chan *MessageRequest
	Done  chan struct{}
	ErrCh chan error
}

func (s *LogServer) Log(ctx context.Context, in *MessageRequest) (*MessageResponse, error) {
	// fmt.Print(CtxGet(ctx), "::")
	s.MsgCh <- in

	return &MessageResponse{Ok: true}, nil
}

func NewLogServer() *LogServer {
	return &LogServer{
		MsgCh: make(chan *MessageRequest),
		Done:  make(chan struct{}),
	}
}

func (s *LogServer) LogStream(stream LogService_LogStreamServer) error {
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				s.ErrCh <- err
				return
			}
			// ctx := stream.Context()
			// fmt.Print(CtxGet(ctx), "::")

			s.MsgCh <- in
			err = stream.Send(&MessageResponse{Ok: true})
			if err != nil {
				s.ErrCh <- err
				return
			}
		}
	}()

	for {
		err := <-s.ErrCh
		fmt.Println(err)
	}
}
