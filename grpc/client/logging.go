package client

import (
	"context"
	"errors"

	"github.com/zalgonoise/zlog/log"
	pb "github.com/zalgonoise/zlog/proto/message"
	"google.golang.org/grpc"
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

		go loggingStreamHandler(logger, clientStream, method, desc)

		return clientStream, err
	}
}

func loggingStreamHandler(logger log.Logger, clientStream grpc.ClientStream, method string, desc *grpc.StreamDesc) {
	var req *pb.MessageRequest
	var res *pb.MessageResponse

	for {
		req = &pb.MessageRequest{}
		err := clientStream.SendMsg(req)

		if err != nil {
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Message("[send] stream RPC logger -- sending message resulted in an error").Metadata(log.Field{
				"error":  err.Error(),
				"method": method,
				"stream": desc.StreamName,
			}).Build())
			continue
		}
		logger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("logger").Message("[send] stream RPC logger -- sent log message to gRPC server").Build())

		res = &pb.MessageResponse{}
		err = clientStream.RecvMsg(res)
		if err != nil {
			logger.Log(log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Message("[send] stream RPC logger -- receiving message resulted in an error").Metadata(log.Field{
				"error":  err.Error(),
				"method": method,
				"stream": desc.StreamName,
			}).Build())
		}

		// check server response for errors
		if err != nil {
			meta := log.Field{
				"error":  err.Error(),
				"method": method,
				"stream": desc.StreamName,
			}

			if res != nil && res.GetReqID() != "" {
				meta["id"] = res.GetReqID()
			}

			logger.Log(
				log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Metadata(meta).Message("[recv] stream RPC logger -- issue receiving message from stream").Build(),
			)
			continue

		}

		// there are no errors in the response; check the response's OK value
		// if not OK, register this as a local bad response error and continue
		if !res.GetOk() {
			var err error
			if res.GetErr() != "" {
				err = errors.New(res.GetErr())
			} else {
				err = ErrBadResponse
			}
			logger.Log(
				log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("logger").Metadata(log.Field{
					"id": res.GetReqID(),
					"response": log.Field{
						"ok":    res.GetOk(),
						"bytes": res.GetBytes(),
						"error": err.Error(),
					},
					"error": err.Error(),
				}).Message("[recv] stream RPC logger -- failed to write log message").Build(),
			)
			continue
		}

		// server response is OK, register this event
		logger.Log(
			log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("logger").Metadata(log.Field{
				"id": res.GetReqID(),
				"response": log.Field{
					"ok":    res.GetOk(),
					"bytes": res.GetBytes(),
				},
			}).Message("[recv] stream RPC logger -- registering server response").Build(),
		)

	}
}
