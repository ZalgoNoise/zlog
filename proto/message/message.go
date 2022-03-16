package message

import (
	context "context"
)

type LogServer struct {
	msgCh chan *MessageRequest
}

func (s *LogServer) Log(ctx context.Context, msg *MessageRequest) (*MessageResponse, error) {
	// process input message
	// logmsg := NewMessage().FromProto(msg).Build()
	// fmt.Println(logmsg)
	s.msgCh <- msg

	return &MessageResponse{Ok: true}, nil
}

func NewLogServer(msgCh chan *MessageRequest) *LogServer {
	return &LogServer{
		msgCh: msgCh,
	}
}
