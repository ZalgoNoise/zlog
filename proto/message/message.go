package message

import (
	"context"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const commPrefix string = "gRPC"

var (
	ErrNoResponse = errors.New("couldn't receive a response for the write request, from the logging module")

	contextCancelledRegexp = regexp.MustCompile(`rpc error: code = Canceled desc = context canceled`)
)

// LogServer struct defines the elements of a gRPC Log Server, used as a log message, internal logs,
// errors and done channel router, for a GRPCLogServer object.
type LogServer struct {
	MsgCh chan *MessageRequest
	Done  chan struct{}
	ErrCh chan error
	Comm  chan *MessageRequest
	Resp  chan *MessageResponse
}

// NewLogServer is a placeholder function to create a LogServer object, which returns
// a pointer to a LogServer with initialized channels
func NewLogServer() *LogServer {
	return &LogServer{
		MsgCh: make(chan *MessageRequest),
		Done:  make(chan struct{}),
		Comm:  make(chan *MessageRequest),
		Resp:  make(chan *MessageResponse),
	}
}

func newComm(level int32, method string, msg ...string) *MessageRequest {
	l := Level(level)
	p := commPrefix
	s := method

	sb := strings.Builder{}

	for _, m := range msg {
		sb.WriteString(m)
	}

	return &MessageRequest{
		Time:   timestamppb.New(time.Now()),
		Prefix: &p,
		Sub:    &s,
		Level:  &l,
		Msg:    sb.String(),
	}
}

// Log method implements the LogServiceClient interface
func (s *LogServer) Log(ctx context.Context, in *MessageRequest) (*MessageResponse, error) {
	fName := "Log"

	// collect request ID from context (if set)
	reqID := getRequestID(ctx)

	s.Comm <- newComm(1, fName, "recv: [", reqID, "]")

	s.MsgCh <- in

	s.Comm <- newComm(1, fName, "send: [", reqID, "]")
	return &MessageResponse{Ok: true}, nil
}

// LogStream method implements the LogServiceClient interface
func (s *LogServer) LogStream(stream LogService_LogStreamServer) error {
	go s.logStream(stream)

	for {
		err := <-s.ErrCh
		return err
	}
}

func (s *LogServer) logStream(stream LogService_LogStreamServer) {
	fName := "logStream"

	for {
		// get incoming messages from stream
		in, err := stream.Recv()

		// collect request ID from context (if set)
		reqID := getRequestID(stream.Context())

		// check for errors
		if err != nil {

			// error is EOF -- stream disconnected
			// break from this loop / keep listening for connections
			if err == io.EOF {
				s.Comm <- newComm(1, fName, "recv: got EOF from [", reqID, "]")
				// break
				continue
			}

			// context cancelled by client -- exit
			if contextCancelledRegexp.MatchString(err.Error()) {
				s.Comm <- newComm(2, fName, "recv: got context closure from [", reqID, "] :: ", err.Error())
				return
			}

			// other errors are logged and sent to the error channel, response sent to client
			// -- then, exit
			s.Comm <- newComm(4, fName, "recv: got error from [", reqID, "] :: ", err.Error())
			s.ErrCh <- err

			// send Not OK message to client
			err = stream.Send(&MessageResponse{Ok: false})
			if err != nil {
				// handle send errors if existing
				// log level warning since it's an issue with the client
				s.Comm <- newComm(3, fName, "send: got error with [", reqID, "] :: ", err.Error())
				s.ErrCh <- err
				return
			}

			return
		}

		// register a recv transaction with request ID
		s.Comm <- newComm(1, fName, "recv: [", reqID, "]")
		// send new (valid) message to the messages channel to be logged
		s.MsgCh <- in

		res, ok := <-s.Resp

		if !ok {
			err := ErrNoResponse.Error()
			res = &MessageResponse{
				Ok:  false,
				Err: &err,
			}
		}

		// register a send transaction with request ID
		s.Comm <- newComm(1, fName, "send: [", reqID, "]")
		// send OK response to client
		err = stream.Send(res)
		if err != nil {
			// handle send errors if existing
			// log level warning since it's an issue with the client
			s.Comm <- newComm(3, fName, "send: got error with [", reqID, "] :: ", err.Error())
			s.ErrCh <- err
			return
		}
	}
}

func getRequestID(ctx context.Context) string {
	reqID := CtxGet(ctx, RequestIDKey)
	if len(reqID) == 0 || reqID[0] == "" {
		return DefaultRequestID
	}
	return reqID[0]
}
