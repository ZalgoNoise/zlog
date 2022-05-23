package server

import (
	"bytes"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/zalgonoise/zlog/grpc/client"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	pb "github.com/zalgonoise/zlog/proto/service"
	"google.golang.org/grpc/metadata"
)

func TestUnaryServerLogging(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "UnaryServerLogging()"

	_ = module
	_ = funcname

	type test struct {
		name     string
		s        *GRPCLogServer
		matchers []string
	}

	var buf = []*bytes.Buffer{{}, {}, {}}

	var tests = []test{
		{
			name: "unary server logging test",
			s: New(
				WithLogger(log.New(log.WithOut(buf[0]), log.SkipExit)),
				WithServiceLoggerV(log.New(log.WithOut(buf[1]), log.CfgTextLevelFirst, log.SkipExit)),
				WithAddr("127.0.0.1:9099"),
				WithGRPCOpts(),
			),
			matchers: []string{
				`^\[trace\]\s+\[.*\]\s+\[gRPC\]\s+\[logger\]\s+\[recv\]\s+unary RPC -- \/logservice.LogService\/Log$`,
				`^\[trace\]\s+\[.*\]\s+\[gRPC\]\s+\[logger\]\s+\[send\]\s+unary RPC -- \/logservice.LogService\/Log.*`,
			},
		},
		{
			name: "unary server logging w/ timer",
			s: New(
				WithLogger(log.New(log.WithOut(buf[0]), log.SkipExit)),
				WithServiceLoggerV(log.New(log.WithOut(buf[1]), log.CfgTextLevelFirst, log.SkipExit)),
				WithTiming(),
				WithAddr("127.0.0.1:9099"),
				WithGRPCOpts(),
			),
			matchers: []string{
				`^\[trace\]\s+\[.*\]\s+\[gRPC\]\s+\[logger\]\s+\[recv\]\s+unary RPC -- \/logservice.LogService\/Log.*`,
				`^\[trace\]\s+\[.*\]\s+\[gRPC\]\s+\[logger\]\s+\[send\]\s+unary RPC -- \/logservice.LogService\/Log.*`,
			},
		},
	}

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	var stop = func(test test) {
		test.s.Stop()
		return
	}

	var initClient = func() (client.GRPCLogger, chan error) {
		return client.New(
			client.WithLogger(log.New(log.WithOut(buf[2]), log.SkipExit)),
			client.WithAddr("127.0.0.1:9099"),
			client.UnaryRPC(),
			client.WithGRPCOpts(),
		)
	}

	var bufferFilter = func(in []byte) [][]byte {
		// split lines
		var line [][]byte
		var buf []byte

		for _, b := range in {
			if b == 10 {
				if len(buf) > 0 {
					copy := buf
					line = append(line, copy)
					buf = []byte{}
				}
				continue
			}
			buf = append(buf, b)
		}

		if len(buf) > 0 {
			copy := buf
			line = append(line, copy)
			buf = []byte{}
		}

		return line
	}

	var bufferMatcher = func(test test, lines [][]byte) bool {
		len := len(test.matchers)

		var pass int = 0

		for _, m := range test.matchers {
			r := regexp.MustCompile(m)

			for _, entry := range lines {
				if r.Match(entry) {
					pass++
					break
				}
			}
		}

		if pass != len {
			return false
		}

		return true
	}

	var verifyServiceLogger = func(
		idx int,
		test test,
		c client.GRPCLogger,
		done chan struct{},
	) {
		c.Info("null")
		time.Sleep(time.Second)

		filter := bufferFilter(buf[1].Bytes())

		var lines []string
		for _, l := range filter {
			lines = append(lines, string(l))
		}

		if !bufferMatcher(test, filter) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] couldn't detect expected interceptor entries (%v): expected %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				len(test.matchers),
				test.matchers,
				lines,
				test.name,
			)
			test.s.ErrCh <- errors.New("couldn't detect expected interceptor entries")
			return
		}

		done <- struct{}{}

	}

	var verify = func(idx int, test test) {
		defer reset()
		defer stop(test)

		var done = make(chan struct{})

		go test.s.Serve()
		time.Sleep(time.Second)

		c, clientErr := initClient()

		go verifyServiceLogger(idx, test, c, done)

		for {
			select {
			case err := <-clientErr:
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unexpected client error: %v -- action: %s",
					idx,
					module,
					funcname,
					err,
					test.name,
				)
				return

			case err := <-test.s.ErrCh:
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unexpected server error: %v -- action: %s",
					idx,
					module,
					funcname,
					err,
					test.name,
				)
				return
			case <-done:
				return
			}
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
		name     string
		s        *GRPCLogServer
		matchers []string
	}

	var buf = []*bytes.Buffer{{}, {}, {}}

	var tests = []test{
		{
			name: "unary server logging test",
			s: New(
				WithLogger(log.New(log.WithOut(buf[0]), log.SkipExit)),
				WithServiceLoggerV(log.New(log.WithOut(buf[1]), log.CfgTextLevelFirst, log.SkipExit)),
				WithAddr("127.0.0.1:9099"),
				WithGRPCOpts(),
			),
			matchers: []string{
				`^\[debug\]\s+\[.*\]\s+\[gRPC\]\s+\[logger\]\s+\[recv\]\s+stream RPC logger -- received message from gRPC client.*`,
				`^\[debug\]\s+\[.*\]\s+\[gRPC\]\s+\[logger\]\s+\[send\]\s+stream RPC logger -- sent message to gRPC client.*`,
			},
		},
		{
			name: "unary server logging w/ timer",
			s: New(
				WithLogger(log.New(log.WithOut(buf[0]), log.SkipExit)),
				WithServiceLoggerV(log.New(log.WithOut(buf[1]), log.CfgTextLevelFirst, log.SkipExit)),
				WithTiming(),
				WithAddr("127.0.0.1:9099"),
				WithGRPCOpts(),
			),
			matchers: []string{
				`^\[debug\]\s+\[.*\]\s+\[gRPC\]\s+\[logger\]\s+\[recv\]\s+stream RPC logger -- received message from gRPC client.*`,
				`^\[debug\]\s+\[.*\]\s+\[gRPC\]\s+\[logger\]\s+\[send\]\s+stream RPC logger -- sent message to gRPC client.*`,
			},
		},
	}

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	var stop = func(test test) {
		test.s.Stop()
		return
	}

	var initClient = func() (client.GRPCLogger, chan error) {
		return client.New(
			client.WithLogger(log.New(log.WithOut(buf[2]), log.SkipExit)),
			client.WithAddr("127.0.0.1:9099"),
			client.StreamRPC(),
			client.WithGRPCOpts(),
		)
	}

	var bufferFilter = func(in []byte) [][]byte {
		// split lines
		var line [][]byte
		var buf []byte

		for _, b := range in {
			if b == 10 {
				if len(buf) > 0 {
					copy := buf
					line = append(line, copy)
					buf = []byte{}
				}
				continue
			}
			buf = append(buf, b)
		}

		if len(buf) > 0 {
			copy := buf
			line = append(line, copy)
			buf = []byte{}
		}

		return line
	}

	var bufferMatcher = func(test test, lines [][]byte) bool {
		len := len(test.matchers)

		var pass int = 0

		for _, m := range test.matchers {
			r := regexp.MustCompile(m)

			for _, entry := range lines {
				if r.Match(entry) {
					pass++
					break
				}
			}
		}

		if pass != len {
			return false
		}

		return true
	}

	var verifyServiceLogger = func(
		idx int,
		test test,
		c client.GRPCLogger,
		done chan struct{},
	) {
		c.Info("null")
		time.Sleep(time.Second)

		filter := bufferFilter(buf[1].Bytes())

		var lines []string
		for _, l := range filter {
			lines = append(lines, string(l)+"\n")
		}

		if !bufferMatcher(test, filter) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] couldn't detect expected interceptor entries (%v): expected %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				len(test.matchers),
				test.matchers,
				lines,
				test.name,
			)
			test.s.ErrCh <- errors.New("couldn't detect expected interceptor entries")
			return
		}

		done <- struct{}{}

	}

	var verify = func(idx int, test test) {
		defer reset()
		defer stop(test)

		var done = make(chan struct{})

		go test.s.Serve()
		time.Sleep(time.Second)

		c, clientErr := initClient()

		go verifyServiceLogger(idx, test, c, done)

		for {
			select {
			case err := <-clientErr:
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unexpected client error: %v -- action: %s",
					idx,
					module,
					funcname,
					err,
					test.name,
				)
				return

			case err := <-test.s.ErrCh:
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unexpected server error: %v -- action: %s",
					idx,
					module,
					funcname,
					err,
					test.name,
				)
				return
			case <-done:
				return
			}
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

	var buf = new(bytes.Buffer)
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
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
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
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
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
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
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
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
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
				stream:    testServerStream{},
				logger:    log.New(log.WithOut(buf), log.SkipExit),
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
				stream:    testServerStream{},
				logger:    log.New(log.WithOut(buf), log.SkipExit),
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

	var buf = new(bytes.Buffer)

	var tests = []test{
		{
			name: "working test",
			t: &loggingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
				method: "testLog",
			},
			m:  event.New().Message("null").Build(),
			ok: true,
		},
		{
			name: "errored test",
			t: &loggingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
				method: "testLog",
			},
			m: event.New().Level(event.Level_error).Message("null").Build(),
		},
		{
			name: "errored with EOF",
			t: &loggingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
				method: "testLog",
			},
			m: event.New().Message("EOF").Build(),
		},
		{
			name: "errored with deadline exceeded",
			t: &loggingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
				method: "testLog",
			},
			m: event.New().Message("deadline").Build(),
		},
		{
			name: "working test w/ timer",
			t: &loggingStream{
				stream:    testServerStream{},
				logger:    log.New(log.WithOut(buf), log.SkipExit),
				method:    "testLog",
				withTimer: true,
			},
			m:  event.New().Message("null").Build(),
			ok: true,
		},
		{
			name: "errored test w/ timer",
			t: &loggingStream{
				stream:    testServerStream{},
				logger:    log.New(log.WithOut(buf), log.SkipExit),
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

	var buf = new(bytes.Buffer)

	var tests = []test{
		{
			name: "working test",
			t: &loggingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
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

	var buf = new(bytes.Buffer)

	var tests = []test{
		{
			name: "working test",
			t: &loggingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
				method: "testLog",
			},
			m: metadata.New(map[string]string{
				"test": "header",
			}),
		},
	}

	var verify = func(idx int, test test) {
		err := test.t.SendHeader(test.m)

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

	var buf = new(bytes.Buffer)

	var tests = []test{
		{
			name: "working test",
			t: &loggingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
				method: "testLog",
			},
			m: metadata.New(map[string]string{
				"test": "trailer",
			}),
		},
	}

	var verify = func(idx int, test test) {
		test.t.SetTrailer(test.m)
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

	var buf = new(bytes.Buffer)

	var tests = []test{
		{
			name: "working test",
			t: &loggingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
				method: "testLog",
			},
		},
	}

	var verify = func(idx int, test test) {
		ctx := test.t.Context()

		if ctx == nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected nil context -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
