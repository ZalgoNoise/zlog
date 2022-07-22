package client

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/grpc/address"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

var errTestLongWrite error = errors.New("test: message is too long")

type testLogClient struct {
	skipExit bool
}

// ChanneledLogger impl
func (l *testLogClient) Close() {}
func (l *testLogClient) Channels() (chan *event.Event, chan struct{}) {
	return make(chan *event.Event), make(chan struct{})
}

// io.Writer impl
func (l *testLogClient) Write(p []byte) (n int, err error) {
	// returning an error
	if len(p) > 100 {
		return -1, errTestLongWrite
	}

	// returning zero bytes written
	if len(p) == 0 {
		return 0, nil
	}

	return 1, nil
}

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
func (l *testLogClient) Output(m *event.Event) (n int, err error) {
	return l.Write(m.Encode())
}
func (l *testLogClient) Log(m ...*event.Event)                  {}
func (l *testLogClient) Print(v ...interface{})                 {}
func (l *testLogClient) Println(v ...interface{})               {}
func (l *testLogClient) Printf(format string, v ...interface{}) {}
func (l *testLogClient) Panic(v ...interface{})                 {}
func (l *testLogClient) Panicln(v ...interface{})               {}
func (l *testLogClient) Panicf(format string, v ...interface{}) {}
func (l *testLogClient) Fatal(v ...interface{})                 {}
func (l *testLogClient) Fatalln(v ...interface{})               {}
func (l *testLogClient) Fatalf(format string, v ...interface{}) {}
func (l *testLogClient) Error(v ...interface{})                 {}
func (l *testLogClient) Errorln(v ...interface{})               {}
func (l *testLogClient) Errorf(format string, v ...interface{}) {}
func (l *testLogClient) Warn(v ...interface{})                  {}
func (l *testLogClient) Warnln(v ...interface{})                {}
func (l *testLogClient) Warnf(format string, v ...interface{})  {}
func (l *testLogClient) Info(v ...interface{})                  {}
func (l *testLogClient) Infoln(v ...interface{})                {}
func (l *testLogClient) Infof(format string, v ...interface{})  {}
func (l *testLogClient) Debug(v ...interface{})                 {}
func (l *testLogClient) Debugln(v ...interface{})               {}
func (l *testLogClient) Debugf(format string, v ...interface{}) {}
func (l *testLogClient) Trace(v ...interface{})                 {}
func (l *testLogClient) Traceln(v ...interface{})               {}
func (l *testLogClient) Tracef(format string, v ...interface{}) {}

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

func TestMultiLoggerWrite(t *testing.T) {
	module := "GRPCLogger"              // module
	funcname := "MultiLogger().Write()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
		msg     []byte
		isErr   bool
	}

	var tests = []test{
		{
			name: "OK write",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
			msg: []byte("short message"),
		},
		{
			name: "not OK write",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
			msg:   []byte("very long message that will surely overflow the set limit of one hundred characters in this test interface"),
			isErr: true,
		},
		{
			name: "zero-bytes write",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
			msg:   []byte(""),
			isErr: true,
		},
		{
			name: "single-error write",
			loggers: MultiLogger(
				&testLogClient{},
				NilClient(),
			),
			msg:   []byte("very long message that will surely overflow the set limit of one hundred characters in this test interface"),
			isErr: true,
		},
	}

	var verify = func(idx int, test test) {
		_, err := test.loggers.Write(test.msg)

		if err != nil && !test.isErr {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerClose(t *testing.T) {
	_ = "GRPCLogger"            // module
	_ = "MultiLogger().Close()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Close() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Close()
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerChannels(t *testing.T) {
	_ = "GRPCLogger"               // module
	_ = "MultiLogger().Channels()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Close() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		log, done := test.loggers.Channels()

		// cover goroutine activity
		log <- event.New().Message("null").Build()
		done <- struct{}{}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerOutput(t *testing.T) {
	module := "GRPCLogger"               // module
	funcname := "MultiLogger().Output()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
		msg     *event.Event
		isErr   bool
	}

	var tests = []test{
		{
			name: "OK write",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
			msg: event.New().Message("short").Build(),
		},
		{
			name: "not OK write",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
			msg: event.New().Message(
				"very long message that will surely overflow the set limit of one hundred characters in this test interface",
			).Build(),
			isErr: true,
		},
	}

	var verify = func(idx int, test test) {
		_, err := test.loggers.Output(test.msg)

		if err != nil && !test.isErr {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerPrint(t *testing.T) {
	_ = "GRPCLogger"            // module
	_ = "MultiLogger().Print()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Print() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Print("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerPrintln(t *testing.T) {
	_ = "GRPCLogger"              // module
	_ = "MultiLogger().Println()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Println() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Println("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerPrintf(t *testing.T) {
	_ = "GRPCLogger"             // module
	_ = "MultiLogger().Printf()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Printf() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Printf("%s", "null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerLog(t *testing.T) {
	_ = "GRPCLogger"          // module
	_ = "MultiLogger().Log()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
		e       []*event.Event
	}

	var tests = []test{
		{
			name: "Log() method call -- one event",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
			e: []*event.Event{
				event.New().Message("null").Build(),
			},
		},
		{
			name: "Log() method call -- multiple events",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
			e: []*event.Event{
				event.New().Message("null").Build(),
				event.New().Message("null").Build(),
				event.New().Message("null").Build(),
			},
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Log(test.e...)
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerPanic(t *testing.T) {
	_ = "GRPCLogger"            // module
	_ = "MultiLogger().Panic()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Panic() method call",
			loggers: MultiLogger(
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: true},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Panic("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerPanicln(t *testing.T) {
	_ = "GRPCLogger"              // module
	_ = "MultiLogger().Panicln()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Panicln() method call",
			loggers: MultiLogger(
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: true},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Panicln("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerPanicf(t *testing.T) {
	_ = "GRPCLogger"             // module
	_ = "MultiLogger().Panicf()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Panicf() method call",
			loggers: MultiLogger(
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: true},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Panicf("%s", "null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerFatal(t *testing.T) {
	_ = "GRPCLogger"            // module
	_ = "MultiLogger().Fatal()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Fatal() method call",
			loggers: MultiLogger(
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: true},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Fatal("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerFatalln(t *testing.T) {
	_ = "GRPCLogger"              // module
	_ = "MultiLogger().Fatalln()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Fatalln() method call",
			loggers: MultiLogger(
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: true},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Fatalln("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerFatalf(t *testing.T) {
	_ = "GRPCLogger"             // module
	_ = "MultiLogger().Fatalf()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Fatalf() method call",
			loggers: MultiLogger(
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: true},
				&testLogClient{skipExit: true},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Fatalf("%s", "null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerError(t *testing.T) {
	_ = "GRPCLogger"            // module
	_ = "MultiLogger().Error()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Error() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Error("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerErrorln(t *testing.T) {
	_ = "GRPCLogger"              // module
	_ = "MultiLogger().Errorln()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Errorln() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Errorln("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerErrorf(t *testing.T) {
	_ = "GRPCLogger"             // module
	_ = "MultiLogger().Errorf()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Errorf() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Errorf("%s", "null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerWarn(t *testing.T) {
	_ = "GRPCLogger"           // module
	_ = "MultiLogger().Warn()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Warn() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Warn("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerWarnln(t *testing.T) {
	_ = "GRPCLogger"             // module
	_ = "MultiLogger().Warnln()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Warnln() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Warnln("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerWarnf(t *testing.T) {
	_ = "GRPCLogger"            // module
	_ = "MultiLogger().Warnf()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Warnf() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Warnf("%s", "null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerInfo(t *testing.T) {
	_ = "GRPCLogger"           // module
	_ = "MultiLogger().Info()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Info() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Info("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerInfoln(t *testing.T) {
	_ = "GRPCLogger"             // module
	_ = "MultiLogger().Infoln()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Infoln() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Infoln("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerInfof(t *testing.T) {
	_ = "GRPCLogger"            // module
	_ = "MultiLogger().Infof()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Infof() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Infof("%s", "null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerDebug(t *testing.T) {
	_ = "GRPCLogger"            // module
	_ = "MultiLogger().Debug()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Debug() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Debug("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerDebugln(t *testing.T) {
	_ = "GRPCLogger"              // module
	_ = "MultiLogger().Debugln()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Debugln() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Debugln("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerDebugf(t *testing.T) {
	_ = "GRPCLogger"             // module
	_ = "MultiLogger().Debugf()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Debugf() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Debugf("%s", "null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerTrace(t *testing.T) {
	_ = "GRPCLogger"            // module
	_ = "MultiLogger().Trace()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Trace() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Trace("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerTraceln(t *testing.T) {
	_ = "GRPCLogger"              // module
	_ = "MultiLogger().Traceln()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Traceln() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Traceln("null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestMultiLoggerTracef(t *testing.T) {
	_ = "GRPCLogger"             // module
	_ = "MultiLogger().Tracef()" // funcname

	type test struct {
		name    string
		loggers GRPCLogger
	}

	var tests = []test{
		{
			name: "Tracef() method call",
			loggers: MultiLogger(
				&testLogClient{},
				&testLogClient{},
				&testLogClient{},
			),
		},
	}

	var verify = func(idx int, test test) {
		test.loggers.Tracef("%s", "null")
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
