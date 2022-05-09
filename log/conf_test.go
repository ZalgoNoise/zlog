package log

import (
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/log/format/protobuf"
	"github.com/zalgonoise/zlog/store/db"
)

func TestMultiConf(t *testing.T) {
	module := "LoggerConfig"
	funcname := "MultiConf()"

	_ = module
	_ = funcname

	type test struct {
		name string
		conf LoggerConfig
		want *LoggerBuilder
	}

	tests := []test{
		{
			name: "default MultiConf()",
			conf: MultiConf(),
			want: &LoggerBuilder{
				Out:         os.Stderr,
				Prefix:      "log",
				Sub:         "",
				Fmt:         TextColorLevelFirst,
				SkipExit:    false,
				LevelFilter: 0,
			},
		},
		{
			name: "MultiConf() w/ SkipExit, JSON format, and StdOut config",
			conf: MultiConf(SkipExit, WithFormat(FormatJSON), StdOut),
			want: &LoggerBuilder{
				Out:         os.Stderr,
				Prefix:      "",
				Sub:         "",
				Fmt:         FormatJSON,
				SkipExit:    true,
				LevelFilter: 0,
			},
		},
		{
			name: "MultiConf() w/ SkipExit, Level filter, and custom prefix",
			conf: MultiConf(SkipExit, FilterInfo, WithPrefix("test")),
			want: &LoggerBuilder{
				Out:         nil,
				Prefix:      "test",
				Sub:         "",
				Fmt:         nil,
				SkipExit:    true,
				LevelFilter: 2,
			},
		},
	}

	var init = func(test test) *LoggerBuilder {
		builder := &LoggerBuilder{}

		MultiConf(test.conf).Apply(builder)

		return builder
	}

	var verify = func(idx int, test test) {
		builder := init(test)

		if !reflect.DeepEqual(*builder, *test.want) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				*test.want,
				*builder,
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

func TestNilLogger(t *testing.T) {
	module := "LoggerConfig"
	funcname := "NilLogger()"

	type test struct {
		name  string
		input []LoggerConfig
		wants Logger
	}

	var tests = []test{
		{
			name: "test nil logger config routine",
			input: []LoggerConfig{
				NilLogger(),
			},
			wants: &nilLogger{},
		},
	}

	var init = func(test test) Logger {
		return New(test.input...)
	}

	var verify = func(idx int, test test) {
		input := init(test)

		if !reflect.DeepEqual(*input.(*nilLogger), *test.wants.(*nilLogger)) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				*test.wants.(*nilLogger),
				*input.(*nilLogger),
				test.wants,
			)
		}
		t.Logf(
			"#%v -- PASSED -- [%s] [%s] -- action: %s",
			idx,
			module,
			funcname,
			test.wants,
		)
	}

	for idx, test := range tests {

		verify(idx, test)

	}
}

func FuzzPrefix(f *testing.F) {
	module := "LoggerConfig"
	funcname := "WithPrefix()"

	f.Add("test-prefix")
	f.Fuzz(func(t *testing.T, a string) {
		e := WithPrefix(a)

		builder := &LoggerBuilder{}

		e.Apply(builder)

		if builder.Prefix != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed prefix mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				builder.Prefix,
			)
		}
	})
}

func FuzzSub(f *testing.F) {
	module := "LoggerConfig"
	funcname := "WithSub()"

	f.Add("test-sub")
	f.Fuzz(func(t *testing.T, a string) {
		e := WithSub(a)

		builder := &LoggerBuilder{}

		e.Apply(builder)

		if builder.Sub != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed sub-prefix mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				builder.Sub,
			)
		}
	})
}

func TestOut(t *testing.T) {
	module := "LoggerConfig"
	funcname := "WithOut()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		outs  []io.Writer
		wants *LCOut
	}

	var tests = []test{
		{
			name:  "test defaults",
			outs:  []io.Writer{},
			wants: &LCOut{out: os.Stderr},
		},
		{
			name:  "test single writer",
			outs:  []io.Writer{os.Stdout},
			wants: &LCOut{out: os.Stdout},
		},
		{
			name:  "test multi writers",
			outs:  []io.Writer{os.Stdout, os.Stderr},
			wants: &LCOut{out: io.MultiWriter(os.Stdout, os.Stderr)},
		},
	}

	var init = func(test test) LoggerConfig {
		return WithOut(test.outs...)
	}
	var verify = func(idx int, test test) {
		conf := init(test)

		if !reflect.DeepEqual(conf.(*LCOut).out, test.wants.out) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				*test.wants,
				*conf.(*LCOut),
				test.name,
			)
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

func TestSkipExit(t *testing.T) {
	module := "LoggerConfig"
	funcname := "SkipExit"

	_ = module
	_ = funcname

	type test struct {
		name string
		conf LoggerConfig
		want bool
	}

	tests := []test{
		{
			name: "SkipExit config",
			conf: SkipExit,
			want: true,
		},
		{
			name: "default config",
			conf: MultiConf(),
			want: false,
		},
	}

	var init = func(test test) *LoggerBuilder {
		builder := &LoggerBuilder{}

		test.conf.Apply(builder)

		return builder
	}

	var verify = func(idx int, test test) {
		builder := init(test)

		if builder.SkipExit != test.want {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.want,
				builder.SkipExit,
				test.name,
			)
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

func TestFilter(t *testing.T) {
	module := "LoggerConfig"
	funcname := "WithFilter()"

	_ = module
	_ = funcname

	type test struct {
		name string
		conf LoggerConfig
		want event.Level
	}

	var tests = []test{
		{
			name: "with level by number",
			conf: WithFilter(3),
			want: event.Level_warn,
		},
		{
			name: "with level by reference",
			conf: WithFilter(event.Level_warn),
			want: event.Level_warn,
		},
	}

	var init = func(test test) *LoggerBuilder {
		builder := &LoggerBuilder{}

		test.conf.Apply(builder)

		return builder
	}

	var verify = func(idx int, test test) {
		builder := init(test)

		if builder.LevelFilter != test.want.Int() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %s ; got %s -- action: %s",
				idx,
				module,
				funcname,
				test.want.String(),
				event.Level_name[builder.LevelFilter],
				test.name,
			)
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

func TestWithDatabase(t *testing.T) {
	module := "LoggerConfig"
	funcname := "WithDatabase()"

	type test struct {
		name  string
		w     []io.WriteCloser
		wants LoggerConfig
	}

	var testWCs = []*testWC{{}, {}}

	var tests = []test{
		{
			name:  "empty slice",
			w:     []io.WriteCloser{},
			wants: nil,
		},
		{
			name:  "nil input",
			w:     nil,
			wants: nil,
		},
		{
			name: "one WriteCloser",
			w:    []io.WriteCloser{testWCs[0]},
			wants: &LCDatabase{
				Out: testWCs[0],
				Fmt: &protobuf.FmtPB{},
			},
		},
		{
			name: "multiple WriteClosers",
			w:    []io.WriteCloser{testWCs[0], testWCs[0]},
			wants: &LCDatabase{
				Out: db.MultiWriteCloser(testWCs[0], testWCs[1]),
				Fmt: &protobuf.FmtPB{},
			},
		},
	}

	var verify = func(idx int, test test) {
		var conf LoggerConfig
		if test.w == nil {
			conf = WithDatabase(nil)
		} else {
			conf = WithDatabase(test.w...)
		}

		if !reflect.DeepEqual(conf, test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				conf,
				test.name,
			)
			return
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
