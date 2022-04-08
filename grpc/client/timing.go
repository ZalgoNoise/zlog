package client

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc"

	"github.com/zalgonoise/zlog/log"
	pb "github.com/zalgonoise/zlog/proto/message"
)

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
