package server

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	pb "github.com/zalgonoise/zlog/proto/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	errInvalidRequest error = errors.New("invalid request object")
	errSentHeader     error = errors.New("already sent header")
	testErrUnexpected error = errors.New("unexpected error")
)

type testServerStream struct {
	hMD        metadata.MD
	sentHeader bool
	tMD        metadata.MD
	ctx        context.Context // for DeepEqual tests
}

func (s *testServerStream) SetHeader(m metadata.MD) error {
	s.hMD = m
	return nil
}

func (s *testServerStream) SendHeader(m metadata.MD) error {
	if s.sentHeader {
		return errSentHeader
	}

	s.hMD = m
	s.sentHeader = true

	return nil
}

func (s *testServerStream) SetTrailer(m metadata.MD) {
	s.tMD = m
}

func (s *testServerStream) Context() context.Context { return s.ctx }

func (s *testServerStream) SendMsg(m interface{}) error {
	msg, ok := m.(*pb.LogResponse)

	if !ok {
		return testErrUnexpected
	}

	if msg.GetReqID() == "000" {
		return testErrUnexpected
	}

	return nil
}

func (s *testServerStream) RecvMsg(m interface{}) error {
	msg, ok := m.(*event.Event)

	if !ok {
		return testErrUnexpected
	}

	if msg.GetLevel().String() == "error" {
		return testErrUnexpected
	}

	if msg.GetMsg() == "deadline" {
		return status.Error(codes.DeadlineExceeded, "")
	}

	if msg.GetMsg() == "EOF" {
		return io.EOF
	}

	return nil
}

func TestUnaryServerLogging(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "UnaryServerLogging()"

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
			name:      "unary server logging test",
			ok:        true,
			logger:    log.New(log.NilConfig),
			withTimer: false,
			req:       event.New().Message("null").Build(),
			reply: &pb.LogResponse{
				Ok:    true,
				ReqID: "test",
				Bytes: &wBytes,
			},
		},
		{
			name:      "unary server logging test w/ timer",
			ok:        true,
			logger:    log.New(log.NilConfig),
			withTimer: true,
			req:       event.New().Message("null").Build(),
			reply: &pb.LogResponse{
				Ok:    true,
				ReqID: "test",
				Bytes: &wBytes,
			},
		},
		{
			name:      "failing unary server logging test",
			ok:        false,
			logger:    log.New(log.NilConfig),
			withTimer: false,
			req:       event.New().Message("fail").Build(),
			reply: &pb.LogResponse{
				Ok:    false,
				Err:   &errStr,
				ReqID: "test",
			},
		},
	}

	var verify = func(idx int, test test) {
		fn := UnaryServerLogging(test.logger, test.withTimer)

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

func TestStreamServerLogging(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "StreamServerLogging()"

	_ = module
	_ = funcname

	type test struct {
		name      string
		ok        bool
		logger    log.Logger
		withTimer bool
		req       string
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
			name:      "unary server logging test",
			ok:        true,
			logger:    log.New(log.NilConfig),
			withTimer: false,
			req:       "ok",
		},
		{
			name:      "unary server logging test w/ timer",
			ok:        true,
			logger:    log.New(log.NilConfig),
			withTimer: true,
			req:       "ok",
		},
		{
			name:      "failing unary server logging test",
			ok:        false,
			logger:    log.New(log.NilConfig),
			withTimer: false,
			req:       "fail",
		},
	}

	var verify = func(idx int, test test) {
		fn := StreamServerLogging(test.logger, test.withTimer)

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

func TestLoggingSendMsg(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "loggingStream.SendMsg()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *loggingStream
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
			t: &loggingStream{
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
			t: &loggingStream{
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
		{
			name: "not-OK test",
			t: &loggingStream{
				stream: &testServerStream{},
				logger: log.New(log.NilConfig),
				method: "testLog",
			},
			m: &pb.LogResponse{
				Ok:    false,
				ReqID: "123",
				Bytes: &bytesResponse[1],
				Err:   &errResponse[1],
			},
		},
		{
			name: "not-OK test, no error in response",
			t: &loggingStream{
				stream: &testServerStream{},
				logger: log.New(log.NilConfig),
				method: "testLog",
			},
			m: &pb.LogResponse{
				Ok:    false,
				ReqID: "123",
				Bytes: &bytesResponse[1],
				Err:   &errResponse[0],
			},
		},
		{
			name: "working test w/ timer",
			t: &loggingStream{
				stream:    &testServerStream{},
				logger:    log.New(log.NilConfig),
				method:    "testLog",
				withTimer: true,
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
			name: "errored test w/ timer",
			t: &loggingStream{
				stream:    &testServerStream{},
				logger:    log.New(log.NilConfig),
				method:    "testLog",
				withTimer: true,
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

func TestLoggingRecvMsg(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "loggingStream.RecvMsg()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *loggingStream
		m    *event.Event
		ok   bool
	}

	var tests = []test{
		{
			name: "working test",
			t: &loggingStream{
				stream: &testServerStream{},
				logger: log.New(log.NilConfig),
				method: "testLog",
			},
			m:  event.New().Message("null").Build(),
			ok: true,
		},
		{
			name: "errored test",
			t: &loggingStream{
				stream: &testServerStream{},
				logger: log.New(log.NilConfig),
				method: "testLog",
			},
			m: event.New().Level(event.Level_error).Message("null").Build(),
		},
		{
			name: "errored with EOF",
			t: &loggingStream{
				stream: &testServerStream{},
				logger: log.New(log.NilConfig),
				method: "testLog",
			},
			m: event.New().Message("EOF").Build(),
		},
		{
			name: "errored with deadline exceeded",
			t: &loggingStream{
				stream: &testServerStream{},
				logger: log.New(log.NilConfig),
				method: "testLog",
			},
			m: event.New().Message("deadline").Build(),
		},
		{
			name: "working test w/ timer",
			t: &loggingStream{
				stream:    &testServerStream{},
				logger:    log.New(log.NilConfig),
				method:    "testLog",
				withTimer: true,
			},
			m:  event.New().Message("null").Build(),
			ok: true,
		},
		{
			name: "errored test w/ timer",
			t: &loggingStream{
				stream:    &testServerStream{},
				logger:    log.New(log.NilConfig),
				method:    "testLog",
				withTimer: true,
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

func TestLoggingSetHeader(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "loggingStream.SetHeader()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *loggingStream
		m    metadata.MD
	}

	var tests = []test{
		{
			name: "working test",
			t: &loggingStream{
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

func TestLoggingSendHeader(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "loggingStream.SendHeader()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *loggingStream
		m    metadata.MD
	}

	var tests = []test{
		{
			name: "working test",
			t: &loggingStream{
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
			t: &loggingStream{
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

func TestLoggingSetTrailer(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "loggingStream.SendHeader()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *loggingStream
		m    metadata.MD
	}

	var tests = []test{
		{
			name: "working test",
			t: &loggingStream{
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

func TestLoggingContext(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "loggingStream.Context()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *loggingStream
	}

	rootCtx := context.Background()

	var tests = []test{
		{
			name: "working test",
			t: &loggingStream{
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
