package client

import (
	"bytes"
	"errors"
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

const decodeLimit = 300

// const decodeWait = time.Millisecond * 150
const maxWaitTime time.Duration = time.Millisecond * 50

var errDecodeDeadlineExceeded = errors.New("deadline exceeded")

func decode(buf *bytes.Buffer, n int) (*event.Event, error) {
	if n == decodeLimit {
		return nil, errDecodeDeadlineExceeded
	}

	n++

	e := buf.Bytes()

	if len(e) == 0 {
		time.Sleep(maxWaitTime * 2)
		return decode(buf, n)
	}

	buf.Reset()
	return jsonpb.Decode(e)
}

func TestNew(t *testing.T) {
	module := "GRPCLogClient"
	funcname := "New()"

	_ = module
	_ = funcname

	type test struct {
		name      string
		cfg       []LogClientConfig
		addr      string
		hasLogger bool
	}

	var testAddr string = "127.0.0.1:9099"

	var tests = []test{
		{
			name: "default config, no input",
			cfg:  []LogClientConfig{},
		},
		{
			name: "with custom config (one entry)",
			cfg: []LogClientConfig{
				WithAddr(testAddr),
			},
			addr: testAddr,
		},
		{
			name: "with custom config (three entries)",
			cfg: []LogClientConfig{
				WithAddr(testAddr),
				WithLogger(log.New(log.NilConfig)),
				UnaryRPC(),
			},
			addr:      testAddr,
			hasLogger: true,
		},
		{
			name: "with custom config (three entries) containing nil values",
			cfg: []LogClientConfig{
				WithAddr(testAddr),
				nil,
				WithLogger(log.New(log.NilConfig)),
				nil,
				UnaryRPC(),
			},
			addr:      testAddr,
			hasLogger: true,
		},
		{
			name: "with nil input",
			cfg:  nil,
		},
	}

	var verify = func(idx int, test test) {
		logger, _ := New(test.cfg...)
		l := logger.(*GRPCLogClient)

		// check address
		if test.addr != "" {
			if l.addr.Len() > 0 {
				addr := l.addr.Keys()

				if addr[0] != test.addr {
					t.Errorf(
						"#%v -- FAILED -- [%s] [%s] address mismatch error: wanted %s ; got %s -- action: %s",
						idx,
						module,
						funcname,
						test.addr,
						addr[0],
						test.name,
					)
					return
				}

			} else {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] expected a configured address but address map was empty -- action: %s",
					idx,
					module,
					funcname,
					test.name,
				)
				return
			}
		}

		// check initialized elements
		if l.msgCh == nil || l.done == nil || l.errCh == nil || l.backoff == nil || l.meta == nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected uninitialized elements error -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		// check if logger is set
		if test.hasLogger && l.svcLogger == nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] expected a configured logger but it is nil -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		// check default prefix and sub
		if l.prefix != "log" || l.sub != "" {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] prefix / sub-prefix mismatch error -- prefix: wanted \"%s\" ; got \"%s\" -- sub: wanted \"%s\" ; got \"%s\" -- action: %s",
				idx,
				module,
				funcname,
				"log",
				l.prefix,
				"",
				l.sub,
				test.name,
			)
			return
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}

}

// func TestNew(t *testing.T) {
// 	module := "GRPCLogClient"
// 	funcname := "New()"

// 	_ = module
// 	_ = funcname

// 	var mockServer = server.New(
// 		server.WithAddr("127.0.0.1:9099"),
// 		server.WithLogger(log.New(log.NilConfig)),
// 	)

// 	go mockServer.Serve()
// 	defer mockServer.Stop()

// 	type testGRPCLogger struct {
// 		l GRPCLogger
// 		e chan error
// 	}

// 	type test struct {
// 		name  string
// 		cfg   []LogClientConfig
// 		wants testGRPCLogger
// 	}

// 	var writers = []log.Logger{
// 		log.New(log.NilConfig),
// 		log.New(),
// 		log.New(log.SkipExit),
// 	}

// 	var expectedLoggers = func() []testGRPCLogger {
// 		var s []testGRPCLogger

// 		defaultL, defaultE := New()
// 		defaultS := testGRPCLogger{
// 			l: defaultL,
// 			e: defaultE,
// 		}
// 		s = append(s, defaultS)

// 		writerL, writerE := New(WithAddr("127.0.0.1:9099"))
// 		writerS := testGRPCLogger{
// 			l: writerL,
// 			e: writerE,
// 		}
// 		s = append(s, writerS)

// 		writerTwoL, writerTwoE := New(
// 			WithAddr("127.0.0.1:9099"),
// 			WithLogger(writers[0]),
// 			UnaryRPC(),
// 		)
// 		writerTwoS := testGRPCLogger{
// 			l: writerTwoL,
// 			e: writerTwoE,
// 		}
// 		s = append(s, writerTwoS)

// 		defaultTwoL, defaultTwoE := New(nil)
// 		defaultTwoS := testGRPCLogger{
// 			l: defaultTwoL,
// 			e: defaultTwoE,
// 		}
// 		s = append(s, defaultTwoS)

// 		return s
// 	}()

// 	var tests = []test{
// 		{
// 			name:  "default config, no input",
// 			cfg:   []LogClientConfig{},
// 			wants: expectedLoggers[0],
// 		},
// 		{
// 			name: "with custom config (one entry)",
// 			cfg: []LogClientConfig{
// 				WithAddr("127.0.0.1:9099"),
// 			},
// 			wants: expectedLoggers[1],
// 		},
// 		{
// 			name: "with custom config (three entries)",
// 			cfg: []LogClientConfig{
// 				WithAddr("127.0.0.1:9099"),
// 				WithLogger(writers[0]),
// 				UnaryRPC(),
// 			},
// 			wants: expectedLoggers[2],
// 		},
// 		{
// 			name:  "with nil input",
// 			cfg:   nil,
// 			wants: expectedLoggers[3],
// 		},
// 	}

// 	var verifyLoggers = func(idx int, test test, client GRPCLogger, errCh chan error, done chan struct{}) {
// 		if client.(*GRPCLogClient).addr.Len() != test.wants.l.(*GRPCLogClient).addr.Len() {
// 			errCh <- fmt.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] connections-addresses length mismatch error: wanted %v ; got %v -- action: %s",
// 				idx,
// 				module,
// 				funcname,
// 				test.wants.l.(*GRPCLogClient).addr.Len(),
// 				client.(*GRPCLogClient).addr.Len(),
// 				test.name,
// 			)
// 			return
// 		}

// 		if len(client.(*GRPCLogClient).opts) != len(test.wants.l.(*GRPCLogClient).opts) {
// 			errCh <- fmt.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] gRPC options length mismatch error: wanted %v ; got %v -- action: %s",
// 				idx,
// 				module,
// 				funcname,
// 				test.wants.l.(*GRPCLogClient).opts,
// 				client.(*GRPCLogClient).opts,
// 				test.name,
// 			)
// 			return
// 		}

// 		if !reflect.DeepEqual(client.(*GRPCLogClient).svcLogger, test.wants.l.(*GRPCLogClient).svcLogger) {
// 			errCh <- fmt.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] logger mismatch error: wanted %v ; got %v -- action: %s",
// 				idx,
// 				module,
// 				funcname,
// 				test.wants.l.(*GRPCLogClient).svcLogger,
// 				client.(*GRPCLogClient).svcLogger,
// 				test.name,
// 			)
// 			return
// 		}

// 		if !reflect.DeepEqual(client.(*GRPCLogClient).backoff, test.wants.l.(*GRPCLogClient).backoff) {
// 			errCh <- fmt.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] backoff module mismatch error: wanted %v ; got %v -- action: %s",
// 				idx,
// 				module,
// 				funcname,
// 				test.wants.l.(*GRPCLogClient).backoff,
// 				client.(*GRPCLogClient).backoff,
// 				test.name,
// 			)
// 			return
// 		}

// 		done <- struct{}{}
// 	}

// 	var verify = func(idx int, test test) {
// 		var done = make(chan struct{})

// 		client, errCh := New(test.cfg...)

// 		// test Channels() execution
// 		client.Channels()

// 		if client == nil || errCh == nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] client or error channel are unexpectedly nil values -- action: %s",
// 				idx,
// 				module,
// 				funcname,
// 				test.name,
// 			)
// 			return
// 		}

// 		go verifyLoggers(idx, test, client, errCh, done)

// 		for {
// 			select {
// 			case err := <-errCh:
// 				t.Error(err.Error())
// 				return
// 			case <-done:
// 				return
// 			}
// 		}

// 	}

// 	// sleep to allow server to start up
// 	time.Sleep(time.Millisecond * 400)

// 	for idx, test := range tests {
// 		verify(idx, test)
// 	}

// }

// E2E
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
		addr string
	}

	var mockAddr = []string{
		"127.0.0.1:45060",
		"127.0.0.1:45061",
	}

	var bufs = []*bytes.Buffer{{}, {}, {}}

	var writers = []log.Logger{
		log.New(log.WithOut(bufs[0]), log.SkipExit, log.CfgFormatJSONSkipNewline),
		log.New(log.WithOut(bufs[1]), log.SkipExit, log.CfgFormatJSONSkipNewline),
		log.New(log.WithOut(bufs[2]), log.SkipExit, log.CfgFormatJSONSkipNewline),
	}

	var tests = []test{
		{
			name: "Unary RPC logger",
			cfg: []LogClientConfig{
				WithAddr(mockAddr[0]),
				WithLogger(writers[1]),
				UnaryRPC(),
			},
			addr: mockAddr[0],
		},
		{
			name: "Stream RPC logger",
			cfg: []LogClientConfig{
				WithAddr(mockAddr[1]),
				WithLogger(writers[2]),
				StreamRPC(),
			},
			addr: mockAddr[1],
		},
	}

	var verifyLoggers = func(idx int, test test, client GRPCLogger, errCh chan error, done chan struct{}) {
		bufs[0].Reset()
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
		time.Sleep(maxWaitTime)

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
		// prepare local server
		var mockServer = server.New(
			server.WithAddr(test.addr),
			server.WithLogger(writers[0]),
		)

		go mockServer.Serve()
		defer mockServer.Stop()

		// sleep to allow server to start up
		time.Sleep(maxWaitTime)

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

	for idx, test := range tests {
		verify(idx, test)
	}
}

// E2E
func TestSetOuts(t *testing.T) {
	module := "GRPCLogClient"
	funcname := "SetOuts()"

	_ = module
	_ = funcname

	var mockAddr = []string{
		"127.0.0.1:45062",
	}

	// prepare local server
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(log.NilConfig)),
	)

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	var testConnAddr = &address.ConnAddr{}
	testConnAddr.Add(mockAddr[0])

	type test struct {
		name   string
		cfg    *address.ConnAddr
		setter []io.Writer
		fails  bool
	}

	var tests = []test{
		{
			name: "setting one valid address",
			cfg:  testConnAddr,
			setter: []io.Writer{
				address.New(mockAddr[0]),
			},
		},
		{
			name: "setting multiple valid addresses",
			cfg:  testConnAddr,
			setter: []io.Writer{
				address.New(mockAddr[0]),
				address.New(mockAddr[0]),
			},
		},
		{
			name: "setting nil values",
			cfg:  testConnAddr,
			setter: []io.Writer{
				nil,
			},
		},
		{
			name: "setting multiple nil values",
			cfg:  testConnAddr,
			setter: []io.Writer{
				nil,
				nil,
				nil,
			},
		},
		{
			name: "setting multiple addresses values mixed with nils",
			cfg:  testConnAddr,
			setter: []io.Writer{
				nil,
				address.New(mockAddr[0]),
				nil,
				address.New(mockAddr[0]),
				nil,
			},
		},
		{
			name:   "nil input",
			cfg:    testConnAddr,
			setter: nil,
		},
		{
			name: "setting one invalid writer",
			cfg:  testConnAddr,
			setter: []io.Writer{
				&fs.Logfile{},
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

		if len(keys) != len(mockAddr) {
			errCh <- fmt.Errorf(
				"#%v -- FAILED -- [%s] [%s] output length mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				len(mockAddr),
				len(keys),
				test.name,
			)
			return
		}

		for _, k := range keys {
			var pass bool
			for _, a := range mockAddr {
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
		logger, errCh := New(UnaryRPC())

		// set the input config
		logger.(*GRPCLogClient).addr = test.cfg

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

	for idx, test := range tests {
		verify(idx, test)
	}

}

// E2E
func TestAddOuts(t *testing.T) {
	module := "GRPCLogClient"
	funcname := "AddOuts()"

	_ = module
	_ = funcname

	var mockAddr = []string{
		"127.0.0.1:45063",
		"127.0.0.1:45064",
		"127.0.0.1:45065",
	}

	var mockServerA = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(log.NilConfig)),
	)

	go mockServerA.Serve()
	defer mockServerA.Stop()

	var mockServerB = server.New(
		server.WithAddr(mockAddr[1]),
		server.WithLogger(log.New(log.NilConfig)),
	)

	go mockServerB.Serve()
	defer mockServerB.Stop()

	var mockServerC = server.New(
		server.WithAddr(mockAddr[2]),
		server.WithLogger(log.New(log.NilConfig)),
	)

	go mockServerC.Serve()
	defer mockServerC.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

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
				WithAddr(mockAddr[0]),
				UnaryRPC(),
			},
			setter: []io.Writer{
				address.New(mockAddr[1]),
			},
			wants: []string{
				mockAddr[0],
				mockAddr[1],
			},
		},
		{
			name: "setting the same address already configured",
			cfg: []LogClientConfig{
				WithAddr(mockAddr[0]),
				UnaryRPC(),
			},
			setter: []io.Writer{
				address.New(mockAddr[0]),
			},
			wants: []string{
				mockAddr[0],
			},
		},
		{
			name: "setting multiple valid addresses",
			cfg: []LogClientConfig{
				WithAddr(mockAddr[0]),
				UnaryRPC(),
			},
			setter: []io.Writer{
				address.New(mockAddr[1]),
				address.New(mockAddr[2]),
			},
			wants: mockAddr,
		},
		{
			name: "setting nil values",
			cfg: []LogClientConfig{
				WithAddr(mockAddr[0]),
				UnaryRPC(),
			},
			setter: []io.Writer{
				nil,
			},
			wants: []string{
				mockAddr[0],
			},
		},
		{
			name: "setting multiple nil values",
			cfg: []LogClientConfig{
				WithAddr(mockAddr[0]),
				UnaryRPC(),
			},
			setter: []io.Writer{
				nil,
				nil,
				nil,
			},
			wants: []string{
				mockAddr[0],
			},
		},
		{
			name: "setting multiple addresses values mixed with nils",
			cfg: []LogClientConfig{
				WithAddr(mockAddr[0]),
				UnaryRPC(),
			},
			setter: []io.Writer{
				nil,
				address.New(mockAddr[1]),
				nil,
				address.New(mockAddr[2]),
				nil,
			},
			wants: mockAddr,
		},
		{
			name: "nil input",
			cfg: []LogClientConfig{
				WithAddr(mockAddr[0]),
				UnaryRPC(),
			},
			setter: nil,
			wants: []string{
				mockAddr[0],
			},
		},
		{
			name: "setting one invalid writer",
			cfg: []LogClientConfig{
				WithAddr(mockAddr[0]),
				UnaryRPC(),
			},
			setter: []io.Writer{
				&fs.Logfile{},
			},
			wants: []string{
				mockAddr[0],
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

	for idx, test := range tests {
		verify(idx, test)
	}

}

// E2E
func TestWrite(t *testing.T) {
	module := "GRPCLogClient"
	funcname := "Write()"

	_ = module
	_ = funcname

	var mockAddr = []string{
		"127.0.0.1:45066",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(log.NilConfig)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	type test struct {
		name string
		cfg  []LogClientConfig
		msg  []byte
		ok   bool
	}

	var tests = []test{
		{
			name: "encoded event test",
			cfg:  connConf,
			msg:  event.New().Message("null").Build().Encode(),
			ok:   true,
		},
		{
			name: "byte string test",
			cfg:  connConf,
			msg:  []byte("test"),
			ok:   true,
		},
		{
			name: "zero byte input",
			cfg:  connConf,
			msg:  []byte{},
		},
		{
			name: "nil input",
			cfg:  connConf,
			msg:  nil,
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

func TestLog(t *testing.T) {
	module := "GRPCLogClient"
	funcname := "Log()"

	_ = module
	_ = funcname

	var mockAddr = []string{
		"127.0.0.1:45067",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(log.NilConfig)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	type test struct {
		name string
		e    []*event.Event
	}

	var tests = []test{
		{
			name: "single event test",
			e: []*event.Event{
				event.New().Message("null").Build(),
			},
		},
		{
			name: "multiple event test",
			e: []*event.Event{
				event.New().Message("null_1").Build(),
				event.New().Message("null_2").Build(),
				event.New().Message("null_3").Build(),
			},
		},
		{
			name: "empty event test",
			e:    []*event.Event{},
		},
		{
			name: "nil event test",
			e:    nil,
		},
		{
			name: "multiple events mixed in with nils test",
			e: []*event.Event{
				nil,
				event.New().Message("null_1").Build(),
				nil,
				event.New().Message("null_2").Build(),
				nil,
				event.New().Message("null_3").Build(),
				nil,
			},
		},
	}

	var verifyLoggers = func(idx int, test test, logger GRPCLogger, errCh chan error, done chan struct{}) {
		if test.e == nil {
			logger.Log(nil)
		} else {
			logger.Log(test.e...)
		}

		done <- struct{}{}
	}

	var verify = func(idx int, test test) {
		logger, errCh := New(connConf...)

		done := make(chan struct{})

		go verifyLoggers(idx, test, logger, errCh, done)

		for {
			select {
			case err := <-errCh:
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
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

func FuzzLoggerPrint(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Print()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45068",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Print(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "info" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"info",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerPrintln(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Println()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45069",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Println(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "info" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"info",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerPrintf(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Printf()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45070",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Printf(":%s", a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "info" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"info",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != ":"+a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerPanic(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Panic()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45071",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Panic(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "panic" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"panic",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerPanicln(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Panicln()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45072",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Panicln(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "panic" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"panic",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerPanicf(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Panicf()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45073",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Panicf(":%s", a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "panic" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"panic",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != ":"+a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerFatal(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Fatal()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45074",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Fatal(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "fatal" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"fatal",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerFatalln(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Fatalln()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45075",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Fatalln(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "fatal" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"fatal",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerFatalf(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Fatalf()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45076",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Fatalf(":%s", a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "fatal" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"fatal",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != ":"+a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerError(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Error()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45077",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Error(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "error" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"error",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerErrorln(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Errorln()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45078",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Errorln(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "error" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"error",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerErrorf(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Errorf()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45079",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Errorf(":%s", a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "error" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"error",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != ":"+a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerWarn(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Warn()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45080",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Warn(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "warn" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"warn",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerWarnln(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Warnln()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45081",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Warnln(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "warn" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"warn",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerWarnf(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Warnf()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45082",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Warnf(":%s", a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "warn" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"warn",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != ":"+a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerInfo(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Info()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45083",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Info(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "info" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"info",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerInfoln(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Infoln()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45084",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Infoln(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "info" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"info",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerInfof(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Infof()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45085",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Infof(":%s", a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "info" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"info",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != ":"+a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerDebug(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Debug()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45086",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Debug(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "debug" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"debug",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerDebugln(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Debugln()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45087",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Debugln(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "debug" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"debug",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerDebugf(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Debugf()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45088",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Debugf(":%s", a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "debug" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"debug",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != ":"+a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerTrace(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Trace()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45089",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Trace(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "trace" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"trace",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerTraceln(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Traceln()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45090",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Traceln(a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "trace" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"trace",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}

func FuzzLoggerTracef(f *testing.F) {
	module := "GRPCLogClient"
	funcname := "Tracef()"

	var buf = &bytes.Buffer{}

	var mockAddr = []string{
		"127.0.0.1:45091",
	}
	var mockServer = server.New(
		server.WithAddr(mockAddr[0]),
		server.WithLogger(log.New(
			log.WithOut(buf),
			log.SkipExit,
			log.CfgFormatJSONSkipNewline,
		)),
	)

	var connConf = []LogClientConfig{WithAddr(mockAddr[0])}

	go mockServer.Serve()
	defer mockServer.Stop()

	// sleep to allow server to start up
	time.Sleep(maxWaitTime)

	l, _ := New(connConf...)

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		l.Tracef(":%s", a)

		logE, err := decode(buf, 0)

		if err != nil {
			t.Errorf(
				"FAILED -- [%s] [%s] unexpected unmarshalling error: %v",
				module,
				funcname,
				err,
			)
			return
		}

		if logE.GetLevel().String() != "trace" {
			t.Errorf(
				"FAILED -- [%s] [%s] log level mismatch: wanted %s ; got %s",
				module,
				funcname,
				"trace",
				logE.GetLevel().String(),
			)
			return
		}

		if logE.GetMsg() != ":"+a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				logE.GetMsg(),
			)
			return
		}

	})
}
