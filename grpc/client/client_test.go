package client

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/zalgonoise/zlog/grpc/address"
	"github.com/zalgonoise/zlog/grpc/server"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	jsonpb "github.com/zalgonoise/zlog/log/format/json"
	"github.com/zalgonoise/zlog/store/fs"
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

func TestSetOuts(t *testing.T) {
	module := "GRPCLogClient"
	funcname := "SetOuts()"

	_ = module
	_ = funcname

	var mockServerA = server.New(
		server.WithAddr("127.0.0.1:9099"),
		server.WithLogger(log.New(log.NilConfig)),
	)

	go mockServerA.Serve()
	defer mockServerA.Stop()

	var mockServerB = server.New(
		server.WithAddr("127.0.0.1:9098"),
		server.WithLogger(log.New(log.NilConfig)),
	)

	go mockServerB.Serve()
	defer mockServerB.Stop()

	type test struct {
		name   string
		cfg    []LogClientConfig
		setter []io.Writer
		wants  []string
		fails  bool
	}

	var tests = []test{
		{
			name: "setting one valid address",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
			},
			setter: []io.Writer{
				address.New("127.0.0.1:9098"),
			},
			wants: []string{
				"127.0.0.1:9098",
			},
		},
		{
			name: "setting multiple valid addresses",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
			},
			setter: []io.Writer{
				address.New("127.0.0.1:9099"),
				address.New("127.0.0.1:9098"),
			},
			wants: []string{
				"127.0.0.1:9099",
				"127.0.0.1:9098",
			},
		},
		{
			name: "setting nil values",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
			},
			setter: []io.Writer{
				nil,
			},
			wants: []string{
				"127.0.0.1:9099",
			},
		},
		{
			name: "setting multiple nil values",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
			},
			setter: []io.Writer{
				nil,
				nil,
				nil,
			},
			wants: []string{
				"127.0.0.1:9099",
			},
		},
		{
			name: "setting multiple addresses values mixed with nils",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
			},
			setter: []io.Writer{
				nil,
				address.New("127.0.0.1:9099"),
				nil,
				address.New("127.0.0.1:9098"),
				nil,
			},
			wants: []string{
				"127.0.0.1:9099",
				"127.0.0.1:9098",
			},
		},
		{
			name: "nil input",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
			},
			setter: nil,
			wants: []string{
				"127.0.0.1:9099",
			},
		},
		{
			name: "setting one invalid writer",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
			},
			setter: []io.Writer{
				&fs.Logfile{},
			},
			wants: []string{
				"127.0.0.1:9099",
			},
			fails: true,
		},
	}

	var verifyLoggers = func(idx int, test test, logger GRPCLogger, errCh chan error, done chan struct{}) {
		var res log.Logger

		if test.setter == nil {
			res = logger.SetOuts()
		} else {
			res = logger.SetOuts(test.setter...)
		}

		if res == nil && !test.fails {
			errCh <- fmt.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error setting writers -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		out := logger.(*GRPCLogClient)
		keys := out.addr.Keys()

		if len(keys) != len(test.wants) {
			errCh <- fmt.Errorf(
				"#%v -- FAILED -- [%s] [%s] output length mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				len(test.wants),
				len(keys),
				test.name,
			)
			return
		}

		for _, k := range keys {
			var pass bool
			for _, a := range test.wants {
				if k == a {
					pass = true
					break
				}
			}

			if !pass {
				errCh <- fmt.Errorf(
					"#%v -- FAILED -- [%s] [%s] output mismatch error: no matches for addr %s -- action: %s",
					idx,
					module,
					funcname,
					k,
					test.name,
				)
				return
			}
		}

		done <- struct{}{}
	}

	var verify = func(idx int, test test) {
		logger, errCh := New(test.cfg...)

		done := make(chan struct{})

		go verifyLoggers(idx, test, logger, errCh, done)

		for {
			select {
			case err := <-errCh:
				if !test.fails {
					t.Error(err.Error())
					return
				}
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

func TestAddOuts(t *testing.T) {
	module := "GRPCLogClient"
	funcname := "AddOuts()"

	_ = module
	_ = funcname

	var mockServerA = server.New(
		server.WithAddr("127.0.0.1:9099"),
		server.WithLogger(log.New(log.NilConfig)),
	)

	go mockServerA.Serve()
	defer mockServerA.Stop()

	var mockServerB = server.New(
		server.WithAddr("127.0.0.1:9098"),
		server.WithLogger(log.New(log.NilConfig)),
	)

	go mockServerB.Serve()
	defer mockServerB.Stop()

	var mockServerC = server.New(
		server.WithAddr("127.0.0.1:9097"),
		server.WithLogger(log.New(log.NilConfig)),
	)

	go mockServerC.Serve()
	defer mockServerC.Stop()

	type test struct {
		name   string
		cfg    []LogClientConfig
		setter []io.Writer
		wants  []string
		fails  bool
	}

	var tests = []test{
		{
			name: "setting one valid address",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9097"),
			},
			setter: []io.Writer{
				address.New("127.0.0.1:9098"),
			},
			wants: []string{
				"127.0.0.1:9097",
				"127.0.0.1:9098",
			},
		},
		{
			name: "setting the same address already configured",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9097"),
			},
			setter: []io.Writer{
				address.New("127.0.0.1:9097"),
			},
			wants: []string{
				"127.0.0.1:9097",
			},
		},
		{
			name: "setting multiple valid addresses",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9097"),
			},
			setter: []io.Writer{
				address.New("127.0.0.1:9099"),
				address.New("127.0.0.1:9098"),
			},
			wants: []string{
				"127.0.0.1:9097",
				"127.0.0.1:9098",
				"127.0.0.1:9099",
			},
		},
		{
			name: "setting nil values",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9097"),
			},
			setter: []io.Writer{
				nil,
			},
			wants: []string{
				"127.0.0.1:9097",
			},
		},
		{
			name: "setting multiple nil values",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9097"),
			},
			setter: []io.Writer{
				nil,
				nil,
				nil,
			},
			wants: []string{
				"127.0.0.1:9097",
			},
		},
		{
			name: "setting multiple addresses values mixed with nils",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9097"),
			},
			setter: []io.Writer{
				nil,
				address.New("127.0.0.1:9099"),
				nil,
				address.New("127.0.0.1:9098"),
				nil,
			},
			wants: []string{
				"127.0.0.1:9097",
				"127.0.0.1:9098",
				"127.0.0.1:9099",
			},
		},
		{
			name: "nil input",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9097"),
			},
			setter: nil,
			wants: []string{
				"127.0.0.1:9097",
			},
		},
		{
			name: "setting one invalid writer",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9097"),
			},
			setter: []io.Writer{
				&fs.Logfile{},
			},
			wants: []string{
				"127.0.0.1:9097",
			},
			fails: true,
		},
	}

	var verifyLoggers = func(idx int, test test, logger GRPCLogger, errCh chan error, done chan struct{}) {
		var res log.Logger

		if test.setter == nil {
			res = logger.AddOuts()
		} else {
			res = logger.AddOuts(test.setter...)
		}

		if res == nil && !test.fails {
			errCh <- fmt.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error setting writers -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		out := logger.(*GRPCLogClient)
		keys := out.addr.Keys()

		if len(keys) != len(test.wants) {
			errCh <- fmt.Errorf(
				"#%v -- FAILED -- [%s] [%s] output length mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				len(test.wants),
				len(keys),
				test.name,
			)
			return
		}

		for _, k := range keys {
			var pass bool
			for _, a := range test.wants {
				if k == a {
					pass = true
					break
				}
			}

			if !pass {
				errCh <- fmt.Errorf(
					"#%v -- FAILED -- [%s] [%s] output mismatch error: no matches for addr %s -- action: %s",
					idx,
					module,
					funcname,
					k,
					test.name,
				)
				return
			}
		}

		done <- struct{}{}
	}

	var verify = func(idx int, test test) {
		logger, errCh := New(test.cfg...)

		done := make(chan struct{})

		go verifyLoggers(idx, test, logger, errCh, done)

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

func TestWrite(t *testing.T) {
	module := "GRPCLogClient"
	funcname := "Write()"

	_ = module
	_ = funcname

	var mockServer = server.New(
		server.WithAddr("127.0.0.1:9099"),
		server.WithLogger(log.New(log.NilConfig)),
	)

	go mockServer.Serve()
	defer mockServer.Stop()

	type test struct {
		name string
		cfg  []LogClientConfig
		msg  []byte
		ok   bool
	}

	var tests = []test{
		{
			name: "encoded event test",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
			},
			msg: event.New().Message("null").Build().Encode(),
			ok:  true,
		},
		{
			name: "byte string test",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
			},
			msg: []byte("test"),
			ok:  true,
		},
		{
			name: "zero byte input",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
			},
			msg: []byte{},
		},
		{
			name: "nil input",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
			},
			msg: nil,
		},
	}

	var verifyLoggers = func(idx int, test test, logger GRPCLogger, errCh chan error, done chan struct{}) {
		n, err := logger.Write(test.msg)

		if err != nil && test.ok {
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

		if n == 0 && test.ok {
			errCh <- fmt.Errorf(
				"#%v -- FAILED -- [%s] [%s] zero bytes written error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		done <- struct{}{}
	}

	var verify = func(idx int, test test) {
		logger, errCh := New(test.cfg...)

		done := make(chan struct{})

		go verifyLoggers(idx, test, logger, errCh, done)

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

func FuzzLoggerPrefix(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Prefix()"

	l := &GRPCLogClient{}

	f.Add("")
	f.Add("test-prefix")
	f.Fuzz(func(t *testing.T, a string) {
		l.Prefix(a)

		if l.prefix != a && a != "" {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed prefix mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				l.prefix,
			)
			return
		}
	})
}

func FuzzLoggerSub(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Sub()"

	l := &GRPCLogClient{}

	f.Add("")
	f.Add("test-subprefix")
	f.Fuzz(func(t *testing.T, a string) {
		l.Sub(a)

		if l.sub != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed prefix mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				l.sub,
			)
			return
		}
	})
}

func TestFields(t *testing.T) {
	module := "GRPCLogClient"
	funcname := "Fields()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		init  map[string]interface{}
		meta  map[string]interface{}
		wants map[string]interface{}
	}

	var tests = []test{
		{
			name:  "setting new metadata from empty",
			init:  nil,
			meta:  map[string]interface{}{"a": true},
			wants: map[string]interface{}{"a": true},
		},
		{
			name:  "setting new metadata from existing",
			init:  map[string]interface{}{"b": false},
			meta:  map[string]interface{}{"a": true},
			wants: map[string]interface{}{"a": true},
		},
		{
			name:  "resetting meta with empty obj",
			init:  map[string]interface{}{"b": false},
			meta:  map[string]interface{}{},
			wants: map[string]interface{}{},
		},
		{
			name:  "resetting meta with nil",
			init:  map[string]interface{}{"b": false},
			meta:  nil,
			wants: map[string]interface{}{},
		},
	}

	var verify = func(idx int, test test) {
		l := &GRPCLogClient{
			meta: test.init,
		}

		l.Fields(test.meta)

		if !reflect.DeepEqual(l.meta, test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				l.meta,
				test.name,
			)
			return

		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}

}

func TestIsSkipExit(t *testing.T) {
	module := "GRPCLogClient"
	funcname := "IsSkipExit()"

	_ = module
	_ = funcname

	type test struct {
		name      string
		svcLogger log.Logger
		wants     bool
	}

	var buf = &bytes.Buffer{}

	var tests = []test{
		{
			name:      "with skip exit logger",
			svcLogger: log.New(log.WithOut(buf), log.SkipExit),
			wants:     true,
		},
		{
			name:      "without skip exit logger",
			svcLogger: log.New(log.WithOut(buf)),
		},
	}

	var verify = func(idx int, test test) {
		l := &GRPCLogClient{
			svcLogger: test.svcLogger,
		}

		if l.IsSkipExit() != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				l.meta,
				test.name,
			)
			return

		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}

}
