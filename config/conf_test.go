package config

import (
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/grpc/address"
	"github.com/zalgonoise/zlog/log"
)

func TestNewConfiguration(t *testing.T) {
	module := "Configuration"
	function := "New()"

	logger := log.New()
	chLogger := log.NewLogCh(logger)

	tests := []struct {
		name   string
		parent interface{}
	}{
		{
			name:   "prefix",
			parent: logger,
		},
		{
			name:   "sub",
			parent: logger,
		},
		{
			name:   "logCh",
			parent: chLogger,
		},
	}

	for id, test := range tests {
		conf := New(test.name, test.parent)

		if conf.Name() != test.name {
			t.Errorf(
				"#%v -- FAILED -- [%s][%s] -- name mismatch: wanted %s ; got %s",
				id,
				module,
				function,
				test.name,
				conf.Name(),
			)
			return
		}

		if !conf.Is(test.parent) {
			t.Errorf(
				"#%v -- FAILED -- [%s][%s] -- type mismatch: wanted %s ; got %s",
				id,
				module,
				function,
				reflect.TypeOf(test.parent),
				reflect.TypeOf(conf),
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [%s][%s]",
			id,
			module,
			function,
		)

	}
}

func TestWithValueConfiguration(t *testing.T) {
	module := "Configuration"
	function := "WithValue()"

	logger := log.New()
	chLogger := log.NewLogCh(logger)

	tests := []struct {
		name   string
		parent interface{}
		value  interface{}
	}{
		{
			name:   "prefix",
			parent: logger,
			value:  int32(15),
		},
		{
			name:   "sub",
			parent: logger,
			value:  address.ConnAddr{},
		},
		{
			name:   "logCh",
			parent: chLogger,
			value:  log.LCOut{},
		},
	}

	for id, test := range tests {
		conf := New(test.name, test.parent)
		conf = WithValue(conf, test.value)

		if conf.Name() != test.name {
			t.Errorf(
				"#%v -- FAILED -- [%s][%s] -- name mismatch: wanted %s ; got %s",
				id,
				module,
				function,
				test.name,
				conf.Name(),
			)
			return
		}

		if !conf.Is(test.parent) {
			t.Errorf(
				"#%v -- FAILED -- [%s][%s] -- type mismatch: wanted %s ; got %s",
				id,
				module,
				function,
				reflect.TypeOf(test.parent),
				reflect.TypeOf(conf),
			)
			return
		}

		if v := conf.(*configuration).value; reflect.TypeOf(v) != reflect.TypeOf(test.value) {
			t.Errorf(
				"#%v -- FAILED -- [%s][%s] -- value mismatch: wanted %s ; got %s",
				id,
				module,
				function,
				test.value,
				v,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [%s][%s]",
			id,
			module,
			function,
		)

	}
}

func TestConfigs(t *testing.T) {
	module := "Configs"
	function := "Mapping"

	logger := log.New()
	chLogger := log.NewLogCh(logger)

	tests := []struct {
		name   []string
		parent []interface{}
		value  []interface{}
	}{
		{
			name:   []string{"prefix", "sub", "level"},
			parent: []interface{}{logger, logger, logger},
			value:  []interface{}{"log", "", int32(1)},
		},
		{
			name:   []string{"prefix", "sub", "level"},
			parent: []interface{}{chLogger, chLogger, chLogger},
			value:  []interface{}{"log", "", int32(1)},
		},
		{
			name:   []string{"prefix", "sub", "level"},
			parent: []interface{}{chLogger, chLogger, chLogger},
			value:  []interface{}{"log", "", nil},
		},
	}

	for id, test := range tests {
		target := &Configs{}
		confList := []Config{}

		for i := 0; i < len(test.name); i++ {
			conf := New(test.name[i], test.parent[i])
			conf = WithValue(conf, test.value[i])

			confList = append(confList, conf)
		}

		for _, opt := range confList {
			opt.Apply(target)
		}

		tmap := target.Map()

		for i := 0; i < len(test.name); i++ {

			v, ok := tmap[test.name[i]]
			if !ok {
				t.Errorf(
					"#%v -- FAILED -- [%s][%s] -- key request failed for key %s",
					id,
					module,
					function,
					test.name[i],
				)
				return
			}
			if v.Name() != test.name[i] {
				t.Errorf(
					"#%v -- FAILED -- [%s][%s] -- config name mismatch: wanted %s ; got %s",
					id,
					module,
					function,
					test.name[i],
					v.Name(),
				)
				return
			}

			if !v.Is(test.parent[i]) {
				t.Errorf(
					"#%v -- FAILED -- [%s][%s] -- type mismatch: wanted %s ; got %s",
					id,
					module,
					function,
					reflect.TypeOf(test.parent),
					reflect.TypeOf(v),
				)
				return
			}
			if val := target.Get(test.name[i]); reflect.TypeOf(val) != reflect.TypeOf(test.value[i]) {
				t.Errorf(
					"#%v -- FAILED -- [%s][%s] -- value mismatch: wanted %s ; got %s",
					id,
					module,
					function,
					test.value,
					v.(*configuration).value,
				)
				return
			}
		}
	}
}
