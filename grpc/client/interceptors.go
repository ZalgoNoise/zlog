package client

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc"

	"github.com/zalgonoise/zlog/log"
	pb "github.com/zalgonoise/zlog/proto/message"
)

// UnaryClientLogging returns a new unary client interceptor that adds a gRPC Client Logger
// which captures inbound / outbound interactions with the service
func UnaryClientLogging(logger log.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("send").Message("unary RPC logger -- " + method).Build())

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
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("recv").Message("unary RPC logger -- message handling failed with an error").Metadata(log.Field{
				"error":    err.Error(),
				"method":   method,
				"response": res,
			}).Build())
		} else if !reply.(*pb.MessageResponse).GetOk() {
			// handle errors in the response; return the error in the message
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("recv").Message("unary RPC logger -- message returned a not-OK status").Metadata(log.Field{
				"error":    reply.(*pb.MessageResponse).GetErr(),
				"method":   method,
				"response": res,
			}).Build())

			return errors.New(reply.(*pb.MessageResponse).GetErr())

		} else {
			// log an OK transaction
			logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("recv").Message("unary RPC logger").Metadata(log.Field{
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

		logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("open").Message("stream RPC logger -- " + method).Build())

		clientStream, err := streamer(ctx, desc, cc, method, opts...)

		if err != nil {
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("LogStream").Message("stream RPC logger -- failed to initialize stream with an error").Metadata(log.Field{
				"error":  err.Error(),
				"method": method,
				"stream": desc.StreamName,
			}).Build())

		}

		logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("conn").Message("stream RPC logger -- connection was established").Build())

		return clientStream, err
	}
}

// UnaryClientTiming returns a new unary client interceptor that shows the time taken to complete RPCs
func UnaryClientTiming(logger log.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		now := time.Now()

		err := invoker(ctx, method, req, reply, cc, opts...)

		if err != nil {
			// handle errors in the transaction
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("recv").Message("unary RPC timer -- message handling failed with an error").Metadata(log.Field{
				"error":  err.Error(),
				"method": method,
				"time":   time.Since(now).String(),
			}).Build())
		} else if !reply.(*pb.MessageResponse).GetOk() {
			// handle errors in the response; return the error in the message
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("recv").Message("unary RPC timer -- message returned a not-OK status").Metadata(log.Field{
				"error":    reply.(*pb.MessageResponse).GetErr(),
				"method":   method,
				"response": log.Field{"id": reply.(*pb.MessageResponse).GetReqID()},
				"time":     time.Since(now).String(),
			}).Build())

			return errors.New(reply.(*pb.MessageResponse).GetErr())

		} else {
			// log an OK transaction
			logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("recv").Message("unary RPC timer").Metadata(log.Field{
				"method":   method,
				"response": log.Field{"id": reply.(*pb.MessageResponse).GetReqID()},
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
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("LogStream").Message("stream RPC timer -- failed to initialize stream with an error").Metadata(log.Field{
				"error":  err.Error(),
				"method": method,
				"stream": desc.StreamName,
				"time":   time.Since(now).String(),
			}).Build())

		}

		logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("conn").Message("stream RPC timer -- connection was established").Metadata(log.Field{
			"time": time.Since(now).String(),
		}).Build())

		return clientStream, err
	}
}
