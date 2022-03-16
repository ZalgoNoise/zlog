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
