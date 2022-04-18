package client

import (
	"context"
	"errors"
	"time"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	pb "github.com/zalgonoise/zlog/proto/message"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryClientLogging returns a new unary client interceptor that adds a gRPC Client Logger
// which captures inbound / outbound interactions with the service
func UnaryClientLogging(logger log.Logger, withTimer bool) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		var now time.Time
		var after time.Duration

		if withTimer {
			now = time.Now()
		}

		logger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("logger").Message("[send] unary RPC logger -- " + method).Build())

		err := invoker(ctx, method, req, reply, cc, opts...)

		if withTimer {
			after = time.Since(now)
		}

		// validate response fields
		var res = event.Field{}
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
		var meta = event.Field{}

		meta["method"] = method
		meta["response"] = res

		if withTimer {
			meta["time"] = after.String()
		}

		if err != nil {
			// handle errors in the transaction
			meta["error"] = err.Error()
			logger.Log(event.New().Level(event.Level_warn).Prefix("gRPC").Sub("logger").Message("[recv] unary RPC logger -- message handling failed with an error").Metadata(meta).Build())
		} else if !reply.(*pb.MessageResponse).GetOk() {
			// handle errors in the response; return the error in the message
			meta["error"] = reply.(*pb.MessageResponse).GetErr()
			logger.Log(event.New().Level(event.Level_warn).Prefix("gRPC").Sub("logger").Message("[recv] unary RPC logger -- message returned a not-OK status").Metadata(meta).Build())

			return errors.New(reply.(*pb.MessageResponse).GetErr())

		} else {
			// log an OK transaction
			logger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("logger").Message("[recv] unary RPC logger").Metadata(meta).Build())
		}
		return err
	}
}

// StreamClientLogging returns a new stream client interceptor that adds a gRPC Client Logger
// which captures inbound / outbound interactions with the service
func StreamClientLogging(logger log.Logger, withTimer bool) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		var now time.Time
		var after time.Duration

		if withTimer {
			now = time.Now()
		}

		logger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("logger").Message("[open] stream RPC logger connection open -- " + method).Build())

		clientStream, err := streamer(ctx, desc, cc, method, opts...)

		if withTimer {
			after = time.Since(now)
		}

		var meta = event.Field{}

		meta["method"] = method
		meta["stream"] = desc.StreamName

		if withTimer {
			meta["time"] = after.String()
		}

		if err != nil {
			meta["error"] = err.Error()
			logger.Log(event.New().Level(event.Level_warn).Prefix("gRPC").Sub("logger").Message("[conn] stream RPC logger -- failed to initialize stream with an error").Metadata(meta).Build())
		}

		logger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("logger").Message("[conn] stream RPC logger -- connection was established").Metadata(meta).Build())

		wStream := loggingStream{
			stream:    clientStream,
			logger:    logger,
			method:    method,
			name:      desc.StreamName,
			withTimer: withTimer,
		}

		return wStream, err
	}
}

type loggingStream struct {
	stream    grpc.ClientStream
	logger    log.Logger
	method    string
	name      string
	withTimer bool
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
	var now time.Time
	var after time.Duration

	if w.withTimer {
		now = time.Now()
	}

	err := w.stream.SendMsg(m)

	if w.withTimer {
		after = time.Since(now)
	}

	var meta = event.Field{}

	meta["method"] = w.method
	meta["stream"] = w.name

	if w.withTimer {
		meta["time"] = after.String()
	}

	if err != nil {
		meta["error"] = err.Error()
		w.logger.Log(event.New().Level(event.Level_warn).Prefix("gRPC").Sub("logger").Message("[send] stream RPC logger -- sending message resulted in an error").Metadata(meta).Build())
		return err
	}
	w.logger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("logger").Message("[send] stream RPC logger -- sent log message to gRPC server").Metadata(meta).Build())

	return err
}

// RecvMsg method is a wrapper for the grpc.ClientStream.RecvMsg(m) method, for which the
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

	var meta = event.Field{}
	var res = event.Field{}
	meta["method"] = w.method
	meta["stream"] = w.name

	if w.withTimer {
		meta["time"] = after.String()
	}

	// check server response for errors
	if err != nil {
		meta["error"] = err.Error()

		if m.(*pb.MessageResponse) != nil && m.(*pb.MessageResponse).GetReqID() != "" {
			res["id"] = m.(*pb.MessageResponse).GetReqID()
			meta["response"] = res
		}

		w.logger.Log(event.New().Level(event.Level_warn).Prefix("gRPC").Sub("logger").Metadata(meta).Message("[recv] stream RPC logger -- issue receiving message from stream").Build())
		return err
	}

	res["id"] = m.(*pb.MessageResponse).GetReqID()
	res["ok"] = m.(*pb.MessageResponse).GetOk()
	res["bytes"] = m.(*pb.MessageResponse).GetBytes()

	// there are no errors in the response; check the response's OK value
	// if not OK, register this as a local bad response error and continue
	if !m.(*pb.MessageResponse).GetOk() {
		var err error
		if m.(*pb.MessageResponse).GetErr() != "" {
			err = errors.New(m.(*pb.MessageResponse).GetErr())
		} else {
			err = ErrBadResponse
		}
		res["error"] = err.Error()
		meta["error"] = err.Error()
		meta["response"] = res

		w.logger.Log(event.New().Level(event.Level_warn).Prefix("gRPC").Sub("logger").Metadata(meta).
			Message("[recv] stream RPC logger -- failed to write log message").Build())
		return err
	}

	// server response is OK, register this event
	meta["response"] = res

	w.logger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("logger").Metadata(meta).
		Message("[recv] stream RPC logger -- registering server response").Build())

	return err

}
