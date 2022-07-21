package client

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/grpc/address"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

type testLogClient struct {
	skipExit bool
}

// ChanneledLogger impl
func (l *testLogClient) Close()                                       {}
func (l *testLogClient) Channels() (chan *event.Event, chan struct{}) { return nil, nil }

// io.Writer impl
func (l *testLogClient) Write(p []byte) (n int, err error) { return 1, nil }

// log.Logger impl
func (l *testLogClient) SetOuts(outs ...io.Writer) log.Logger {
	if len(outs) == 0 {
		return nil
	}

	return l
}
func (l *testLogClient) AddOuts(outs ...io.Writer) log.Logger {
	if len(outs) == 0 {
		return nil
	}

	return l
}
func (l *testLogClient) Prefix(prefix string) log.Logger                 { return l }
func (l *testLogClient) Sub(sub string) log.Logger                       { return l }
func (l *testLogClient) Fields(fields map[string]interface{}) log.Logger { return l }
func (l *testLogClient) IsSkipExit() bool                                { return l.skipExit }

// log.Printer impl
func (l *testLogClient) Output(m *event.Event) (n int, err error) { return 1, nil }
func (l *testLogClient) Log(m ...*event.Event)                    {}
func (l *testLogClient) Print(v ...interface{})                   {}
func (l *testLogClient) Println(v ...interface{})                 {}
func (l *testLogClient) Printf(format string, v ...interface{})   {}
func (l *testLogClient) Panic(v ...interface{})                   {}
func (l *testLogClient) Panicln(v ...interface{})                 {}
func (l *testLogClient) Panicf(format string, v ...interface{})   {}
func (l *testLogClient) Fatal(v ...interface{})                   {}
func (l *testLogClient) Fatalln(v ...interface{})                 {}
func (l *testLogClient) Fatalf(format string, v ...interface{})   {}
func (l *testLogClient) Error(v ...interface{})                   {}
func (l *testLogClient) Errorln(v ...interface{})                 {}
func (l *testLogClient) Errorf(format string, v ...interface{})   {}
func (l *testLogClient) Warn(v ...interface{})                    {}
func (l *testLogClient) Warnln(v ...interface{})                  {}
func (l *testLogClient) Warnf(format string, v ...interface{})    {}
func (l *testLogClient) Info(v ...interface{})                    {}
func (l *testLogClient) Infoln(v ...interface{})                  {}
func (l *testLogClient) Infof(format string, v ...interface{})    {}
func (l *testLogClient) Debug(v ...interface{})                   {}
func (l *testLogClient) Debugln(v ...interface{})                 {}
func (l *testLogClient) Debugf(format string, v ...interface{})   {}
func (l *testLogClient) Trace(v ...interface{})                   {}
func (l *testLogClient) Traceln(v ...interface{})                 {}
func (l *testLogClient) Tracef(format string, v ...interface{})   {}

func TestMultiLogger(t *testing.T) {
	module := "GRPCLogger"
	funcname := "MultiLogger()"

	type test struct {
		name  string
		l     []GRPCLogger
		wants GRPCLogger
	}

	var loggers = []GRPCLogger{
		NilClient(),
		NilClient(),
		NilClient(),
		NilClient(),
	}

	var tests = []test{
		{
			name:  "empty call",
			l:     []GRPCLogger{},
			wants: nil,
		},
		{
			name:  "nil call",
			l:     nil,
			wants: nil,
		},
		{
			name:  "one logger",
			l:     []GRPCLogger{loggers[0]},
			wants: loggers[0],
		},
		{
			name: "multiple loggers",
			l:    []GRPCLogger{loggers[0], loggers[1], loggers[2]},
			wants: &multiLogger{
				loggers: []GRPCLogger{loggers[0], loggers[1], loggers[2]},
			},
		},
		{
			name: "nested multiloggers",
			l:    []GRPCLogger{loggers[0], MultiLogger(loggers[1], loggers[2])},
			wants: &multiLogger{
				loggers: []GRPCLogger{loggers[0], loggers[1], loggers[2]},
			},
		},
		{
			name: "add nil logger in the mix",
			l:    []GRPCLogger{loggers[0], loggers[1], nil},
			wants: &multiLogger{
				loggers: []GRPCLogger{loggers[0], loggers[1]},
			},
		},
		{
			name:  "add nil loggers in the mix",
			l:     []GRPCLogger{loggers[0], nil, nil},
			wants: loggers[0],
		},
		{
			name:  "add only nil loggers",
			l:     []GRPCLogger{nil, nil, nil},
			wants: nil,
		},
	}

	var verify = func(idx int, test test) {
		ml := MultiLogger(test.l...)

		if !reflect.DeepEqual(ml, test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				ml,
				test.name,
			)
		}

		// test interface calls --
		// TODO (zalgonoise): add individual tests for these
		if ml != nil {

			// log.Logger impl
			ml.SetOuts(&bytes.Buffer{})
			ml.AddOuts(&bytes.Buffer{})
			ml.Prefix("null")
			ml.Sub("null")
			ml.Fields(map[string]interface{}{"ok": true})
			ml.IsSkipExit()

			// ChanneledLogger impl
			ml.Close()
			ml.Channels()

			// io.Writer impl
			ml.Write(event.New().Message("null").Build().Encode())

			// log.Printer impl
			ml.Output(event.New().Message("null").Build())
			ml.Log(event.New().Message("null").Build())
			ml.Print("null")
			ml.Println("null")
			ml.Printf("null")
			ml.Panic("null")
			ml.Panicln("null")
			ml.Panicf("null")
			ml.Fatal("null")
			ml.Fatalln("null")
			ml.Fatalf("null")
			ml.Error("null")
			ml.Errorln("null")
			ml.Errorf("null")
			ml.Warn("null")
			ml.Warnln("null")
			ml.Warnf("null")
			ml.Info("null")
			ml.Infoln("null")
			ml.Infof("null")
			ml.Debug("null")
			ml.Debugln("null")
			ml.Debugf("null")
			ml.Trace("null")
			ml.Traceln("null")
			ml.Tracef("null")
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerSetOuts(t *testing.T) {
	module := "GRPCLogger"                // module
	funcname := "MultiLogger().SetOuts()" // funcname

	type test struct {
		name  string
		input []io.Writer
		isNil bool
	}

	ml := MultiLogger(
		&testLogClient{},
		&testLogClient{},
		&testLogClient{},
		&testLogClient{},
	)

	var tests = []test{
		{
			name:  "empty call",
			input: []io.Writer{},
			isNil: true,
		},
		{
			name:  "nil call",
			input: nil,
			isNil: true,
		},
		{
			name: "one ConnAddr",
			input: []io.Writer{
				address.New("localhost:9099"),
			},
		},
		{
			name: "one multi-addr ConnAddr",
			input: []io.Writer{
				address.New(
					"localhost:9097",
					"localhost:9098",
					"localhost:9099",
				),
			},
		},
		{
			name: "multiple ConnAddr",
			input: []io.Writer{
				address.New("localhost:9097"),
				address.New("localhost:9098"),
				address.New("localhost:9099"),
			},
		},
		{
			name: "multiple ConnAddr w/ nils",
			input: []io.Writer{
				address.New("localhost:9097"),
				nil,
				address.New("localhost:9098"),
				nil,
				address.New("localhost:9099"),
			},
		},
		{
			name: "only nils",
			input: []io.Writer{
				nil,
				nil,
				nil,
			},
			isNil: true,
		},
		{
			name: "invalid writers",
			input: []io.Writer{
				new(bytes.Buffer),
				new(bytes.Buffer),
			},
			isNil: true,
		},
	}

	var verify = func(idx int, test test) {
		// use a copy of the multilogger
		logger := ml

		r := logger.SetOuts(test.input...)

		if r == nil && !test.isNil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected nil output -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerAddOuts(t *testing.T) {
	module := "GRPCLogger"                // module
	funcname := "MultiLogger().AddOuts()" // funcname

	type test struct {
		name  string
		input []io.Writer
		isNil bool
	}

	ml := MultiLogger(
		&testLogClient{},
		&testLogClient{},
		&testLogClient{},
		&testLogClient{},
	)

	var tests = []test{
		{
			name:  "empty call",
			input: []io.Writer{},
			isNil: true,
		},
		{
			name:  "nil call",
			input: nil,
			isNil: true,
		},
		{
			name: "one ConnAddr",
			input: []io.Writer{
				address.New("localhost:9099"),
			},
		},
		{
			name: "one multi-addr ConnAddr",
			input: []io.Writer{
				address.New(
					"localhost:9097",
					"localhost:9098",
					"localhost:9099",
				),
			},
		},
		{
			name: "multiple ConnAddr",
			input: []io.Writer{
				address.New("localhost:9097"),
				address.New("localhost:9098"),
				address.New("localhost:9099"),
			},
		},
		{
			name: "multiple ConnAddr w/ nils",
			input: []io.Writer{
				address.New("localhost:9097"),
				nil,
				address.New("localhost:9098"),
				nil,
				address.New("localhost:9099"),
			},
		},
		{
			name: "only nils",
			input: []io.Writer{
				nil,
				nil,
				nil,
			},
			isNil: true,
		},
		{
			name: "invalid writers",
			input: []io.Writer{
				new(bytes.Buffer),
				new(bytes.Buffer),
			},
			isNil: true,
		},
	}

	var verify = func(idx int, test test) {
		// use a copy of the multilogger
		logger := ml

		r := logger.AddOuts(test.input...)

		if r == nil && !test.isNil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected nil output -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerPrefix(t *testing.T) {
	_ = "GRPCLogger"             // module
	_ = "MultiLogger().Prefix()" // funcname

	type test struct {
		name string
		p    string
	}

	ml := MultiLogger(
		NilClient(),
		NilClient(),
		NilClient(),
		NilClient(),
	)

	var tests = []test{
		{
			name: "no input",
			p:    "",
		},
		{
			name: "any input",
			p:    "something",
		},
	}

	var verify = func(idx int, test test) {
		// use a copy of the multilogger
		logger := ml

		logger.Prefix(test.p)
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerSub(t *testing.T) {
	_ = "GRPCLogger"          // module
	_ = "MultiLogger().Sub()" // funcname

	type test struct {
		name string
		s    string
	}

	ml := MultiLogger(
		NilClient(),
		NilClient(),
		NilClient(),
		NilClient(),
	)

	var tests = []test{
		{
			name: "no input",
			s:    "",
		},
		{
			name: "any input",
			s:    "something",
		},
	}

	var verify = func(idx int, test test) {
		// use a copy of the multilogger
		logger := ml

		logger.Sub(test.s)
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerFields(t *testing.T) {
	_ = "GRPCLogger"             // module
	_ = "MultiLogger().Fields()" // funcname

	type test struct {
		name string
		f    map[string]interface{}
	}

	ml := MultiLogger(
		NilClient(),
		NilClient(),
		NilClient(),
		NilClient(),
	)

	var tests = []test{
		{
			name: "no input",
			f:    map[string]interface{}{},
		},
		{
			name: "nil input",
			f:    nil,
		},
		{
			name: "any input",
			f: map[string]interface{}{
				"ok": true,
			},
		},
	}

	var verify = func(idx int, test test) {
		// use a copy of the multilogger
		logger := ml

		logger.Fields(test.f)
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerIsSkipExit(t *testing.T) {
	module := "GRPCLogger"                   // module
	funcname := "MultiLogger().IsSkipExit()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
		wants   bool
	}

	var tests = []test{
		{
			name: "all skip-exit loggers",
			loggers: MultiLogger(
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: true},
			),
			wants: true,
		},
		{
			name: "all no-skip-exit loggers",
			loggers: MultiLogger(
				&testLogClient{skipExit: false},
				&testLogClient{skipExit: false},
				&testLogClient{skipExit: false},
			),
			wants: false,
		},
		{
			name: "one no-skip-exit logger",
			loggers: MultiLogger(
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: false},
			),
			wants: false,
		},
	}

	var verify = func(idx int, test test) {
		if test.loggers.IsSkipExit() != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				test.loggers.IsSkipExit(),
				test.name,
			)
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
