package client

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/zalgonoise/zlog/log"
	"google.golang.org/grpc"
)

func getEnv(env string) (val string, ok bool) {
	v := os.Getenv(env)

	if v == "" {
		return v, false
	}

	return v, true
}

type testTLS struct {
	keyCert []string
	caCert  string
}

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

		// test Apply()
		builder := &gRPCLogClientBuilder{
			interceptors: clientInterceptors{
				streamItcp: make(map[string]grpc.StreamClientInterceptor),
				unaryItcp:  make(map[string]grpc.UnaryClientInterceptor),
			},
		}
		cfg.Apply(builder)
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

		// test Apply()
		builder := &gRPCLogClientBuilder{
			interceptors: clientInterceptors{
				streamItcp: make(map[string]grpc.StreamClientInterceptor),
				unaryItcp:  make(map[string]grpc.UnaryClientInterceptor),
			},
		}
		cfg.Apply(builder)
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

		// test Apply()
		builder := &gRPCLogClientBuilder{
			interceptors: clientInterceptors{
				streamItcp: make(map[string]grpc.StreamClientInterceptor),
				unaryItcp:  make(map[string]grpc.UnaryClientInterceptor),
			},
		}

		cfg.Apply(builder)
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

		// test Apply()
		builder := &gRPCLogClientBuilder{
			interceptors: clientInterceptors{
				streamItcp: make(map[string]grpc.StreamClientInterceptor),
				unaryItcp:  make(map[string]grpc.UnaryClientInterceptor),
			},
		}
		cfg.Apply(builder)
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

		// test Apply()
		builder := &gRPCLogClientBuilder{
			interceptors: clientInterceptors{
				streamItcp: make(map[string]grpc.StreamClientInterceptor),
				unaryItcp:  make(map[string]grpc.UnaryClientInterceptor),
			},
		}
		cfg.Apply(builder)
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestWithBackoff(t *testing.T) {
	module := "LogClientConfig"
	funcname := "WithBackoff()"

	_ = module
	_ = funcname

	type test struct {
		name     string
		deadline time.Duration
		fn       BackoffFunc
		wants    *Backoff
	}

	b := []*Backoff{
		NewBackoff(), NewBackoff(), NewBackoff(),
	}

	b[0].BackoffFunc(NoBackoff())

	b[1].Time(defaultRetryTime)
	b[1].BackoffFunc(BackoffLinear(time.Second))

	b[2].Time(time.Minute)
	b[2].BackoffFunc(BackoffExponential())

	var tests = []test{
		{
			name:     "no deadline and no backoff function",
			deadline: 0,
			fn:       nil,
			wants:    b[0],
		},
		{
			name:     "no deadline and linear backoff function",
			deadline: 0,
			fn:       BackoffLinear(time.Second),
			wants:    b[1],
		},
		{
			name:     "with deadline but no backoff function",
			deadline: time.Minute,
			fn:       nil,
			wants:    b[2],
		},
	}

	var verify = func(idx int, test test) {
		var cfg = WithBackoff(test.deadline, test.fn)

		config := cfg.(*LSExpBackoff)

		// check if backoffFunc is not nil; then make it nil
		// for a quick DeepEqual
		//
		// BackoffFunc tests will be found in backoff_test.go
		if config.backoff.backoffFunc == nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] backoff func cannot be nil -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		config.backoff.backoffFunc = nil
		test.wants.backoffFunc = nil

		if !reflect.DeepEqual(config.backoff, test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				config.backoff,
				test.name,
			)
			return
		}

		// test Apply()
		builder := &gRPCLogClientBuilder{
			interceptors: clientInterceptors{
				streamItcp: make(map[string]grpc.StreamClientInterceptor),
				unaryItcp:  make(map[string]grpc.UnaryClientInterceptor),
			},
		}
		cfg.Apply(builder)
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestWithTiming(t *testing.T) {
	module := "LogClientConfig"
	funcname := "WithTiming()"

	_ = module
	_ = funcname

	type test struct {
		name      string
		base      *gRPCLogClientBuilder
		hasLogger bool
	}

	var logger = log.New(log.NilConfig)

	var tests = []test{
		{
			name: "without logger",
			base: &gRPCLogClientBuilder{
				interceptors: clientInterceptors{
					streamItcp: make(map[string]grpc.StreamClientInterceptor),
					unaryItcp:  make(map[string]grpc.UnaryClientInterceptor),
				},
			},
		},
		{
			name: "with logger",
			base: &gRPCLogClientBuilder{
				interceptors: clientInterceptors{
					streamItcp: map[string]grpc.StreamClientInterceptor{
						"logging": StreamClientLogging(logger, false),
					},
					unaryItcp: map[string]grpc.UnaryClientInterceptor{
						"logging": UnaryClientLogging(logger, false),
					},
				},
				svcLogger: logger,
			},
			hasLogger: true,
		},
	}

	var verify = func(idx int, test test) {
		var cfg = WithTiming()

		cfg.Apply(test.base)

		if test.hasLogger {
			if _, ok := test.base.interceptors.streamItcp["logging"]; !ok {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] stream interceptor map does not contain a logging entry -- action: %s",
					idx,
					module,
					funcname,
					test.name,
				)
				return
			}
			if _, ok := test.base.interceptors.unaryItcp["logging"]; !ok {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unary interceptor map does not contain a logging entry -- action: %s",
					idx,
					module,
					funcname,
					test.name,
				)
				return
			}
		} else {
			if _, ok := test.base.interceptors.streamItcp["timing"]; !ok {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] stream interceptor map does not contain a timing entry -- action: %s",
					idx,
					module,
					funcname,
					test.name,
				)
				return
			}
			if _, ok := test.base.interceptors.unaryItcp["timing"]; !ok {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unary interceptor map does not contain a timing entry -- action: %s",
					idx,
					module,
					funcname,
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

func TestWithTLS(t *testing.T) {
	module := "LogClientConfig"
	funcname := "WithTLS()"

	_ = module
	_ = funcname

	type test struct {
		name       string
		caPathEnv  string
		keyCertEnv []string
		isMutual   bool
		ok         bool
	}

	var tests = []test{
		{
			name:      "simple TLS",
			caPathEnv: "TLS_CA_CERT",
			ok:        true,
		},
		{
			name:      "mutual TLS",
			caPathEnv: "TLS_CA_CERT",
			keyCertEnv: []string{
				"TLS_CLIENT_CERT",
				"TLS_CLIENT_KEY",
			},
			isMutual: true,
			ok:       true,
		},
		{
			name:      "no arguments provided",
			caPathEnv: "",
		},
		{
			name:      "mTLS, but not enough arguments",
			caPathEnv: "TLS_CA_CERT",
			keyCertEnv: []string{
				"TLS_CLIENT_CERT",
				"",
			},
			isMutual: true,
		},
	}

	var catchPanics = func(idx int, test test) {
		r := recover()

		if r != nil && test.ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] execution error (panic): %v -- action: %s",
				idx,
				module,
				funcname,
				r,
				test.name,
			)
		}
	}

	var init = func(test test) (*testTLS, error) {

		out := new(testTLS)

		caCert, ok := getEnv(test.caPathEnv)
		if !ok && test.ok {
			return nil, errors.New("missing server certificate env variable")
		}
		out.caCert = caCert

		if test.isMutual {
			out.keyCert = []string{"", ""}

			clientCert, ok := getEnv(test.keyCertEnv[0])
			if !ok && test.ok {
				return nil, errors.New("missing server certificate env variable")
			}
			out.keyCert[0] = clientCert

			clientKey, ok := getEnv(test.keyCertEnv[1])
			if !ok && test.ok {
				return nil, errors.New("missing server key env variable")
			}
			out.keyCert[1] = clientKey
		}

		return out, nil

	}

	var verify = func(idx int, test test) {

		defer catchPanics(idx, test)

		cfg, err := init(test)
		if err != nil {
			t.Logf(
				"#%v -- SKIPPED -- [%s] [%s] error when loading environment variables: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		if test.isMutual {
			_ = WithTLS(cfg.caCert, cfg.keyCert...)
		} else {
			_ = WithTLS(cfg.caCert)
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
