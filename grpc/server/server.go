package server

import (
	"errors"
	"net"

	"github.com/google/uuid"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	pb "github.com/zalgonoise/zlog/proto/service"
)

var (
	ErrMessageParse error = errors.New("failed to parse message")
	ErrAddrListen   error = errors.New("failed to listen to input address")

	grpcServer *grpc.Server
)

// LogServer interface defines the behavior of a gRPC Log Server, which focuses
// mostly on its (background) runtime operations.
//
// The interface exposes a `Channels()` method nevertheless, which allows for local
// control of log messages (and error handling) without having to send a message
// over-the-wire
type LogServer interface {
	Serve()
	Stop()
	Channels() (logCh, logSvCh chan *event.Event, errCh chan error)
}

// GRPCLogServer struct will define the elements required to build and work with
// a gRPC Log Server.
//
// Besides the gRPC-related elements, this struct will contain two Loggers (Logger
// and SvcLogger). This allows the gRPC Server to both do its job -- and register
// any (important) log events to a different output, with its own configuration
// and requirements.
type GRPCLogServer struct {
	addr      string
	opts      []grpc.ServerOption
	logger    log.Logger
	svcLogger log.Logger
	errCh     chan error
	logSv     *pb.LogServer
}

// gRPCLogServerBuilder is a helper data structure to spawn new GRPCLogServers
type gRPCLogServerBuilder struct {
	addr         string
	opts         []grpc.ServerOption
	interceptors serverInterceptors
	logger       log.Logger
	svcLogger    log.Logger
	errCh        chan error
	logSv        *pb.LogServer
}

// build method will merge the stream / unary gRPC interceptors as []grpc.ServerOption
func (b *gRPCLogServerBuilder) build() *GRPCLogServer {
	// auto merge stream / unary interceptors as []grpc.ServerOption
	var opts []grpc.ServerOption

	if len(b.interceptors.streamItcp) > 0 {
		var interceptors []grpc.UnaryServerInterceptor

		for _, i := range b.interceptors.unaryItcp {
			interceptors = append(interceptors, i)
		}

		uItcp := grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...))
		opts = append(b.opts, uItcp)
	}

	if len(b.interceptors.unaryItcp) > 0 {
		var interceptors []grpc.StreamServerInterceptor

		for _, i := range b.interceptors.streamItcp {
			interceptors = append(interceptors, i)
		}

		sItcp := grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(interceptors...))
		opts = append(opts, sItcp)
	}

	return &GRPCLogServer{
		addr:      b.addr,
		opts:      append(b.opts, opts...),
		logger:    b.logger,
		svcLogger: b.svcLogger,
		errCh:     make(chan error),
		logSv:     pb.NewLogServer(),
	}

}

// serverInterceptors struct is a placeholder for different interceptors to be added
// to the GRPCLogServer
type serverInterceptors struct {
	streamItcp map[string]grpc.StreamServerInterceptor
	unaryItcp  map[string]grpc.UnaryServerInterceptor
}

// New function will create a new gRPC Log Server, ensuring that at least the default
// settings are applied.
//
// Once the Log Server is configured, a goroutine is kicked off to listen to internal
// comms (the registerComms() method), which will route runtime-related log messages
// to its SvcLogger
func New(confs ...LogServerConfig) *GRPCLogServer {
	builder := &gRPCLogServerBuilder{
		interceptors: serverInterceptors{
			streamItcp: make(map[string]grpc.StreamServerInterceptor),
			unaryItcp:  make(map[string]grpc.UnaryServerInterceptor),
		},
	}

	// enforce defaults
	defaultConfig.Apply(builder)

	// apply input configs
	if len(confs) > 0 {
		MultiConf(confs...).Apply(builder)
	}

	// merge configurations / server options & interceptors
	server := builder.build()

	go server.registerComms()

	return server
}

// registerComms method will listen to messages from the Log Server's Comm channel, and register
// them in the service logger accordingly.
func (s GRPCLogServer) registerComms() {
	for {
		msg := <-s.logSv.Comm
		s.svcLogger.Log(msg)
	}
}

// listen method will start listening on the provided address, sending any errors to the Log Server's
// error channel.
//
// If there are no errors setting up this listener, the function returns a net.Listener
func (s GRPCLogServer) listen() net.Listener {
	lis, err := net.Listen("tcp", s.addr)

	if err != nil {
		s.errCh <- err

		s.svcLogger.Log(event.New().Level(event.Level_fatal).Prefix("gRPC").Sub("listen").
			Message("couldn't listen to input address").Metadata(event.Field{
			"error": err.Error(),
			"addr":  s.addr,
		}).Build())

		return nil
	}

	s.svcLogger.Log(event.New().Level(event.Level_info).Prefix("gRPC").Sub("listen").
		Message("gRPC server is listening to connections").Metadata(event.Field{
		"addr": s.addr,
	}).Build())

	return lis
}

// handleResponses method will take care of registering an input log message in the
// (actual) target Logger.
//
// This is done via the Output() method which, like the io.Writer, returns the
// number of bytes written and an error. From this point, depending on the outcome,
// a pb.LogResponse object is built and sent to the Responses channel
func (s GRPCLogServer) handleResponses(logmsg *event.Event) {
	n, err := s.logger.Output(logmsg)
	n32 := int32(n)

	// generate request ID
	reqID := uuid.New().String()

	// handle write errors or zero-bytes-written errors
	if err != nil || n == 0 {
		var errStr string
		if err == nil {
			errStr = "zero bytes written"
		} else {
			errStr = err.Error()
		}

		s.svcLogger.Log(event.New().Level(event.Level_warn).Prefix("gRPC").Sub("handler").Message("issue writting log message").Metadata(event.Field{"error": errStr, "bytesWritten": n}).Build())

		// send not OK response
		s.logSv.Resp <- &pb.LogResponse{
			Ok:    false,
			ReqID: reqID,
			Err:   &errStr,
			Bytes: &n32,
		}
		return
	}

	s.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("handler").Message("input log message parsed and registered").Build())

	// send OK response
	s.logSv.Resp <- &pb.LogResponse{
		Ok:    true,
		ReqID: reqID,
		Bytes: &n32,
	}
}

// handleMessages method will be a (blocking) function kicked off as a go-routine
// which will take in messages from the Log Server's message channel and register them
func (s GRPCLogServer) handleMessages() {
	s.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("handler").Message("message handler is running").Build())

	// avoid calling Done() method repeatedly
	done := s.logSv.Done()

	for {
		select {
		// new message is received
		case msg := <-s.logSv.MsgCh:

			// send message to be written in a goroutine
			go s.handleResponses(msg)

		// done signal is received
		case <-done:
			s.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("handler").Message("received done signal").Build())
			return
		}
	}
}

// Serve method will be a long-running, blocking function which will launch the gRPC server
//
// It will start listening to the resgistered address and launch its internal message handler routine.
// Then, the gRPC Server is created (as a package-level instance), registered for reflection. Finally,
// the grpc.Server's own Serve() method is executed and persisted unless an error occurs.
func (s GRPCLogServer) Serve() {
	lis := s.listen()
	if lis == nil {
		return
	}

	go s.handleMessages()

	grpcServer = grpc.NewServer(s.opts...)
	pb.RegisterLogServiceServer(grpcServer, s.logSv)

	// gRPC reflection
	reflection.Register(grpcServer)

	s.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("serve").
		Message("gRPC server is running").Metadata(event.Field{
		"addr": s.addr,
	}).Build())

	// tests for grpcServer.Serve() are out-of-scope as it is part of the
	// gRPC framework logic
	if err := grpcServer.Serve(lis); err != nil {
		s.errCh <- err

		s.svcLogger.Log(event.New().Level(event.Level_fatal).Prefix("gRPC").Sub("serve").
			Message("gRPC server crashed with an error").Metadata(event.Field{
			"error": err.Error(),
			"addr":  s.addr,
		}).Build())
		return
	}

}

// Stop method will be a wrapper for the routine involved to (gracefully) stop this gRPC
// Log Server.
func (s GRPCLogServer) Stop() {

	if grpcServer != nil {
		grpcServer.Stop()
	}

	s.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("Stop").Message("srv: received done signal").Build())
}

// Channels method returns channels for a Log Server's I/O. It returns a channel for
// log messages (for actual log event writes), a channel for the service logger
// (the server's own logger), and an error channel to collect Log Server errors from.
func (s GRPCLogServer) Channels() (logCh, logSvCh chan *event.Event, errCh chan error) {
	// create new channels to route messages
	logCh = make(chan *event.Event)
	logSvCh = make(chan *event.Event)
	errCh = s.errCh

	// launch logger channel goroutine (direct access to writing logs)
	go func(ch chan *event.Event) {
		for {
			msg := <-ch
			s.logger.Log(msg)
		}
	}(logCh)

	// launch service logger channel goroutine (service logs write access)
	go func(ch chan *event.Event) {
		for {
			msg := <-ch
			s.svcLogger.Log(msg)
		}
	}(logSvCh)

	// return channels
	return logCh, logSvCh, errCh
}
