package server

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

func TestUnaryServerTiming(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "UnaryServerTiming()"

	_ = module
	_ = funcname

	type test struct {
		name   string
		ok     bool
		logger log.Logger
		req    *event.Event
		reply  *pb.LogResponse
	}

	var sInfo = &grpc.UnaryServerInfo{
		FullMethod: "test",
	}

	var errStr = "failed to write message"
	var wBytes int32 = 100

	var handler grpc.UnaryHandler = func(ctx context.Context, req interface{}) (interface{}, error) {
		r, ok := req.(*event.Event)

		if !ok {
			return nil, errInvalidRequest
		}

		if r.GetMsg() == "fail" {
			res := &pb.LogResponse{
				Ok:    false,
				Err:   &errStr,
				ReqID: "test",
			}

			return res, errInvalidRequest
		}

		res := &pb.LogResponse{
			Ok:    true,
			ReqID: "test",
			Bytes: &wBytes,
		}

		return res, nil
	}

	var tests = []test{
		{
			name:   "unary server timing test",
			ok:     true,
			logger: log.New(log.NilConfig),
			req:    event.New().Message("null").Build(),
			reply: &pb.LogResponse{
				Ok:    true,
				ReqID: "test",
				Bytes: &wBytes,
			},
		},
		{
			name:   "failing unary server timing test",
			ok:     false,
			logger: log.New(log.NilConfig),
			req:    event.New().Message("fail").Build(),
			reply: &pb.LogResponse{
				Ok:    false,
				Err:   &errStr,
				ReqID: "test",
			},
		},
	}

	var verify = func(idx int, test test) {
		fn := UnaryServerTiming(test.logger)

		res, err := fn(
			context.Background(),
			test.req,
			sInfo,
			handler,
		)

		if err != nil {
			if test.ok {
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

			if !errors.Is(err, errInvalidRequest) {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unexpected error: wanted %v ; got %v -- action: %s",
					idx,
					module,
					funcname,
					errInvalidRequest,
					err,
					test.name,
				)
				return
			}

			return
		}

		r, ok := res.(*pb.LogResponse)

		if !ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] response is not of type *pb.LogResponse -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		if !reflect.DeepEqual(r, test.reply) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.reply,
				r,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestStreamServerTiming(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "StreamServerTiming()"

	_ = module
	_ = funcname

	type test struct {
		name   string
		ok     bool
		logger log.Logger
		req    string
	}

	var sInfo = &grpc.StreamServerInfo{
		FullMethod:     "test",
		IsClientStream: true,
		IsServerStream: true,
	}

	var handler grpc.StreamHandler = func(srv interface{}, stream grpc.ServerStream) error {
		r, ok := srv.(string)

		if !ok {
			return errInvalidRequest
		}

		if r == "fail" {
			return errInvalidRequest
		}

		return nil
	}

	var tests = []test{
		{
			name:   "unary server timing test",
			ok:     true,
			logger: log.New(log.NilConfig),
			req:    "ok",
		},

		{
			name:   "failing unary timing logging test",
			ok:     false,
			logger: log.New(log.NilConfig),
			req:    "fail",
		},
	}

	var verify = func(idx int, test test) {
		fn := StreamServerTiming(test.logger)

		err := fn(
			test.req,
			&testServerStream{
				ctx: context.Background(),
			},
			sInfo,
			handler,
		)

		if err != nil {
			if test.ok {
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

			if !errors.Is(err, errInvalidRequest) {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unexpected error: wanted %v ; got %v -- action: %s",
					idx,
					module,
					funcname,
					errInvalidRequest,
					err,
					test.name,
				)
				return
			}

			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestTimingSendMsg(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "timingStream.SendMsg()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *timingStream
		m    *pb.LogResponse
		ok   bool
	}

	var bytesResponse = []int32{
		203,
		1008,
	}
	var errResponse = []string{
		"",
		testErrUnexpected.Error(),
	}

	var tests = []test{
		{
			name: "working test",
			t: &timingStream{
				stream: &testServerStream{},
				logger: log.New(log.NilConfig),
				method: "testLog",
			},
			m: &pb.LogResponse{
				Ok:    true,
				ReqID: "123",
				Bytes: &bytesResponse[0],
				Err:   &errResponse[0],
			},
			ok: true,
		},
		{
			name: "errored test",
			t: &timingStream{
				stream: &testServerStream{},
				logger: log.New(log.NilConfig),
				method: "testLog",
			},
			m: &pb.LogResponse{
				Ok:    false,
				ReqID: "000",
				Bytes: &bytesResponse[1],
				Err:   &errResponse[1],
			},
		},
	}

	var verify = func(idx int, test test) {
		err := test.t.SendMsg(test.m)

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

func TestTimingRecvMsg(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "timingStream.RecvMsg()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *timingStream
		m    *event.Event
		ok   bool
	}

	var tests = []test{
		{
			name: "working test",
			t: &timingStream{
				stream: &testServerStream{},
				logger: log.New(log.NilConfig),
				method: "testLog",
			},
			m:  event.New().Message("null").Build(),
			ok: true,
		},
		{
			name: "errored test",
			t: &timingStream{
				stream: &testServerStream{},
				logger: log.New(log.NilConfig),
				method: "testLog",
			},
			m: event.New().Level(event.Level_error).Message("null").Build(),
		},
	}

	var verify = func(idx int, test test) {
		err := test.t.RecvMsg(test.m)

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

func TestTimingSetHeader(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "timingStream.SetHeader()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *timingStream
		m    metadata.MD
	}

	var tests = []test{
		{
			name: "working test",
			t: &timingStream{
				stream: &testServerStream{},
				logger: log.New(log.NilConfig),
				method: "testLog",
			},
			m: metadata.New(map[string]string{
				"test": "header",
			}),
		},
	}

	var verify = func(idx int, test test) {
		err := test.t.SetHeader(test.m)

		if err != nil {
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

		if !reflect.DeepEqual(test.t.stream.(*testServerStream).hMD, test.m) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.t.stream.(*testServerStream).hMD,
				test.m,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestTimingSendHeader(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "timingStream.SendHeader()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *timingStream
		m    metadata.MD
	}

	var tests = []test{
		{
			name: "working test",
			t: &timingStream{
				stream: &testServerStream{},
				logger: log.New(log.NilConfig),
				method: "testLog",
			},
			m: metadata.New(map[string]string{
				"test": "header",
			}),
		},
		{
			name: "error test",
			t: &timingStream{
				stream: &testServerStream{
					sentHeader: true,
				},
				logger: log.New(log.NilConfig),
				method: "testLog",
			},
			m: metadata.New(map[string]string{
				"test": "header",
			}),
		},
	}

	var verify = func(idx int, test test) {
		err := test.t.SendHeader(test.m)

		if err != nil && !errors.Is(err, errSentHeader) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] error mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				errSentHeader,
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

func TestTimingSetTrailer(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "timingStream.SendHeader()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *timingStream
		m    metadata.MD
	}

	var tests = []test{
		{
			name: "working test",
			t: &timingStream{
				stream: &testServerStream{},
				logger: log.New(log.NilConfig),
				method: "testLog",
			},
			m: metadata.New(map[string]string{
				"test": "trailer",
			}),
		},
	}

	var verify = func(idx int, test test) {
		test.t.SetTrailer(test.m)

		if !reflect.DeepEqual(test.t.stream.(*testServerStream).tMD, test.m) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.t.stream.(*testServerStream).tMD,
				test.m,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestTimingContext(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "timingStream.Context()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *timingStream
	}

	rootCtx := context.Background()

	var tests = []test{
		{
			name: "working test",
			t: &timingStream{
				stream: &testServerStream{
					ctx: rootCtx,
				},
				logger: log.New(log.NilConfig),
				method: "testLog",
			},
		},
	}

	var verify = func(idx int, test test) {
		ctx := test.t.Context()

		if !reflect.DeepEqual(ctx, rootCtx) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				rootCtx,
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
