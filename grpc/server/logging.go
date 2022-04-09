package server

import (
	"context"
	"errors"

	"google.golang.org/grpc"

	"github.com/zalgonoise/zlog/log"
	pb "github.com/zalgonoise/zlog/proto/message"
)

// UnaryServerLogging returns a new unary server interceptor that adds a gRPC Server Logger
// which captures inbound / outbound interactions with the service
func UnaryServerLogging(logger log.Logger) grpc.UnaryServerInterceptor {
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

// StreamServerLogging returns a new stream server interceptor that adds a gRPC Server Logger
// which captures inbound / outbound interactions with the service
func StreamServerLogging(logger log.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		logger.Log(log.NewMessage().Level(log.LLTrace).Prefix("gRPC").Sub("logger").Message("[open] stream RPC -- " + info.FullMethod).Build())

		err := handler(srv, stream)

		if err != nil {
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Message("[conn] stream RPC -- failed to initialize stream with an error").Metadata(log.Field{
				"error":  err.Error(),
				"method": info.FullMethod,
			}).Build())
		} else {
			logger.Log(log.NewMessage().Level(log.LLTrace).Prefix("gRPC").Sub("logger").Message("[conn] stream RPC -- " + info.FullMethod).Build())

			go loggingStreamHandler(logger, stream, info.FullMethod)
		}
		return err
	}
}

func loggingStreamHandler(logger log.Logger, stream grpc.ServerStream, method string) {
	var req *pb.MessageRequest
	var res *pb.MessageResponse

	for {
		req = &pb.MessageRequest{}
		err := stream.RecvMsg(req)
		if err != nil {
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Message("[recv] stream RPC logger -- error receiving message").Metadata(log.Field{
				"error":  err.Error(),
				"method": method,
			}).Build())
			continue
		}
		logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("logger").Message("[recv] stream RPC logger -- received message from gRPC client").Build())

		res = &pb.MessageResponse{}
		err = stream.SendMsg(res)

		if err != nil {
			meta := log.Field{
				"error":  err.Error(),
				"method": method,
			}
			if res != nil && res.GetReqID() != "" {
				meta["id"] = res.GetReqID()
			}

			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Message("[send] stream RPC logger -- sending message resulted in an error").Metadata(meta).Build())
			continue
		}

		if !res.GetOk() {
			var err error
			if res.GetErr() != "" {
				err = errors.New(res.GetErr())
			} else {
				err = ErrMessageParse
			}
			logger.Log(
				log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Metadata(log.Field{
					"id":     res.GetReqID(),
					"method": method,
					"response": log.Field{
						"ok":    res.GetOk(),
						"bytes": res.GetBytes(),
						"error": err.Error(),
					},
					"error": err.Error(),
				}).Message("[send] stream RPC logger -- failed to send response message").Build(),
			)
			continue

		}

		logger.Log(
			log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("logger").Metadata(log.Field{
				"id":     res.GetReqID(),
				"method": method,
				"response": log.Field{
					"ok":    res.GetOk(),
					"bytes": res.GetBytes(),
				},
			}).Message("[send] stream RPC logger -- sent server response").Build(),
		)

	}
}
