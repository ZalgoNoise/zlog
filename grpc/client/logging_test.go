package client

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	pb "github.com/zalgonoise/zlog/proto/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type testClientStream struct {
	hMD metadata.MD
	tMD metadata.MD
	err error
	ctx context.Context // for DeepEqual tests
}

func (s *testClientStream) Header() (metadata.MD, error) { return s.hMD, s.err }
func (s *testClientStream) Trailer() metadata.MD         { return s.tMD }
func (s *testClientStream) CloseSend() error             { return s.err }
func (s *testClientStream) Context() context.Context     { return s.ctx }
func (s *testClientStream) SendMsg(m interface{}) error  { return s.err }
func (s *testClientStream) RecvMsg(m interface{}) error  { return s.err }

func TestUnaryClientLogging(t *testing.T) {
	module := "LogClient Interceptors"
	funcname := "UnaryClientLogging()"

	_ = module
	_ = funcname

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

func TestStreamClientLogging(t *testing.T) {
	module := "LogClient Interceptors"
	funcname := "StreamClientLogging()"

	_ = module
	_ = funcname

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
		var stream grpc.ClientStream = &testClientStream{}

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

func FuzzLoggingStreamHeader(f *testing.F) {
	module := "loggingStream"
	funcname := "Header()"

	f.Add("test-meta")
	f.Fuzz(func(t *testing.T, a string) {

		var md = metadata.MD{}

		md["test"] = []string{a}

		logStream := &loggingStream{
			stream: &testClientStream{
				hMD: md,
			},
			logger:    log.New(log.NilConfig),
			method:    "test",
			name:      "test",
			withTimer: false,
		}

		meta, err := logStream.Header()

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if meta["test"][0] != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				meta["test"][0],
			)
			return
		}
	})
}

func FuzzLoggingStreamTrailer(f *testing.F) {
	module := "loggingStream"
	funcname := "Trailer()"

	f.Add("test-meta")
	f.Fuzz(func(t *testing.T, a string) {

		var md = metadata.MD{}

		md["test"] = []string{a}

		logStream := &loggingStream{
			stream: &testClientStream{
				tMD: md,
			},
			logger:    log.New(log.NilConfig),
			method:    "test",
			name:      "test",
			withTimer: false,
		}

		meta := logStream.Trailer()

		if meta["test"][0] != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				meta["test"][0],
			)
			return
		}
	})
}

func TestLoggingStreamCloseSend(t *testing.T) {
	module := "loggingStream"
	funcname := "CloseSend()"

	_ = module
	_ = funcname

	type test struct {
		name string
		err  error
	}

	var tests = []test{
		{
			name: "no error",
		},
		{
			name: "with error",
			err:  errors.New("test error"),
		},
	}

	var verify = func(idx int, test test) {
		logStream := &loggingStream{
			stream: &testClientStream{
				err: test.err,
			},
			logger:    log.New(log.NilConfig),
			method:    "test",
			name:      "test",
			withTimer: false,
		}

		err := logStream.CloseSend()

		if err != nil && !errors.Is(err, test.err) {
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

func TestLoggingStreamContext(t *testing.T) {
	module := "loggingStream"
	funcname := "Context()"

	_ = module
	_ = funcname

	type test struct {
		name string
		ctx  context.Context
	}

	var tests = []test{
		{
			name: "no context",
		},
		{
			name: "with context",
			ctx:  context.Background(),
		},
	}

	var verify = func(idx int, test test) {
		logStream := &loggingStream{
			stream: &testClientStream{
				ctx: test.ctx,
			},
			logger:    log.New(log.NilConfig),
			method:    "test",
			name:      "test",
			withTimer: false,
		}

		ctx := logStream.Context()

		if ctx != nil && !reflect.DeepEqual(ctx, test.ctx) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.ctx,
				ctx,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestLoggingSendMsg(t *testing.T) {
	module := "loggingStream"
	funcname := "SendMsg()"

	_ = module
	_ = funcname

	type test struct {
		name      string
		withTimer bool
		err       error
	}

	var tests = []test{
		{
			name: "no timer",
		},
		{
			name:      "with timer",
			withTimer: true,
		},
		{
			name:      "with error",
			withTimer: true,
			err:       errors.New("test error"),
		},
	}

	var verify = func(idx int, test test) {
		logStream := &loggingStream{
			stream: &testClientStream{
				err: test.err,
			},
			logger:    log.New(log.NilConfig),
			method:    "test",
			name:      "test",
			withTimer: test.withTimer,
		}

		err := logStream.SendMsg(event.New().Message("null").Build())

		if err != nil && !errors.Is(err, test.err) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] error mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.err,
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

func TestLoggingRecvMsg(t *testing.T) {
	module := "loggingStream"
	funcname := "RecvMsg()"

	_ = module
	_ = funcname

	type test struct {
		name      string
		withTimer bool
		err       error
		wants     string
		res       *pb.LogResponse
	}

	var nBytes int32 = 12
	var nErrorStr string = "test error"
	var nError error = errors.New(nErrorStr)

	var tests = []test{
		{
			name: "no timer",
			res: &pb.LogResponse{
				Ok:    true,
				ReqID: "test_id",
				Bytes: &nBytes,
			},
		},
		{
			name:      "with timer",
			withTimer: true,
			res: &pb.LogResponse{
				Ok:    true,
				ReqID: "test_id",
				Bytes: &nBytes,
			},
		},
		{
			name:      "with error from recv",
			withTimer: true,
			err:       nError,
			wants:     nErrorStr,
			res: &pb.LogResponse{
				ReqID: "test_id",
			},
		},
		{
			name:      "with error from message",
			withTimer: true,
			wants:     nErrorStr,
			res: &pb.LogResponse{
				Ok:    false,
				ReqID: "test_id",
				Bytes: &nBytes,
				Err:   &nErrorStr,
			},
		},
		{
			name:      "with error from message",
			withTimer: true,
			wants:     nErrorStr,
			res: &pb.LogResponse{
				Ok:    false,
				ReqID: "test_id",
				Bytes: &nBytes,
				Err:   &nErrorStr,
			},
		},
		{
			name:      "with error from message w/ blank error field",
			withTimer: true,
			wants:     "failed to write log message in remote gRPC server",
			res: &pb.LogResponse{
				Ok:    false,
				ReqID: "test_id",
				Bytes: &nBytes,
			},
		},
	}

	var verify = func(idx int, test test) {
		logStream := &loggingStream{
			stream: &testClientStream{
				err: test.err,
			},
			logger:    log.New(log.NilConfig),
			method:    "test",
			name:      "test",
			withTimer: test.withTimer,
		}

		err := logStream.RecvMsg(test.res)

		if err != nil && err.Error() != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] error mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.err,
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
