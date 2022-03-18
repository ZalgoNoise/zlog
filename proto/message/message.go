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
	// process input message
	// logmsg := NewMessage().FromProto(msg).Build()
	// fmt.Println(logmsg)
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

// type LogStreamServer struct {
// 	MsgCh chan *MessageRequest
// 	Done  chan struct{}
// }

// func (s *LogServer) LogStream(stream LogService_LogStreamClient) error {
// 	for {
// 		in, err := stream.Recv()
// 		if err == io.EOF {
// 			return nil
// 		}
// 		if err != nil {
// 			return err
// 		}
// 		if !in.GetOk() {
// 			return errors.New("gRPC server failed to write the log message")
// 		}
// 	}
// }

// func NewLogStreamServer() *LogStreamServer {
// 	return &LogStreamServer{
// 		MsgCh: make(chan *MessageRequest),
// 		Done:  make(chan struct{}),
// 	}
// }
