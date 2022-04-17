package log

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/log/format/json"
	"github.com/zalgonoise/zlog/store"
)

func TestMultiConf(t *testing.T) {
	type test struct {
		conf LoggerConfig
		want *LoggerBuilder
	}

	tests := []test{
		{
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

	var verify = func(id int, test test, builder *LoggerBuilder) {
		if builder.Out != test.want.Out {
			t.Errorf(
				"#%v -- FAILED -- [Conf] MultiConf(...confs) -- mismatching outputs: got %v ; expected %v",
				id,
				builder.Out,
				test.want.Out,
			)
			return
		}

		if builder.Prefix != test.want.Prefix {
			t.Errorf(
				"#%v -- FAILED -- [Conf] MultiConf(...confs) -- mismatching prefixes: got %s ; expected %s",
				id,
				builder.Prefix,
				test.want.Prefix,
			)
			return
		}

		if builder.Sub != test.want.Sub {
			t.Errorf(
				"#%v -- FAILED -- [Conf] MultiConf(...confs) -- mismatching sub-prefixes: got %s ; expected %s",
				id,
				builder.Sub,
				test.want.Sub,
			)
			return
		}

		if builder.Fmt != test.want.Fmt {
			t.Errorf(
				"#%v -- FAILED -- [Conf] MultiConf(...confs) -- mismatching formats: got %v ; expected %v",
				id,
				builder.Fmt,
				test.want.Fmt,
			)
			return
		}

		if builder.SkipExit != test.want.SkipExit {
			t.Errorf(
				"#%v -- FAILED -- [Conf] MultiConf(...confs) -- mismatching skip-exit opts: got %v ; expected %v",
				id,
				builder.SkipExit,
				test.want.SkipExit,
			)
			return
		}

		if builder.LevelFilter != test.want.LevelFilter {
			t.Errorf(
				"#%v -- FAILED -- [Conf] MultiConf(...confs) -- mismatching level filters: got %v ; expected %v",
				id,
				builder.LevelFilter,
				test.want.LevelFilter,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [Conf] MultiConf(...confs)",
			id,
		)
	}

	for id, test := range tests {
		builder := &LoggerBuilder{}

		MultiConf(test.conf).Apply(builder)

		verify(id, test, builder)

	}

}

func TestNilLogger(t *testing.T) {
	module := "LoggerConfig"
	funcname := "NilLogger()"

	type test struct {
		input []LoggerConfig
		wants Logger
	}

	var tests = []test{
		{
			input: []LoggerConfig{
				WithOut(),
				WithPrefix("test"),
				WithSub("new"),
				WithFormat(FormatJSON),
			},
			wants: &logger{
				out:         os.Stderr,
				prefix:      "test",
				sub:         "new",
				fmt:         &json.FmtJSON{},
				skipExit:    false,
				levelFilter: 0,
			},
		},
		{
			input: []LoggerConfig{
				NilLogger(),
			},
			wants: &nilLogger{},
		},
	}

	var verify = func(id int, test test, input Logger) {
		switch v := input.(type) {
		case *logger:
			if !reflect.DeepEqual(v, test.wants.(*logger)) {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- logger mismatch: wanted %v ; got %v",
					id,
					module,
					funcname,
					test.wants.(*logger),
					v,
				)
				return
			}

		case *nilLogger:
			var checkWriter bool = false
			var checkExit bool = false
			var isNil bool = false

			for _, v := range test.input {
				if multi, ok := v.(*multiconf); ok {
					for _, conf := range multi.confs {
						if out, ok := conf.(*LCOut); ok && out.out == store.EmptyWriter {
							checkWriter = true
						}
						if _, ok := conf.(*LCSkipExit); ok {
							checkExit = true
						}
					}
				}
			}

			_, isNil = input.(*nilLogger)

			if !checkWriter || !checkExit || !isNil {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- nilLogger checks failed: check writer: %v ; check exit: %v ; check nilLogger type: %v",
					id,
					module,
					funcname,
					checkWriter,
					checkExit,
					isNil,
				)
				return
			}
		}
	}

	for id, test := range tests {
		logger := New(test.input...)

		verify(id, test, logger)

	}
}

func TestLCPrefix(t *testing.T) {
	type test struct {
		conf LoggerConfig
		want *LoggerBuilder
	}

	tests := []test{
		{
			conf: WithPrefix(""),
			want: &LoggerBuilder{
				Prefix: "",
			},
		},
		{
			conf: WithPrefix("log"),
			want: &LoggerBuilder{
				Prefix: "log",
			},
		},
		{
			conf: WithPrefix("test"),
			want: &LoggerBuilder{
				Prefix: "test",
			},
		},
	}

	var verify = func(id int, test test, builder *LoggerBuilder) {

		if builder.Prefix != test.want.Prefix {
			t.Errorf(
				"#%v -- FAILED -- [Conf] WithPrefix(prefix) -- mismatching prefixes: got %s ; expected %s",
				id,
				builder.Prefix,
				test.want.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [Conf] WithPrefix(prefix)",
			id,
		)
	}

	for id, test := range tests {
		builder := &LoggerBuilder{}

		test.conf.Apply(builder)

		verify(id, test, builder)
	}
}

func TestLCSub(t *testing.T) {
	type test struct {
		conf LoggerConfig
		want *LoggerBuilder
	}

	tests := []test{
		{
			conf: WithSub(""),
			want: &LoggerBuilder{
				Sub: "",
			},
		},
		{
			conf: WithSub("log"),
			want: &LoggerBuilder{
				Sub: "log",
			},
		},
		{
			conf: WithSub("test"),
			want: &LoggerBuilder{
				Sub: "test",
			},
		},
	}

	var verify = func(id int, test test, builder *LoggerBuilder) {

		if builder.Sub != test.want.Sub {
			t.Errorf(
				"#%v -- FAILED -- [Conf] WithSub(sub) -- mismatching sub-prefixes: got %s ; expected %s",
				id,
				builder.Prefix,
				test.want.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [Conf] WithSub(sub)",
			id,
		)
	}

	for id, test := range tests {
		builder := &LoggerBuilder{}

		test.conf.Apply(builder)

		verify(id, test, builder)
	}
}

func TestLCOut(t *testing.T) {
	type test struct {
		conf LoggerConfig
		want *LoggerBuilder
	}

	buf := &bytes.Buffer{}

	tests := []test{
		{
			conf: WithOut(),
			want: &LoggerBuilder{
				Out: os.Stderr,
			},
		},
		{
			conf: WithOut(os.Stderr),
			want: &LoggerBuilder{
				Out: os.Stderr,
			},
		},
		{
			conf: WithOut(buf),
			want: &LoggerBuilder{
				Out: buf,
			},
		},
	}

	var verify = func(id int, test test, builder *LoggerBuilder) {

		if builder.Out != test.want.Out {
			t.Errorf(
				"#%v -- FAILED -- [Conf] WithOut(...Outs) -- mismatching outputs: got %v ; expected %v",
				id,
				builder.Out,
				test.want.Out,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [Conf] WithOut(...Outs)",
			id,
		)
	}

	for id, test := range tests {
		builder := &LoggerBuilder{}

		test.conf.Apply(builder)

		verify(id, test, builder)
	}
}

func TestLCSkipExit(t *testing.T) {
	type test struct {
		conf LoggerConfig
		want *LoggerBuilder
	}

	tests := []test{
		{
			conf: SkipExit,
			want: &LoggerBuilder{
				SkipExit: true,
			},
		},
		{
			conf: MultiConf(),
			want: &LoggerBuilder{
				SkipExit: false,
			},
		},
		{
			conf: DefaultConfig,
			want: &LoggerBuilder{
				SkipExit: false,
			},
		},
	}

	var verify = func(id int, test test, builder *LoggerBuilder) {

		if builder.SkipExit != test.want.SkipExit {
			t.Errorf(
				"#%v -- FAILED -- [Conf] SkipExit() -- mismatching skip-exit opts: got %v ; expected %v",
				id,
				builder.SkipExit,
				test.want.SkipExit,
			)
		}

		t.Logf(
			"#%v -- PASSED -- [Conf] SkipExit()",
			id,
		)
	}

	for id, test := range tests {
		builder := &LoggerBuilder{}

		test.conf.Apply(builder)

		verify(id, test, builder)
	}
}

func TestLCFilter(t *testing.T) {
	type test struct {
		conf LoggerConfig
		want *LoggerBuilder
	}

	tests := []test{
		{
			conf: WithFilter(0),
			want: &LoggerBuilder{
				LevelFilter: 0,
			},
		},
		{
			conf: WithFilter(5),
			want: &LoggerBuilder{
				LevelFilter: 5,
			},
		},
		{
			conf: WithFilter(9),
			want: &LoggerBuilder{
				LevelFilter: 9,
			},
		},
		{
			conf: FilterInfo,
			want: &LoggerBuilder{
				LevelFilter: 2,
			},
		},
		{
			conf: FilterWarn,
			want: &LoggerBuilder{
				LevelFilter: 3,
			},
		},
		{
			conf: FilterError,
			want: &LoggerBuilder{
				LevelFilter: 4,
			},
		},
		{
			conf: DefaultCfg,
			want: &LoggerBuilder{
				LevelFilter: 0,
			},
		},
	}

	var verify = func(id int, test test, builder *LoggerBuilder) {

		if builder.LevelFilter != test.want.LevelFilter {
			t.Errorf(
				"#%v -- FAILED -- [Conf] LevelFilter(level) -- mismatching level filters: got %v ; expected %v",
				id,
				builder.LevelFilter,
				test.want.LevelFilter,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [Conf] LevelFilter(level)",
			id,
		)
	}

	for id, test := range tests {
		builder := &LoggerBuilder{}

		test.conf.Apply(builder)

		verify(id, test, builder)
	}
}
