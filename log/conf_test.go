package log

import (
	"bytes"
	"os"
	"testing"
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
				out:         os.Stdout,
				prefix:      "log",
				sub:         "",
				fmt:         TextFormat,
				skipExit:    false,
				levelFilter: 0,
			},
		},
		{
			conf: MultiConf(SkipExitCfg, JSONCfg, StdOutCfg),
			want: &LoggerBuilder{
				out:         os.Stdout,
				prefix:      "",
				sub:         "",
				fmt:         JSONFormat,
				skipExit:    true,
				levelFilter: 0,
			},
		},
		{
			conf: MultiConf(SkipExitCfg, InfoFilter, WithPrefix("test")),
			want: &LoggerBuilder{
				out:         nil,
				prefix:      "test",
				sub:         "",
				fmt:         nil,
				skipExit:    true,
				levelFilter: 2,
			},
		},
	}

	var verify = func(id int, test test, builder *LoggerBuilder) {
		if builder.out != test.want.out {
			t.Errorf(
				"#%v -- FAILED -- [Conf] MultiConf(...confs) -- mismatching outputs: got %v ; expected %v",
				id,
				builder.out,
				test.want.out,
			)
			return
		}

		if builder.prefix != test.want.prefix {
			t.Errorf(
				"#%v -- FAILED -- [Conf] MultiConf(...confs) -- mismatching prefixes: got %s ; expected %s",
				id,
				builder.prefix,
				test.want.prefix,
			)
			return
		}

		if builder.sub != test.want.sub {
			t.Errorf(
				"#%v -- FAILED -- [Conf] MultiConf(...confs) -- mismatching sub-prefixes: got %s ; expected %s",
				id,
				builder.sub,
				test.want.sub,
			)
			return
		}

		if builder.fmt != test.want.fmt {
			t.Errorf(
				"#%v -- FAILED -- [Conf] MultiConf(...confs) -- mismatching formats: got %v ; expected %v",
				id,
				builder.fmt,
				test.want.fmt,
			)
			return
		}

		if builder.skipExit != test.want.skipExit {
			t.Errorf(
				"#%v -- FAILED -- [Conf] MultiConf(...confs) -- mismatching skip-exit opts: got %v ; expected %v",
				id,
				builder.skipExit,
				test.want.skipExit,
			)
			return
		}

		if builder.levelFilter != test.want.levelFilter {
			t.Errorf(
				"#%v -- FAILED -- [Conf] MultiConf(...confs) -- mismatching level filters: got %v ; expected %v",
				id,
				builder.levelFilter,
				test.want.levelFilter,
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

func TestLCPrefix(t *testing.T) {
	type test struct {
		conf LoggerConfig
		want *LoggerBuilder
	}

	tests := []test{
		{
			conf: WithPrefix(""),
			want: &LoggerBuilder{
				prefix: "",
			},
		},
		{
			conf: WithPrefix("log"),
			want: &LoggerBuilder{
				prefix: "log",
			},
		},
		{
			conf: WithPrefix("test"),
			want: &LoggerBuilder{
				prefix: "test",
			},
		},
	}

	var verify = func(id int, test test, builder *LoggerBuilder) {

		if builder.prefix != test.want.prefix {
			t.Errorf(
				"#%v -- FAILED -- [Conf] WithPrefix(prefix) -- mismatching prefixes: got %s ; expected %s",
				id,
				builder.prefix,
				test.want.prefix,
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
				sub: "",
			},
		},
		{
			conf: WithSub("log"),
			want: &LoggerBuilder{
				sub: "log",
			},
		},
		{
			conf: WithSub("test"),
			want: &LoggerBuilder{
				sub: "test",
			},
		},
	}

	var verify = func(id int, test test, builder *LoggerBuilder) {

		if builder.sub != test.want.sub {
			t.Errorf(
				"#%v -- FAILED -- [Conf] WithSub(sub) -- mismatching sub-prefixes: got %s ; expected %s",
				id,
				builder.prefix,
				test.want.prefix,
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
				out: os.Stdout,
			},
		},
		{
			conf: WithOut(os.Stderr),
			want: &LoggerBuilder{
				out: os.Stderr,
			},
		},
		{
			conf: WithOut(buf),
			want: &LoggerBuilder{
				out: buf,
			},
		},
	}

	var verify = func(id int, test test, builder *LoggerBuilder) {

		if builder.out != test.want.out {
			t.Errorf(
				"#%v -- FAILED -- [Conf] WithOut(...outs) -- mismatching outputs: got %v ; expected %v",
				id,
				builder.out,
				test.want.out,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [Conf] WithOut(...outs)",
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
			conf: SkipExitCfg,
			want: &LoggerBuilder{
				skipExit: true,
			},
		},
		{
			conf: MultiConf(),
			want: &LoggerBuilder{
				skipExit: false,
			},
		},
		{
			conf: defaultConfig,
			want: &LoggerBuilder{
				skipExit: false,
			},
		},
	}

	var verify = func(id int, test test, builder *LoggerBuilder) {

		if builder.skipExit != test.want.skipExit {
			t.Errorf(
				"#%v -- FAILED -- [Conf] SkipExit() -- mismatching skip-exit opts: got %v ; expected %v",
				id,
				builder.skipExit,
				test.want.skipExit,
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
				levelFilter: 0,
			},
		},
		{
			conf: WithFilter(5),
			want: &LoggerBuilder{
				levelFilter: 5,
			},
		},
		{
			conf: WithFilter(9),
			want: &LoggerBuilder{
				levelFilter: 9,
			},
		},
		{
			conf: InfoFilter,
			want: &LoggerBuilder{
				levelFilter: 2,
			},
		},
		{
			conf: WarnFilter,
			want: &LoggerBuilder{
				levelFilter: 3,
			},
		},
		{
			conf: ErrorFilter,
			want: &LoggerBuilder{
				levelFilter: 4,
			},
		},
		{
			conf: DefaultCfg,
			want: &LoggerBuilder{
				levelFilter: 0,
			},
		},
	}

	var verify = func(id int, test test, builder *LoggerBuilder) {

		if builder.levelFilter != test.want.levelFilter {
			t.Errorf(
				"#%v -- FAILED -- [Conf] LevelFilter(level) -- mismatching level filters: got %v ; expected %v",
				id,
				builder.levelFilter,
				test.want.levelFilter,
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
