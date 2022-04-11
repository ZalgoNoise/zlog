package client

import (
	"context"
	"errors"

	"github.com/zalgonoise/zlog/log"
	pb "github.com/zalgonoise/zlog/proto/message"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryClientLogging returns a new unary client interceptor that adds a gRPC Client Logger
// which captures inbound / outbound interactions with the service
func UnaryClientLogging(logger log.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("logger").Message("[send] unary RPC logger -- " + method).Build())

		err := invoker(ctx, method, req, reply, cc, opts...)

		// validate response fields
		var res = log.Field{}
		if r, ok := reply.(*pb.MessageResponse); ok && r != nil {
			res["ok"] = r.GetOk()
			res["id"] = r.GetReqID()

			if r.GetBytes() > 0 {
				res["bytes"] = r.GetBytes()
			}

			if r.GetErr() != "" {
				res["error"] = r.GetErr()
				if err == nil {
					err = errors.New(r.GetErr())
				}
			}
		}

		if err != nil {
			// handle errors in the transaction
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Message("[recv] unary RPC logger -- message handling failed with an error").Metadata(log.Field{
				"error":    err.Error(),
				"method":   method,
				"response": res,
			}).Build())
		} else if !reply.(*pb.MessageResponse).GetOk() {
			// handle errors in the response; return the error in the message
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Message("[recv] unary RPC logger -- message returned a not-OK status").Metadata(log.Field{
				"error":    reply.(*pb.MessageResponse).GetErr(),
				"method":   method,
				"response": res,
			}).Build())

			return errors.New(reply.(*pb.MessageResponse).GetErr())

		} else {
			// log an OK transaction
			logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("logger").Message("[recv] unary RPC logger").Metadata(log.Field{
				"method":   method,
				"response": res,
			}).Build())
		}
		return err
	}
}

// StreamClientLogging returns a new stream client interceptor that adds a gRPC Client Logger
// which captures inbound / outbound interactions with the service
func StreamClientLogging(logger log.Logger) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

		logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("logger").Message("[open] stream RPC logger connection open -- " + method).Build())

		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Message("[conn] stream RPC logger -- failed to initialize stream with an error").Metadata(log.Field{
				"error":  err.Error(),
				"method": method,
				"stream": desc.StreamName,
			}).Build())

		}

		logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("logger").Message("[conn] stream RPC logger -- connection was established").Build())

		wStream := loggingStream{
			stream: clientStream,
			logger: logger,
			method: method,
			name:   desc.StreamName,
		}

		return wStream, err
	}
}

type loggingStream struct {
	stream grpc.ClientStream
	logger log.Logger
	method string
	name   string
}

// Header method is a wrapper for the grpc.ClientStream.Header() method
func (w loggingStream) Header() (metadata.MD, error) { return w.stream.Header() }

// Trailer method is a wrapper for the grpc.ClientStream.Trailer() method
func (w loggingStream) Trailer() metadata.MD { return w.stream.Trailer() }

// CloseSend method is a wrapper for the grpc.ClientStream.CloseSend() method
func (w loggingStream) CloseSend() error { return w.stream.CloseSend() }

// Context method is a wrapper for the grpc.ClientStream.Context() method
func (w loggingStream) Context() context.Context { return w.stream.Context() }

// SendMsg method is a wrapper for the grpc.ClientStream.SendMsg(m) method, for which the
// configured logger will register outbound messages or errors
func (w loggingStream) SendMsg(m interface{}) error {
	err := w.stream.SendMsg(m)

	if err != nil {
		w.logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Message("[send] stream RPC logger -- sending message resulted in an error").Metadata(log.Field{
			"error":  err.Error(),
			"method": w.method,
			"stream": w.name,
		}).Build())
		return err
	}
	w.logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("logger").Message("[send] stream RPC logger -- sent log message to gRPC server").Build())

	return err
}

// RecvMsg method is a wrapper for the grpc.ClientStream.RecvMsg(m) method, for which the
// configured logger will register inbound messages or errors
func (w loggingStream) RecvMsg(m interface{}) error {
	err := w.stream.RecvMsg(m)

	// check server response for errors
	if err != nil {
		meta := log.Field{
			"error":  err.Error(),
			"method": w.method,
			"stream": w.name,
		}

		if m.(*pb.MessageResponse) != nil && m.(*pb.MessageResponse).GetReqID() != "" {
			meta["id"] = m.(*pb.MessageResponse).GetReqID()
		}

		w.logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Metadata(meta).Message("[recv] stream RPC logger -- issue receiving message from stream").Build())
		return err
	}

	// there are no errors in the response; check the response's OK value
	// if not OK, register this as a local bad response error and continue
	if !m.(*pb.MessageResponse).GetOk() {
		var err error
		if m.(*pb.MessageResponse).GetErr() != "" {
			err = errors.New(m.(*pb.MessageResponse).GetErr())
		} else {
			err = ErrBadResponse
		}
		w.logger.Log(
			log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Metadata(log.Field{
				"id": m.(*pb.MessageResponse).GetReqID(),
				"response": log.Field{
					"ok":    m.(*pb.MessageResponse).GetOk(),
					"bytes": m.(*pb.MessageResponse).GetBytes(),
					"error": err.Error(),
				},
				"error": err.Error(),
			}).Message("[recv] stream RPC logger -- failed to write log message").Build(),
		)
		return err
	}

	// server response is OK, register this event
	w.logger.Log(
		log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("logger").Metadata(log.Field{
			"id": m.(*pb.MessageResponse).GetReqID(),
			"response": log.Field{
				"ok":    m.(*pb.MessageResponse).GetOk(),
				"bytes": m.(*pb.MessageResponse).GetBytes(),
			},
		}).Message("[recv] stream RPC logger -- registering server response").Build(),
	)

	return err

}
