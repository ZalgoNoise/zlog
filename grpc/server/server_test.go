package server

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/zalgonoise/zlog/log"
)

func TestNew(t *testing.T) {
	module := "GRPCLogServer"
	funcname := "New()"

	_ = module
	_ = funcname

	type test struct {
		name    string
		cfg     []LogServerConfig
		wants   *GRPCLogServer
		optsLen int
	}

	var writers = []log.Logger{
		log.New(log.NilConfig),
		log.New(),
		log.New(log.SkipExit),
	}

	var tests = []test{
		{
			name:    "default config, no input",
			cfg:     []LogServerConfig{},
			wants:   New(defaultConfig),
			optsLen: 0,
		},
		{
			name: "with custom config (one entry)",
			cfg: []LogServerConfig{
				WithLogger(writers[0]),
			},
			wants:   New(WithLogger(writers[0])),
			optsLen: 0,
		},
		{
			name: "with custom config (two entries)",
			cfg: []LogServerConfig{
				WithLogger(writers[1]),
				WithServiceLoggerV(writers[1]),
			},
			wants: New(
				WithLogger(writers[1]),
				WithServiceLoggerV(writers[1]),
			),
			optsLen: 2,
		},
		{
			name:    "with nil input",
			cfg:     nil,
			wants:   New(defaultConfig),
			optsLen: 0,
		},
	}

	var verify = func(idx int, test test) {
		wants := test.wants
		server := New(test.cfg...)

		// skip error channel deepequal
		if server.ErrCh == nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected nil error channel -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}
		server.ErrCh = nil
		wants.ErrCh = nil

		// skip LogSv deepequal (tested in pb package)
		if server.LogSv == nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected nil log server -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}
		server.LogSv = nil
		wants.LogSv = nil

		// check options length first
		if len(server.opts) != test.optsLen {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] grpc.ServerOption length mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.optsLen,
				len(server.opts),
				test.name,
			)
			return
		}

		server.opts = nil
		test.wants.opts = nil

		if !reflect.DeepEqual(server, test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				server,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestServe(t *testing.T) {
	module := "GRPCLogServer"
	funcname := "Serve()"

	_ = module
	_ = funcname

	type test struct {
		name string
		s    *GRPCLogServer
		ok   bool
	}

	var bufs = []*bytes.Buffer{{}, {}}

	var writers = []log.Logger{
		log.New(log.WithOut(bufs[0]), log.SkipExit, log.CfgFormatJSON),
		log.New(log.WithOut(bufs[1]), log.SkipExit, log.CfgFormatJSON),
	}

	var tests = []test{
		{
			name: "working config",
			s: New(
				WithLogger(writers[0]),
				WithServiceLoggerV(writers[1]),
				WithAddr("127.0.0.1:9099"),
				WithGRPCOpts(),
			),
			ok: true,
		},
		{
			name: "invalid port error",
			s: New(
				WithLogger(writers[0]),
				WithServiceLoggerV(writers[1]),
				WithAddr("127.0.0.1:999099"),
				WithGRPCOpts(),
			),
		},
	}

	var reset = func() {
		for _, b := range bufs {
			b.Reset()
		}
	}

	var verify = func(idx int, test test) {
		defer test.s.Stop()
		defer reset()
		go test.s.Serve()
		go func() {
			for {
				select {
				case err := <-test.s.ErrCh:
					if test.ok {
						t.Errorf(
							"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
							idx,
							module,
							funcname,
							err,
							test.name,
						)
						return
					}
				}
			}
		}()
		time.Sleep(time.Second * 2)

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
