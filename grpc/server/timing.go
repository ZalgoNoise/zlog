package server

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	pb "github.com/zalgonoise/zlog/proto/message"
)

// UnaryServerTiming returns a new unary server interceptor that adds a gRPC Server Logger
// which times inbound / outbound interactions with the service
func UnaryServerTiming(logger log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		now := time.Now()
		method := info.FullMethod

		res, err := handler(ctx, req)

		after := time.Since(now)

		meta := event.Field{
			"method": method,
			"time":   after.String(),
		}

		if err != nil {
			meta["error"] = err.Error()

			logger.Log(event.New().Level(event.Level_warn).Prefix("gRPC").Sub("timer").
				Message("[send] unary RPC -- message handling failed with an error").Metadata(meta).Build())
		} else {
			meta["response"] = event.Field{"id": res.(*pb.MessageResponse).GetReqID()}

			logger.Log(event.New().Level(event.Level_trace).Prefix("gRPC").Sub("timer").
				Message("[send] unary RPC").Metadata(meta).Build())
		}

		return res, err
	}
}

// StreamServerTiming returns a new stream server interceptor that adds a gRPC Server Logger
// which times inbound / outbound interactions with the service
//
// To be able to safely capture the message exchange within the stream, a wrapper is created
// containing the logger, the stream and the method name. This wrapper will implement the
// grpc.ServerStream interface, to add new actions when sending and receiving a message.
//
// This assures that the stream is untouched while still adding a new feature to each exchange.
func StreamServerTiming(logger log.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		now := time.Now()
		method := info.FullMethod

		wStream := loggingStream{
			stream: stream,
			logger: logger,
			method: method,
		}

		err := handler(srv, wStream)

		after := time.Since(now)

		var meta = event.Field{
			"method": method,
			"time":   after.String(),
		}

		if err != nil {
			meta["error"] = err.Error()

			logger.Log(event.New().Level(event.Level_warn).Prefix("gRPC").Sub("timer").
				Message("[conn] stream RPC -- failed to initialize stream with an error").Metadata(meta).Build())
			return err
		}

		logger.Log(event.New().Level(event.Level_trace).Prefix("gRPC").Sub("timer").
			Message("[conn] stream RPC ").Metadata(meta).Build())

		return err
	}
}

type timingStream struct {
	stream grpc.ServerStream
	logger log.Logger
	method string
}

// Header method is a wrapper for the grpc.ServerStream.Header() method
func (w timingStream) SetHeader(m metadata.MD) error { return w.stream.SetHeader(m) }

// Trailer method is a wrapper for the grpc.ServerStream.Trailer() method
func (w timingStream) SendHeader(m metadata.MD) error { return w.stream.SendHeader(m) }

// CloseSend method is a wrapper for the grpc.ServerStream.CloseSend() method
func (w timingStream) SetTrailer(m metadata.MD) { w.stream.SetTrailer(m) }

// Context method is a wrapper for the grpc.ServerStream.Context() method
func (w timingStream) Context() context.Context { return w.stream.Context() }

// SendMsg method is a wrapper for the grpc.ServerStream.SendMsg(m) method, for which the
// configured logger will register outbound messages or errors
func (w timingStream) SendMsg(m interface{}) error {
	now := time.Now()

	err := w.stream.SendMsg(m)

	after := time.Since(now)

	var res = event.Field{}
	var meta = event.Field{
		"method": w.method,
		"time":   after.String(),
	}

	if err != nil {
		meta["error"] = err.Error()

		if m.(*pb.MessageResponse).GetReqID() != "" {
			res["id"] = m.(*pb.MessageResponse).GetReqID()
			meta["response"] = res
		}

		w.logger.Log(event.New().Level(event.Level_warn).Prefix("gRPC").Sub("timer").
			Message("[send] stream RPC logger -- error sending message").Metadata(meta).Build())
		return err
	}

	res["id"] = m.(*pb.MessageResponse).GetReqID()
	meta["response"] = res

	w.logger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("timer").
		Message("[send] stream RPC logger -- sent message to gRPC client").Metadata(meta).Build())
	return err
}

// RecvMsg method is a wrapper for the grpc.ServerStream.RecvMsg(m) method, for which the
// configured logger will register inbound messages or errors
func (w timingStream) RecvMsg(m interface{}) error {

	now := time.Now()

	err := w.stream.RecvMsg(m)

	after := time.Since(now)

	var meta = event.Field{
		"method": w.method,
		"time":   after.String(),
	}

	if err != nil {
		meta["error"] = err.Error()

		// default error handling
		w.logger.Log(event.New().Level(event.Level_warn).Prefix("gRPC").Sub("timer").
			Message("[recv] stream RPC logger -- error receiving message").Metadata(meta).Build())
		return err

	}
	w.logger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("logger").
		Message("[recv] stream RPC logger -- received message from gRPC client").Metadata(meta).Build())
	return err
}
