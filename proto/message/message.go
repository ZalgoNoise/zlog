package message

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const commPrefix string = "gRPC"

var (
	contextCancelledRegexp = regexp.MustCompile(`rpc error: code = Canceled desc = context canceled`)
)

type LogServer struct {
	MsgCh chan *MessageRequest
	Done  chan struct{}
	ErrCh chan error
	Comm  chan *MessageRequest
}

func newComm(level int32, method, message string) *MessageRequest {
	l := level
	p := commPrefix
	s := method

	return &MessageRequest{
		Time:   timestamppb.New(time.Now()),
		Prefix: &p,
		Sub:    &s,
		Level:  &l,
		Msg:    message,
	}
}

func (s *LogServer) Log(ctx context.Context, in *MessageRequest) (*MessageResponse, error) {

	reqID := CtxGet(ctx)
	s.Comm <- newComm(1, "Log", fmt.Sprintf("recv: %s", reqID))

	s.MsgCh <- in

	s.Comm <- newComm(1, "Log", fmt.Sprintf("send: %s", reqID))
	return &MessageResponse{Ok: true}, nil
}

func NewLogServer() *LogServer {
	return &LogServer{
		MsgCh: make(chan *MessageRequest),
		Done:  make(chan struct{}),
		Comm:  make(chan *MessageRequest),
	}
}

func (s *LogServer) LogStream(stream LogService_LogStreamServer) error {
	go func() {
		for {
			in, err := stream.Recv()
			ctx := stream.Context()
			reqID := CtxGet(ctx)

			if err == io.EOF {
				s.Comm <- newComm(1, "LogStream", fmt.Sprintf("recv: got EOF from %s", reqID))
				break
			}
			if err != nil {
				if contextCancelledRegexp.MatchString(err.Error()) {
					s.Comm <- newComm(2, "LogStream", fmt.Sprintf("recv: got context closure from %s :: %s", reqID, err))
					return
				}

				s.Comm <- newComm(4, "LogStream", fmt.Sprintf("recv: got error from %s :: %s", reqID, err))

				s.ErrCh <- err
				return
			}

			s.Comm <- newComm(1, "LogStream", fmt.Sprintf("recv: %s", reqID))

			s.MsgCh <- in

			s.Comm <- newComm(1, "LogStream", fmt.Sprintf("send: %s", reqID))
			err = stream.Send(&MessageResponse{Ok: true})
			if err != nil {
				s.Comm <- newComm(4, "LogStream", fmt.Sprintf("send: got error with %s :: %s", reqID, err))
				s.ErrCh <- err
				return
			}
		}
	}()

	for {
		err := <-s.ErrCh
		return err
	}
}
