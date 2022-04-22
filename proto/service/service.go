package service

import (
	"context"
	"errors"
	"io"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/zalgonoise/zlog/log/event"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const commPrefix string = "gRPC"

var (
	ErrNoResponse = errors.New("couldn't receive a response for the write request, from the logging module")

	contextCancelledRegexp = regexp.MustCompile(`rpc error: code = Canceled desc = context canceled`)

	// closedchan is a reusable closed channel.
	closedchan = make(chan struct{})
)

func init() {
	close(closedchan)
}

// LogServer struct defines the elements of a gRPC Log Server, used as a event, internal logs,
// errors and done channel router, for a GRPCLogServer object.
type LogServer struct {
	ErrCh chan error
	MsgCh chan *event.Event
	Comm  chan *event.Event
	Resp  chan *LogResponse
	done  atomic.Value
}

// NewLogServer is a placeholder function to create a LogServer object, which returns
// a pointer to a LogServer with initialized channels
func NewLogServer() *LogServer {
	return &LogServer{
		MsgCh: make(chan *event.Event),
		Comm:  make(chan *event.Event),
		Resp:  make(chan *LogResponse),
	}
}

func newComm(level int32, method string, msg ...string) *event.Event {
	l := event.Level(level)
	p := commPrefix
	s := method

	sb := strings.Builder{}

	for _, m := range msg {
		sb.WriteString(m)
	}

	body := sb.String()

	return &event.Event{
		Time:   timestamppb.New(time.Now()),
		Prefix: &p,
		Sub:    &s,
		Level:  &l,
		Msg:    &body,
	}
}

// Log method implements the LogServiceClient interface
func (s *LogServer) Log(ctx context.Context, in *event.Event) (*LogResponse, error) {
	// send message to be written
	s.MsgCh <- in

	// receive Logger's response
	res, ok := <-s.Resp

	// handle bad responses
	if !ok {
		return nil, ErrNoResponse
	}

	// send OK response
	return res, nil
}

// LogStream method implements the LogServiceClient interface
func (s *LogServer) LogStream(stream LogService_LogStreamServer) error {
	fName := "LogStream"

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// goroutine to listen to stream -- it is closed with s.Stop()
	go s.logStream(ctx, stream)

	done := s.Done()

	for {
		select {
		case <-done:
			s.Comm <- newComm(0, fName, "done signal received; closing goroutine contexts")
			cancel()
			return nil
		case err := <-s.ErrCh:
			s.Comm <- newComm(3, fName, "error received; closing goroutine contexts")
			cancel()
			return err

		}
	}
}

func (s *LogServer) logStream(ctx context.Context, stream LogService_LogStreamServer) {
	fName := "logStream"
	done := ctx.Done()

	// local channel to route input messages and errors
	localCh := make(chan struct {
		in  *event.Event
		err error
	})

	// get incoming messages from stream
	// send to local channel
	go func() {
		for {
			in, err := stream.Recv()

			localCh <- struct {
				in  *event.Event
				err error
			}{
				in:  in,
				err: err,
			}
		}
	}()

	// blocking long-running operation to switch on:
	// - input messages from the local message channel (localCh)
	// - done signals from the input context's done channel
	for {
		select {
		case msg := <-localCh:
			in := msg.in
			err := msg.err

			fallbackUUID := uuid.New().String()

			// check for errors
			if err != nil {

				// error is EOF -- stream disconnected
				// keep listening for connections
				if err == io.EOF {
					continue
				}

				// context cancelled by client -- exit
				if contextCancelledRegexp.MatchString(err.Error()) {
					return
				}

				// other errors are sent to the error channel, response sent to client
				// -- then, exit
				s.ErrCh <- err

				// send Not OK message to client
				errStr := err.Error()
				err = stream.Send(&LogResponse{Ok: false, ReqID: fallbackUUID, Err: &errStr})
				if err != nil {
					// handle send errors if existing
					s.ErrCh <- err
					return
				}

				return
			}

			// send new (valid) message to the messages channel to be logged
			s.MsgCh <- in

			res, ok := <-s.Resp

			if !ok {
				err := ErrNoResponse.Error()
				res = &LogResponse{
					Ok:    false,
					ReqID: fallbackUUID,
					Err:   &err,
				}
			}

			// send OK response to client
			err = stream.Send(res)
			if err != nil {
				// handle send errors if existing
				s.ErrCh <- err
				return
			}

		// context closure ; exit goroutine
		case <-done:
			s.Comm <- newComm(0, fName, "exiting log stream goroutine")
			return
		}
	}

}

// Done method will be similar to context.Context's Done() implementation of the
// cancelCtx. It allocates the done struct as an atomic value, which is created or
// loaded when this method is called.
//
// Just like the context package, this can be used to select over and act upon (for a
// graceful exit request).
//
//     for {
//         select {
//             (...)
//             case <-server.Done():
//                 return
//         }
//     }
//
func (s *LogServer) Done() <-chan struct{} {
	fName := "Done"

	s.Comm <- newComm(0, fName, "listening to done signal")

	d := s.done.Load()

	if d == nil {
		d = make(chan struct{})
		s.done.Store(d)
	}
	return d.(chan struct{})
}

// Stop method will close the LogServer's done channel, which ensures it will halt
// any on-going goroutines gracefully.
func (s *LogServer) Stop() {
	fName := "Stop"
	s.Comm <- newComm(0, fName, "msg: received done signal")

	d := s.done.Load()
	if d == nil {
		s.done.Store(closedchan)
		return
	}
	close(d.(chan struct{}))

}
