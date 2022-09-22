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

func TestUnaryClientTiming(t *testing.T) {
	module := "LogClient Interceptors"
	funcname := "UnaryClientTiming()"

	_ = module
	_ = funcname

	type test struct {
		name    string
		ok      bool
		logger  log.Logger
		req     *event.Event
		reply   *pb.LogResponse
		invoker grpc.UnaryInvoker
	}

	var invoker grpc.UnaryInvoker = func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return nil
	}

	l := log.New(log.NilConfig)
	var bVal int32 = 20
	var eVal string = "some error"

	var invokerErr grpc.UnaryInvoker = func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return errors.New(eVal)
	}

	var tests = []test{
		{
			name:   "unary client timing test",
			ok:     true,
			logger: l,
			req:    event.New().Message("null").Build(),
			reply: &pb.LogResponse{
				Ok:    true,
				ReqID: "something",
				Bytes: &bVal,
			},
			invoker: invoker,
		},
		{
			name:   "unary client timing test w/ err response",
			logger: l,
			req:    event.New().Message("null").Build(),
			reply: &pb.LogResponse{
				Ok:    false,
				Err:   &eVal,
				ReqID: "something",
				Bytes: &bVal,
			},
			invoker: invoker,
		},
		{
			name:   "unary client timing test w/ err response but no err message",
			logger: l,
			req:    event.New().Message("null").Build(),
			reply: &pb.LogResponse{
				Ok:    false,
				ReqID: "something",
				Bytes: &bVal,
			},
			invoker: invoker,
		},
		{
			name:   "unary client timing test w/ invoker error",
			ok:     false,
			logger: l,
			req:    event.New().Message("null").Build(),
			reply: &pb.LogResponse{
				Ok:    true,
				ReqID: "something",
				Bytes: &bVal,
			},
			invoker: invokerErr,
		},
	}

	var verify = func(idx int, test test) {
		fn := UnaryClientTiming(test.logger)

		err := fn(
			context.Background(),
			"test",
			test.req,
			test.reply,
			new(grpc.ClientConn),
			test.invoker,
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

func TestStreamClientTiming(t *testing.T) {
	module := "LogClient Interceptors"
	funcname := "StreamClientTiming()"

	_ = module
	_ = funcname

	type test struct {
		name   string
		ok     bool
		logger log.Logger
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
			name:   "unary client logging test",
			ok:     true,
			logger: l,
		},
		{
			name:   "unary client logging test",
			ok:     true,
			logger: l,
		},
		{
			name:   "unary client logging test",
			ok:     false,
			logger: l,
		},
	}

	var verify = func(idx int, test test) {
		fn := StreamClientTiming(test.logger)

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

func FuzzTimingStreamHeader(f *testing.F) {
	module := "timingStream"
	funcname := "Header()"

	f.Add("test-meta")
	f.Fuzz(func(t *testing.T, a string) {

		var md = metadata.MD{}

		md["test"] = []string{a}

		logStream := &timingStream{
			stream: &testClientStream{
				hMD: md,
			},
			logger: log.New(log.NilConfig),
			method: "test",
			name:   "test",
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

func FuzzTimingStreamTrailer(f *testing.F) {
	module := "timingStream"
	funcname := "Trailer()"

	f.Add("test-meta")
	f.Fuzz(func(t *testing.T, a string) {

		var md = metadata.MD{}

		md["test"] = []string{a}

		logStream := &timingStream{
			stream: &testClientStream{
				tMD: md,
			},
			logger: log.New(log.NilConfig),
			method: "test",
			name:   "test",
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

func TestTimingStreamCloseSend(t *testing.T) {
	module := "timingStream"
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
		logStream := &timingStream{
			stream: &testClientStream{
				err: test.err,
			},
			logger: log.New(log.NilConfig),
			method: "test",
			name:   "test",
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

func TestTimingStreamContext(t *testing.T) {
	module := "timingStream"
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
		logStream := &timingStream{
			stream: &testClientStream{
				ctx: test.ctx,
			},
			logger: log.New(log.NilConfig),
			method: "test",
			name:   "test",
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

func TestTimingSendMsg(t *testing.T) {
	module := "timingStream"
	funcname := "SendMsg()"

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
		logStream := &timingStream{
			stream: &testClientStream{
				err: test.err,
			},
			logger: log.New(log.NilConfig),
			method: "test",
			name:   "test",
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

func TestTimingRecvMsg(t *testing.T) {
	module := "timingStream"
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
		logStream := &timingStream{
			stream: &testClientStream{
				err: test.err,
			},
			logger: log.New(log.NilConfig),
			method: "test",
			name:   "test",
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
