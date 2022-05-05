package log

import (
	"errors"
	"fmt"
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

type testLogger struct {
	outs []io.Writer
}

func (l *testLogger) Write(p []byte) (n int, err error) {
	if len(p) < 10 {
		return
	}

	if len(p) > 10 && len(p) < 20 {
		n = 10
		return
	}

	if len(p) > 40 {
		n = len(p)
		err = errors.New("testing a write error")
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
func (l *testLogger) Output(m *event.Event) (n int, err error)    { return 1, nil }
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

func TestMultiLoggerSetOuts(t *testing.T) {
	module := "MultiLogger"
	funcname := "SetOuts()"

	type test struct {
		name string
		w    []io.Writer
	}

	var fakeConnAddr = &testLogger{}
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

		if !reflect.DeepEqual(out, tw) {
			pass = true
		}

		return pass
	}

	var verifyConnAddr = func(l *testLogger, tw io.Writer) (pass bool) {
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
				l, ok := logu.(*testLogger)
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

func TestMultiLoggerAddOuts(t *testing.T) {
	module := "MultiLogger"
	funcname := "AddOuts()"

	type test struct {
		name string
		w    []io.Writer
	}

	var fakeConnAddr = &testLogger{}
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

		ml.AddOuts(test.w...)

		return ml.(*multiLogger)
	}

	var getTestWriter = func(test test) (tw []io.Writer, isAddr bool) {
		if len(test.w) == 1 {
			tw = append(tw, test.w...)
		}

		if len(tw) > 0 {
			_, ok := tw[0].(*address.ConnAddr)

			if ok {
				isAddr = true
			}
		}
		tw = append(tw, os.Stderr)

		return
	}

	var verifyLogger = func(l *logger, tw io.Writer) (pass bool) {
		out := l.out

		if !reflect.DeepEqual(out, tw) {
			pass = true
		}

		return pass
	}

	var verifyConnAddr = func(l *testLogger, tw io.Writer) (pass bool) {
		if len(l.outs) > 0 {
			out := l.outs[0]

			if tw != nil && out == tw {
				pass = true
			}
		}
		return pass
	}

	var scanMultiLogger = func(ml *multiLogger, tw []io.Writer, isAddr bool) (pass bool) {
		for _, logu := range ml.loggers {
			l, ok := logu.(*logger)

			if ok && !isAddr {
				pass = verifyLogger(l, tw[0])

			} else if !ok && isAddr {
				l, ok := logu.(*testLogger)
				if ok {
					pass = verifyConnAddr(l, tw[0])
				}
			}
		}

		return
	}

	var verify = func(idx int, test test) {
		reset()
		defer reset()

		ml := init(test)

		var tw []io.Writer
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

func FuzzMultiLoggerPrefix(f *testing.F) {
	module := "MultiLogger"
	funcname := "Prefix()"

	ml := MultiLogger(loggers...)

	f.Add("test-prefix")
	f.Fuzz(func(t *testing.T, a string) {
		ml.Prefix(a)

		for idx, l := range ml.(*multiLogger).loggers {
			if l.(*logger).prefix != a {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] fuzzed prefix mismatch: wanted %s ; got %s",
					idx,
					module,
					funcname,
					a,
					l.(*logger).prefix,
				)
			}
		}
	})
}

func FuzzMultiLoggerSub(f *testing.F) {
	module := "MultiLogger"
	funcname := "Sub()"

	ml := MultiLogger(loggers...)

	f.Add("test-prefix")
	f.Fuzz(func(t *testing.T, a string) {
		ml.Sub(a)

		for idx, l := range ml.(*multiLogger).loggers {
			if l.(*logger).sub != a {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] fuzzed subprefix mismatch: wanted %s ; got %s",
					idx,
					module,
					funcname,
					a,
					l.(*logger).sub,
				)
			}
		}
	})
}

func TestMultiLoggerFields(t *testing.T) {
	module := "MultiLogger"
	funcname := "Fields()"

	type test struct {
		name  string
		init  map[string]interface{}
		input map[string]interface{}
		want  map[string]interface{}
	}

	var tests = []test{
		{
			name:  "default blank call",
			init:  map[string]interface{}{},
			input: map[string]interface{}{},
			want:  map[string]interface{}{},
		},
		{
			name:  "overwrite with blank",
			init:  map[string]interface{}{"a": true},
			input: map[string]interface{}{},
			want:  map[string]interface{}{},
		},
		{
			name:  "overwrite with nil",
			init:  map[string]interface{}{"a": true},
			input: nil,
			want:  map[string]interface{}{},
		},
		{
			name:  "write",
			init:  map[string]interface{}{},
			input: map[string]interface{}{"a": true},
			want:  map[string]interface{}{"a": true},
		},
		{
			name:  "overwrite",
			init:  map[string]interface{}{"a": false},
			input: map[string]interface{}{"a": true},
			want:  map[string]interface{}{"a": true},
		},
	}

	var reset = func(ml *multiLogger) {
		for _, l := range ml.loggers {
			l.Fields(nil)
		}
	}

	var init = func(test test) *multiLogger {
		ml := MultiLogger(loggers...)

		ml.Fields(test.init)

		return ml.(*multiLogger)
	}

	var verify = func(idx int, test test) {
		ml := init(test)
		defer reset(ml)

		ml.Fields(test.input)

		for _, l := range ml.loggers {
			if !reflect.DeepEqual(l.(*logger).meta, test.want) {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
					idx,
					module,
					funcname,
					test.want,
					l.(*logger).meta,
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

func TestMultiLoggerSkipExit(t *testing.T) {
	module := "MultiLogger"
	funcname := "SkipExit()"

	type test struct {
		name  string
		input []Logger
		want  bool
	}

	var tests = []test{
		{
			name:  "none skip exit",
			input: []Logger{New(), New()},
			want:  false,
		},
		{
			name:  "first is skip exit",
			input: []Logger{New(), New(SkipExit)},
			want:  false,
		},
		{
			name:  "middle is skip exit",
			input: []Logger{New(), New(SkipExit), New()},
			want:  false,
		},
		{
			name:  "last is skip exit",
			input: []Logger{New(), New(SkipExit)},
			want:  false,
		},
		{
			name:  "middle is not skip exit",
			input: []Logger{New(SkipExit), New(), New(SkipExit)},
			want:  false,
		},
		{
			name:  "all skip exit",
			input: []Logger{New(SkipExit), New(SkipExit), New(SkipExit)},
			want:  true,
		},
	}

	var reset = func(ml *multiLogger) {
		for _, l := range ml.loggers {
			l.(*logger).skipExit = false
		}
	}

	var init = func(test test) *multiLogger {
		return MultiLogger(test.input...).(*multiLogger)
	}

	var verify = func(idx int, test test) {
		ml := init(test)
		defer reset(ml)

		ok := ml.IsSkipExit()

		if ok != test.want {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.want,
				ok,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func recUnwrap(err error, errs *[]error) {
	if err == nil {
		return
	}

	*errs = append(*errs, err)

	e := errors.Unwrap(err)

	fmt.Println(e)

	recUnwrap(e, errs)
}

func TestMultiLoggerWrite(t *testing.T) {
	module := "MultiLogger"
	funcname := "Write()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		l     []Logger
		input []byte
		n     int
		errs  int
	}

	var tests = []test{
		{
			name:  "OK write",
			l:     []Logger{&testLogger{}, &testLogger{}},
			input: []byte("this is a long string to write"),
			n:     30,
			errs:  0,
		},
		{
			name:  "write error",
			l:     []Logger{&testLogger{}, New(NilConfig)},
			input: []byte("this is a long string to write that will error out"),
			n:     -1,
			errs:  1,
		},
		{
			name:  "multiple write errors",
			l:     []Logger{&testLogger{}, &testLogger{}},
			input: []byte("this is a long string to write that will error out"),
			n:     -1,
			errs:  3,
		},
		{
			name:  "short write error",
			l:     []Logger{&testLogger{}, &testLogger{}},
			input: []byte(""),
			n:     -1,
			errs:  3,
		},
	}

	var init = func(test test) *multiLogger {
		ml := MultiLogger(test.l...)

		return ml.(*multiLogger)
	}

	var unwrapErr = func(err error) []error {
		if err == nil {
			return []error{}
		}

		var errs []error

		recUnwrap(err, &errs)

		return errs
	}

	var verify = func(idx int, test test) {
		ml := init(test)

		n, err := ml.Write(test.input)

		if n != test.n {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] written bytes length mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.n,
				n,
				test.name,
			)
			return
		}

		errs := unwrapErr(err)

		if len(errs) != test.errs {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] expected errors length mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.errs,
				len(errs),
				test.name,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [%s] [%s] num errs: %v ; errs: %v -- action: %s",
			idx,
			module,
			funcname,
			len(errs),
			errs,
			test.name,
		)
		return
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
