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

type testWC struct{}

func (testWC) Write(p []byte) (n int, err error) { return 1, nil }
func (testWC) Close() error                      { return nil }

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

		// two digit margin for timestamp micros
		if n != test.n && n != (test.n-1) && n != (test.n-2) {
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

func FuzzLoggerLog(f *testing.F) {
	module := "Logger"
	funcname := "Log()"
	action := "fuzz testing logger.Log(...*event.Event)"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoTimestamp().NoLevel().Build()))

	f.Add("test-prefix", "test-sub", "test-message")
	f.Fuzz(func(t *testing.T, a, b, c string) {
		defer buf.Reset()

		e := event.New().Prefix(a).Sub(b).Message(c).Build()

		logger.Log(e)

		var sb strings.Builder
		sb.WriteString(`[`)
		sb.WriteString(a)
		sb.WriteString("]\t[")
		sb.WriteString(b)
		sb.WriteString("]\t")
		sb.WriteString(c)
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

func TestLoggerLog(t *testing.T) {
	module := "Logger"
	funcname := "Log()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		e     []*event.Event
		wants string
		panic bool
	}

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), WithFormat(text.New().NoTimestamp().NoLevel().Build()))

	var tests = []test{
		{
			name:  "no messages sent -- nil",
			e:     nil,
			wants: "",
			panic: false,
		},
		{
			name:  "no messages sent -- empty slice",
			e:     []*event.Event{},
			wants: "",
			panic: false,
		},
		{
			name:  "no messages sent -- slice w/ nil events",
			e:     []*event.Event{nil, nil, nil},
			wants: "",
			panic: false,
		},
		{
			name:  "no messages sent -- empty slice",
			e:     []*event.Event{},
			wants: "",
			panic: false,
		},
		{
			name: "no messages sent -- empty slice",
			e: []*event.Event{
				event.New().Level(event.Level_panic).Message("null").Build(),
			},
			wants: "null",
			panic: true,
		},
	}

	var handlePanic = func(idx int, test test) {
		e := recover()

		if e != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] panicking output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				e,
				test.name,
			)
		}
	}

	var verify = func(idx int, test test) {
		defer buf.Reset()

		if test.panic {
			defer handlePanic(idx, test)
		}

		logger.Log(test.e...)

		if buf.String() != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				buf.String(),
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func FuzzLoggerPanic(f *testing.F) {
	module := "Logger"
	funcname := "Panic()"
	action := "fuzz testing logger.Panic(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {

		var handlePanic = func() {
			e := recover()

			if e != a {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					a,
					e,
					action,
				)
				return
			}
		}

		defer handlePanic()
		defer buf.Reset()

		logger.Panic(a)

	})
}

func FuzzLoggerPanicln(f *testing.F) {
	module := "Logger"
	funcname := "Panicln()"
	action := "fuzz testing logger.Panicln(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {

		var handlePanic = func() {
			e := recover()

			var sb strings.Builder
			sb.WriteString(a)
			sb.WriteByte(10)

			if e != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					sb.String(),
					e,
					action,
				)
				return
			}
		}

		defer handlePanic()
		defer buf.Reset()

		logger.Panicln(a)

	})
}

func FuzzLoggerPanicf(f *testing.F) {
	module := "Logger"
	funcname := "Panicf()"
	action := "fuzz testing logger.Panicf(format string, v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {

		var handlePanic = func() {
			e := recover()

			var sb strings.Builder
			sb.WriteString(`"`)
			sb.WriteString(a)
			sb.WriteString(`"`)

			if e != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					sb.String(),
					e,
					action,
				)
				return
			}
		}

		defer handlePanic()
		defer buf.Reset()

		logger.Panicf(`"%s"`, a)

	})
}

func FuzzLoggerFatal(f *testing.F) {
	module := "Logger"
	funcname := "Fatal()"
	action := "fuzz testing logger.Fatal(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Fatal(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[fatal]\t")
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

func FuzzLoggerFatalln(f *testing.F) {
	module := "Logger"
	funcname := "Fatalln()"
	action := "fuzz testing logger.Fatalln(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Fatalln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[fatal]\t")
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

func FuzzLoggerFatalf(f *testing.F) {
	module := "Logger"
	funcname := "Fatalf()"
	action := "fuzz testing logger.Fatalf(format string, v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Fatalf(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[fatal]\t")
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

func FuzzLoggerError(f *testing.F) {
	module := "Logger"
	funcname := "Error()"
	action := "fuzz testing logger.Error(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Error(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[error]\t")
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

func FuzzLoggerErrorln(f *testing.F) {
	module := "Logger"
	funcname := "Errorln()"
	action := "fuzz testing logger.Errorln(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Errorln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[error]\t")
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

func FuzzLoggerErrorf(f *testing.F) {
	module := "Logger"
	funcname := "Errorf()"
	action := "fuzz testing logger.Errorf(format string, v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Errorf(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[error]\t")
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

func FuzzLoggerWarn(f *testing.F) {
	module := "Logger"
	funcname := "Warn()"
	action := "fuzz testing logger.Warn(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Warn(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[warn]\t")
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

func FuzzLoggerWarnln(f *testing.F) {
	module := "Logger"
	funcname := "Warnln()"
	action := "fuzz testing logger.Warnln(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Warnln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[warn]\t")
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

func FuzzLoggerWarnf(f *testing.F) {
	module := "Logger"
	funcname := "Warnf()"
	action := "fuzz testing logger.Warnf(format string, v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Warnf(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[warn]\t")
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

func FuzzLoggerInfo(f *testing.F) {
	module := "Logger"
	funcname := "Info()"
	action := "fuzz testing logger.Info(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Info(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[info]\t")
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

func FuzzLoggerInfoln(f *testing.F) {
	module := "Logger"
	funcname := "Infoln()"
	action := "fuzz testing logger.Infoln(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Infoln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[info]\t")
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

func FuzzLoggerInfof(f *testing.F) {
	module := "Logger"
	funcname := "Infof()"
	action := "fuzz testing logger.Infof(format string, v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Infof(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[info]\t")
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

func FuzzLoggerDebug(f *testing.F) {
	module := "Logger"
	funcname := "Debug()"
	action := "fuzz testing logger.Debug(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Debug(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[debug]\t")
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

func FuzzLoggerDebugln(f *testing.F) {
	module := "Logger"
	funcname := "Debugln()"
	action := "fuzz testing logger.Debugln(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Debugln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[debug]\t")
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

func FuzzLoggerDebugf(f *testing.F) {
	module := "Logger"
	funcname := "Debugf()"
	action := "fuzz testing logger.Debugf(format string, v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Debugf(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[debug]\t")
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

func FuzzLoggerTrace(f *testing.F) {
	module := "Logger"
	funcname := "Trace()"
	action := "fuzz testing logger.Trace(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Trace(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[trace]\t")
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

func FuzzLoggerTraceln(f *testing.F) {
	module := "Logger"
	funcname := "Debugln()"
	action := "fuzz testing logger.Debugln(v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Traceln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[trace]\t")
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

func FuzzLoggerTracef(f *testing.F) {
	module := "Logger"
	funcname := "Tracef()"
	action := "fuzz testing logger.Tracef(format string, v ...interface{})"

	var buf = &bytes.Buffer{}
	var logger = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		logger.Tracef(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[trace]\t")
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

func TestMultiLoggerOutput(t *testing.T) {
	module := "MultiLogger"
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

	var buf = []*bytes.Buffer{{}, {}}

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	var tests = []test{
		{
			name: "default working Output() call",
			l:    MultiLogger(New(WithOut(buf[0]), CfgFormatJSON), New(WithOut(buf[1]), CfgFormatJSON)),
			e:    event.New().Message("null").Build(),
			n:    94,
			err:  nil,
		},
		{
			name: "logger filtering a message due to its level",
			l:    MultiLogger(New(WithOut(buf[0]), CfgFormatJSON, WithFilter(event.Level_error)), New(WithOut(buf[1]), CfgFormatJSON, WithFilter(event.Level_error))),
			e:    event.New().Message("null").Build(),
			n:    0,
			err:  nil,
		},
		{
			name: "logger gets a formatter error",
			l:    MultiLogger(New(WithOut(buf[0]), CfgFormatJSON, testFormat), New(WithOut(buf[1]), CfgFormatJSON, testFormat)),
			e:    event.New().Message("null").Build(),
			n:    -1,
			err:  testErrFormat,
		},
		{
			name: "logger gets a write error",
			l:    MultiLogger(&testLogger{}, &testLogger{}),
			e:    event.New().Message("very long message that must cause an error").Build(),
			n:    67,
			err:  testErrWrite,
		},
	}

	var verify = func(idx int, test test) {
		defer reset()

		n, err := test.l.Output(test.e)

		// one digit margin for timestamp micros
		if n != test.n && n != (test.n-1) && n != (test.n-2) {
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

func FuzzMultiLoggerPrint(f *testing.F) {
	module := "MultiLogger"
	funcname := "Print()"
	action := "fuzz testing multiLogger.Print(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Print(a)

		// add newline
		var sb strings.Builder
		sb.WriteString(a)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}

	})
}

func FuzzMultiLoggerPrintln(f *testing.F) {
	module := "MultiLogger"
	funcname := "Println()"
	action := "fuzz testing multiLogger.Println(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Println(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString(a)
		sb.WriteByte(10)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerPrintf(f *testing.F) {
	module := "MultiLogger"
	funcname := "Printf()"
	action := "fuzz testing multiLogger.Printf(format string, v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Printf(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString(`"`)
		sb.WriteString(a)
		sb.WriteString(`"`)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerLog(f *testing.F) {
	module := "MultiLogger"
	funcname := "Log()"
	action := "fuzz testing multiLogger.Log(...*event.Event)"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoTimestamp().NoLevel().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoTimestamp().NoLevel().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-prefix", "test-sub", "test-message")
	f.Fuzz(func(t *testing.T, a, b, c string) {
		defer reset()

		e := event.New().Prefix(a).Sub(b).Message(c).Build()

		logger.Log(e)

		var sb strings.Builder
		sb.WriteString(`[`)
		sb.WriteString(a)
		sb.WriteString("]\t[")
		sb.WriteString(b)
		sb.WriteString("]\t")
		sb.WriteString(c)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func TestMultiLoggerLog(t *testing.T) {
	module := "MultiLogger"
	funcname := "Log()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		e     []*event.Event
		wants string
		panic bool
	}

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build())),
		New(WithOut(buf[1]), WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	var tests = []test{
		{
			name:  "no messages sent -- nil",
			e:     nil,
			wants: "",
			panic: false,
		},
		{
			name:  "no messages sent -- empty slice",
			e:     []*event.Event{},
			wants: "",
			panic: false,
		},
		{
			name:  "no messages sent -- slice w/ nil events",
			e:     []*event.Event{nil, nil, nil},
			wants: "",
			panic: false,
		},
		{
			name:  "no messages sent -- empty slice",
			e:     []*event.Event{},
			wants: "",
			panic: false,
		},
		{
			name: "no messages sent -- empty slice",
			e: []*event.Event{
				event.New().Level(event.Level_panic).Message("null").Build(),
			},
			wants: "null",
			panic: true,
		},
	}

	var handlePanic = func(idx int, test test) {
		e := recover()

		if e != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] panicking output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				e,
				test.name,
			)
		}
	}

	var verify = func(idx int, test test) {
		defer reset()

		if test.panic {
			defer handlePanic(idx, test)
		}

		logger.Log(test.e...)

		for lidx, b := range buf {
			if b.String() != test.wants {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %v ; got %v -- action: %s",
					idx,
					module,
					funcname,
					lidx,
					test.wants,
					b.String(),
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

func FuzzMultiLoggerPanic(f *testing.F) {
	module := "MultiLogger"
	funcname := "Panic()"
	action := "fuzz testing multiLogger.Panic(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build())),
		New(WithOut(buf[1]), WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {

		var handlePanic = func() {
			e := recover()

			if e != a {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					a,
					e,
					action,
				)
				return
			}
		}

		defer handlePanic()
		defer reset()

		logger.Panic(a)

	})
}

func FuzzMultiLoggerPanicln(f *testing.F) {
	module := "MultiLogger"
	funcname := "Panicln()"
	action := "fuzz testing multiLogger.Panicln(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {

		var handlePanic = func() {
			e := recover()

			var sb strings.Builder
			sb.WriteString(a)
			sb.WriteByte(10)

			if e != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					sb.String(),
					e,
					action,
				)
				return
			}
		}

		defer handlePanic()
		defer reset()

		logger.Panicln(a)

	})
}

func FuzzMultiLoggerPanicf(f *testing.F) {
	module := "MultiLogger"
	funcname := "Panicf()"
	action := "fuzz testing multiLogger.Panicf(format string, v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {

		var handlePanic = func() {
			e := recover()

			var sb strings.Builder
			sb.WriteString(`"`)
			sb.WriteString(a)
			sb.WriteString(`"`)

			if e != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					sb.String(),
					e,
					action,
				)
				return
			}
		}

		defer handlePanic()
		defer reset()

		logger.Panicf(`"%s"`, a)

	})
}

func FuzzMultiLoggerFatal(f *testing.F) {
	module := "MultiLogger"
	funcname := "Fatal()"
	action := "fuzz testing multiLogger.Fatal(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Fatal(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[fatal]\t")
		sb.WriteString(a)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerFatalln(f *testing.F) {
	module := "MultiLogger"
	funcname := "Fatalln()"
	action := "fuzz testing multiLogger.Fatalln(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Fatalln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[fatal]\t")
		sb.WriteString(a)
		sb.WriteByte(10)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerFatalf(f *testing.F) {
	module := "MultiLogger"
	funcname := "Fatalf()"
	action := "fuzz testing multiLogger.Fatalf(format string, v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Fatalf(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[fatal]\t")
		sb.WriteString(`"`)
		sb.WriteString(a)
		sb.WriteString(`"`)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerError(f *testing.F) {
	module := "MultiLogger"
	funcname := "Error()"
	action := "fuzz testing multiLogger.Error(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Error(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[error]\t")
		sb.WriteString(a)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerErrorln(f *testing.F) {
	module := "MultiLogger"
	funcname := "Errorln()"
	action := "fuzz testing multiLogger.Errorln(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Errorln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[error]\t")
		sb.WriteString(a)
		sb.WriteByte(10)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerErrorf(f *testing.F) {
	module := "MultiLogger"
	funcname := "Errorf()"
	action := "fuzz testing multiLogger.Errorf(format string, v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Errorf(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[error]\t")
		sb.WriteString(`"`)
		sb.WriteString(a)
		sb.WriteString(`"`)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerWarn(f *testing.F) {
	module := "MultiLogger"
	funcname := "Warn()"
	action := "fuzz testing multiLogger.Warn(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Warn(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[warn]\t")
		sb.WriteString(a)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerWarnln(f *testing.F) {
	module := "MultiLogger"
	funcname := "Warnln()"
	action := "fuzz testing multiLogger.Warnln(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Warnln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[warn]\t")
		sb.WriteString(a)
		sb.WriteByte(10)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerWarnf(f *testing.F) {
	module := "MultiLogger"
	funcname := "Warnf()"
	action := "fuzz testing multiLogger.Warnf(format string, v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Warnf(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[warn]\t")
		sb.WriteString(`"`)
		sb.WriteString(a)
		sb.WriteString(`"`)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerInfo(f *testing.F) {
	module := "MultiLogger"
	funcname := "Info()"
	action := "fuzz testing multiLogger.Info(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Info(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[info]\t")
		sb.WriteString(a)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerInfoln(f *testing.F) {
	module := "MultiLogger"
	funcname := "Infoln()"
	action := "fuzz testing multiLogger.Infoln(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Infoln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[info]\t")
		sb.WriteString(a)
		sb.WriteByte(10)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerInfof(f *testing.F) {
	module := "MultiLogger"
	funcname := "Infof()"
	action := "fuzz testing multiLogger.Infof(format string, v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Infof(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[info]\t")
		sb.WriteString(`"`)
		sb.WriteString(a)
		sb.WriteString(`"`)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerDebug(f *testing.F) {
	module := "MultiLogger"
	funcname := "Debug()"
	action := "fuzz testing multiLogger.Debug(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Debug(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[debug]\t")
		sb.WriteString(a)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerDebugln(f *testing.F) {
	module := "MultiLogger"
	funcname := "Debugln()"
	action := "fuzz testing multiLogger.Debugln(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Debugln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[debug]\t")
		sb.WriteString(a)
		sb.WriteByte(10)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerDebugf(f *testing.F) {
	module := "MultiLogger"
	funcname := "Debugf()"
	action := "fuzz testing multiLogger.Debugf(format string, v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Debugf(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[debug]\t")
		sb.WriteString(`"`)
		sb.WriteString(a)
		sb.WriteString(`"`)
		sb.WriteByte(10)

		for idx, b := range buf {

			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerTrace(f *testing.F) {
	module := "MultiLogger"
	funcname := "Trace()"
	action := "fuzz testing multiLogger.Trace(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Trace(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[trace]\t")
		sb.WriteString(a)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerTraceln(f *testing.F) {
	module := "MultiLogger"
	funcname := "Debugln()"
	action := "fuzz testing multiLogger.Debugln(v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Traceln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[trace]\t")
		sb.WriteString(a)
		sb.WriteByte(10)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzMultiLoggerTracef(f *testing.F) {
	module := "MultiLogger"
	funcname := "Tracef()"
	action := "fuzz testing multiLogger.Tracef(format string, v ...interface{})"

	var buf = []*bytes.Buffer{{}, {}}
	var logger = MultiLogger(
		New(WithOut(buf[0]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
		New(WithOut(buf[1]), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build())),
	)

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer reset()

		logger.Tracef(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[trace]\t")
		sb.WriteString(`"`)
		sb.WriteString(a)
		sb.WriteString(`"`)
		sb.WriteByte(10)

		for idx, b := range buf {
			if b.String() != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error in logger #%v: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					idx,
					a,
					b.String(),
					action,
				)
			}
		}
	})
}

func FuzzPrint(f *testing.F) {
	module := "StdLogger"
	funcname := "Print()"
	action := "fuzz testing Print(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Print(a)

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

func FuzzPrintln(f *testing.F) {
	module := "StdLogger"
	funcname := "Println()"
	action := "fuzz testing Println(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Println(a)

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

func FuzzPrintf(f *testing.F) {
	module := "StdLogger"
	funcname := "Printf()"
	action := "fuzz testing Printf(format string, v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Printf(`"%s"`, a)

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

func FuzzLog(f *testing.F) {
	module := "StdLogger"
	funcname := "Log()"
	action := "fuzz testing Log(...*event.Event)"

	oldstd := std
	buf := &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoTimestamp().NoLevel().Build()))
	defer func() { std = oldstd }()

	f.Add("test-prefix", "test-sub", "test-message")
	f.Fuzz(func(t *testing.T, a, b, c string) {
		defer buf.Reset()

		e := event.New().Prefix(a).Sub(b).Message(c).Build()

		Log(e)

		var sb strings.Builder
		sb.WriteString(`[`)
		sb.WriteString(a)
		sb.WriteString("]\t[")
		sb.WriteString(b)
		sb.WriteString("]\t")
		sb.WriteString(c)
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

func TestLog(t *testing.T) {
	module := "StdLogger"
	funcname := "Log()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		e     []*event.Event
		wants string
		panic bool
	}

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), WithFormat(text.New().NoTimestamp().NoLevel().Build()))
	defer func() { std = oldstd }()

	var tests = []test{
		{
			name:  "no messages sent -- nil",
			e:     nil,
			wants: "",
			panic: false,
		},
		{
			name:  "no messages sent -- empty slice",
			e:     []*event.Event{},
			wants: "",
			panic: false,
		},
		{
			name:  "no messages sent -- slice w/ nil events",
			e:     []*event.Event{nil, nil, nil},
			wants: "",
			panic: false,
		},
		{
			name:  "no messages sent -- empty slice",
			e:     []*event.Event{},
			wants: "",
			panic: false,
		},
		{
			name: "no messages sent -- empty slice",
			e: []*event.Event{
				event.New().Level(event.Level_panic).Message("null").Build(),
			},
			wants: "null",
			panic: true,
		},
	}

	var handlePanic = func(idx int, test test) {
		e := recover()

		if e != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] panicking output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				e,
				test.name,
			)
		}
	}

	var verify = func(idx int, test test) {
		defer buf.Reset()

		if test.panic {
			defer handlePanic(idx, test)
		}

		Log(test.e...)

		if buf.String() != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				buf.String(),
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func FuzzPanic(f *testing.F) {
	module := "StdLogger"
	funcname := "Panic()"
	action := "fuzz testing Panic(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {

		var handlePanic = func() {
			e := recover()

			if e != a {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					a,
					e,
					action,
				)
				return
			}
		}

		defer handlePanic()
		defer buf.Reset()

		Panic(a)

	})
}

func FuzzPanicln(f *testing.F) {
	module := "StdLogger"
	funcname := "Panicln()"
	action := "fuzz testing Panicln(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {

		var handlePanic = func() {
			e := recover()

			var sb strings.Builder
			sb.WriteString(a)
			sb.WriteByte(10)

			if e != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					sb.String(),
					e,
					action,
				)
				return
			}
		}

		defer handlePanic()
		defer buf.Reset()

		Panicln(a)

	})
}

func FuzzPanicf(f *testing.F) {
	module := "StdLogger"
	funcname := "Panicf()"
	action := "fuzz testing Panicf(format string, v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {

		var handlePanic = func() {
			e := recover()

			var sb strings.Builder
			sb.WriteString(`"`)
			sb.WriteString(a)
			sb.WriteString(`"`)

			if e != sb.String() {
				t.Errorf(
					"FAILED -- [%s] [%s] output mismatch error: wanted %s ; got %s -- action: %s",
					module,
					funcname,
					sb.String(),
					e,
					action,
				)
				return
			}
		}

		defer handlePanic()
		defer buf.Reset()

		Panicf(`"%s"`, a)

	})
}

func FuzzFatal(f *testing.F) {
	module := "Logger"
	funcname := "Fatal()"
	action := "fuzz testing Fatal(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Fatal(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[fatal]\t")
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

func FuzzFatalln(f *testing.F) {
	module := "StdLogger"
	funcname := "Fatalln()"
	action := "fuzz testing Fatalln(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Fatalln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[fatal]\t")
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

func FuzzFatalf(f *testing.F) {
	module := "StdLogger"
	funcname := "Fatalf()"
	action := "fuzz testing Fatalf(format string, v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Fatalf(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[fatal]\t")
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

func FuzzError(f *testing.F) {
	module := "StdLogger"
	funcname := "Error()"
	action := "fuzz testing Error(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Error(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[error]\t")
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

func FuzzErrorln(f *testing.F) {
	module := "StdLogger"
	funcname := "Errorln()"
	action := "fuzz testing Errorln(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Errorln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[error]\t")
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

func FuzzErrorf(f *testing.F) {
	module := "StdLogger"
	funcname := "Errorf()"
	action := "fuzz testing Errorf(format string, v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Errorf(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[error]\t")
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

func FuzzWarn(f *testing.F) {
	module := "StdLogger"
	funcname := "Warn()"
	action := "fuzz testing Warn(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Warn(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[warn]\t")
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

func FuzzWarnln(f *testing.F) {
	module := "StdLogger"
	funcname := "Warnln()"
	action := "fuzz testing Warnln(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Warnln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[warn]\t")
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

func FuzzWarnf(f *testing.F) {
	module := "StdLogger"
	funcname := "Warnf()"
	action := "fuzz testing Warnf(format string, v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Warnf(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[warn]\t")
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

func FuzzInfo(f *testing.F) {
	module := "StdLogger"
	funcname := "Info()"
	action := "fuzz testing Info(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Info(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[info]\t")
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

func FuzzInfoln(f *testing.F) {
	module := "StdLogger"
	funcname := "Infoln()"
	action := "fuzz testing Infoln(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Infoln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[info]\t")
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

func FuzzInfof(f *testing.F) {
	module := "StdLogger"
	funcname := "Infof()"
	action := "fuzz testing Infof(format string, v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Infof(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[info]\t")
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

func FuzzDebug(f *testing.F) {
	module := "StdLogger"
	funcname := "Debug()"
	action := "fuzz testing Debug(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Debug(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[debug]\t")
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

func FuzzDebugln(f *testing.F) {
	module := "StdLogger"
	funcname := "Debugln()"
	action := "fuzz testing Debugln(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Debugln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[debug]\t")
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

func FuzzDebugf(f *testing.F) {
	module := "StdLogger"
	funcname := "Debugf()"
	action := "fuzz testing Debugf(format string, v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Debugf(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[debug]\t")
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

func FuzzTrace(f *testing.F) {
	module := "StdLogger"
	funcname := "Trace()"
	action := "fuzz testing Trace(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Trace(a)

		// add newline
		var sb strings.Builder
		sb.WriteString("[trace]\t")
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

func FuzzTraceln(f *testing.F) {
	module := "StdLogger"
	funcname := "Debugln()"
	action := "fuzz testing Debugln(v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Traceln(a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[trace]\t")
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

func FuzzTracef(f *testing.F) {
	module := "StdLogger"
	funcname := "Tracef()"
	action := "fuzz testing Tracef(format string, v ...interface{})"

	oldstd := std
	var buf = &bytes.Buffer{}
	std = New(WithOut(buf), SkipExit, WithFormat(text.New().NoHeaders().NoTimestamp().Build()))
	defer func() { std = oldstd }()

	f.Add("test-message")
	f.Fuzz(func(t *testing.T, a string) {
		defer buf.Reset()

		Tracef(`"%s"`, a)

		// add newline x2
		var sb strings.Builder
		sb.WriteString("[trace]\t")
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

func TestNilLoggerPrint(t *testing.T) {
	module := "NilLogger"

	_ = module

	type test struct {
		name string
		e    *event.Event
	}

	var tests = []test{
		{
			name: "default message",
			e:    event.New().Message("null").Build(),
		},
		{
			name: "nil message",
			e:    nil,
		},
	}

	nl := New(NilConfig)

	var verify = func(idx int, test test) {
		var msg string

		if msg = test.e.GetMsg(); msg == "" {
			msg = "null"
		}

		n, err := nl.Output(test.e)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] call returned an unexpected error: %v -- action: %s",
				idx,
				module,
				err,
				test.name,
			)
			return
		}

		if n != 1 {
			t.Errorf(
				"#%v -- FAILED -- [%s] call returned an unexpected byte length: wanted %v ; got %v -- action: %s",
				idx,
				module,
				1,
				n,
				test.name,
			)
			return
		}

		// zero action calls
		nl.Log(test.e)

		nl.Print(msg)
		nl.Println(msg)
		nl.Printf("%s", msg)

		nl.Panic(msg)
		nl.Panicln(msg)
		nl.Panicf("%s", msg)

		nl.Fatal(msg)
		nl.Fatalln(msg)
		nl.Fatalf("%s", msg)

		nl.Error(msg)
		nl.Errorln(msg)
		nl.Errorf("%s", msg)

		nl.Warn(msg)
		nl.Warnln(msg)
		nl.Warnf("%s", msg)

		nl.Info(msg)
		nl.Infoln(msg)
		nl.Infof("%s", msg)

		nl.Debug(msg)
		nl.Debugln(msg)
		nl.Debugf("%s", msg)

		nl.Trace(msg)
		nl.Traceln(msg)
		nl.Tracef("%s", msg)
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
