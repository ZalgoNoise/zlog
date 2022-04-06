package grpclog

import (
	"context"

	"google.golang.org/grpc"

	"github.com/zalgonoise/zlog/log"
	pb "github.com/zalgonoise/zlog/proto/message"
)

// UnaryServerInterceptor returns a new unary server interceptor that adds a gRPC Server Logger
// which captures inbound / outbound interactions with the service
func UnaryServerInterceptor(logger log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger.Log(log.NewMessage().Level(log.LLTrace).Prefix("gRPC").Sub("recv").Message("unary RPC -- " + info.FullMethod).Build())

		res, err := handler(ctx, req)

		if err != nil {
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("Log").Message("unary RPC -- message handling failed with an error").Metadata(log.Field{
				"error":  err.Error(),
				"method": info.FullMethod,
			}).Build())
		} else {
			logger.Log(log.NewMessage().Level(log.LLTrace).Prefix("gRPC").Sub("Log").Message("unary RPC -- " + info.FullMethod).Metadata(log.Field{
				"id": res.(*pb.MessageResponse).GetReqID(),
				"ok": res.(*pb.MessageResponse).GetOk(),
			}).Build())
		}

		return res, err
	}
}

// StreamServerInterceptor returns a new stream server interceptor that adds a gRPC Server Logger
// which captures inbound / outbound interactions with the service
func StreamServerInterceptor(logger log.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		logger.Log(log.NewMessage().Level(log.LLTrace).Prefix("gRPC").Sub("recv").Message("stream RPC -- " + info.FullMethod).Build())

		err := handler(srv, stream)

		if err != nil {
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("LogStream").Message("stream RPC -- failed to initialize stream with an error").Metadata(log.Field{
				"error":  err.Error(),
				"method": info.FullMethod,
			}).Build())
		} else {
			logger.Log(log.NewMessage().Level(log.LLTrace).Prefix("gRPC").Sub("LogStream").Message("stream RPC -- " + info.FullMethod).Build())
		}
		return err
	}
}
