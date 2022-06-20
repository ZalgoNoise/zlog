package client

import (
	"context"
	"testing"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	pb "github.com/zalgonoise/zlog/proto/service"
	"google.golang.org/grpc"
)

func TestUnaryClientLogging(t *testing.T) {
	module := "LogClient Interceptors"
	funcname := "UnaryClientLogging()"

	_ = module
	_ = funcname

	type testGRPCLogger struct {
		l GRPCLogger
		e chan error
	}

	type test struct {
		name      string
		ok        bool
		logger    log.Logger
		withTimer bool
		req       *event.Event
		reply     *pb.LogResponse
	}

	var invoker grpc.UnaryInvoker = func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return nil
	}

	l := log.New(log.NilConfig)
	var bVal int32 = 20
	var eVal string = "some error"

	var tests = []test{
		{
			name:      "unary client logging test",
			ok:        true,
			logger:    l,
			withTimer: false,
			req:       event.New().Message("null").Build(),
			reply: &pb.LogResponse{
				Ok:    true,
				ReqID: "something",
				Bytes: &bVal,
			},
		},
		{
			name:      "unary client logging test w/ timer",
			ok:        true,
			logger:    l,
			withTimer: true,
			req:       event.New().Message("null").Build(),
			reply: &pb.LogResponse{
				Ok:    true,
				ReqID: "something",
				Bytes: &bVal,
			},
		},
		{
			name:      "unary client logging test w/ err response",
			logger:    l,
			withTimer: false,
			req:       event.New().Message("null").Build(),
			reply: &pb.LogResponse{
				Ok:    false,
				Err:   &eVal,
				ReqID: "something",
				Bytes: &bVal,
			},
		},
		{
			name:      "unary client logging test w/ err response but no err message",
			logger:    l,
			withTimer: false,
			req:       event.New().Message("null").Build(),
			reply: &pb.LogResponse{
				Ok:    false,
				ReqID: "something",
				Bytes: &bVal,
			},
		},
	}

	var verify = func(idx int, test test) {
		fn := UnaryClientLogging(test.logger, test.withTimer)

		err := fn(
			context.Background(),
			"test",
			test.req,
			test.reply,
			new(grpc.ClientConn),
			invoker,
		)

		if err != nil && test.ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
