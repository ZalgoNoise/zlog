package log

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/log/format/text"
	"github.com/zalgonoise/zlog/store"
)

func TestNew(t *testing.T) {
	module := "Logger"
	funcname := "New()"

	type test struct {
		name  string
		conf  []LoggerConfig
		wants *logger
	}

	var tests = []test{
		{
			name: "default config",
			wants: &logger{
				fmt:    TextColorLevelFirst,
				out:    os.Stderr,
				prefix: "log",
			},
		},
		{
			name: "with empty config: New()",
			conf: nil,
			wants: &logger{
				fmt:    TextColorLevelFirst,
				out:    os.Stderr,
				prefix: "log",
			},
		},
		{
			name: "with empty config: New([]LoggerConfig{})",
			conf: []LoggerConfig{},
			wants: &logger{
				fmt:    TextColorLevelFirst,
				out:    os.Stderr,
				prefix: "log",
			},
		},
		{
			name: "with custom config: JSON, SkipExit",
			conf: []LoggerConfig{
				CfgFormatJSON,
				SkipExit,
			},
			wants: &logger{
				fmt:      FormatJSON,
				out:      os.Stderr,
				prefix:   "log",
				skipExit: true,
			},
		},
	}

	var initNil = func() *logger {
		return New(nil).(*logger)
	}

	var init = func(test test) *logger {
		var l Logger

		if test.conf == nil {
			l = New()
		} else {
			l = New(test.conf...)
		}

		return l.(*logger)
	}

	var verify = func(idx int, test test) {
		logger := init(test)

		if !reflect.DeepEqual(logger, test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				logger,
				test.name,
			)
			return
		}
		t.Logf(
			"#%v -- PASSED -- [%s] [%s] -- action: %s",
			idx,
			module,
			funcname,
			test.name,
		)
	}

	for idx, test := range tests {
		verify(idx, test)
	}

	// verify New(nil)
	action := "with empty config: New(nil)"
	nilOptLogger := initNil()
	wants := &logger{
		fmt:    TextColorLevelFirst,
		out:    os.Stderr,
		prefix: "log",
	}

	if !reflect.DeepEqual(nilOptLogger, wants) {
		t.Errorf(
			"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
			0,
			module,
			funcname,
			wants,
			nilOptLogger,
			action,
		)
		return
	}
	t.Logf(
		"#%v -- PASSED -- [%s] [%s] -- action: %s",
		0,
		module,
		funcname,
		action,
	)
}

func TestNewNilLogger(t *testing.T) {
	module := "Logger"
	funcname := "New()"

	type test struct {
		name string
		conf []LoggerConfig
	}

	var tests = []test{
		{
			name: "NilConfig method",
			conf: []LoggerConfig{
				NilConfig,
			},
		},
		{
			name: "EmptyConfig method",
			conf: []LoggerConfig{
				EmptyConfig,
			},
		},
		{
			name: "NilLogger() function call",
			conf: []LoggerConfig{
				NilLogger(),
			},
		},
		{
			name: "manual config",
			conf: []LoggerConfig{
				WithOut(store.EmptyWriter),
				WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build()),
				SkipExit,
			},
		},
	}

	var init = func(test test) Logger {
		return New(test.conf...)
	}

	var verify = func(idx int, test test) {
		l := init(test)

		logger, ok := l.(*nilLogger)

		if !ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output logger isn't nilLogger -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		if !reflect.DeepEqual(*logger, nilLogger{}) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				nilLogger{},
				*logger,
				test.name,
			)
			return
		}
		t.Logf(
			"#%v -- PASSED -- [%s] [%s] -- action: %s",
			idx,
			module,
			funcname,
			test.name,
		)
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestSetOuts(t *testing.T) {
	module := "Logger"
	funcname := "SetOuts()"

	type test struct {
		name string
		w    []io.Writer
	}

	var defBuf = &bytes.Buffer{}
	var bufs = []*bytes.Buffer{{}, {}, {}}
	var l Logger = New(WithOut(defBuf))

	var tests = []test{
		{
			name: "set a different writer",
			w:    []io.Writer{bufs[0]},
		},
		{
			name: "add multiple writers",
			w:    []io.Writer{bufs[0], bufs[1], bufs[2]},
		},
		{
			name: "add no writers",
			w:    []io.Writer{},
		},
		{
			name: "add nil writers",
			w:    []io.Writer{nil, nil, nil},
		},
		{
			name: "add a good writer mixed in nil writers",
			w:    []io.Writer{nil, bufs[0], nil},
		},
	}

	var init = func(test test) io.Writer {
		var outs []io.Writer

		if len(test.w) == 0 {
			return stdout
		} else if len(test.w) > 0 {
			for _, w := range test.w {
				if w != nil {
					outs = append(outs, w)
				}
			}
		}

		if len(outs) > 0 {
			return io.MultiWriter(outs...)

		}

		return stdout
	}

	var verify = func(idx int, test test) {
		out := init(test)

		lcopy := l

		lcopy.SetOuts(test.w...)

		if !reflect.DeepEqual(lcopy.(*logger).out, out) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				out,
				lcopy.(*logger).out,
				test.name,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [%s] [%s] -- action: %s",
			idx,
			module,
			funcname,
			test.name,
		)
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestAddOuts(t *testing.T) {
	module := "Logger"
	funcname := "AddOuts()"

	type test struct {
		name string
		w    []io.Writer
		out  io.Writer
	}

	var defBuf = &bytes.Buffer{}
	var bufs = []*bytes.Buffer{{}, {}, {}}
	var l Logger = New(WithOut(defBuf))

	var tests = []test{
		{
			name: "set a different writer",
			w:    []io.Writer{bufs[0]},
			out:  io.MultiWriter(bufs[0], defBuf),
		},
		{
			name: "add multiple writers",
			w:    []io.Writer{bufs[0], bufs[1], bufs[2]},
			out:  io.MultiWriter(bufs[0], bufs[1], bufs[2], defBuf),
		},
		{
			name: "add no writers",
			w:    []io.Writer{},
			out:  io.MultiWriter(defBuf),
		},
		{
			name: "add nil writers",
			w:    []io.Writer{nil, nil, nil},
			out:  io.MultiWriter(defBuf),
		},
		{
			name: "add a good writer mixed in nil writers",
			w:    []io.Writer{nil, bufs[0], nil},
			out:  io.MultiWriter(bufs[0], defBuf),
		},
	}

	var reset = func() {
		l.SetOuts(defBuf)
	}

	var init = func(test test) io.Writer {
		var outs []io.Writer

		if len(test.w) == 0 {
			return io.MultiWriter(defBuf)
		} else if len(test.w) > 0 {
			for _, w := range test.w {
				if w != nil {
					outs = append(outs, w)
				}
			}
		}

		if len(outs) > 0 {
			outs = append(outs, defBuf)
			return io.MultiWriter(outs...)

		}

		return io.MultiWriter(defBuf)
	}

	var verify = func(idx int, test test) {

		out := init(test)

		lcopy := l

		lcopy.AddOuts(test.w...)

		if !reflect.DeepEqual(lcopy.(*logger).out, out) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				out,
				lcopy.(*logger).out,
				test.name,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [%s] [%s] -- action: %s",
			idx,
			module,
			funcname,
			test.name,
		)

	}

	for idx, test := range tests {
		reset()
		verify(idx, test)
		reset()
	}
}

func FuzzLoggerPrefix(f *testing.F) {
	module := "Logger"
	funcname := "Prefix()"

	l := New(WithPrefix("seed"))

	f.Add("")
	f.Add("test-prefix")
	f.Fuzz(func(t *testing.T, a string) {
		l.Prefix(a)

		if l.(*logger).prefix != a && a != "" {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed prefix mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				l.(*logger).prefix,
			)
			return
		}
	})
}

func FuzzLoggerSub(f *testing.F) {
	module := "Logger"
	funcname := "Sub()"

	l := New(WithSub("seed"))

	f.Add("test-sub-prefix")
	f.Fuzz(func(t *testing.T, a string) {
		l.Sub(a)

		if l.(*logger).sub != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed sub-prefix mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				l.(*logger).sub,
			)
			return
		}
	})
}

func TestLoggerFields(t *testing.T) {
	module := "Logger"
	funcname := "Fields()"

	type test struct {
		name  string
		init  map[string]interface{}
		input map[string]interface{}
		wants map[string]interface{}
	}

	l := New()

	var tests = []test{
		{
			name:  "set simple metadata",
			init:  nil,
			input: map[string]interface{}{"a": true},
			wants: map[string]interface{}{"a": true},
		},
		{
			name:  "replace simple metadata",
			init:  map[string]interface{}{"a": true},
			input: map[string]interface{}{"b": false},
			wants: map[string]interface{}{"b": false},
		},
		{
			name:  "reset simple metadata",
			init:  map[string]interface{}{"a": true},
			input: nil,
			wants: map[string]interface{}{},
		},
	}

	var init = func(test test) {
		if test.init == nil {
			return
		}

		l.Fields(test.init)
	}

	var reset = func() {
		l.Fields(nil)
	}

	var verify = func(idx int, test test) {
		reset()
		defer reset()

		init(test)

		l.Fields(test.input)

		if !reflect.DeepEqual(l.(*logger).meta, test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] fuzzed sub-prefix mismatch: wanted %s ; got %s -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				l.(*logger).sub,
				test.name,
			)
			return
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}

}

func TestLoggerIsSkipExit(t *testing.T) {
	module := "Logger"
	funcname := "IsSkipExit()"

	type test struct {
		name  string
		conf  []LoggerConfig
		wants bool
	}

	var tests = []test{
		{
			name:  "default config",
			wants: false,
		},
		{
			name:  "with SkipExit opt",
			conf:  []LoggerConfig{SkipExit},
			wants: true,
		},
	}

	var init = func(test test) Logger {
		return New(test.conf...)
	}

	var verify = func(idx int, test test) {
		l := init(test)

		skip := l.IsSkipExit()

		if skip != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] is-skip-exit mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				skip,
				test.name,
			)
			return
		}
		t.Logf(
			"#%v -- PASSED -- [%s] [%s] -- action: %s",
			idx,
			module,
			funcname,
			test.name,
		)
	}

	for idx, test := range tests {
		verify(idx, test)
	}

}

func TestLoggerWrite(t *testing.T) {
	module := "Logger"
	funcname := "Write()"

	type test struct {
		name string
		msg  []byte
	}

	var tests = []test{
		{
			name: "non-encoded message",
			msg:  []byte("null"),
		},
		{
			name: "non-encoded message",
			msg:  event.New().Message("null").Build().Encode(),
		},
	}

	buf := &bytes.Buffer{}

	l := New(WithOut(buf), SkipExit)

	var verify = func(idx int, test test) {
		buf.Reset()
		defer buf.Reset()

		n, err := l.Write(test.msg)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] error writing message: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		if n == 0 {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] zero-bytes written error -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		if n != buf.Len() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] invalid write length: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				buf.Len(),
				n,
				test.name,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [%s] [%s] -- action: %s",
			idx,
			module,
			funcname,
			test.name,
		)
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestNilLoggerSetOuts(t *testing.T) {
	module := "nilLogger"
	funcname := "SetOuts()"

	type test struct {
		name string
		w    []io.Writer
	}

	var bufs = []*bytes.Buffer{{}, {}, {}}

	var tests = []test{
		{
			name: "no writers",
			w:    []io.Writer{},
		},
		{
			name: "nil writers",
			w:    nil,
		},
		{
			name: "w/ writers",
			w:    []io.Writer{bufs[0], bufs[1], bufs[2]},
		},
	}

	var verify = func(idx int, test test) {
		nl := New(NilConfig)

		new := nl.SetOuts(test.w...)

		if !reflect.DeepEqual(new, nl) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				nl,
				new,
				test.name,
			)
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestNilLoggerAddOuts(t *testing.T) {
	module := "nilLogger"
	funcname := "AddOuts()"

	type test struct {
		name string
		w    []io.Writer
	}

	var bufs = []*bytes.Buffer{{}, {}, {}}

	var tests = []test{
		{
			name: "no writers",
			w:    []io.Writer{},
		},
		{
			name: "nil writers",
			w:    nil,
		},
		{
			name: "w/ writers",
			w:    []io.Writer{bufs[0], bufs[1], bufs[2]},
		},
	}

	var verify = func(idx int, test test) {
		nl := New(NilConfig)

		new := nl.AddOuts(test.w...)

		if !reflect.DeepEqual(new, nl) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				nl,
				new,
				test.name,
			)
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestNilLoggerPrefix(t *testing.T) {
	module := "nilLogger"
	funcname := "Prefix()"

	type test struct {
		name string
		p    string
	}

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
		nl := New(NilConfig)

		new := nl.Prefix(test.p)

		if !reflect.DeepEqual(new, nl) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				nl,
				new,
				test.name,
			)
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestNilLoggerSub(t *testing.T) {
	module := "nilLogger"
	funcname := "Sub()"

	type test struct {
		name string
		s    string
	}

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
		nl := New(NilConfig)

		new := nl.Sub(test.s)

		if !reflect.DeepEqual(new, nl) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				nl,
				new,
				test.name,
			)
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestNilLoggerFields(t *testing.T) {
	module := "nilLogger"
	funcname := "Fields()"

	type test struct {
		name string
		m    map[string]interface{}
	}

	var tests = []test{
		{
			name: "no input",
			m:    map[string]interface{}{},
		},
		{
			name: "nil input",
			m:    nil,
		},
		{
			name: "any input",
			m:    map[string]interface{}{"a": true},
		},
	}

	var verify = func(idx int, test test) {
		nl := New(NilConfig)

		new := nl.Fields(test.m)

		if !reflect.DeepEqual(new, nl) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				nl,
				new,
				test.name,
			)
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestNilLoggerIsSkipExit(t *testing.T) {
	module := "nilLogger"
	funcname := "IsSkipExit()"

	type test struct {
		name  string
		wants bool
	}

	var tests = []test{
		{
			name:  "default",
			wants: true,
		},
	}

	var verify = func(idx int, test test) {
		nl := New(NilConfig)

		ok := nl.IsSkipExit()

		if ok != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				ok,
				test.name,
			)
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestNilLoggerWrite(t *testing.T) {
	module := "nilLogger"
	funcname := "Write()"

	type test struct {
		name string
		b    []byte
	}

	var tests = []test{
		{
			name: "no input",
			b:    []byte{},
		},
		{
			name: "nil input",
			b:    nil,
		},
		{
			name: "event input",
			b:    event.New().Message("null").Build().Encode(),
		},
		{
			name: "byte message input",
			b:    []byte("null"),
		},
	}

	var verify = func(idx int, test test) {
		nl := New(NilConfig)

		n, err := nl.Write(test.b)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] write op returned an unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
		}

		if n != 1 {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] write op returned an unexpected write length: %v -- action: %s",
				idx,
				module,
				funcname,
				n,
				test.name,
			)
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
