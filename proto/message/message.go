package message

import (
	context "context"
)

type LogServer struct {
	MsgCh chan *MessageRequest
	Done  chan struct{}
}

func (s *LogServer) Log(ctx context.Context, msg *MessageRequest) (*MessageResponse, error) {
	// process input message
	// logmsg := NewMessage().FromProto(msg).Build()
	// fmt.Println(logmsg)
	s.MsgCh <- msg

	return &MessageResponse{Ok: true}, nil
}

func NewLogServer() *LogServer {
	return &LogServer{
		MsgCh: make(chan *MessageRequest),
		Done:  make(chan struct{}),
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
