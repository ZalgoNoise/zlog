package client

import (
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/log"
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
