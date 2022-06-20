package client

import (
	"context"
	"errors"
	"testing"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	pb "github.com/zalgonoise/zlog/proto/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

type testServerStream struct{}

func (s *testServerStream) Header() (metadata.MD, error) { return nil, nil }
func (s *testServerStream) Trailer() metadata.MD         { return nil }
func (s *testServerStream) CloseSend() error             { return nil }
func (s *testServerStream) Context() context.Context     { return context.Background() }
func (s *testServerStream) SendMsg(m interface{}) error  { return nil }
func (s *testServerStream) RecvMsg(m interface{}) error  { return nil }

func TestStreamClientLogging(t *testing.T) {
	module := "LogClient Interceptors"
	funcname := "StreamClientLogging()"

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
	}

	l := log.New(log.NilConfig)

	var streamHandler = func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	var streamDesc = &grpc.StreamDesc{
		StreamName:    "test stream",
		Handler:       streamHandler,
		ServerStreams: true,
		ClientStreams: true,
	}

	var streamer grpc.Streamer = func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		var stream grpc.ClientStream = &testServerStream{}

		var err error

		if method == "fail" {
			err = errors.New("errored")
		}

		return stream, err
	}

	var tests = []test{
		{
			name:      "unary client logging test",
			ok:        true,
			logger:    l,
			withTimer: false,
		},
		{
			name:      "unary client logging test",
			ok:        true,
			logger:    l,
			withTimer: true,
		},
		{
			name:      "unary client logging test",
			ok:        false,
			logger:    l,
			withTimer: false,
		},
	}

	var verify = func(idx int, test test) {
		fn := StreamClientLogging(test.logger, test.withTimer)

		var err error

		if test.ok {
			_, err = fn(
				context.Background(),
				streamDesc,
				new(grpc.ClientConn),
				"test",
				streamer,
			)

		} else {
			_, err = fn(
				context.Background(),
				streamDesc,
				new(grpc.ClientConn),
				"fail",
				streamer,
			)
		}

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
