package server

import (
	"context"
	"errors"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/zalgonoise/zlog/log"
	pb "github.com/zalgonoise/zlog/proto/message"
)

var (
// ErrContextCancelledRegexp = regexp.MustCompile(`rpc error: code = Canceled desc = context canceled`)
)

// UnaryServerLogging returns a new unary server interceptor that adds a gRPC Server Logger
// which captures inbound / outbound interactions with the service
func UnaryServerLogging(logger log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method := info.FullMethod

		logger.Log(log.NewMessage().Level(log.LLTrace).Prefix("gRPC").Sub("logger").Message("[recv] unary RPC -- " + method).Build())

		res, err := handler(ctx, req)

		if err != nil {
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Message("[send] unary RPC -- message handling failed with an error").Metadata(log.Field{
				"error":  err.Error(),
				"method": method,
			}).Build())
		} else {
			logger.Log(log.NewMessage().Level(log.LLTrace).Prefix("gRPC").Sub("logger").Message("[send] unary RPC -- " + method).Metadata(log.Field{
				"id": res.(*pb.MessageResponse).GetReqID(),
				"ok": res.(*pb.MessageResponse).GetOk(),
			}).Build())
		}

		return res, err
	}
}

// StreamServerLogging returns a new stream server interceptor that adds a gRPC Server Logger
// which captures inbound / outbound interactions with the service
//
// To be able to safely capture the message exchange within the stream, a wrapper is created
// containing the logger, the stream and the method name. This wrapper will implement the
// grpc.ServerStream interface, to add new actions when sending and receiving a message.
//
// This assures that the stream is untouched while still adding a new feature to each exchange.
func StreamServerLogging(logger log.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		method := info.FullMethod

		logger.Log(log.NewMessage().Level(log.LLTrace).Prefix("gRPC").Sub("logger").Message("[open] stream RPC -- " + method).Build())

		wStream := loggingStream{
			stream: stream,
			logger: logger,
			method: method,
		}

		err := handler(srv, wStream)

		if err != nil {
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Message("[conn] stream RPC -- failed to initialize stream with an error").Metadata(log.Field{
				"error":  err.Error(),
				"method": method,
			}).Build())
			return err
		}

		logger.Log(log.NewMessage().Level(log.LLTrace).Prefix("gRPC").Sub("logger").Message("[conn] stream RPC -- " + method).Build())

		return err
	}
}

type loggingStream struct {
	stream grpc.ServerStream
	logger log.Logger
	method string
}

// Header method is a wrapper for the grpc.ServerStream.Header() method
func (w loggingStream) SetHeader(m metadata.MD) error { return w.stream.SetHeader(m) }

// Trailer method is a wrapper for the grpc.ServerStream.Trailer() method
func (w loggingStream) SendHeader(m metadata.MD) error { return w.stream.SendHeader(m) }

// CloseSend method is a wrapper for the grpc.ServerStream.CloseSend() method
func (w loggingStream) SetTrailer(m metadata.MD) { w.stream.SetTrailer(m) }

// Context method is a wrapper for the grpc.ServerStream.Context() method
func (w loggingStream) Context() context.Context { return w.stream.Context() }

// SendMsg method is a wrapper for the grpc.ServerStream.SendMsg(m) method, for which the
// configured logger will register outbound messages or errors
func (w loggingStream) SendMsg(m interface{}) error {
	err := w.stream.SendMsg(m)
	if err != nil {
		meta := log.Field{
			"error":  err.Error(),
			"method": w.method,
		}

		if m.(*pb.MessageResponse).GetReqID() != "" {
			meta["id"] = m.(*pb.MessageResponse).GetReqID()
		}

		w.logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Message("[send] stream RPC logger -- error sending message").Metadata(meta).Build())
		return err
	}

	if !m.(*pb.MessageResponse).GetOk() {
		var err error
		if m.(*pb.MessageResponse).GetErr() != "" {
			err = errors.New(m.(*pb.MessageResponse).GetErr())
		} else {
			err = ErrMessageParse
		}
		w.logger.Log(
			log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Metadata(log.Field{
				"id":     m.(*pb.MessageResponse).GetReqID(),
				"method": w.method,
				"response": log.Field{
					"ok":    m.(*pb.MessageResponse).GetOk(),
					"bytes": m.(*pb.MessageResponse).GetBytes(),
					"error": err.Error(),
				},
				"error": err.Error(),
			}).Message("[send] stream RPC logger -- failed to send response message").Build(),
		)
		return err

	}

	w.logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("logger").Message("[send] stream RPC logger -- sent message to gRPC client").Metadata(log.Field{
		"id":     m.(*pb.MessageResponse).GetReqID(),
		"method": w.method,
		"response": log.Field{
			"ok":    m.(*pb.MessageResponse).GetOk(),
			"bytes": m.(*pb.MessageResponse).GetBytes(),
		},
	}).Build())
	return err
}

// RecvMsg method is a wrapper for the grpc.ServerStream.RecvMsg(m) method, for which the
// configured logger will register inbound messages or errors
func (w loggingStream) RecvMsg(m interface{}) error {
	err := w.stream.RecvMsg(m)
	if err != nil {

		// handle EOF
		if errors.Is(err, io.EOF) {
			w.logger.Log(log.NewMessage().Level(log.LLInfo).Prefix("gRPC").Sub("logger").Message("[recv] stream RPC logger -- received EOF from client").Metadata(log.Field{
				"error":  err.Error(),
				"method": w.method,
			}).Build())
			return err
		}

		// handle context cancelled
		if errCode := status.Code(err); errCode == codes.DeadlineExceeded {
			w.logger.Log(log.NewMessage().Level(log.LLInfo).Prefix("gRPC").Sub("logger").Message("[recv] stream RPC logger -- received context closure from client").Metadata(log.Field{
				"error":  err.Error(),
				"method": w.method,
			}).Build())
			return err
		}

		// default error handling
		w.logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Message("[recv] stream RPC logger -- error receiving message").Metadata(log.Field{
			"error":  err.Error(),
			"method": w.method,
		}).Build())
		return err

	}
	w.logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("logger").Message("[recv] stream RPC logger -- received message from gRPC client").Build())
	return err
}
