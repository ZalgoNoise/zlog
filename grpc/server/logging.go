package server

import (
	"context"
	"errors"
	"io"
	"time"

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
func UnaryServerLogging(logger log.Logger, withTimer bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var now time.Time
		var after time.Duration

		if withTimer {
			now = time.Now()
		}

		method := info.FullMethod

		logger.Log(log.NewMessage().Level(log.LLTrace).Prefix("gRPC").Sub("logger").Message("[recv] unary RPC -- " + method).Build())

		res, err := handler(ctx, req)

		if withTimer {
			after = time.Since(now)
		}

		var meta = log.Field{}
		meta["method"] = method

		if withTimer {
			meta["time"] = after.String()
		}

		if err != nil {
			meta["error"] = err.Error()

			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Message("[send] unary RPC -- message handling failed with an error").Metadata(meta).Build())
		} else {
			meta["id"] = res.(*pb.MessageResponse).GetReqID()
			meta["ok"] = res.(*pb.MessageResponse).GetOk()

			logger.Log(log.NewMessage().Level(log.LLTrace).Prefix("gRPC").Sub("logger").Message("[send] unary RPC -- " + method).Metadata(meta).Build())
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
func StreamServerLogging(logger log.Logger, withTimer bool) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var now time.Time
		var after time.Duration

		if withTimer {
			now = time.Now()
		}

		method := info.FullMethod

		logger.Log(log.NewMessage().Level(log.LLTrace).Prefix("gRPC").Sub("logger").Message("[open] stream RPC -- " + method).Build())

		wStream := loggingStream{
			stream:    stream,
			logger:    logger,
			method:    method,
			withTimer: withTimer,
		}

		err := handler(srv, wStream)

		if withTimer {
			after = time.Since(now)
		}

		var meta = log.Field{}
		meta["method"] = method

		if withTimer {
			meta["time"] = after.String()
		}

		if err != nil {
			meta["error"] = err.Error()

			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Message("[conn] stream RPC -- failed to initialize stream with an error").Metadata(meta).Build())
			return err
		}

		logger.Log(log.NewMessage().Level(log.LLTrace).Prefix("gRPC").Sub("logger").Message("[conn] stream RPC ").Metadata(meta).Build())

		return err
	}
}

type loggingStream struct {
	stream    grpc.ServerStream
	logger    log.Logger
	method    string
	withTimer bool
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
	var now time.Time
	var after time.Duration

	if w.withTimer {
		now = time.Now()
	}

	err := w.stream.SendMsg(m)

	if w.withTimer {
		after = time.Since(now)
	}

	var meta = log.Field{}
	var res = log.Field{}
	meta["method"] = w.method

	if w.withTimer {
		meta["time"] = after.String()
	}

	if err != nil {
		meta["error"] = err.Error()

		if m.(*pb.MessageResponse).GetReqID() != "" {
			res["id"] = m.(*pb.MessageResponse).GetReqID()
			meta["response"] = res
		}

		w.logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").
			Message("[send] stream RPC logger -- error sending message").Metadata(meta).Build())
		return err
	}

	res["id"] = m.(*pb.MessageResponse).GetReqID()
	res["ok"] = m.(*pb.MessageResponse).GetOk()
	res["bytes"] = m.(*pb.MessageResponse).GetBytes()

	if !m.(*pb.MessageResponse).GetOk() {
		var err error
		if m.(*pb.MessageResponse).GetErr() != "" {
			err = errors.New(m.(*pb.MessageResponse).GetErr())
		} else {
			err = ErrMessageParse
		}
		res["error"] = err.Error()
		meta["error"] = err.Error()
		meta["response"] = res

		w.logger.Log(
			log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Metadata(meta).
				Message("[send] stream RPC logger -- failed to send response message").Build(),
		)
		return err

	}

	meta["response"] = res

	w.logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("logger").
		Message("[send] stream RPC logger -- sent message to gRPC client").Metadata(meta).Build())
	return err
}

// RecvMsg method is a wrapper for the grpc.ServerStream.RecvMsg(m) method, for which the
// configured logger will register inbound messages or errors
func (w loggingStream) RecvMsg(m interface{}) error {

	var now time.Time
	var after time.Duration

	if w.withTimer {
		now = time.Now()
	}

	err := w.stream.RecvMsg(m)

	if w.withTimer {
		after = time.Since(now)
	}

	var meta = log.Field{}
	meta["method"] = w.method

	if w.withTimer {
		meta["time"] = after.String()
	}

	if err != nil {
		meta["error"] = err.Error()

		// handle EOF
		if errors.Is(err, io.EOF) {
			w.logger.Log(log.NewMessage().Level(log.LLInfo).Prefix("gRPC").Sub("logger").
				Message("[recv] stream RPC logger -- received EOF from client").Metadata(meta).Build())
			return err
		}

		// handle context cancelled
		if errCode := status.Code(err); errCode == codes.DeadlineExceeded {
			w.logger.Log(log.NewMessage().Level(log.LLInfo).Prefix("gRPC").Sub("logger").
				Message("[recv] stream RPC logger -- received context closure from client").Metadata(meta).Build())
			return err
		}

		// default error handling
		w.logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").
			Message("[recv] stream RPC logger -- error receiving message").Metadata(meta).Build())
		return err

	}
	w.logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("logger").Message("[recv] stream RPC logger -- received message from gRPC client").Metadata(meta).Build())
	return err
}
