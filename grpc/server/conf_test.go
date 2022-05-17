package server

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/zalgonoise/zlog/log"
	"google.golang.org/grpc"
)

func TestMultiConf(t *testing.T) {
	module := "LogServerConfig"
	funcname := "MultiConf()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		cfg   []LogServerConfig
		wants LogServerConfig
	}

	var tests = []test{
		{
			name:  "default config, no input",
			cfg:   []LogServerConfig{},
			wants: defaultConfig,
		},
		{
			name: "one config as input",
			cfg: []LogServerConfig{
				WithAddr("127.0.0.1:9099"),
			},
			wants: WithAddr("127.0.0.1:9099"),
		},
		{
			name: "multiple config as input",
			cfg: []LogServerConfig{
				WithAddr("127.0.0.1:9099"),
				WithLogger(log.New(log.NilConfig)),
			},
			wants: &multiconf{confs: []LogServerConfig{
				WithAddr("127.0.0.1:9099"),
				WithLogger(log.New(log.NilConfig)),
			}},
		},
		{
			name: "multiple config as input, with nil values",
			cfg: []LogServerConfig{
				nil,
				WithAddr("127.0.0.1:9099"),
				nil,
				nil,
				WithLogger(log.New(log.NilConfig)),
			},
			wants: &multiconf{confs: []LogServerConfig{
				WithAddr("127.0.0.1:9099"),
				WithLogger(log.New(log.NilConfig)),
			}},
		},
		{
			name: "multiple config as input, with nil values, only one valid config",
			cfg: []LogServerConfig{
				nil,
				WithAddr("127.0.0.1:9099"),
				nil,
				nil,
			},
			wants: WithAddr("127.0.0.1:9099"),
		},
		{
			name: "multiple config as input, all nil values",
			cfg: []LogServerConfig{
				nil,
				nil,
				nil,
			},
			wants: defaultConfig,
		},
		{
			name:  "nil input",
			cfg:   nil,
			wants: defaultConfig,
		},
	}

	var verify = func(idx int, test test) {
		cfg := MultiConf(test.cfg...)

		if !reflect.DeepEqual(cfg, test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				cfg,
				test.name,
			)
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestWithAddr(t *testing.T) {
	module := "LogServerConfig"
	funcname := "WithAddr()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		input string
		wants string
	}

	var tests = []test{
		{
			name:  "valid input",
			input: "127.0.0.1:9099",
			wants: "127.0.0.1:9099",
		},
		{
			name:  "empty input",
			input: "",
			wants: ":9099",
		},
		{
			name:  "invalid input",
			input: ":",
			wants: ":9099",
		},
	}

	var verify = func(idx int, test test) {
		cfg := WithAddr(test.input)

		if cfg.(*LSAddr).addr != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				cfg,
				test.name,
			)
			return
		}

		var builder = new(gRPCLogServerBuilder)

		cfg.Apply(builder)

		if builder.addr != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				builder.addr,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestWithLogger(t *testing.T) {
	module := "LogServerConfig"
	funcname := "WithLogger()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		input []log.Logger
		wants log.Logger
	}

	var bufs = []*bytes.Buffer{{}, {}}

	var loggers = []log.Logger{
		log.New(log.WithOut(bufs[0]), log.SkipExit),
		log.New(log.WithOut(bufs[1]), log.SkipExit),
	}

	var tests = []test{
		{
			name:  "one logger as input",
			input: []log.Logger{loggers[0]},
			wants: loggers[0],
		},
		{
			name:  "multiple loggers as input",
			input: []log.Logger{loggers[0], loggers[1]},
			wants: log.MultiLogger(loggers[0], loggers[1]),
		},
		{
			name:  "no input",
			input: []log.Logger{},
			wants: log.New(),
		},
		{
			name:  "nil input",
			input: nil,
			wants: log.New(),
		},
	}

	var verify = func(idx int, test test) {
		cfg := WithLogger(test.input...)

		if fmt.Sprintf("%T", cfg.(*LSLogger).logger) != fmt.Sprintf("%T", test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- config mismatch error: wanted %v of type %T ; got %v of type %T -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				test.wants,
				cfg.(*LSLogger).logger,
				cfg.(*LSLogger).logger,
				test.name,
			)
			return
		}

		var builder = new(gRPCLogServerBuilder)

		cfg.Apply(builder)

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestWithServiceLogger(t *testing.T) {
	module := "LogServerConfig"
	funcname := "WithServiceLogger()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		input []log.Logger
		wants log.Logger
	}

	var bufs = []*bytes.Buffer{{}, {}}

	var loggers = []log.Logger{
		log.New(log.WithOut(bufs[0]), log.SkipExit),
		log.New(log.WithOut(bufs[1]), log.SkipExit),
	}

	var tests = []test{
		{
			name:  "one logger as input",
			input: []log.Logger{loggers[0]},
			wants: loggers[0],
		},
		{
			name:  "multiple loggers as input",
			input: []log.Logger{loggers[0], loggers[1]},
			wants: log.MultiLogger(loggers[0], loggers[1]),
		},
		{
			name:  "no input",
			input: []log.Logger{},
			wants: log.New(log.NilConfig),
		},
		{
			name:  "nil input",
			input: nil,
			wants: log.New(log.NilConfig),
		},
	}

	var verify = func(idx int, test test) {
		cfg := WithServiceLogger(test.input...)

		if fmt.Sprintf("%T", cfg.(*LSServiceLogger).logger) != fmt.Sprintf("%T", test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- config mismatch error: wanted %v of type %T ; got %v of type %T -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				test.wants,
				cfg.(*LSServiceLogger).logger,
				cfg.(*LSServiceLogger).logger,
				test.name,
			)
			return
		}

		var builder = new(gRPCLogServerBuilder)

		cfg.Apply(builder)

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestWithServiceLoggerV(t *testing.T) {
	module := "LogServerConfig"
	funcname := "WithServiceLoggerV()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		input []log.Logger
		wants log.Logger
	}

	var bufs = []*bytes.Buffer{{}, {}}

	var loggers = []log.Logger{
		log.New(log.WithOut(bufs[0]), log.SkipExit),
		log.New(log.WithOut(bufs[1]), log.SkipExit),
	}

	var tests = []test{
		{
			name:  "one logger as input",
			input: []log.Logger{loggers[0]},
			wants: loggers[0],
		},
		{
			name:  "multiple loggers as input",
			input: []log.Logger{loggers[0], loggers[1]},
			wants: log.MultiLogger(loggers[0], loggers[1]),
		},
		{
			name:  "no input",
			input: []log.Logger{},
			wants: log.New(log.NilConfig),
		},
		{
			name:  "nil input",
			input: nil,
			wants: log.New(log.NilConfig),
		},
	}

	var verify = func(idx int, test test) {
		cfg := WithServiceLoggerV(test.input...)

		if fmt.Sprintf("%T", cfg.(*LSServiceLogger).logger) != fmt.Sprintf("%T", test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- config mismatch error: wanted %v of type %T ; got %v of type %T -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				test.wants,
				cfg.(*LSServiceLogger).logger,
				cfg.(*LSServiceLogger).logger,
				test.name,
			)
			return
		}

		if !cfg.(*LSServiceLogger).verbose {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- expected verbose logger -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		var builder = &gRPCLogServerBuilder{
			interceptors: serverInterceptors{
				streamItcp: make(map[string]grpc.StreamServerInterceptor),
				unaryItcp:  make(map[string]grpc.UnaryServerInterceptor),
			},
		}

		cfg.Apply(builder)

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestWithTiming(t *testing.T) {
	module := "LogServerConfig"
	funcname := "WithTiming()"

	_ = module
	_ = funcname

	type test struct {
		name       string
		entrypoint LogServerConfig
	}

	var bufs = []*bytes.Buffer{{}}

	var loggers = []log.Logger{
		log.New(log.WithOut(bufs[0]), log.SkipExit),
	}

	var tests = []test{
		{
			name: "WithTiming() execution test",
		},
		{
			name:       "WithTiming() execution test, with logger",
			entrypoint: WithServiceLoggerV(loggers[0]),
		},
	}

	var verify = func(idx int, test test) {
		cfg := WithTiming()

		var builder = &gRPCLogServerBuilder{
			interceptors: serverInterceptors{
				streamItcp: make(map[string]grpc.StreamServerInterceptor),
				unaryItcp:  make(map[string]grpc.UnaryServerInterceptor),
			},
		}

		fmt.Println(test.entrypoint)

		if test.entrypoint != nil {
			test.entrypoint.Apply(builder)

			if builder.interceptors.streamItcp["logging"] == nil || builder.interceptors.unaryItcp["logging"] == nil {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] missing logging interceptors -- action: %s",
					idx,
					module,
					funcname,
					test.name,
				)
				return
			}
		}

		cfg.Apply(builder)

		if test.entrypoint == nil &&
			(builder.interceptors.streamItcp["timing"] == nil || builder.interceptors.unaryItcp["timing"] == nil) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] missing timing interceptors -- action: %s",
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

func TestWithGRPCOpts(t *testing.T) {
	module := "LogServerConfig"
	funcname := "WithGRPCOpts()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		input []grpc.ServerOption
		wants []grpc.ServerOption
	}

	var cfgs = []grpc.ServerOption{
		grpc.ConnectionTimeout(time.Second * 3),
		grpc.ConnectionTimeout(time.Second * 10),
	}

	var tests = []test{
		{
			name:  "one option input",
			input: []grpc.ServerOption{cfgs[0]},
			wants: []grpc.ServerOption{cfgs[0]},
		},
		{
			name: "multiple options input",
			input: []grpc.ServerOption{
				cfgs[0],
				cfgs[1],
			},
			wants: []grpc.ServerOption{
				cfgs[0],
				cfgs[1],
			},
		},
		{
			name: "multiple options with nils",
			input: []grpc.ServerOption{
				nil,
				cfgs[0],
				nil,
				nil,
				cfgs[1],
				nil,
			},
			wants: []grpc.ServerOption{
				cfgs[0],
				cfgs[1],
			},
		},
		{
			name: "multiple options with all nils",
			input: []grpc.ServerOption{
				nil,
				nil,
				nil,
			},
			wants: []grpc.ServerOption{},
		},
		{
			name:  "nil input",
			input: nil,
			wants: []grpc.ServerOption{},
		},
	}

	var verify = func(idx int, test test) {
		cfg := WithGRPCOpts(test.input...)

		for iidx, c := range cfg.(*LSOpts).opts {
			if !reflect.DeepEqual(c, test.wants[iidx]) {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- config mismatch error in index %v: wanted %v ; got %v -- action: %s",
					idx,
					module,
					funcname,
					iidx,
					test.wants,
					cfg.(*LSOpts).opts,
					test.name,
				)
				return
			}
		}

		var builder = new(gRPCLogServerBuilder)

		cfg.Apply(builder)

		for iidx, c := range builder.opts {
			if !reflect.DeepEqual(c, test.wants[iidx]) {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- output mismatch error: wanted %v ; got %v -- action: %s",
					idx,
					module,
					funcname,
					test.wants,
					builder.addr,
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
