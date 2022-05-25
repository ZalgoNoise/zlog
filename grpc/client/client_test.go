package client

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/zalgonoise/zlog/grpc/server"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	jsonpb "github.com/zalgonoise/zlog/log/format/json"
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
		name  string
		cfg   []LogClientConfig
		wants testGRPCLogger
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
			name:  "default config, no input",
			cfg:   []LogClientConfig{},
			wants: expectedLoggers[0],
		},
		{
			name: "with custom config (one entry)",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
			},
			wants: expectedLoggers[1],
		},
		{
			name: "with custom config (three entries)",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
				WithLogger(writers[0]),
				UnaryRPC(),
			},
			wants: expectedLoggers[2],
		},
		{
			name:  "with nil input",
			cfg:   nil,
			wants: expectedLoggers[3],
		},
	}

	var verifyLoggers = func(idx int, test test, client GRPCLogger, errCh chan error, done chan struct{}) {
		if client.(*GRPCLogClient).addr.Len() != test.wants.l.(*GRPCLogClient).addr.Len() {
			errCh <- fmt.Errorf(
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
			errCh <- fmt.Errorf(
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
			errCh <- fmt.Errorf(
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
			errCh <- fmt.Errorf(
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

		done <- struct{}{}
	}

	var verify = func(idx int, test test) {
		var done = make(chan struct{})

		client, errCh := New(test.cfg...)

		// test Channels() execution
		client.Channels()

		if client == nil || errCh == nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] client or error channel are unexpectedly nil values -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		go verifyLoggers(idx, test, client, errCh, done)

		for {
			select {
			case err := <-errCh:
				t.Error(err.Error())
				return
			case <-done:
				return
			}
		}

	}

	// sleep to allow server to start up
	time.Sleep(time.Millisecond * 400)

	for idx, test := range tests {
		verify(idx, test)
	}

}

func TestGRPCClientAction(t *testing.T) {
	module := "GRPCLogClient"
	funcname := "log() / stream()"

	_ = module
	_ = funcname

	type testGRPCLogger struct {
		l GRPCLogger
		e chan error
	}

	type test struct {
		name string
		cfg  []LogClientConfig
	}

	var bufs = []*bytes.Buffer{{}, {}, {}}

	var writers = []log.Logger{
		log.New(log.WithOut(bufs[0]), log.SkipExit, log.CfgFormatJSONSkipNewline),
		log.New(log.WithOut(bufs[1]), log.SkipExit, log.CfgFormatJSONSkipNewline),
		log.New(log.WithOut(bufs[2]), log.SkipExit, log.CfgFormatJSONSkipNewline),
	}

	var mockServer = server.New(
		server.WithAddr("127.0.0.1:9099"),
		server.WithLogger(writers[0]),
	)

	go mockServer.Serve()
	defer mockServer.Stop()

	var tests = []test{
		{
			name: "Unary RPC logger",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
				WithLogger(writers[1]),
				UnaryRPC(),
			},
		},
		{
			name: "Stream RPC logger",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
				WithLogger(writers[2]),
				StreamRPC(),
			},
		},
	}

	var verifyLoggers = func(idx int, test test, client GRPCLogger, errCh chan error, done chan struct{}) {
		defer bufs[0].Reset()

		f := jsonpb.FmtJSON{SkipNewline: true}
		in := event.New().Message("test").Build()
		out, err := f.Format(in)

		if err != nil {
			errCh <- fmt.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected JSON formatter error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		n, err := client.Output(in)
		time.Sleep(time.Millisecond * 50)

		if err != nil {
			errCh <- fmt.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		if n == 0 {
			errCh <- fmt.Errorf(
				"#%v -- FAILED -- [%s] [%s] zero bytes written error -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		buf := bufs[0].Bytes()

		if !reflect.DeepEqual(buf, out) {
			errCh <- fmt.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				string(out),
				string(buf),
				test.name,
			)
			return
		}

		done <- struct{}{}

	}

	var verify = func(idx int, test test) {
		var done = make(chan struct{})

		client, errCh := New(test.cfg...)
		defer client.Close()

		if client == nil || errCh == nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] client or error channel are unexpectedly nil values -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		go verifyLoggers(idx, test, client, errCh, done)

		for {
			select {
			case err := <-errCh:
				t.Error(err.Error())
				return
			case <-done:
				return
			}
		}
	}

	// sleep to allow server to start up
	time.Sleep(time.Millisecond * 400)

	for idx, test := range tests {
		verify(idx, test)
	}
}
