package log

import (
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/grpc/address"
	"github.com/zalgonoise/zlog/log/event"
)

var loggers = []Logger{
	New(), New(), New(), New(),
}

func TestMultiLogger(t *testing.T) {
	module := "MultiLogger"
	funcname := "MultiLogger()"

	type test struct {
		name  string
		l     []Logger
		wants Logger
	}

	var tests = []test{
		{
			name:  "empty call",
			l:     []Logger{},
			wants: nil,
		},
		{
			name:  "nil call",
			l:     nil,
			wants: nil,
		},
		{
			name:  "one logger",
			l:     []Logger{loggers[0]},
			wants: loggers[0],
		},
		{
			name: "multiple loggers",
			l:    []Logger{loggers[0], loggers[1], loggers[2]},
			wants: &multiLogger{
				loggers: []Logger{loggers[0], loggers[1], loggers[2]},
			},
		},
		{
			name: "nested multiloggers",
			l:    []Logger{loggers[0], MultiLogger(loggers[1], loggers[2])},
			wants: &multiLogger{
				loggers: []Logger{loggers[0], loggers[1], loggers[2]},
			},
		},
		{
			name: "add nil logger in the mix",
			l:    []Logger{loggers[0], loggers[1], nil},
			wants: &multiLogger{
				loggers: []Logger{loggers[0], loggers[1]},
			},
		},
		{
			name:  "add nil loggers in the mix",
			l:     []Logger{loggers[0], nil, nil},
			wants: loggers[0],
		},
		{
			name:  "add only loggers",
			l:     []Logger{nil, nil, nil},
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
	}

	for idx, test := range tests {
		verify(idx, test)
	}

}

type connAddrTestLogger struct {
	outs []io.Writer
}

func (l *connAddrTestLogger) Write(p []byte) (n int, err error) { return 1, nil }
func (l *connAddrTestLogger) SetOuts(outs ...io.Writer) Logger {
	l.outs = outs
	return l
}
func (l *connAddrTestLogger) AddOuts(outs ...io.Writer) Logger {
	l.outs = append(l.outs, outs...)
	return l
}
func (l *connAddrTestLogger) Prefix(prefix string) Logger                 { return l }
func (l *connAddrTestLogger) Sub(sub string) Logger                       { return l }
func (l *connAddrTestLogger) Fields(fields map[string]interface{}) Logger { return l }
func (l *connAddrTestLogger) IsSkipExit() bool                            { return true }
func (l *connAddrTestLogger) Output(m *event.Event) (n int, err error)    { return 1, nil }
func (l *connAddrTestLogger) Log(m ...*event.Event)                       {}
func (l *connAddrTestLogger) Print(v ...interface{})                      {}
func (l *connAddrTestLogger) Println(v ...interface{})                    {}
func (l *connAddrTestLogger) Printf(format string, v ...interface{})      {}
func (l *connAddrTestLogger) Panic(v ...interface{})                      {}
func (l *connAddrTestLogger) Panicln(v ...interface{})                    {}
func (l *connAddrTestLogger) Panicf(format string, v ...interface{})      {}
func (l *connAddrTestLogger) Fatal(v ...interface{})                      {}
func (l *connAddrTestLogger) Fatalln(v ...interface{})                    {}
func (l *connAddrTestLogger) Fatalf(format string, v ...interface{})      {}
func (l *connAddrTestLogger) Error(v ...interface{})                      {}
func (l *connAddrTestLogger) Errorln(v ...interface{})                    {}
func (l *connAddrTestLogger) Errorf(format string, v ...interface{})      {}
func (l *connAddrTestLogger) Warn(v ...interface{})                       {}
func (l *connAddrTestLogger) Warnln(v ...interface{})                     {}
func (l *connAddrTestLogger) Warnf(format string, v ...interface{})       {}
func (l *connAddrTestLogger) Info(v ...interface{})                       {}
func (l *connAddrTestLogger) Infoln(v ...interface{})                     {}
func (l *connAddrTestLogger) Infof(format string, v ...interface{})       {}
func (l *connAddrTestLogger) Debug(v ...interface{})                      {}
func (l *connAddrTestLogger) Debugln(v ...interface{})                    {}
func (l *connAddrTestLogger) Debugf(format string, v ...interface{})      {}
func (l *connAddrTestLogger) Trace(v ...interface{})                      {}
func (l *connAddrTestLogger) Traceln(v ...interface{})                    {}
func (l *connAddrTestLogger) Tracef(format string, v ...interface{})      {}

func TestMultiLoggerSetOuts(t *testing.T) {
	module := "MultiLogger"
	funcname := "SetOuts()"

	type test struct {
		name string
		w    []io.Writer
	}

	var fakeConnAddr = &connAddrTestLogger{}
	var loggers = []Logger{
		New(), New(), New(), New(),
	}

	var addr = address.New("test")

	var tests = []test{
		{
			name: "empty config",
			w:    []io.Writer{},
		},
		{
			name: "nil config",
			w:    nil,
		},
		{
			name: "one writer",
			w:    []io.Writer{os.Stdout},
		},
		{
			name: "one ConnAddr",
			w:    []io.Writer{addr},
		},
		{
			name: "one nil writer",
			w:    []io.Writer{nil},
		},
	}

	var reset = func() {
		for _, l := range loggers {
			l.SetOuts(os.Stderr)
		}
		fakeConnAddr.outs = []io.Writer{}
	}

	var init = func(test test) *multiLogger {
		ml := MultiLogger(loggers[0], fakeConnAddr)

		ml.SetOuts(test.w...)

		return ml.(*multiLogger)
	}

	var getTestWriter = func(test test) (tw io.Writer, isAddr bool) {
		if len(test.w) == 1 {
			tw = test.w[0]
		}

		if tw != nil {
			_, ok := tw.(*address.ConnAddr)

			if ok {
				isAddr = true
			}
		}

		return
	}

	var verifyLogger = func(l *logger, tw io.Writer) (pass bool) {
		out := l.out

		if tw == nil && !reflect.DeepEqual(out, tw) {
			pass = true
		}

		if tw != nil && !reflect.DeepEqual(out, tw) {
			pass = true
		}

		return pass
	}

	var verifyConnAddr = func(l *connAddrTestLogger, tw io.Writer) (pass bool) {
		if len(l.outs) > 0 {
			out := l.outs[0]

			if tw != nil && out == tw {
				pass = true
			}
		}
		return pass
	}

	var scanMultiLogger = func(ml *multiLogger, tw io.Writer, isAddr bool) (pass bool) {
		for _, logu := range ml.loggers {
			l, ok := logu.(*logger)

			if ok && !isAddr {
				pass = verifyLogger(l, tw)

			} else if !ok && isAddr {
				l, ok := logu.(*connAddrTestLogger)
				if ok {
					pass = verifyConnAddr(l, tw)
				}
			}
		}

		return
	}

	var verify = func(idx int, test test) {
		reset()
		defer reset()

		ml := init(test)

		var tw io.Writer
		var isAddr bool
		var pass bool

		tw, isAddr = getTestWriter(test)

		pass = scanMultiLogger(ml, tw, isAddr)

		if !pass {

			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] test failed -- action: %s",
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

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"os"
// 	"reflect"
// 	"testing"

// 	"github.com/zalgonoise/zlog/grpc/address"
// 	"github.com/zalgonoise/zlog/log/event"
// 	"github.com/zalgonoise/zlog/store"
// )

// var mockMultiPrefixes = []string{
// 	"multilog-test-01",
// 	"multilog-test-02",
// 	"multilog-test-03",
// 	"multilog-test-04",
// 	"multilog-test-05",
// 	"multilog-test-06",
// }

// const msg string = "multilogger test message"

// // add 6 mock buffers
// var mockBufs = []*bytes.Buffer{
// 	{},
// 	{},
// 	{},
// 	{},
// 	{},
// 	{},
// }

// var mockLoggers = []Logger{
// 	New(WithPrefix(mockMultiPrefixes[0]), WithFormat(FormatJSON), WithOut(mockBufs[0])),
// 	New(WithPrefix(mockMultiPrefixes[1]), WithFormat(FormatJSON), WithOut(mockBufs[1])),
// 	New(WithPrefix(mockMultiPrefixes[2]), WithFormat(FormatJSON), WithOut(mockBufs[2])),
// 	New(WithPrefix(mockMultiPrefixes[3]), WithFormat(FormatJSON), WithOut(mockBufs[3])),
// 	New(WithPrefix(mockMultiPrefixes[4]), WithFormat(FormatJSON), WithOut(mockBufs[4])),
// 	New(WithPrefix(mockMultiPrefixes[5]), WithFormat(FormatJSON), WithOut(mockBufs[5])),
// }

// var mockMultiLogger = struct {
// 	log Logger
// 	buf []*bytes.Buffer
// }{
// 	log: MultiLogger(mockLoggers...),
// 	buf: mockBufs,
// }

// func TestNewMultiLogger(t *testing.T) {
// 	type test struct {
// 		input  []Logger
// 		bufs   []*bytes.Buffer
// 		prefix []string
// 		msg    string
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockLoggers); a++ {

// 		var test = test{}
// 		test.msg = msg
// 		for b := 0; b <= a; b++ {
// 			test.input = append(test.input, mockLoggers[b])
// 			test.bufs = append(test.bufs, mockBufs[b])
// 			test.prefix = append(test.prefix, mockMultiPrefixes[b])
// 		}
// 		tests = append(tests, test)

// 	}

// 	var verify = func(id int, test test) {
// 		for bufID, buf := range test.bufs {
// 			logEntry := &event.Event{}

// 			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg,
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg,
// 					test.msg,
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_info.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg,
// 					event.Level_info.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.prefix[bufID] {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- log prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg,
// 					test.prefix,
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg,
// 				test.msg,
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, buf := range test.bufs {
// 			buf.Reset()
// 		}

// 		ml := MultiLogger(test.input...)
// 		ml.Info(test.msg)

// 		verify(id, test)
// 	}
// }

// func TestMultiLoggerOutput(t *testing.T) {

// 	var testAllObjects []map[string]interface{}
// 	testAllObjects = append(testAllObjects, testObjects...)
// 	testAllObjects = append(testAllObjects, testEmptyObjects...)

// 	var testAllMessages []string
// 	testAllMessages = append(testAllMessages, mockMessages...)
// 	for _, fmtMsg := range mockFmtMessages {
// 		testAllMessages = append(testAllMessages, fmt.Sprintf(fmtMsg.format, fmtMsg.v...))
// 	}

// 	var tests []*event.Event

// 	for a := 0; a < len(mockLogLevelsOK); a++ {

// 		for b := 0; b < len(mockPrefixes); b++ {

// 			for c := 0; c < len(testAllMessages); c++ {

// 				for d := 0; d < len(testAllObjects); d++ {
// 					msg := event.New().
// 						Level(mockLogLevelsOK[a]).
// 						Prefix(mockPrefixes[b]).
// 						Message(testAllMessages[c]).
// 						Metadata(testAllObjects[d]).
// 						Build()

// 					tests = append(tests, msg)
// 				}
// 			}
// 		}
// 	}

// 	var verify = func(id int, test *event.Event) {
// 		for bufID, buf := range mockMultiLogger.buf {
// 			logEntry := &event.Event{}

// 			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.GetMsg(),
// 					test.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != test.GetLevel().String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.GetMsg(),
// 					test.GetLevel().String(),
// 					logEntry.GetLevel().String(),
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- log prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.GetMsg(),
// 					test.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.GetMsg(),
// 					test.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.GetMsg(),
// 					test.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.GetMsg(),
// 						len(logEntry.Meta.AsMap()),
// 						len(test.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- %s",
// 				id,
// 				bufID,
// 				test.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, msg := range tests {
// 		for _, buf := range mockMultiLogger.buf {
// 			buf.Reset()
// 		}

// 		mockMultiLogger.log.Output(msg)

// 		verify(id, msg)
// 	}

// }

// func TestMultiLoggerSetOuts(t *testing.T) {
// 	module := "MultiLogger"
// 	funcname := "SetOuts()"

// 	t1logger := New(
// 		WithPrefix("test-new-logger"),
// 		WithFormat(FormatText),
// 		WithOut(mockBufs[5]),
// 	)
// 	t2logger := New(
// 		WithPrefix("test-new-logger-2"),
// 		WithFormat(FormatText),
// 		WithOut(mockBufs[4]),
// 	)
// 	innerML := MultiLogger(t2logger)

// 	nilLogger := New(EmptyConfig)

// 	ml := MultiLogger(t1logger, innerML, nilLogger)

// 	type test struct {
// 		name  string
// 		input []io.Writer
// 		wants io.Writer
// 	}

// 	var tests = []test{
// 		{
// 			name:  "switching to buffer #0",
// 			input: []io.Writer{mockBufs[0]},
// 			wants: io.MultiWriter(mockBufs[0]),
// 		},
// 		{
// 			name:  "switching to multi-buffer #0",
// 			input: []io.Writer{mockBufs[0], mockBufs[1], mockBufs[3]},
// 			wants: io.MultiWriter(mockBufs[0], mockBufs[1], mockBufs[3]),
// 		},
// 		{
// 			name:  "ConnAddr flow test",
// 			input: []io.Writer{mockBufs[0], &address.ConnAddr{}},
// 			wants: io.MultiWriter(mockBufs[0]),
// 		},
// 		{
// 			name:  "switching to default writer with zero arguments",
// 			input: nil,
// 			wants: bufs[3],
// 		},
// 		{
// 			name:  "switching to default writer with nil writers",
// 			input: []io.Writer{nil, nil, nil},
// 			wants: bufs[3],
// 		},
// 		{
// 			name:  "ensure the empty writer works",
// 			input: []io.Writer{store.EmptyWriter},
// 			wants: io.MultiWriter(store.EmptyWriter),
// 		},
// 	}

// 	var verify = func(id int, logw, w io.Writer, action string) {
// 		if !reflect.DeepEqual(logw, w) {
// 			t.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] writer mismatch: wanted %v ; got %v -- action: %s",
// 				id,
// 				module,
// 				funcname,
// 				w,
// 				logw,
// 				action,
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [%s] [%s]",
// 			id,
// 			module,
// 			funcname,
// 		)
// 	}

// 	for id, test := range tests {
// 		if test.input != nil {
// 			ml.SetOuts(test.input...)
// 		} else {
// 			ml.SetOuts()
// 		}

// 		for _, l := range ml.(*multiLogger).loggers {
// 			if _, ok := l.(*logger); ok {
// 				logw := l.(*logger).out
// 				verify(id, logw, test.wants, test.name)
// 			}
// 		}

// 	}

// }

// func TestMultiLoggerAddOuts(t *testing.T) {

// 	module := "MultiLogger"
// 	funcname := "AddOuts()"

// 	t1logger := New(
// 		WithPrefix("test-new-logger"),
// 		WithFormat(FormatText),
// 		WithOut(mockBufs[5]),
// 	)
// 	t2logger := New(
// 		WithPrefix("test-new-logger-2"),
// 		WithFormat(FormatText),
// 		WithOut(mockBufs[4]),
// 	)
// 	innerML := MultiLogger(t2logger)

// 	nilLogger := New(EmptyConfig)

// 	ml := MultiLogger(t1logger, innerML, nilLogger)

// 	type test struct {
// 		name  string
// 		input []io.Writer
// 		wants io.Writer
// 	}

// 	var tests = []test{
// 		{
// 			name:  "adding buffer #0",
// 			input: []io.Writer{mockBufs[0]},
// 			wants: io.MultiWriter(mockBufs[0], mockBufs[5]),
// 		},
// 		{
// 			name:  "adding multi-buffer #0",
// 			input: []io.Writer{mockBufs[0], mockBufs[1], mockBufs[3]},
// 			wants: io.MultiWriter(mockBufs[0], mockBufs[1], mockBufs[3], mockBufs[5]),
// 		},
// 		{
// 			name:  "ConnAddr flow test",
// 			input: []io.Writer{mockBufs[0], &address.ConnAddr{}},
// 			wants: io.MultiWriter(mockBufs[0], mockBufs[5]),
// 		},
// 		{
// 			name:  "adding default writer with zero arguments",
// 			input: nil,
// 			wants: io.MultiWriter(mockBufs[5]),
// 		},
// 		{
// 			name:  "adding default writer with nil writers",
// 			input: []io.Writer{nil, nil, nil},
// 			wants: io.MultiWriter(mockBufs[5]),
// 		},
// 		{
// 			name:  "ensure the empty writer works",
// 			input: []io.Writer{store.EmptyWriter},
// 			wants: io.MultiWriter(store.EmptyWriter, mockBufs[5]),
// 		},
// 	}

// 	var verify = func(id int, logw, w io.Writer, action string) {
// 		if !reflect.DeepEqual(logw, w) {
// 			t.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] writer mismatch: wanted %v ; got %v -- action: %s",
// 				id,
// 				module,
// 				funcname,
// 				w,
// 				logw,
// 				action,
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [%s] [%s]",
// 			id,
// 			module,
// 			funcname,
// 		)
// 	}

// 	for id, test := range tests {
// 		if test.input != nil {
// 			ml.AddOuts(test.input...)
// 		} else {
// 			ml.AddOuts()
// 		}

// 		for _, l := range ml.(*multiLogger).loggers {
// 			if _, ok := l.(*logger); ok {
// 				logw := l.(*logger).out
// 				verify(id, logw, test.wants, test.name)
// 			}
// 		}
// 		// reset
// 		ml.SetOuts(mockBufs[5])

// 	}
// }

// func TestMultiLoggerSub(t *testing.T) {

// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
// 		sub string
// 	}

// 	var newSubPrefixes = []string{
// 		"Prefix()",
// 		"new prefix",
// 		"awesome service",
// 		"alert!!",
// 		"@whatever",
// 		"01101001101",
// 	}

// 	var testAllObjects []map[string]interface{}
// 	testAllObjects = append(testAllObjects, testObjects...)
// 	testAllObjects = append(testAllObjects, testEmptyObjects...)

// 	var testAllMessages []string
// 	testAllMessages = append(testAllMessages, mockMessages...)
// 	for _, fmtMsg := range mockFmtMessages {
// 		testAllMessages = append(testAllMessages, fmt.Sprintf(fmtMsg.format, fmtMsg.v...))
// 	}

// 	var tests []test

// 	for a := 0; a < len(newSubPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for d := 0; d < len(newSubPrefixes); d++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix("log"), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					sub: newSubPrefixes[a],
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Message(testAllMessages[b]).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 				}

// 				tests = append(tests, obj)
// 			}

// 		}

// 	}

// 	var verify = func(id int, test test) {
// 		defer func() {
// 			for _, b := range test.ml.buf {
// 				b.Reset()
// 			}
// 		}()

// 		for bufID, buf := range test.ml.buf {
// 			logEntry := &event.Event{}

// 			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.sub,
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.sub,
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetSub() != test.sub {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- sub-prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.sub,
// 					test.sub,
// 					logEntry.GetSub(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_info.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.sub,
// 					event.Level_info.String(),
// 					logEntry.GetLevel().String(),
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.sub,
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.sub,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.sub,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.sub,
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.sub,
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- %s",
// 				id,
// 				bufID,
// 				test.sub,
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}
// 		test.ml.log.Sub(test.sub).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Info(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerFields(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg    *event.Event
// 		ml     ml
// 		prefix string
// 	}

// 	var newPrefixes = []string{
// 		"Prefix()",
// 		"new prefix",
// 		"awesome service",
// 		"alert!!",
// 		"@whatever",
// 		"01101001101",
// 	}

// 	var testAllObjects []map[string]interface{}
// 	testAllObjects = append(testAllObjects, testObjects...)
// 	testAllObjects = append(testAllObjects, testEmptyObjects...)

// 	var testAllMessages []string
// 	testAllMessages = append(testAllMessages, mockMessages...)
// 	for _, fmtMsg := range mockFmtMessages {
// 		testAllMessages = append(testAllMessages, fmt.Sprintf(fmtMsg.format, fmtMsg.v...))
// 	}

// 	var tests []test

// 	for a := 0; a < len(newPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for d := 0; d < len(mockMultiPrefixes); d++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[d]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					prefix: newPrefixes[a],
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Message(testAllMessages[b]).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 				}

// 				tests = append(tests, obj)
// 			}

// 		}

// 	}

// 	var verify = func(id int, test test) {
// 		defer func() {
// 			for _, b := range test.ml.buf {
// 				b.Reset()
// 			}
// 		}()

// 		for bufID, buf := range test.ml.buf {
// 			logEntry := &event.Event{}

// 			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.prefix {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.prefix,
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_info.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					event.Level_info.String(),
// 					logEntry.GetLevel().String(),
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- %s",
// 				id,
// 				bufID,
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}
// 		test.ml.log.Prefix(test.prefix).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Info(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerWrite(t *testing.T) {
// 	type test struct {
// 		msg    []byte
// 		prefix string
// 		sub    string
// 		level  event.Level
// 		body   string
// 	}

// 	var tests = []test{
// 		{
// 			msg:    event.New().Level(event.Level_info).Prefix("test").Sub("tester").Message("write test").Build().Encode(),
// 			prefix: "test",
// 			sub:    "tester",
// 			level:  event.Level_info,
// 			body:   "write test",
// 		},
// 		{
// 			msg:    []byte("hello world"),
// 			prefix: "log",
// 			sub:    "",
// 			level:  event.Level_info,
// 			body:   "hello world",
// 		},
// 	}

// 	bufs := []*bytes.Buffer{{}, {}, {}}

// 	logger := MultiLogger(
// 		New(WithFormat(FormatJSON), WithOut(bufs[0])),
// 		New(WithFormat(FormatJSON), WithOut(bufs[1])),
// 		New(WithFormat(FormatJSON), WithOut(bufs[2])),
// 	)
// 	var verify = func(id int, test test) {

// 		for bid, buffer := range bufs {

// 			buf := buffer.Bytes()

// 			if len(buf) <= 0 {
// 				t.Errorf(
// 					"#%v [Logger] -- FAILED -- [MultiLogger] Write([]byte) [buffer #%v] -- empty buffer error: %v bytes written",
// 					id,
// 					bid,
// 					len(buf),
// 				)
// 				return
// 			}

// 			logEntry := &event.Event{}

// 			err := json.Unmarshal(buf, logEntry)
// 			if err != nil {
// 				t.Errorf(
// 					"#%v [Logger] -- FAILED -- [MultiLogger] Write([]byte) [buffer #%v] -- JSON decoding error: %s ; buf: %s",
// 					id,
// 					bid,
// 					err,
// 					string(buf),
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.prefix {
// 				t.Errorf(
// 					"#%v [Logger] -- FAILED -- [MultiLogger] Write([]byte) [buffer #%v] -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bid,
// 					logEntry.GetPrefix(),
// 					test.prefix,
// 				)
// 				return
// 			}

// 			if logEntry.GetSub() != test.sub {
// 				t.Errorf(
// 					"#%v [Logger] -- FAILED -- [MultiLogger] Write([]byte) [buffer #%v] -- sub-prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bid,
// 					logEntry.GetSub(),
// 					test.sub,
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != test.level.String() {
// 				t.Errorf(
// 					"#%v [Logger] -- FAILED -- [MultiLogger] Write([]byte) [buffer #%v] -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bid,
// 					logEntry.Level.String(),
// 					test.level.String(),
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.body {
// 				t.Errorf(
// 					"#%v [Logger] -- FAILED -- [MultiLogger] Write([]byte) [buffer #%v] -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bid,
// 					logEntry.GetMsg(),
// 					test.body,
// 				)
// 				return
// 			}

// 			t.Logf(
// 				"#%v [Logger] -- PASSED -- [MultiLogger] Write([]byte) [buffer #%v]",
// 				id,
// 				bid,
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range bufs {
// 			b.Reset()
// 		}
// 		n, err := logger.Write(test.msg)

// 		if err != nil {
// 			t.Errorf(
// 				"#%v [Logger] -- FAILED -- Write([]byte) -- write error: %s",
// 				id,
// 				err,
// 			)
// 		}

// 		if n <= 0 {
// 			t.Errorf(
// 				"#%v [Logger] -- FAILED -- Write([]byte) -- no bytes written: %v",
// 				id,
// 				n,
// 			)
// 		}

// 		verify(id, test)

// 		for _, b := range bufs {
// 			b.Reset()
// 		}
// 	}

// 	// failing tests:
// 	tmpf, err := os.Create(`tmp.log`)
// 	if err != nil {
// 		t.Errorf(
// 			"#%v [Logger] -- FAILED -- Write([]byte) -- failed to create temp file: %s",
// 			0,
// 			err,
// 		)
// 	}
// 	tmpf.Close()
// 	defer os.RemoveAll(`tmp.log`)

// 	closedBuf, err := os.OpenFile(`tmp.log`, os.O_RDONLY, 0o000)
// 	if err != nil {
// 		t.Errorf(
// 			"#%v [Logger] -- FAILED -- Write([]byte) -- failed to open temp file: %s",
// 			0,
// 			err,
// 		)
// 	}
// 	logger.SetOuts(closedBuf)

// 	for id, test := range tests {
// 		n, err := logger.Write(test.msg)
// 		if err == nil && n <= 0 {
// 			t.Errorf(
// 				"#%v [Logger] -- FAILED -- Write([]byte) -- write succeeded when it shouldn't",
// 				id,
// 			)
// 		}
// 		t.Logf(
// 			"#%v [Logger] -- PASSED -- Write([]byte) -- write failed as expected: error: %s ; bytes written: %v",
// 			id,
// 			err,
// 			n,
// 		)
// 	}

// }
