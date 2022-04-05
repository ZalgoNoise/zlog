package server

import (
	"errors"
	"net"

	"github.com/google/uuid"
	"github.com/zalgonoise/zlog/log"
	pb "github.com/zalgonoise/zlog/proto/message"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	ErrMessageParse error = errors.New("failed to parse message")
	ErrAddrListen   error = errors.New("failed to listen to input address")

	grpcServer *grpc.Server
)

// GRPCLogServer struct will define the elements required to build and work with
// a gRPC Log Server.
//
// Besides the gRPC-related elements, this struct will contain two Loggers (Logger
// and SvcLogger). This allows the gRPC Server to both do its job -- and register
// any (important) log messages to a different output, with its own configuration
// and requirements.
type GRPCLogServer struct {
	Addr      string
	opts      []grpc.ServerOption
	Logger    log.Logger
	SvcLogger log.Logger
	ErrCh     chan error
	LogSv     *pb.LogServer
	// Server    *grpc.Server
}

// New function will create a new gRPC Log Server, ensuring that at least the default
// settings are applied.
//
// Once the Log Server is configured, a goroutine is kicked off to listen to internal
// comms (the registerComms() method), which will route runtime-related log messages
// to its SvcLogger
func New(confs ...LogServerConfig) *GRPCLogServer {
	server := &GRPCLogServer{
		ErrCh: make(chan error),
		LogSv: pb.NewLogServer(),
	}

	// enforce defaults
	defaultConfig.Apply(server)

	// apply input configs
	for _, config := range confs {
		config.Apply(server)
	}

	go server.registerComms()

	return server

}

// registerComms method will listen to messages from the Log Server's Comm channel, and register
// them in the service logger accordingly.
func (s GRPCLogServer) registerComms() {
	for {
		msg, ok := <-s.LogSv.Comm
		if !ok {
			s.SvcLogger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("LogServer.Comm").Message("couldn't parse message from LogServer").Metadata(log.Field{"error": ErrMessageParse.Error()}).Build())
			continue
		}

		s.SvcLogger.Log(log.NewMessage().FromProto(msg).Build())
	}
}

// listen method will start listening on the provided address, sending any errors to the Log Server's
// error channel.
//
// If there are no errors setting up this listener, the function returns a net.Listener
func (s GRPCLogServer) listen() net.Listener {
	lis, err := net.Listen("tcp", s.Addr)

	if err != nil {
		s.ErrCh <- err

		s.SvcLogger.Log(log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("listen").
			Message("couldn't listen to input address").Metadata(log.Field{
			"error": err.Error(),
			"addr":  s.Addr,
		}).Build())

		return nil
	}

	s.SvcLogger.Log(log.NewMessage().Level(log.LLInfo).Prefix("gRPC").Sub("listen").
		Message("gRPC server is listening to connections").Metadata(log.Field{
		"addr": s.Addr,
	}).Build())

	return lis
}

// handleResponses method will take care of registering an input log message in the
// (actual) target Logger.
//
// This is done via the Output() method which, like the io.Writer, returns the
// number of bytes written and an error. From this point, depending on the outcome,
// a pb.MessageResponse object is built and sent to the Responses channel
func (s GRPCLogServer) handleResponses(logmsg *log.LogMessage) {
	n, err := s.Logger.Output(logmsg)
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

		s.SvcLogger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("handler").Message("issue writting log message").Metadata(log.Field{"error": errStr, "bytesWritten": n}).Build())

		// send not OK response
		s.LogSv.Resp <- &pb.MessageResponse{
			Ok:    false,
			ReqID: reqID,
			Err:   &errStr,
			Bytes: &n32,
		}
		return
	}

	s.SvcLogger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("handler").Message("input log message parsed and registered").Build())

	// send OK response
	s.LogSv.Resp <- &pb.MessageResponse{
		Ok:    true,
		ReqID: reqID,
		Bytes: &n32,
	}
}

// handleMessages method will be a (blocking) function kicked off as a go-routine
// which will take in messages from the Log Server's message channel and register them
func (s GRPCLogServer) handleMessages() {
	s.SvcLogger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("handler").Message("message handler is running").Build())

	// avoid calling Done() method repeatedly
	done := s.LogSv.Done()

	for {
		select {
		// new message is received
		case msg := <-s.LogSv.MsgCh:

			// convert pb.MessageRequest to log.LogMessage
			logmsg := log.NewMessage().FromProto(msg).Build()

			// send message to be written in a goroutine
			go s.handleResponses(logmsg)

		// done signal is received
		case <-done:
			s.SvcLogger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("handler").Message("received done signal").Build())
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
	pb.RegisterLogServiceServer(grpcServer, s.LogSv)

	// gRPC reflection
	reflection.Register(grpcServer)

	s.SvcLogger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("serve").
		Message("gRPC server is running").Metadata(log.Field{
		"addr": s.Addr,
	}).Build())

	if err := grpcServer.Serve(lis); err != nil {
		s.ErrCh <- err

		s.SvcLogger.Log(log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("serve").
			Message("gRPC server crashed with an error").Metadata(log.Field{
			"error": err.Error(),
			"addr":  s.Addr,
		}).Build())
		return
	}

}

// Stop method will be a wrapper for the routine involved to (gracefully) stop this gRPC
// Log Server. It will first call the
func (s GRPCLogServer) Stop() {
	s.LogSv.Stop()
	grpcServer.Stop()

	s.SvcLogger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("Stop").Message("srv: received done signal").Build())
}
