package bson

import (
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/log/event"
)

func TestFormat(t *testing.T) {
	module := "FmtBSON"
	funcname := "Format()"

	type test struct {
		name string
		e    *event.Event
	}

	var tests = []test{
		{
			name: "simple event",
			e:    event.New().Message("null\n").Build(),
		},
		{
			name: "simple event; metadata with empty map",
			e:    event.New().Message("null\n").Metadata(map[string]interface{}{"empty": map[string]interface{}{}}).Build(),
		},
		{
			name: "complete event",
			e:    event.New().Prefix("test").Sub("testing").Level(event.Level_warn).Message("null").Metadata(event.Field{"a": true}).Build(),
		},
	}

	var f = new(FmtBSON)

	var verify = func(idx int, test test) {
		b, err := f.Format(test.e)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- marshalling error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		e, err := Decode(b)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- unmarshalling error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		if e.GetTime().Seconds != test.e.GetTime().Seconds {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- time mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				e.GetTime().Seconds,
				test.e.GetTime().Seconds,
				test.name,
			)
			return
		}

		if e.GetPrefix() != test.e.GetPrefix() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- prefix mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				e.GetPrefix(),
				test.e.GetPrefix(),
				test.name,
			)
		}

		if e.GetSub() != test.e.GetSub() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- sub-prefix mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				e.GetSub(),
				test.e.GetSub(),
				test.name,
			)
		}

		if e.GetLevel().String() != test.e.GetLevel().String() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- level mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				e.GetLevel().String(),
				test.e.GetLevel().String(),
				test.name,
			)
		}

		if e.GetMsg() != test.e.GetMsg() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- message mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				e.GetMsg(),
				test.e.GetMsg(),
				test.name,
			)
		}

		meta := e.GetMeta().AsMap()
		if testMeta := test.e.GetMeta().AsMap(); !reflect.DeepEqual(meta, testMeta) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- metadata mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				testMeta,
				meta,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
