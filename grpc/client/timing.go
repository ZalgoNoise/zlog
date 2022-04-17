package client

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	pb "github.com/zalgonoise/zlog/proto/message"
)

// UnaryClientTiming returns a new unary client interceptor that shows the time taken to complete RPCs
func UnaryClientTiming(logger log.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		now := time.Now()

		err := invoker(ctx, method, req, reply, cc, opts...)

		if err != nil {
			// handle errors in the transaction
			logger.Log(event.New().Level(event.LLWarn).Prefix("gRPC").Sub("timer").Message("[recv] unary RPC -- message handling failed with an error").Metadata(event.Field{
				"error":  err.Error(),
				"method": method,
				"time":   time.Since(now).String(),
			}).Build())
		} else if !reply.(*pb.MessageResponse).GetOk() {
			// handle errors in the response; return the error in the message
			logger.Log(event.New().Level(event.LLWarn).Prefix("gRPC").Sub("timer").Message("[recv] unary RPC -- message returned a not-OK status").Metadata(event.Field{
				"error":    reply.(*pb.MessageResponse).GetErr(),
				"method":   method,
				"response": event.Field{"id": reply.(*pb.MessageResponse).GetReqID()},
				"time":     time.Since(now).String(),
			}).Build())

			return errors.New(reply.(*pb.MessageResponse).GetErr())

		} else {
			// log an OK transaction
			logger.Log(event.New().Level(event.LLDebug).Prefix("gRPC").Sub("timer").Message("[recv] unary RPC").Metadata(event.Field{
				"method":   method,
				"response": event.Field{"id": reply.(*pb.MessageResponse).GetReqID()},
				"time":     time.Since(now).String(),
			}).Build())
		}
		return err
	}
}

// StreamClientTiming returns a new stream client interceptor that shows the time taken to complete RPCs
func StreamClientTiming(logger log.Logger) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		now := time.Now()

		clientStream, err := streamer(ctx, desc, cc, method, opts...)

		if err != nil {
			logger.Log(event.New().Level(event.LLWarn).Prefix("gRPC").Sub("timer").Message("[conn] stream RPC -- failed to initialize stream with an error").Metadata(event.Field{
				"error":  err.Error(),
				"method": method,
				"stream": desc.StreamName,
				"time":   time.Since(now).String(),
			}).Build())

		}

		logger.Log(event.New().Level(event.LLDebug).Prefix("gRPC").Sub("timer").Message("[conn] stream RPC -- connection was established").Metadata(event.Field{
			"time": time.Since(now).String(),
		}).Build())

		wStream := timingStream{
			stream: clientStream,
			logger: logger,
			method: method,
			name:   desc.StreamName,
		}

		return wStream, err
	}
}

type timingStream struct {
	stream grpc.ClientStream
	logger log.Logger
	method string
	name   string
}

// Header method is a wrapper for the grpc.ClientStream.Header() method
func (w timingStream) Header() (metadata.MD, error) { return w.stream.Header() }

// Trailer method is a wrapper for the grpc.ClientStream.Trailer() method
func (w timingStream) Trailer() metadata.MD { return w.stream.Trailer() }

// CloseSend method is a wrapper for the grpc.ClientStream.CloseSend() method
func (w timingStream) CloseSend() error { return w.stream.CloseSend() }

// Context method is a wrapper for the grpc.ClientStream.Context() method
func (w timingStream) Context() context.Context { return w.stream.Context() }

// SendMsg method is a wrapper for the grpc.ClientStream.SendMsg(m) method, for which the
// configured logger will register outbound messages' timing
func (w timingStream) SendMsg(m interface{}) error {
	start := time.Now()
	err := w.stream.SendMsg(m)
	w.logger.Log(event.New().Level(event.LLDebug).Prefix("gRPC").Sub("timer").Message("[send] stream RPC").Metadata(event.Field{
		"time":   time.Since(start).String(),
		"method": w.method,
		"name":   w.name,
	}).Build())

	return err
}

// RecvMsg method is a wrapper for the grpc.ClientStream.RecvMsg(m) method, for which the
// configured logger will register inbound messages' timing
func (w timingStream) RecvMsg(m interface{}) error {
	start := time.Now()
	err := w.stream.RecvMsg(m)
	w.logger.Log(event.New().Level(event.LLDebug).Prefix("gRPC").Sub("timer").Message("[recv] stream RPC").Metadata(event.Field{
		"time":   time.Since(start).String(),
		"method": w.method,
		"name":   w.name,
	}).Build())
	return err

}
