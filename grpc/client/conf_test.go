package client

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/log"
	"google.golang.org/grpc"
)

func TestMultiConf(t *testing.T) {
	module := "LogClientConfig"
	funcname := "MultiConf()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		cfg   []LogClientConfig
		wants LogClientConfig
	}

	var tests = []test{
		{
			name:  "default config, no input",
			cfg:   []LogClientConfig{},
			wants: defaultConfig,
		},
		{
			name: "one config as input",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
			},
			wants: WithAddr("127.0.0.1:9099"),
		},
		{
			name: "multiple config as input",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
				WithLogger(log.New(log.NilConfig)),
			},
			wants: &multiconf{confs: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
				WithLogger(log.New(log.NilConfig)),
			}},
		},
		{
			name: "multiple config as input, with nil values",
			cfg: []LogClientConfig{
				nil,
				WithAddr("127.0.0.1:9099"),
				nil,
				nil,
				WithLogger(log.New(log.NilConfig)),
			},
			wants: &multiconf{confs: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
				WithLogger(log.New(log.NilConfig)),
			}},
		},
		{
			name: "multiple config as input, with nil values, only one valid config",
			cfg: []LogClientConfig{
				nil,
				WithAddr("127.0.0.1:9099"),
				nil,
				nil,
			},
			wants: WithAddr("127.0.0.1:9099"),
		},
		{
			name: "multiple config as input, all nil values",
			cfg: []LogClientConfig{
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

		_, _ = New(cfg)
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestWithAddr(t *testing.T) {
	module := "LogClientConfig"
	funcname := "WithAddr()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		input []string
		wants []string
	}

	var tests = []test{
		{
			name:  "with one valid address",
			input: []string{"127.0.0.1:9099"},
			wants: []string{"127.0.0.1:9099"},
		},
		{
			name: "with multiple valid addresses",
			input: []string{
				"127.0.0.1:9098",
				"127.0.0.1:9099",
			},
			wants: []string{
				"127.0.0.1:9098",
				"127.0.0.1:9099",
			},
		},
		{
			name:  "with one empty address",
			input: []string{""},
			wants: []string{":9099"},
		},
		{
			name: "with multiple addresses, some empty",
			input: []string{
				"",
				"127.0.0.1:9098",
				"",
				"127.0.0.1:9099",
				"",
			},
			wants: []string{
				"127.0.0.1:9098",
				"127.0.0.1:9099",
			},
		},
		{
			name: "with multiple addresses, all empty",
			input: []string{
				"",
				"",
				"",
			},
			wants: []string{":9099"},
		},
		{
			name:  "with nil input",
			input: nil,
			wants: []string{":9099"},
		},
	}

	var verify = func(idx int, test test) {
		var cfg = WithAddr(test.input...)

		config := cfg.(*LSAddr)
		keys := config.addr.Keys()

		for _, k := range keys {
			var ok bool
			for _, v := range test.wants {
				if k == v {
					ok = true
					break
				}
			}
			if !ok {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] keys do not match the expected values: key %s ; expected range: %v -- action: %s",
					idx,
					module,
					funcname,
					k,
					test.wants,
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

func TestStreamOrUnaryRPC(t *testing.T) {
	module := "LogClientConfig"
	funcname := "StreamRPC() / UnaryRPC()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		call  func() LogClientConfig
		wants bool
	}

	var tests = []test{
		{
			name:  "Unary RPC call",
			call:  UnaryRPC,
			wants: true,
		},
		{
			name:  "Stream RPC call",
			call:  StreamRPC,
			wants: false,
		},
	}

	var verify = func(idx int, test test) {
		var cfg = test.call()

		config := cfg.(*LSType)

		if config.isUnary != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] type config's isUnary property value mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				config.isUnary,
				test.wants,
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
	module := "LogClientConfig"
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
		log.New(log.NilConfig),
	}

	var tests = []test{
		{
			name:  "one valid logger",
			input: []log.Logger{loggers[0]},
			wants: loggers[0],
		},
		{
			name:  "several valid loggers",
			input: []log.Logger{loggers[0], loggers[1]},
			wants: log.MultiLogger(loggers[0], loggers[1]),
		},
		{
			name:  "no loggers as input",
			input: []log.Logger{},
			wants: log.New(log.NilConfig),
		},
		{
			name:  "nil value as input",
			input: []log.Logger{nil},
			wants: log.New(log.NilConfig),
		},
		{
			name: "several loggers, mixed in with nil values",
			input: []log.Logger{
				nil,
				loggers[0],
				nil,
				loggers[1],
				nil,
			},
			wants: log.MultiLogger(loggers[0], loggers[1]),
		},
		{
			name: "several nil values",
			input: []log.Logger{
				nil,
				nil,
				nil,
			},
			wants: log.MultiLogger(loggers[2]),
		},
		{
			name:  "nil input",
			input: nil,
			wants: log.MultiLogger(loggers[2]),
		},
	}

	var verify = func(idx int, test test) {
		var cfg = WithLogger(test.input...)

		config := cfg.(*LSLogger)

		if !reflect.DeepEqual(config.logger, test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				config.logger,
				test.name,
			)
			return
		}

		if config.verbose {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] expected a non-verbose logger, verbose attribute was true -- action: %s",
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

func TestWithLoggerV(t *testing.T) {
	module := "LogClientConfig"
	funcname := "WithLoggerV()"

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
		log.New(log.NilConfig),
	}

	var tests = []test{
		{
			name:  "one valid logger",
			input: []log.Logger{loggers[0]},
			wants: loggers[0],
		},
		{
			name:  "several valid loggers",
			input: []log.Logger{loggers[0], loggers[1]},
			wants: log.MultiLogger(loggers[0], loggers[1]),
		},
		{
			name:  "no loggers as input",
			input: []log.Logger{},
			wants: log.New(log.NilConfig),
		},
		{
			name:  "nil value as input",
			input: []log.Logger{nil},
			wants: log.New(log.NilConfig),
		},
		{
			name: "several loggers, mixed in with nil values",
			input: []log.Logger{
				nil,
				loggers[0],
				nil,
				loggers[1],
				nil,
			},
			wants: log.MultiLogger(loggers[0], loggers[1]),
		},
		{
			name: "several nil values",
			input: []log.Logger{
				nil,
				nil,
				nil,
			},
			wants: log.MultiLogger(loggers[2]),
		},
		{
			name:  "nil input",
			input: nil,
			wants: log.MultiLogger(loggers[2]),
		},
	}

	var verify = func(idx int, test test) {
		var cfg = WithLoggerV(test.input...)

		config := cfg.(*LSLogger)

		if !reflect.DeepEqual(config.logger, test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				config.logger,
				test.name,
			)
			return
		}

		if !config.verbose {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] expected a verbose logger, verbose attribute was false -- action: %s",
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
	module := "LogClientConfig"
	funcname := "WithGRPCOpts()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		input []grpc.DialOption
		wants []grpc.DialOption
	}

	opt := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.FailOnNonTempDialError(true),
	}

	opts := [][]grpc.DialOption{
		{
			opt[0],
		},
		{
			opt[0],
			opt[1],
		},
		{
			nil,
			opt[0],
			nil,
			opt[1],
			nil,
		},
		{
			nil,
			nil,
			nil,
		},
	}

	var tests = []test{
		{
			name:  "one valid option",
			input: opts[0],
			wants: opts[0],
		},
		{
			name:  "multiple valid options",
			input: opts[1],
			wants: opts[1],
		},
		{
			name:  "multiple valid options mixed with nil values",
			input: opts[2],
			wants: opts[1],
		},
		{
			name:  "multiple nil options",
			input: opts[3],
			wants: defaultDialOptions,
		},
		{
			name:  "zero options",
			input: []grpc.DialOption{},
			wants: defaultDialOptions,
		},
		{
			name:  "nil options",
			input: nil,
			wants: defaultDialOptions,
		},
	}

	var verify = func(idx int, test test) {
		var cfg = WithGRPCOpts(test.input...)

		config := cfg.(*LSOpts)

		if !reflect.DeepEqual(config.opts, test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				config.opts,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
