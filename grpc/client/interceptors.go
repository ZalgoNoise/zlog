package client

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc"

	"github.com/zalgonoise/zlog/log"
	pb "github.com/zalgonoise/zlog/proto/message"
)

// UnaryServerInterceptor returns a new unary server interceptor that adds a gRPC Client Logger
// which captures inbound / outbound interactions with the service
func UnaryClientLogging(logger log.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// TODO: add an interceptor to do measure time
		now := time.Now()

		logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("send").Message("unary RPC -- " + method).Build())

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
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("recv").Message("unary RPC -- message handling failed with an error").Metadata(log.Field{
				"error":    err.Error(),
				"method":   method,
				"response": res,
				"time":     time.Since(now).String(),
			}).Build())
		} else if !reply.(*pb.MessageResponse).GetOk() {
			// handle errors in the response; return the error in the message
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("recv").Message("unary RPC -- message returned a not-OK status").Metadata(log.Field{
				"error":    reply.(*pb.MessageResponse).GetErr(),
				"method":   method,
				"response": res,
				"time":     time.Since(now).String(),
			}).Build())

			return errors.New(reply.(*pb.MessageResponse).GetErr())

		} else {
			// log an OK transaction
			logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("recv").Message("unary RPC").Metadata(log.Field{
				"method":   method,
				"response": res,
				"time":     time.Since(now).String(),
			}).Build())
		}
		return err
	}
}

// StreamServerInterceptor returns a new stream server interceptor that adds a gRPC Client Logger
// which captures inbound / outbound interactions with the service
func StreamClientLogging(logger log.Logger) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// TODO: add an interceptor to do measure time
		now := time.Now()

		logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("open").Message("stream RPC -- " + method).Build())

		clientStream, err := streamer(ctx, desc, cc, method, opts...)

		if err != nil {
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("LogStream").Message("stream RPC -- failed to initialize stream with an error").Metadata(log.Field{
				"error":    err.Error(),
				"method":   method,
				"stream":   desc.StreamName,
				"duration": time.Since(now).String(),
			}).Build())

		}

		logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("closed").Message("stream RPC was closed").Metadata(log.Field{
			"duration": time.Since(now).String(),
		}).Build())

		return clientStream, err
	}
}
