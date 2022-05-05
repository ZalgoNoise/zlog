package log

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/log/format/text"
)

var (
	testErrFormat error = errors.New("failed to format message")
	testErrWrite  error = errors.New("failed to write to output")
)

type testLogger struct {
	outs []io.Writer
}

func (l *testLogger) Write(p []byte) (n int, err error) {
	if len(p) < 10 {
		return
	}

	if len(p) > 40 {
		n = len(p)
		err = testErrWrite

		return
	}

	n = len(p)
	return
}
func (l *testLogger) SetOuts(outs ...io.Writer) Logger {
	l.outs = outs
	return l
}
func (l *testLogger) AddOuts(outs ...io.Writer) Logger {
	l.outs = append(l.outs, outs...)
	return l
}
func (l *testLogger) Prefix(prefix string) Logger                 { return l }
func (l *testLogger) Sub(sub string) Logger                       { return l }
func (l *testLogger) Fields(fields map[string]interface{}) Logger { return l }
func (l *testLogger) IsSkipExit() bool                            { return true }
func (l *testLogger) Output(m *event.Event) (n int, err error)    { return l.Write(m.Encode()) }
func (l *testLogger) Log(m ...*event.Event)                       {}
func (l *testLogger) Print(v ...interface{})                      {}
func (l *testLogger) Println(v ...interface{})                    {}
func (l *testLogger) Printf(format string, v ...interface{})      {}
func (l *testLogger) Panic(v ...interface{})                      {}
func (l *testLogger) Panicln(v ...interface{})                    {}
func (l *testLogger) Panicf(format string, v ...interface{})      {}
func (l *testLogger) Fatal(v ...interface{})                      {}
func (l *testLogger) Fatalln(v ...interface{})                    {}
func (l *testLogger) Fatalf(format string, v ...interface{})      {}
func (l *testLogger) Error(v ...interface{})                      {}
func (l *testLogger) Errorln(v ...interface{})                    {}
func (l *testLogger) Errorf(format string, v ...interface{})      {}
func (l *testLogger) Warn(v ...interface{})                       {}
func (l *testLogger) Warnln(v ...interface{})                     {}
func (l *testLogger) Warnf(format string, v ...interface{})       {}
func (l *testLogger) Info(v ...interface{})                       {}
func (l *testLogger) Infoln(v ...interface{})                     {}
func (l *testLogger) Infof(format string, v ...interface{})       {}
func (l *testLogger) Debug(v ...interface{})                      {}
func (l *testLogger) Debugln(v ...interface{})                    {}
func (l *testLogger) Debugf(format string, v ...interface{})      {}
func (l *testLogger) Trace(v ...interface{})                      {}
func (l *testLogger) Traceln(v ...interface{})                    {}
func (l *testLogger) Tracef(format string, v ...interface{})      {}

func recUnwrap(err error, errs *[]error) {
	if err == nil {
		return
	}

	*errs = append(*errs, err)

	e := errors.Unwrap(err)

	recUnwrap(e, errs)
}

type testFormatter struct{}

func (f testFormatter) Format(e *event.Event) (buf []byte, err error) {
	return nil, testErrFormat
}

var testFormat LoggerConfig = &formatConfig{f: &testFormatter{}}

func TestLoggerCheckDefaults(t *testing.T) {
	module := "Logger"
	funcname := "checkDefaults()"

	_ = module
	_ = funcname

	type test struct {
		name string
		l    *logger
		e    *event.Event
		want *event.Event
	}

	var tests = []test{
		{
			name: "both w/ defaults",
			l:    New().(*logger),
			e:    event.New().Message("null").Build(),
			want: event.New().Message("null").Build(),
		},
		{
			name: "logger w/ custom prefix",
			l:    New(WithPrefix("test")).(*logger),
			e:    event.New().Message("null").Build(),
			want: event.New().Prefix("test").Message("null").Build(),
		},
		{
			name: "logger w/ custom sub-prefix",
			l:    New(WithSub("test")).(*logger),
			e:    event.New().Message("null").Build(),
			want: event.New().Sub("test").Message("null").Build(),
		},
		{
			name: "logger w/ custom metadata",
			l:    New(WithSub("test")).Fields(map[string]interface{}{"a": true}).(*logger),
			e:    event.New().Message("null").Build(),
			want: event.New().Sub("test").Message("null").Metadata(map[string]interface{}{"a": true}).Build(),
		},
		{
			name: "logger w/ custom metadata -- merge with event's metadata",
			l:    New(WithSub("test")).Fields(map[string]interface{}{"a": true}).(*logger),
			e:    event.New().Message("null").Metadata(map[string]interface{}{"b": false}).Build(),
			want: event.New().Sub("test").Message("null").Metadata(map[string]interface{}{"a": true, "b": false}).Build(),
		},
	}

	var verify = func(idx int, test test) {
		e := test.e

		test.l.checkDefaults(e)
		test.want.Time = e.Time

		if !reflect.DeepEqual(e, test.want) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.want,
				e,
				test.name,
			)
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}

}

func TestLoggerOutput(t *testing.T) {
	module := "Logger"
	funcname := "Output()"

	_ = module
	_ = funcname

	type test struct {
		name string
		l    Logger
		e    *event.Event
		n    int
		err  error
	}

	var buf = &bytes.Buffer{}

	var tests = []test{
		{
			name: "default working Output() call",
			l:    New(WithOut(buf), CfgFormatJSON),
			e:    event.New().Message("null").Build(),
			n:    94,
			err:  nil,
		},
		{
			name: "logger filtering a message due to its level",
			l:    New(WithOut(buf), CfgFormatJSON, WithFilter(event.Level_error)),
			e:    event.New().Message("null").Build(),
			n:    0,
			err:  nil,
		},
		{
			name: "logger gets a formatter error",
			l:    New(WithOut(buf), CfgFormatJSON, testFormat),
			e:    event.New().Message("null").Build(),
			n:    -1,
			err:  testErrFormat,
		},
		{
			name: "logger gets a write error",
			l:    &testLogger{},
			e:    event.New().Message("very long message that must cause an error").Build(),
			n:    67,
			err:  testErrWrite,
		},
	}

	var verify = func(idx int, test test) {
		n, err := test.l.Output(test.e)

		// one digit margin for timestamp micros
		if n != test.n && n != (test.n-1) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] written byte length mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.n,
				n,
				test.name,
			)
			return
		}

		if err != nil {
			if test.err == nil {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] got an error when no errors were expected: %v -- action: %s",
					idx,
					module,
					funcname,
					err,
					test.name,
				)
				return
			}

			if !errors.Is(err, test.err) {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] error mismatch: wanted %v ; got %v -- action: %s",
					idx,
					module,
					funcname,
					test.err,
					err,
					test.name,
				)
				return
			}
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func FuzzLoggerPrint(f *testing.F) {
	module := "Logger"
	funcname := "Print()"
	action := "fuzz testing logger.Print(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Print(a)

		// add newline
		var sb strings.Builder
		sb.WriteString(a)
		sb.WriteByte(10)

		if buf.String() != sb.String() {
			t.Errorf(
				"FAILED -- [%s] [%s] output mismatch error: wanted %s ; got %s -- action: %s",
				module,
				funcname,
				a,
				buf.String(),
				action,
			)
		}
	})
}

func FuzzLoggerPrintln(f *testing.F) {
	module := "Logger"
	funcname := "Println()"
	action := "fuzz testing logger.Println(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Println(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString(a)
		sb.WriteByte(10)
		sb.WriteByte(10)

		if buf.String() != sb.String() {
			t.Errorf(
				"FAILED -- [%s] [%s] output mismatch error: wanted %s ; got %s -- action: %s",
				module,
				funcname,
				a,
				buf.String(),
				action,
			)
		}
	})
}

func FuzzLoggerPrintf(f *testing.F) {
	module := "Logger"
	funcname := "Printf()"
	action := "fuzz testing logger.Printf(format string, v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Printf(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString(`"`)
		sb.WriteString(a)
		sb.WriteString(`"`)
		sb.WriteByte(10)

		if buf.String() != sb.String() {
			t.Errorf(
				"FAILED -- [%s] [%s] output mismatch error: wanted %s ; got %s -- action: %s",
				module,
				funcname,
				a,
				buf.String(),
				action,
			)
		}
	})
}

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"reflect"
// 	"testing"

// 	"github.com/zalgonoise/zlog/log/event"
// 	"github.com/zalgonoise/zlog/store"
// )

// func TestLoggerPrint(t *testing.T) {
// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Print(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Print(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Print(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Print(test.msg)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerPrintln(t *testing.T) {
// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Println(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Println(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Println(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Println(test.msg)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerPrintf(t *testing.T) {
// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		test := test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		test := test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Printf(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.format,
// 				test.v,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Printf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Printf(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Printf(test.format, test.v...)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerOutputLevelFilter(t *testing.T) {
// 	module := "Logger"
// 	funcname := "WithFilter() - Output()"

// 	type test struct {
// 		levelFilter event.Level
// 		msg         *event.Event
// 		n           int
// 	}

// 	var tests = []test{
// 		{
// 			levelFilter: 0,
// 			msg:         event.New().Message("hi").Build(),
// 			n:           3,
// 		},
// 		{
// 			levelFilter: 3,
// 			msg:         event.New().Message("hi").Build(),
// 			n:           0,
// 		},
// 		{
// 			levelFilter: 3,
// 			msg:         event.New().Level(event.Level_error).Message("hi").Build(),
// 			n:           3,
// 		},
// 		{
// 			levelFilter: 9,
// 			msg:         event.New().Message("hi").Build(),
// 			n:           0,
// 		},
// 	}

// 	var verify = func(id int, test test, logger Logger) {
// 		n, err := logger.Output(test.msg)
// 		if err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] error writting message: %s",
// 				id,
// 				module,
// 				funcname,
// 				err,
// 			)
// 			return
// 		}
// 		if n != test.n {
// 			t.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] byte count mismatch: wanted %v ; got %v",
// 				id,
// 				module,
// 				funcname,
// 				test.n,
// 				n,
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
// 	mockBufs[0].Reset()
// 	for id, test := range tests {
// 		logger := New(WithFilter(test.levelFilter), SkipExit, WithFormat(TextOnly), WithOut(mockBufs[0]))

// 		verify(id, test, logger)
// 		mockBufs[0].Reset()
// 	}
// }

// func TestLoggerCheckDefaults(t *testing.T) {
// 	module := "Logger"
// 	funcname := "checkDefaults()"

// 	tlogger := New(WithOut(store.EmptyWriter), WithFormat(FormatText))

// 	type test struct {
// 		name      string
// 		logMeta   map[string]interface{}
// 		logPrefix string
// 		logSub    string
// 		input     *event.Event
// 		wants     *event.Event
// 	}

// 	var tests = []test{
// 		{
// 			name:      "check defaults -- basic",
// 			logMeta:   nil,
// 			logPrefix: "",
// 			logSub:    "",
// 			input:     event.New().Message("hi").Build(),
// 			wants:     event.New().Message("hi").Build(),
// 		},
// 		{
// 			name:      "check defaults -- keep message metadata",
// 			logMeta:   nil,
// 			logPrefix: "",
// 			logSub:    "",
// 			input:     event.New().Message("hi").Metadata(event.Field{"a": 0}).Build(),
// 			wants:     event.New().Message("hi").Metadata(event.Field{"a": 0}).Build(),
// 		},
// 		{
// 			name:      "apply logger metadata",
// 			logMeta:   event.Field{"a": 0},
// 			logPrefix: "",
// 			logSub:    "",
// 			input:     event.New().Message("hi").Build(),
// 			wants:     event.New().Message("hi").Metadata(event.Field{"a": 0}).Build(),
// 		},
// 		{
// 			name:      "append logger metadata",
// 			logMeta:   event.Field{"a": 0},
// 			logPrefix: "",
// 			logSub:    "",
// 			input:     event.New().Message("hi").Metadata(event.Field{"b": 1}).Build(),
// 			wants:     event.New().Message("hi").Metadata(event.Field{"a": 0, "b": 1}).Build(),
// 		},
// 		{
// 			name:      "apply logger sub-prefix",
// 			logMeta:   nil,
// 			logPrefix: "",
// 			logSub:    "new",
// 			input:     event.New().Message("hi").Build(),
// 			wants:     event.New().Message("hi").Sub("new").Build(),
// 		},
// 		{
// 			name:      "apply logger prefix",
// 			logMeta:   nil,
// 			logPrefix: "new",
// 			logSub:    "",
// 			input:     event.New().Message("hi").Build(),
// 			wants:     event.New().Message("hi").Prefix("new").Build(),
// 		},
// 	}

// 	var verify = func(id int, test test, input *event.Event) {
// 		if input.GetPrefix() != test.wants.GetPrefix() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] prefix mismatch: wanted %s ; got %s -- action: %s",
// 				id,
// 				module,
// 				funcname,
// 				test.wants.GetPrefix(),
// 				input.GetPrefix(),
// 				test.name,
// 			)
// 			return
// 		}

// 		if input.GetSub() != test.wants.GetSub() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] sub-prefix mismatch: wanted %s ; got %s -- action: %s",
// 				id,
// 				module,
// 				funcname,
// 				test.wants.GetSub(),
// 				input.GetSub(),
// 				test.name,
// 			)
// 			return
// 		}
// 		if input.GetMsg() != test.wants.GetMsg() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] message mismatch: wanted %s ; got %s -- action: %s",
// 				id,
// 				module,
// 				funcname,
// 				test.wants.GetMsg(),
// 				input.GetMsg(),
// 				test.name,
// 			)
// 			return
// 		}

// 		if !reflect.DeepEqual(input.Meta.AsMap(), test.wants.Meta.AsMap()) {
// 			t.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] metadata mismatch: wanted %v ; got %v -- action: %s",
// 				id,
// 				module,
// 				funcname,
// 				test.wants.Meta.AsMap(),
// 				input.Meta.AsMap(),
// 				test.name,
// 			)
// 			return
// 		}
// 		t.Logf(
// 			"#%v -- PASSED -- [%s] [%s] -- action: %s",
// 			id,
// 			module,
// 			funcname,
// 			test.name,
// 		)
// 	}

// 	for id, test := range tests {
// 		tlogger.Prefix(test.logPrefix).Sub(test.logSub).Fields(test.logMeta)

// 		msg := test.input

// 		tlogger.(*logger).checkDefaults(msg)

// 		verify(id, test, msg)
// 		tlogger.Prefix("").Sub("").Fields(nil)
// 	}

// }

// func TestLoggerLog(t *testing.T) {
// 	type test struct {
// 		level     event.Level
// 		msg       string
// 		wantLevel string
// 		wantMsg   string
// 		meta      map[string]interface{}
// 		ok        bool
// 		panics    bool
// 	}

// 	var tests []test

// 	// metadata appendage test
// 	tests = append(tests, test{
// 		level:     event.Level_info,
// 		wantLevel: event.Level_info.String(),
// 		msg:       "meta",
// 		wantMsg:   "meta",
// 		meta: map[string]interface{}{
// 			"works": true,
// 		},
// 		ok:     true,
// 		panics: false,
// 	})

// 	for a := 0; a < len(mockLogLevelsOK); a++ {
// 		if a == 5 {
// 			continue // skip event.Level_fatal, or os.Exit(1)
// 		}
// 		for b := 0; b < len(mockMessages); b++ {
// 			test := test{
// 				level:     mockLogLevelsOK[a],
// 				msg:       mockMessages[b],
// 				wantLevel: mockLogLevelsOK[a].String(),
// 				wantMsg:   mockMessages[b],
// 				meta:      nil,
// 				ok:        true,
// 				panics:    false,
// 			}

// 			if a == 9 {
// 				test.panics = true
// 			}

// 			tests = append(tests, test)
// 		}
// 		for c := 0; c < len(mockFmtMessages); c++ {
// 			test := test{
// 				level:     mockLogLevelsOK[a],
// 				msg:       fmt.Sprintf(mockFmtMessages[c].format, mockFmtMessages[c].v...),
// 				wantLevel: mockLogLevelsOK[a].String(),
// 				wantMsg:   fmt.Sprintf(mockFmtMessages[c].format, mockFmtMessages[c].v...),
// 				ok:        true,
// 				meta:      nil,
// 				panics:    false,
// 			}

// 			if a == 9 {
// 				test.panics = true
// 			}

// 			tests = append(tests, test)
// 		}
// 	}

// 	for d := 0; d < len(mockLogLevelsNOK); d++ {
// 		if d == 5 {
// 			continue // skip event.Level_fatal, or os.Exit(1)
// 		}
// 		for e := 0; e < len(mockMessages); e++ {
// 			test := test{
// 				level:     mockLogLevelsNOK[d],
// 				msg:       mockMessages[e],
// 				wantLevel: mockLogLevelsNOK[d].String(),
// 				wantMsg:   mockMessages[e],
// 				meta:      nil,
// 				ok:        false,
// 				panics:    false,
// 			}

// 			if d == 9 {
// 				test.panics = true
// 			}

// 			tests = append(tests, test)
// 		}
// 		for f := 0; f < len(mockFmtMessages); f++ {
// 			test := test{
// 				level:     mockLogLevelsNOK[d],
// 				msg:       fmt.Sprintf(mockFmtMessages[f].format, mockFmtMessages[f].v...),
// 				wantLevel: mockLogLevelsNOK[d].String(),
// 				wantMsg:   fmt.Sprintf(mockFmtMessages[f].format, mockFmtMessages[f].v...),
// 				meta:      nil,
// 				ok:        true,
// 				panics:    false,
// 			}

// 			if d == 9 {
// 				test.panics = true
// 			}

// 			tests = append(tests, test)
// 		}
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		r := recover()

// 		if test.level == event.Level_panic {
// 			if r == nil {
// 				t.Errorf(
// 					"#%v -- FAILED -- LoggerMessage] Log(%s, %s) -- panic did not occur",
// 					id,
// 					test.level.String(),
// 					test.msg,
// 				)
// 				return
// 			}

// 			if r != test.wantMsg {
// 				t.Errorf(
// 					"#%v -- FAILED -- LoggerMessage] Log(%s, %s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					test.level.String(),
// 					test.msg,
// 					test.wantMsg,
// 					r,
// 				)
// 				return
// 			}
// 			t.Logf(
// 				"#%v -- PASSED -- LoggerMessage] Log(%s, %s) : %s",
// 				id,
// 				test.level.String(),
// 				test.msg,
// 				mockLogger.buf.String(),
// 			)

// 			mockLogger.buf.Reset()
// 			return
// 		}

// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Log(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.level.String(),
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != test.wantLevel {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Log(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.level.String(),
// 				test.msg,
// 				test.wantLevel,
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Log(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.level.String(),
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if len(logEntry.Meta.AsMap()) > 0 {
// 			for k, v := range logEntry.Meta.AsMap() {
// 				if test.meta[k] != v {
// 					t.Errorf(
// 						"#%v -- FAILED -- [LoggerMessage] Log(%s, %s) -- metadata mismatch: key %s mismatch: wanted %s ; got %s",
// 						id,
// 						test.level.String(),
// 						test.msg,
// 						k,
// 						k,
// 						test.meta[k],
// 					)
// 					return
// 				}
// 			}

// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Log(%s, %s) : %s",
// 			id,
// 			test.level.String(),
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		if test.level == event.Level_panic {
// 			defer verify(id, test, logEntry)
// 		}

// 		mockLogger.logger.Fields(test.meta)

// 		logMessage := event.New().Level(test.level).Message(test.msg).Build()

// 		mockLogger.logger.Log(logMessage)

// 		if test.level != event.Level_panic {
// 			verify(id, test, logEntry)
// 		}
// 		mockLogger.logger.Fields(nil)
// 	}

// }

// func TestLoggerLogMultiMessage(t *testing.T) {
// 	module := "Logger"
// 	funcname := "Log()"

// 	tlogger := New(WithOut(mockBufs[0]), WithFormat(TextOnly), SkipExit)

// 	type test struct {
// 		name string
// 		msgs []*event.Event
// 		n    int
// 	}

// 	var tests = []test{
// 		{
// 			name: "log one message",
// 			msgs: []*event.Event{event.New().Message("hi").Build()},
// 			n:    3,
// 		},
// 		{
// 			name: "log three messages",
// 			msgs: []*event.Event{event.New().Message("hi").Build(), event.New().Message("hi").Build(), event.New().Message("hi").Build()},
// 			n:    9,
// 		},
// 		{
// 			name: "log three messages -- first and last are empty",
// 			msgs: []*event.Event{{}, event.New().Message("hi").Build(), {}},
// 			n:    5,
// 		},
// 		{
// 			name: "log three empty messages",
// 			msgs: []*event.Event{{}, {}, {}},
// 			n:    3,
// 		},
// 		{
// 			name: "nil input",
// 			msgs: nil,
// 			n:    0,
// 		},
// 	}

// 	var verify = func(id int, test test) {
// 		if mockBufs[0].Len() != test.n {
// 			t.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] byte count mismatch: wanted %v ; got %v -- action: %s",
// 				id,
// 				module,
// 				funcname,
// 				test.n,
// 				mockBufs[0].Len(),
// 				test.name,
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [%s] [%s] -- action: %s",
// 			id,
// 			module,
// 			funcname,
// 			test.name,
// 		)

// 	}

// 	mockBufs[0].Reset()
// 	for id, test := range tests {

// 		if test.msgs == nil {
// 			tlogger.Log()
// 		} else {
// 			tlogger.Log(test.msgs...)
// 		}

// 		verify(id, test)
// 		mockBufs[0].Reset()
// 	}

// 	// test nil input
// 	tlogger.Log(nil, nil, nil)
// 	verify(0, test{name: "three nil messages", n: 0})

// 	// test nil input with a real message
// 	tlogger.Log(nil, nil, event.New().Message("hi").Build())
// 	verify(0, test{name: "two nil messages, then a real message", n: 3})

// }

// func TestLoggerPanic(t *testing.T) {
// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 		panics  bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		test := test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 			panics:  true,
// 		}

// 		tests = append(tests, test)
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		test := test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 			panics:  true,
// 		}

// 		tests = append(tests, test)
// 	}

// 	var verify = func(id int, test test) {
// 		r := recover()

// 		if r == nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- LoggerMessage] Panic(%s) -- panic did not occur",
// 				id,
// 				test.msg,
// 			)
// 			return
// 		}

// 		if r != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- LoggerMessage] Panic(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				r,
// 			)
// 			return
// 		}
// 		t.Logf(
// 			"#%v -- PASSED -- LoggerMessage] Panic(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		mockLogger.buf.Reset()

// 		defer verify(id, test)

// 		mockLogger.logger.Panic(test.msg)

// 	}
// }

// func TestLoggerPanicln(t *testing.T) {
// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 		panics  bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		test := test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 			panics:  true,
// 		}

// 		tests = append(tests, test)
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		test := test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 			panics:  true,
// 		}

// 		tests = append(tests, test)
// 	}

// 	var verify = func(id int, test test) {
// 		r := recover()

// 		if r == nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Panicln(%s) -- panic did not occur",
// 				id,
// 				test.msg,
// 			)
// 			return
// 		}

// 		if r != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Panicln(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				r,
// 			)
// 			return
// 		}
// 		t.Logf(
// 			"#%v -- PASSED -- LoggerMessage] Panicln(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()

// 	}

// 	for id, test := range tests {

// 		mockLogger.buf.Reset()

// 		defer verify(id, test)

// 		mockLogger.logger.Panicln(test.msg)

// 	}
// }

// func TestLoggerPanicf(t *testing.T) {
// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 		panics  bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		test := test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 			panics:  true,
// 		}

// 		tests = append(tests, test)
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		test := test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 			panics:  true,
// 		}

// 		tests = append(tests, test)
// 	}

// 	var verify = func(id int, test test) {
// 		r := recover()

// 		if r == nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- LoggerMessage] Panicf(%s, %s) -- panic did not occur",
// 				id,
// 				test.format,
// 				test.v,
// 			)
// 			return
// 		}

// 		if r != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- LoggerMessage] Panicf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				r,
// 			)
// 			return
// 		}
// 		t.Logf(
// 			"#%v -- PASSED -- LoggerMessage] Panicf(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()

// 	}

// 	for id, test := range tests {

// 		mockLogger.buf.Reset()

// 		defer verify(id, test)

// 		mockLogger.logger.Panicf(test.format, test.v...)

// 	}
// }

// func TestLoggerFatal(t *testing.T) {
// 	var noExitLogger = struct {
// 		logger Logger
// 		buf    *bytes.Buffer
// 	}{
// 		logger: New(
// 			WithPrefix("test-message"),
// 			WithFormat(FormatJSON),
// 			WithOut(mockBuffer),
// 			SkipExit,
// 		),
// 		buf: mockBuffer,
// 	}

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(noExitLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Fatal(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_fatal.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Fatal(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_fatal.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Fatal(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Fatal(%s) : %s",
// 			id,
// 			test.msg,
// 			noExitLogger.buf.String(),
// 		)

// 		noExitLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		noExitLogger.buf.Reset()

// 		noExitLogger.logger.Fatal(test.msg)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerFatalln(t *testing.T) {
// 	var noExitLogger = struct {
// 		logger Logger
// 		buf    *bytes.Buffer
// 	}{
// 		logger: New(
// 			WithPrefix("test-message"),
// 			WithFormat(FormatJSON),
// 			WithOut(mockBuffer),
// 			SkipExit,
// 		),
// 		buf: mockBuffer,
// 	}

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(noExitLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Fatalln(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_fatal.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Fatalln(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_fatal.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Fatalln(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Fatalln(%s) : %s",
// 			id,
// 			test.msg,
// 			noExitLogger.buf.String(),
// 		)

// 		noExitLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		noExitLogger.buf.Reset()

// 		noExitLogger.logger.Fatalln(test.msg)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerFatalf(t *testing.T) {
// 	var noExitLogger = struct {
// 		logger Logger
// 		buf    *bytes.Buffer
// 	}{
// 		logger: New(
// 			WithPrefix("test-message"),
// 			WithFormat(FormatJSON),
// 			WithOut(mockBuffer),
// 			SkipExit,
// 		),
// 		buf: mockBuffer,
// 	}

// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		test := test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		test := test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(noExitLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Fatalf(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.format,
// 				test.v,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_fatal.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Fatalf(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				event.Level_fatal.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Fatalf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Fatalf(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			noExitLogger.buf.String(),
// 		)

// 		noExitLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		noExitLogger.buf.Reset()

// 		noExitLogger.logger.Fatalf(test.format, test.v...)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerError(t *testing.T) {
// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Error(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_error.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Error(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_error.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Error(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Error(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Error(test.msg)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerErrorln(t *testing.T) {
// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Errorln(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_error.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Errorln(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_error.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Errorln(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Errorln(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Errorln(test.msg)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerErrorf(t *testing.T) {
// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		test := test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		test := test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Errorf(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.format,
// 				test.v,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_error.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Errorf(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				event.Level_error.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Errorf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Errorf(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Errorf(test.format, test.v...)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerWarn(t *testing.T) {
// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Warn(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_warn.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Warn(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_warn.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Warn(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Warn(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Warn(test.msg)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerWarnln(t *testing.T) {
// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Warnln(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_warn.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Warnln(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_warn.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Warnln(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Warnln(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Warnln(test.msg)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerWarnf(t *testing.T) {
// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		test := test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		test := test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Warnf(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.format,
// 				test.v,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_warn.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Warnf(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				event.Level_warn.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Warnf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Warnf(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Warnf(test.format, test.v...)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerInfo(t *testing.T) {
// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Info(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_info.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Info(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_info.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Info(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Info(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Info(test.msg)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerInfoln(t *testing.T) {
// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Infoln(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_info.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Infoln(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_info.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Infoln(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Infoln(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Infoln(test.msg)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerInfof(t *testing.T) {
// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		test := test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		test := test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Infof(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.format,
// 				test.v,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_info.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Infof(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				event.Level_info.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Infof(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Infof(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Infof(test.format, test.v...)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerDebug(t *testing.T) {
// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Debug(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_debug.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Debug(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_debug.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Debug(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Debug(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Debug(test.msg)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerDebugln(t *testing.T) {
// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Debugln(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_debug.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Debugln(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_debug.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Debugln(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Debugln(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Debugln(test.msg)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerDebugf(t *testing.T) {
// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		test := test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		test := test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Infof(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.format,
// 				test.v,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_debug.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Debugf(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				event.Level_debug.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Debugf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Debugf(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Debugf(test.format, test.v...)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerTrace(t *testing.T) {
// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Trace(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_trace.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Trace(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_trace.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Trace(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Trace(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Trace(test.msg)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerTraceln(t *testing.T) {
// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Traceln(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_trace.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Traceln(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_trace.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Traceln(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Traceln(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Traceln(test.msg)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestLoggerTracef(t *testing.T) {
// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		test := test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		test := test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}

// 	var verify = func(id int, test test, logEntry *event.Event) {
// 		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Tracef(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.format,
// 				test.v,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_trace.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Tracef(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				event.Level_trace.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [LoggerMessage] Tracef(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [LoggerMessage] Tracef(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			mockLogger.buf.String(),
// 		)

// 		mockLogger.buf.Reset()
// 	}

// 	for id, test := range tests {

// 		logEntry := &event.Event{}
// 		mockLogger.buf.Reset()

// 		mockLogger.logger.Tracef(test.format, test.v...)

// 		verify(id, test, logEntry)
// 	}
// }

// func TestMultiLoggerPrint(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Print(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Print(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_info.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Print(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_info.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Print(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Print(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Print(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Print(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Print(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Print(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Print(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerPrintln(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Println(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Println(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_info.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Println(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_info.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Println(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Println(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Println(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Println(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Println(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Println(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Println(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerPrintf(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		format string
// 		v      []interface{}
// 		msg    *event.Event
// 		ml     ml
// 	}

// 	var testAllObjects []map[string]interface{}
// 	testAllObjects = append(testAllObjects, testObjects...)
// 	testAllObjects = append(testAllObjects, testEmptyObjects...)

// 	var tests []test

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(mockMessages[b]).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: "%s",
// 					v:      []interface{}{mockMessages[b]},
// 				}

// 				tests = append(tests, obj)
// 			}
// 		}
// 	}

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockFmtMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: mockFmtMessages[b].format,
// 					v:      mockFmtMessages[b].v,
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Printf(%s, %s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Printf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_info.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Printf(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					event.Level_info.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Printf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Printf(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Printf(%s, %s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Printf(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.format,
// 							test.v,
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Printf(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.format,
// 						test.v,
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Printf(%s, %s) -- %s",
// 				id,
// 				bufID,
// 				test.format,
// 				test.v,
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Printf(test.format, test.v...)

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerLog(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
// 	}

// 	var testAllMessages []string
// 	testAllMessages = append(testAllMessages, mockMessages...)
// 	for _, fmtMsg := range mockFmtMessages {
// 		testAllMessages = append(testAllMessages, fmt.Sprintf(fmtMsg.format, fmtMsg.v...))
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testObjects); c++ {
// 				for d := 0; d < len(mockLogLevelsOK); d++ {

// 					// skip event.Level_fatal -- os.Exit(1)
// 					// if mockLogLevelsOK[d] == event.Level_fatal {
// 					// 	continue
// 					// }

// 					var bufs []*bytes.Buffer
// 					var logs []Logger
// 					for e := 0; e < len(mockMultiPrefixes); e++ {
// 						buf := &bytes.Buffer{}
// 						bufs = append(bufs, buf)
// 						logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf), SkipExit))
// 					}
// 					mlogger := MultiLogger(logs...)

// 					obj := test{
// 						ml: ml{
// 							log: mlogger,
// 							buf: bufs,
// 						},
// 						msg: event.New().
// 							Level(mockLogLevelsOK[d]).
// 							Prefix(mockPrefixes[a]).
// 							Message(testAllMessages[b]).
// 							Metadata(testObjects[c]).
// 							Build(),
// 					}

// 					tests = append(tests, obj)
// 				}
// 			}
// 		}
// 	}

// 	for a := 0; a < len(mockEmptyPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testEmptyObjects); c++ {
// 				for d := 0; d < len(mockLogLevelsNOK); d++ {

// 					var bufs []*bytes.Buffer
// 					var logs []Logger
// 					for e := 0; e < len(mockMultiPrefixes); e++ {
// 						buf := &bytes.Buffer{}
// 						bufs = append(bufs, buf)
// 						logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf), SkipExit))
// 					}
// 					mlogger := MultiLogger(logs...)

// 					obj := test{
// 						ml: ml{
// 							log: mlogger,
// 							buf: bufs,
// 						},
// 						msg: event.New().
// 							Level(mockLogLevelsNOK[d]).
// 							Prefix(mockEmptyPrefixes[a]).
// 							Message(testAllMessages[b]).
// 							Metadata(testEmptyObjects[c]).
// 							Build(),
// 					}

// 					tests = append(tests, obj)
// 				}
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Log(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Log(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != test.msg.GetLevel().String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Log(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Level,
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Log(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Log(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Log(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Log(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Log(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Log(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Log(test.msg)

// 		verify(id, test)
// 	}
// }

// func TestMultiLoggerPanic(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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

// 		r := recover()
// 		if r == nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger).Panic(%s) -- test did not panic",
// 				id,
// 				test.msg.GetMsg(),
// 			)
// 			return
// 		}

// 		if r != test.msg.GetMsg() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger).Panic(%s) -- panic message %s does not match input %s",
// 				id,
// 				test.msg.GetMsg(),
// 				r,
// 				test.msg.GetMsg(),
// 			)
// 		}

// 		for bufID, buf := range test.ml.buf {

// 			logEntry := &event.Event{}

// 			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panic(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panic(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_panic.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panic(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_info.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panic(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panic(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panic(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panic(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panic(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Panic(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())

// 		defer verify(id, test)
// 		test.ml.log.Panic(test.msg.GetMsg())

// 	}
// }

// func TestMultiLoggerPanicln(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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

// 		r := recover()
// 		if r == nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger).Panicln(%s) -- test did not panic",
// 				id,
// 				test.msg.GetMsg(),
// 			)
// 			return
// 		}

// 		if r != test.msg.GetMsg()+"\n" {
// 			t.Errorf(
// 				"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger).Panicln(%s) -- panic message %s does not match input %s",
// 				id,
// 				test.msg.GetMsg(),
// 				r,
// 				test.msg.GetMsg(),
// 			)
// 		}

// 		for bufID, buf := range test.ml.buf {

// 			logEntry := &event.Event{}

// 			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicln(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicln(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_panic.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicln(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_info.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicln(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicln(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicln(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicln(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicln(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicln(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())

// 		defer verify(id, test)
// 		test.ml.log.Panicln(test.msg.GetMsg())

// 	}
// }

// func TestMultiLoggerPanicf(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		format string
// 		v      []interface{}
// 		msg    *event.Event
// 		ml     ml
// 	}

// 	var testAllObjects []map[string]interface{}
// 	testAllObjects = append(testAllObjects, testObjects...)
// 	testAllObjects = append(testAllObjects, testEmptyObjects...)

// 	var tests []test

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(mockMessages[b]).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: "%s",
// 					v:      []interface{}{mockMessages[b]},
// 				}

// 				tests = append(tests, obj)
// 			}
// 		}
// 	}
// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockFmtMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: mockFmtMessages[b].format,
// 					v:      mockFmtMessages[b].v,
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

// 		r := recover()
// 		if r == nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger).Panicf(%s, %s) -- test did not panic",
// 				id,
// 				test.format,
// 				test.v,
// 			)
// 			return
// 		}

// 		if r != test.msg.GetMsg() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger).Panicf(%s, %s) -- panic message %s does not match input %s",
// 				id,
// 				test.format,
// 				test.v,
// 				r,
// 				test.msg.GetMsg(),
// 			)
// 		}

// 		for bufID, buf := range test.ml.buf {

// 			logEntry := &event.Event{}

// 			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicf(%s, %s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_panic.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicf(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					event.Level_info.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicf(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicf(%s, %s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicf(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.format,
// 							test.v,
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicf(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.format,
// 						test.v,
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Panicf(%s, %s) -- %s",
// 				id,
// 				bufID,
// 				test.format,
// 				test.v,
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())

// 		defer verify(id, test)
// 		test.ml.log.Panicf(test.format, test.v...)

// 	}
// }

// func TestMultiLoggerFatal(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf), SkipExit))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatal(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatal(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_fatal.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatal(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_fatal.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatal(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatal(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatal(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatal(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatal(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatal(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Fatal(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerFatalln(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf), SkipExit))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalln(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalln(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_fatal.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalln(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_fatal.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalln(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalln(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalln(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalln(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalln(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalln(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Fatalln(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerFatalf(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		format string
// 		v      []interface{}
// 		msg    *event.Event
// 		ml     ml
// 	}

// 	var testAllObjects []map[string]interface{}
// 	testAllObjects = append(testAllObjects, testObjects...)
// 	testAllObjects = append(testAllObjects, testEmptyObjects...)

// 	var tests []test

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf), SkipExit))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(mockMessages[b]).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: "%s",
// 					v:      []interface{}{mockMessages[b]},
// 				}

// 				tests = append(tests, obj)
// 			}
// 		}
// 	}

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockFmtMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf), SkipExit))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: mockFmtMessages[b].format,
// 					v:      mockFmtMessages[b].v,
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalf(%s, %s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_fatal.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalf(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					event.Level_fatal.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalf(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalf(%s, %s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalf(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.format,
// 							test.v,
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalf(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.format,
// 						test.v,
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Fatalf(%s, %s) -- %s",
// 				id,
// 				bufID,
// 				test.format,
// 				test.v,
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Fatalf(test.format, test.v...)

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerError(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Error(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Error(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_error.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Error(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_error.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Error(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Error(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Error(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Error(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Error(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Error(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Error(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerErrorln(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorln(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorln(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_error.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorln(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_error.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorln(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorln(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorln(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorln(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorln(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorln(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Errorln(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerErrorf(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		format string
// 		v      []interface{}
// 		msg    *event.Event
// 		ml     ml
// 	}

// 	var testAllObjects []map[string]interface{}
// 	testAllObjects = append(testAllObjects, testObjects...)
// 	testAllObjects = append(testAllObjects, testEmptyObjects...)

// 	var tests []test

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(mockMessages[b]).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: "%s",
// 					v:      []interface{}{mockMessages[b]},
// 				}

// 				tests = append(tests, obj)
// 			}
// 		}
// 	}

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockFmtMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: mockFmtMessages[b].format,
// 					v:      mockFmtMessages[b].v,
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorf(%s, %s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_error.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorf(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					event.Level_error.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorf(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorf(%s, %s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorf(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.format,
// 							test.v,
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorf(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.format,
// 						test.v,
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Errorf(%s, %s) -- %s",
// 				id,
// 				bufID,
// 				test.format,
// 				test.v,
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Errorf(test.format, test.v...)

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerWarn(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warn(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warn(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_warn.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warn(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_warn.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warn(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warn(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warn(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warn(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warn(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Warn(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Warn(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerWarnln(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnln(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnln(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_warn.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnln(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_warn.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnln(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnln(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnln(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnln(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnln(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnln(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Warnln(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerWarnf(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		format string
// 		v      []interface{}
// 		msg    *event.Event
// 		ml     ml
// 	}

// 	var testAllObjects []map[string]interface{}
// 	testAllObjects = append(testAllObjects, testObjects...)
// 	testAllObjects = append(testAllObjects, testEmptyObjects...)

// 	var tests []test

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(mockMessages[b]).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: "%s",
// 					v:      []interface{}{mockMessages[b]},
// 				}

// 				tests = append(tests, obj)
// 			}
// 		}
// 	}

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockFmtMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: mockFmtMessages[b].format,
// 					v:      mockFmtMessages[b].v,
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnf(%s, %s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_warn.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnf(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					event.Level_warn.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnf(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnf(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.format,
// 							test.v,
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnf(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.format,
// 						test.v,
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Warnf(%s, %s) -- %s",
// 				id,
// 				bufID,
// 				test.format,
// 				test.v,
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Warnf(test.format, test.v...)

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggernfo(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_info.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_info.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Info(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggernfoln(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infoln(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infoln(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_info.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infoln(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_info.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infoln(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infoln(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infoln(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infoln(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infoln(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Infoln(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Infoln(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggernfof(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		format string
// 		v      []interface{}
// 		msg    *event.Event
// 		ml     ml
// 	}

// 	var testAllObjects []map[string]interface{}
// 	testAllObjects = append(testAllObjects, testObjects...)
// 	testAllObjects = append(testAllObjects, testEmptyObjects...)

// 	var tests []test

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(mockMessages[b]).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: "%s",
// 					v:      []interface{}{mockMessages[b]},
// 				}

// 				tests = append(tests, obj)
// 			}
// 		}
// 	}

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockFmtMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: mockFmtMessages[b].format,
// 					v:      mockFmtMessages[b].v,
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infof(%s, %s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infof(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_info.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infof(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					event.Level_info.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infof(%s, %s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infof(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infof(%s, %s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infof(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.format,
// 							test.v,
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Infof(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.format,
// 						test.v,
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Infof(%s, %s) -- %s",
// 				id,
// 				bufID,
// 				test.format,
// 				test.v,
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Infof(test.format, test.v...)

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerDebug(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debug(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debug(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_debug.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debug(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_debug.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debug(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debug(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debug(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debug(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debug(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Debug(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Debug(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerDebugln(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugln(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugln(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_debug.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugln(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_debug.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugln(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugln(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugln(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugln(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugln(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugln(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Debugln(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerDebugf(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		format string
// 		v      []interface{}
// 		msg    *event.Event
// 		ml     ml
// 	}

// 	var testAllObjects []map[string]interface{}
// 	testAllObjects = append(testAllObjects, testObjects...)
// 	testAllObjects = append(testAllObjects, testEmptyObjects...)

// 	var tests []test

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(mockMessages[b]).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: "%s",
// 					v:      []interface{}{mockMessages[b]},
// 				}

// 				tests = append(tests, obj)
// 			}
// 		}
// 	}

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockFmtMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: mockFmtMessages[b].format,
// 					v:      mockFmtMessages[b].v,
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugf(%s, %s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_debug.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugf(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					event.Level_debug.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugf(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugf(%s, %s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugf(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.format,
// 							test.v,
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugf(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.format,
// 						test.v,
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Debugf(%s, %s) -- %s",
// 				id,
// 				bufID,
// 				test.format,
// 				test.v,
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Debugf(test.format, test.v...)

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerTrace(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Trace(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Trace(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_trace.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Trace(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_trace.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Trace(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Trace(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Trace(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Trace(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Trace(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Trace(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Trace(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerTraceln(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		msg *event.Event
// 		ml  ml
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

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(testAllMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Traceln(%s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Traceln(%s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_trace.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Traceln(%s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					event.Level_trace.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Traceln(%s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Traceln(%s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Traceln(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.msg.GetMsg(),
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Traceln(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.msg.GetMsg(),
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Traceln(%s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.msg.GetMsg(),
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Traceln(%s) -- %s",
// 				id,
// 				bufID,
// 				test.msg.GetMsg(),
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Traceln(test.msg.GetMsg())

// 		verify(id, test)

// 	}
// }

// func TestMultiLoggerTracef(t *testing.T) {
// 	type ml struct {
// 		log Logger
// 		buf []*bytes.Buffer
// 	}

// 	type test struct {
// 		format string
// 		v      []interface{}
// 		msg    *event.Event
// 		ml     ml
// 	}

// 	var testAllObjects []map[string]interface{}
// 	testAllObjects = append(testAllObjects, testObjects...)
// 	testAllObjects = append(testAllObjects, testEmptyObjects...)

// 	var tests []test

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(mockMessages[b]).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: "%s",
// 					v:      []interface{}{mockMessages[b]},
// 				}

// 				tests = append(tests, obj)
// 			}
// 		}
// 	}

// 	for a := 0; a < len(mockPrefixes); a++ {
// 		for b := 0; b < len(mockFmtMessages); b++ {
// 			for c := 0; c < len(testAllObjects); c++ {

// 				var bufs []*bytes.Buffer
// 				var logs []Logger
// 				for e := 0; e < len(mockMultiPrefixes); e++ {
// 					buf := &bytes.Buffer{}
// 					bufs = append(bufs, buf)
// 					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), WithFormat(FormatJSON), WithOut(buf)))
// 				}
// 				mlogger := MultiLogger(logs...)

// 				obj := test{
// 					ml: ml{
// 						log: mlogger,
// 						buf: bufs,
// 					},
// 					msg: event.New().
// 						Prefix(mockPrefixes[a]).
// 						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
// 						Metadata(testAllObjects[c]).
// 						Build(),
// 					format: mockFmtMessages[b].format,
// 					v:      mockFmtMessages[b].v,
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
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Tracef(%s, %s) -- unmarshal error: %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					err,
// 				)
// 				return
// 			}

// 			if logEntry.GetPrefix() != test.msg.GetPrefix() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Tracef(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetPrefix(),
// 					logEntry.GetPrefix(),
// 				)
// 				return
// 			}

// 			if logEntry.GetLevel().String() != event.Level_trace.String() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Tracef(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					event.Level_trace.String(),
// 					logEntry.Level,
// 				)
// 				return
// 			}

// 			if logEntry.GetMsg() != test.msg.GetMsg() {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Tracef(%s, %s) -- message mismatch: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.GetMsg(),
// 					logEntry.GetMsg(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) == 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Tracef(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			} else if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) == 0 {
// 				t.Errorf(
// 					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Tracef(%s, %s) -- retrieved unexpected metadata object: wanted %s ; got %s",
// 					id,
// 					bufID,
// 					test.format,
// 					test.v,
// 					test.msg.Meta.AsMap(),
// 					logEntry.Meta.AsMap(),
// 				)
// 				return
// 			}

// 			if len(logEntry.Meta.AsMap()) > 0 && len(test.msg.Meta.AsMap()) > 0 {
// 				for k, v := range logEntry.Meta.AsMap() {
// 					if v != nil && test.msg.Meta.AsMap()[k] == nil {
// 						t.Errorf(
// 							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Tracef(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
// 							id,
// 							bufID,
// 							test.format,
// 							test.v,
// 							k,
// 							k,
// 						)
// 						return
// 					}
// 				}

// 				if len(logEntry.Meta.AsMap()) != len(test.msg.Meta.AsMap()) {
// 					t.Errorf(
// 						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Tracef(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
// 						id,
// 						bufID,
// 						test.format,
// 						test.v,
// 						len(test.msg.Meta.AsMap()),
// 						len(logEntry.Meta.AsMap()),
// 					)
// 					return
// 				}
// 			}

// 			t.Logf(
// 				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Tracef(%s, %s) -- %s",
// 				id,
// 				bufID,
// 				test.format,
// 				test.v,
// 				buf.String(),
// 			)
// 		}

// 	}

// 	for id, test := range tests {
// 		for _, b := range test.ml.buf {
// 			b.Reset()
// 		}

// 		test.ml.log.Prefix(test.msg.GetPrefix()).Fields(test.msg.Meta.AsMap())
// 		test.ml.log.Tracef(test.format, test.v...)

// 		verify(id, test)

// 	}
// }

// func TestPrint(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Print(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Print(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_info.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Print(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_info.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Print(%s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Print(%s) : %s",
// 			id,
// 			test.msg,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Print(test.msg)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd

// }

// func TestPrintln(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Println(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Println(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_info.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Println(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_info.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Println(%s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Println(%s) : %s",
// 			id,
// 			test.msg,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Println(test.msg)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd

// }

// func TestPrintf(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Printf(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.format,
// 				test.v,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Printf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_info.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Printf(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				event.Level_info.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Printf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Printf(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Printf(test.format, test.v...)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd
// }

// func TestLog(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	defer func() {
// 		if r := recover(); r != nil {
// 			t.Logf(
// 				"# -- TEST -- [Log()] Intended panic recovery -- %s",
// 				r,
// 			)
// 		}
// 	}()

// 	type test struct {
// 		level     event.Level
// 		msg       string
// 		wantLevel string
// 		wantMsg   string
// 		ok        bool
// 		panics    bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockLogLevelsOK); a++ {
// 		if a == 5 {
// 			continue // skip event.Level_fatal, or os.Exit(1)
// 		}
// 		for b := 0; b < len(mockMessages); b++ {
// 			test := test{
// 				level:     mockLogLevelsOK[a],
// 				msg:       mockMessages[b],
// 				wantLevel: mockLogLevelsOK[a].String(),
// 				wantMsg:   mockMessages[b],
// 				ok:        true,
// 				panics:    false,
// 			}

// 			if a == 9 {
// 				test.panics = true
// 			}

// 			tests = append(tests, test)
// 		}
// 		for c := 0; c < len(mockFmtMessages); c++ {
// 			test := test{
// 				level:     mockLogLevelsOK[a],
// 				msg:       fmt.Sprintf(mockFmtMessages[c].format, mockFmtMessages[c].v...),
// 				wantLevel: mockLogLevelsOK[a].String(),
// 				wantMsg:   fmt.Sprintf(mockFmtMessages[c].format, mockFmtMessages[c].v...),
// 				ok:        true,
// 				panics:    false,
// 			}

// 			if a == 9 {
// 				test.panics = true
// 			}

// 			tests = append(tests, test)
// 		}
// 	}
// 	for d := 0; d < len(mockLogLevelsNOK); d++ {
// 		if d == 5 {
// 			continue // skip event.Level_fatal, or os.Exit(1)
// 		}
// 		for e := 0; e < len(mockMessages); e++ {
// 			test := test{
// 				level:     mockLogLevelsNOK[d],
// 				msg:       mockMessages[e],
// 				wantLevel: mockLogLevelsNOK[d].String(),
// 				wantMsg:   mockMessages[e],
// 				ok:        false,
// 				panics:    false,
// 			}

// 			if d == 9 {
// 				test.panics = true
// 			}

// 			tests = append(tests, test)
// 		}
// 		for f := 0; f < len(mockFmtMessages); f++ {
// 			test := test{
// 				level:     mockLogLevelsNOK[d],
// 				msg:       fmt.Sprintf(mockFmtMessages[f].format, mockFmtMessages[f].v...),
// 				wantLevel: mockLogLevelsNOK[d].String(),
// 				wantMsg:   fmt.Sprintf(mockFmtMessages[f].format, mockFmtMessages[f].v...),
// 				ok:        true,
// 				panics:    false,
// 			}

// 			if d == 9 {
// 				test.panics = true
// 			}

// 			tests = append(tests, test)
// 		}
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Log(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.level.String(),
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Log(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.level.String(),
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != test.wantLevel {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Log(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.level.String(),
// 				test.msg,
// 				test.wantLevel,
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Log(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.level.String(),
// 				test.msg,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Log(%s, %s) : %s",
// 			id,
// 			test.level.String(),
// 			test.msg,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	defer func() { std = oldstd }()

// 	for id, test := range tests {

// 		buf.Reset()

// 		if test.panics {
// 			defer verify(id, test, buf.Bytes())
// 		}

// 		logMessage := event.New().Level(test.level).Message(test.msg).Build()

// 		Log(logMessage)

// 		if !test.panics {
// 			verify(id, test, buf.Bytes())
// 		}

// 	}

// }

// func TestPanic(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		test := test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		test := test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}

// 	var verify = func(id int, test test) {
// 		r := recover()

// 		if r == nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Panic(%s) -- panic did not occur",
// 				id,
// 				test.msg,
// 			)
// 			return
// 		}

// 		if r != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Panic(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				r,
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Panic(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		buf.Reset()

// 	}

// 	defer func() { std = oldstd }()

// 	for id, test := range tests {

// 		buf.Reset()

// 		defer verify(id, test)
// 		Panic(test.msg)

// 	}
// }

// func TestPanicln(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		test := test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		test := test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}

// 	var verify = func(id int, test test) {
// 		r := recover()

// 		if r == nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Panicln(%s) -- panic did not occur",
// 				id,
// 				test.msg,
// 			)
// 			return
// 		}

// 		if r != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Panicln(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				r,
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Panicln(%s) : %s",
// 			id,
// 			test.msg,
// 			mockLogger.buf.String(),
// 		)

// 		buf.Reset()

// 	}

// 	defer func() { std = oldstd }()

// 	for id, test := range tests {

// 		buf.Reset()

// 		defer verify(id, test)
// 		Panicln(test.msg)

// 	}
// }

// func TestPanicf(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		test := test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		test := test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		}

// 		tests = append(tests, test)
// 	}

// 	var verify = func(id int, test test) {
// 		r := recover()

// 		if r == nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Panicf(%s, %s) -- panic did not occur",
// 				id,
// 				test.format,
// 				test.v,
// 			)
// 			return
// 		}

// 		if r != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Panicf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				r,
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Panicf(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			mockLogger.buf.String(),
// 		)

// 		buf.Reset()

// 	}

// 	defer func() { std = oldstd }()

// 	for id, test := range tests {

// 		buf.Reset()

// 		defer verify(id, test)
// 		Panicf(test.format, test.v...)

// 	}
// }

// func TestFatal(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 		SkipExit,
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Fatal(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Fatal(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_fatal.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Fatal(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_fatal.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Fatal(%s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Fatal(%s) : %s",
// 			id,
// 			test.msg,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Fatal(test.msg)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd

// }

// func TestFatalln(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 		SkipExit,
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Fatalln(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Fatalln(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_fatal.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Fatalln(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_fatal.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Fatalln(%s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Fatalln(%s) : %s",
// 			id,
// 			test.msg,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Fatalln(test.msg)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd

// }

// func TestFatalf(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 		SkipExit,
// 	)

// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Fatalf(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.format,
// 				test.v,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Fatalf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_fatal.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Fatalf(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				event.Level_fatal.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Fatalf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Fatalf(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Fatalf(test.format, test.v...)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd
// }

// func TestError(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Error(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Error(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_error.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Error(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_error.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Error(%s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Error(%s) : %s",
// 			id,
// 			test.msg,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Error(test.msg)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd

// }

// func TestErrorln(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Errorln(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Errorln(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_error.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Errorln(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_error.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Errorln(%s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Errorln(%s) : %s",
// 			id,
// 			test.msg,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Errorln(test.msg)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd

// }

// func TestErrorf(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Errorf(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.format,
// 				test.v,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Errorf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_error.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Errorf(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				event.Level_error.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Errorf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Errorf(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Errorf(test.format, test.v...)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd
// }

// func TestWarn(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Warn(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Warn(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_warn.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Warn(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_warn.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Warn(%s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Warn(%s) : %s",
// 			id,
// 			test.msg,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Warn(test.msg)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd

// }

// func TestWarnln(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Warnln(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Warnln(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_warn.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Warnln(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_warn.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Warnln(%s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Warnln(%s) : %s",
// 			id,
// 			test.msg,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Warnln(test.msg)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd

// }

// func TestWarnf(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Warnf(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.format,
// 				test.v,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Warnf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_warn.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Warnf(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				event.Level_warn.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Warnf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Warnf(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Warnf(test.format, test.v...)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd
// }

// func TestInfo(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Info(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Info(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_info.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Info(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_info.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Info(%s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Info(%s) : %s",
// 			id,
// 			test.msg,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Info(test.msg)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd

// }

// func TestInfoln(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Infoln(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Infoln(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_info.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Infoln(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_info.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Infoln(%s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Infoln(%s) : %s",
// 			id,
// 			test.msg,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Infoln(test.msg)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd

// }

// func TestInfof(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Infof(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.format,
// 				test.v,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Infof(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_info.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Infof(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				event.Level_info.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Infof(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Infof(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Infof(test.format, test.v...)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd
// }

// func TestDebug(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Debug(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Debug(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_debug.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Debug(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_debug.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Debug(%s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Debug(%s) : %s",
// 			id,
// 			test.msg,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Debug(test.msg)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd

// }

// func TestDebugln(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Debugln(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Debugln(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_debug.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Debugln(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_debug.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Debugln(%s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Debugln(%s) : %s",
// 			id,
// 			test.msg,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Debugln(test.msg)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd

// }

// func TestDebugf(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Debugf(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.format,
// 				test.v,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Debugf(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_debug.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Debugf(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				event.Level_debug.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Debugf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Debugf(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Debugf(test.format, test.v...)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd
// }

// func TestTrace(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Trace(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Trace(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_trace.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Trace(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_trace.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Trace(%s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Trace(%s) : %s",
// 			id,
// 			test.msg,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Trace(test.msg)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd

// }

// func TestTraceln(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		msg     string
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			msg:     mockMessages[a],
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Traceln(%s) -- unmarshal error: %s",
// 				id,
// 				test.msg,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Traceln(%s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_trace.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Traceln(%s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				event.Level_trace.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Traceln(%s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.msg,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Traceln(%s) : %s",
// 			id,
// 			test.msg,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Traceln(test.msg)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd

// }

// func TestTracef(t *testing.T) {
// 	// std logger override
// 	oldstd := std
// 	buf := &bytes.Buffer{}
// 	std = New(
// 		WithPrefix("log"),
// 		WithFormat(FormatJSON),
// 		WithOut(buf),
// 	)

// 	type test struct {
// 		format  string
// 		v       []interface{}
// 		wantMsg string
// 		ok      bool
// 	}

// 	var tests []test

// 	for a := 0; a < len(mockMessages); a++ {
// 		tests = append(tests, test{
// 			format:  "%s",
// 			v:       []interface{}{mockMessages[a]},
// 			wantMsg: mockMessages[a],
// 			ok:      true,
// 		})
// 	}
// 	for b := 0; b < len(mockFmtMessages); b++ {
// 		tests = append(tests, test{
// 			format:  mockFmtMessages[b].format,
// 			v:       mockFmtMessages[b].v,
// 			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
// 			ok:      true,
// 		})
// 	}

// 	var verify = func(id int, test test, result []byte) {
// 		logEntry := &event.Event{}

// 		if err := json.Unmarshal(result, logEntry); err != nil {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Tracef(%s, %s) -- unmarshal error: %s",
// 				id,
// 				test.format,
// 				test.v,
// 				err,
// 			)
// 			return
// 		}

// 		if logEntry.GetMsg() != test.wantMsg {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Tracef(%s, %s) -- message mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				test.wantMsg,
// 				logEntry.GetMsg(),
// 			)
// 			return
// 		}

// 		if logEntry.GetLevel().String() != event.Level_trace.String() {
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Tracef(%s, %s) -- log level mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				event.Level_trace.String(),
// 				logEntry.Level,
// 			)
// 			return
// 		}

// 		if logEntry.GetPrefix() != "log" { // std logger prefix
// 			t.Errorf(
// 				"#%v -- FAILED -- [DefaultLogger] Tracef(%s, %s) -- prefix mismatch: wanted %s ; got %s",
// 				id,
// 				test.format,
// 				test.v,
// 				"log",
// 				logEntry.GetPrefix(),
// 			)
// 			return
// 		}

// 		t.Logf(
// 			"#%v -- PASSED -- [DefaultLogger] Tracef(%s, %s) : %s",
// 			id,
// 			test.format,
// 			test.v,
// 			string(result),
// 		)

// 		buf.Reset()

// 	}

// 	for id, test := range tests {

// 		buf.Reset()

// 		Tracef(test.format, test.v...)

// 		verify(id, test, buf.Bytes())

// 	}

// 	std = oldstd
// }

// // func (l *nilLogger) Output(m *event.Event) (n int, err error)     { return 0, nil }
// func TestNilLoggerOutput(t *testing.T) {
// 	module := "NilLogger"
// 	funcname := "Output()"

// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input *event.Event
// 		n     int
// 		err   error
// 	}{
// 		{
// 			input: event.New().Message("test").Build(),
// 			n:     1,
// 			err:   nil,
// 		},
// 		{
// 			input: event.New().Message("for").Build(),
// 			n:     1,
// 			err:   nil,
// 		},
// 		{
// 			input: event.New().Message("nothing").Build(),
// 			n:     1,
// 			err:   nil,
// 		},
// 	}

// 	for id, test := range tests {
// 		n, err := nillog.Output(test.input)

// 		if n != test.n {
// 			t.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] bytes written mismatch: wanted %v ; got %v",
// 				id,
// 				module,
// 				funcname,
// 				test.n,
// 				n,
// 			)
// 			return

// 		}

// 		if err != test.err {
// 			t.Errorf(
// 				"#%v -- FAILED -- [%s] [%s] returning error mismatch: wanted %v ; got %v",
// 				id,
// 				module,
// 				funcname,
// 				test.err,
// 				err,
// 			)
// 			return
// 		}
// 	}
// }

// // func (l *nilLogger) Log(m ...*event.Event)                        {}
// func TestNilLoggerLog(t *testing.T) {
// 	// module := "NilLogger"
// 	// funcname := "Log()"

// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input *event.Event
// 	}{
// 		{
// 			input: event.New().Message("test").Build(),
// 		},
// 		{
// 			input: event.New().Message("for").Build(),
// 		},
// 		{
// 			input: event.New().Message("nothing").Build(),
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Log(test.input)
// 	}
// }

// // func (l *nilLogger) Print(v ...interface{})                      {}
// func TestNilLoggerPrint(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Print(test.input)
// 	}
// }

// // func (l *nilLogger) Println(v ...interface{})                    {}
// func TestNilLoggerPrintln(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Println(test.input)
// 	}
// }

// // func (l *nilLogger) Printf(format string, v ...interface{})      {}
// func TestNilLoggerPrintf(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Printf("%s", test.input)
// 	}
// }

// // func (l *nilLogger) Panic(v ...interface{})                      {}
// func TestNilLoggerPanic(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Panic(test.input)
// 	}
// }

// // func (l *nilLogger) Panicln(v ...interface{})                    {}
// func TestNilLoggerPanicln(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Panicln(test.input)
// 	}
// }

// // func (l *nilLogger) Panicf(format string, v ...interface{})      {}
// func TestNilLoggerPanicf(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Panicf("%s", test.input)
// 	}
// }

// // func (l *nilLogger) Fatal(v ...interface{})                      {}
// func TestNilLoggerFatal(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Fatal(test.input)
// 	}
// }

// // func (l *nilLogger) Fatalln(v ...interface{})                    {}
// func TestNilLoggerFatalln(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Fatalln(test.input)
// 	}
// }

// // func (l *nilLogger) Fatalf(format string, v ...interface{})      {}
// func TestNilLoggerFatalf(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Fatalf("%s", test.input)
// 	}
// }

// // func (l *nilLogger) Error(v ...interface{})                      {}
// func TestNilLoggerError(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Error(test.input)
// 	}
// }

// // func (l *nilLogger) Errorln(v ...interface{})                    {}
// func TestNilLoggerErrorln(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Errorln(test.input)
// 	}
// }

// // func (l *nilLogger) Errorf(format string, v ...interface{})      {}
// func TestNilLoggerErrorf(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Errorf("%s", test.input)
// 	}
// }

// // func (l *nilLogger) Warn(v ...interface{})                       {}
// func TestNilLoggerWarn(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Warn(test.input)
// 	}
// }

// // func (l *nilLogger) Warnln(v ...interface{})                     {}
// func TestNilLoggerWarnln(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Warnln(test.input)
// 	}
// }

// // func (l *nilLogger) Warnf(format string, v ...interface{})       {}
// func TestNilLoggerWarnf(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Warnf("%s", test.input)
// 	}
// }

// // func (l *nilLogger) Info(v ...interface{})                       {}
// func TestNilLoggerInfo(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Info(test.input)
// 	}
// }

// // func (l *nilLogger) Infoln(v ...interface{})                     {}
// func TestNilLoggerInfoln(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Infoln(test.input)
// 	}
// }

// // func (l *nilLogger) Infof(format string, v ...interface{})       {}
// func TestNilLoggerInfof(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Infof("%s", test.input)
// 	}
// }

// // func (l *nilLogger) Debug(v ...interface{})                      {}
// func TestNilLoggerDebug(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Debug(test.input)
// 	}
// }

// // func (l *nilLogger) Debugln(v ...interface{})                    {}
// func TestNilLoggerDebugln(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Debugln(test.input)
// 	}
// }

// // func (l *nilLogger) Debugf(format string, v ...interface{})      {}
// func TestNilLoggerDebugf(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Debugf("%s", test.input)
// 	}
// }

// // func (l *nilLogger) Trace(v ...interface{})                      {}
// func TestNilLoggerTrace(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Trace(test.input)
// 	}
// }

// // func (l *nilLogger) Traceln(v ...interface{})                    {}
// func TestNilLoggerTraceln(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Traceln(test.input)
// 	}
// }

// // func (l *nilLogger) Tracef(format string, v ...interface{})      {}
// func TestNilLoggerTracef(t *testing.T) {
// 	nillog := New(NilConfig)

// 	var tests = []struct {
// 		input string
// 	}{
// 		{
// 			input: "test",
// 		},
// 		{
// 			input: "for",
// 		},
// 		{
// 			input: "nothing",
// 		},
// 	}

// 	for _, test := range tests {
// 		nillog.Tracef("%s", test.input)
// 	}
// }
