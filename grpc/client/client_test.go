package client

import (
	"reflect"
	"testing"
	"time"

	"github.com/zalgonoise/zlog/grpc/server"
	"github.com/zalgonoise/zlog/log"
)

func TestNew(t *testing.T) {
	module := "GRPCLogClient"
	funcname := "New()"

	_ = module
	_ = funcname

	var mockServer = server.New(
		server.WithAddr("127.0.0.1:9099"),
		server.WithLogger(log.New(log.NilConfig)),
	)

	go mockServer.Serve()
	defer mockServer.Stop()

	type testGRPCLogger struct {
		l GRPCLogger
		e chan error
	}

	type test struct {
		name    string
		cfg     []LogClientConfig
		wants   testGRPCLogger
		optsLen int
	}

	var writers = []log.Logger{
		log.New(log.NilConfig),
		log.New(),
		log.New(log.SkipExit),
	}

	var expectedLoggers = func() []testGRPCLogger {
		var s []testGRPCLogger

		defaultL, defaultE := New()
		defaultS := testGRPCLogger{
			l: defaultL,
			e: defaultE,
		}
		s = append(s, defaultS)

		writerL, writerE := New(WithAddr("127.0.0.1:9099"))
		writerS := testGRPCLogger{
			l: writerL,
			e: writerE,
		}
		s = append(s, writerS)

		writerTwoL, writerTwoE := New(
			WithAddr("127.0.0.1:9099"),
			WithLogger(writers[0]),
			UnaryRPC(),
		)
		writerTwoS := testGRPCLogger{
			l: writerTwoL,
			e: writerTwoE,
		}
		s = append(s, writerTwoS)

		defaultTwoL, defaultTwoE := New(nil)
		defaultTwoS := testGRPCLogger{
			l: defaultTwoL,
			e: defaultTwoE,
		}
		s = append(s, defaultTwoS)

		return s
	}()

	var tests = []test{
		{
			name:    "default config, no input",
			cfg:     []LogClientConfig{},
			wants:   expectedLoggers[0],
			optsLen: 0,
		},
		{
			name: "with custom config (one entry)",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
			},
			wants:   expectedLoggers[1],
			optsLen: 1,
		},
		{
			name: "with custom config (three entries)",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
				WithLogger(writers[0]),
				UnaryRPC(),
			},
			wants:   expectedLoggers[2],
			optsLen: 3,
		},
		{
			name:    "with nil input",
			cfg:     nil,
			wants:   expectedLoggers[3],
			optsLen: 0,
		},
	}

	var verify = func(idx int, test test) {
		client, errCh := New(test.cfg...)

		if client == nil || errCh == nil {
			t.Error()
			return
		}

		if client.(*GRPCLogClient).addr.Len() != test.wants.l.(*GRPCLogClient).addr.Len() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] connections-addresses length mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants.l.(*GRPCLogClient).addr.Len(),
				client.(*GRPCLogClient).addr.Len(),
				test.name,
			)
			return
		}

		if len(client.(*GRPCLogClient).opts) != len(test.wants.l.(*GRPCLogClient).opts) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] gRPC options length mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants.l.(*GRPCLogClient).opts,
				client.(*GRPCLogClient).opts,
				test.name,
			)
			return
		}

		if !reflect.DeepEqual(client.(*GRPCLogClient).svcLogger, test.wants.l.(*GRPCLogClient).svcLogger) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] logger mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants.l.(*GRPCLogClient).svcLogger,
				client.(*GRPCLogClient).svcLogger,
				test.name,
			)
			return
		}

		if !reflect.DeepEqual(client.(*GRPCLogClient).backoff, test.wants.l.(*GRPCLogClient).backoff) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] backoff module mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants.l.(*GRPCLogClient).backoff,
				client.(*GRPCLogClient).backoff,
				test.name,
			)
			return
		}
	}

	// sleep to allow server to start up
	time.Sleep(time.Millisecond * 400)

	for idx, test := range tests {
		verify(idx, test)
	}
}
