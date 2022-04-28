package protobuf

import (
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/log/event"
)

func TestFormat(t *testing.T) {
	module := "FmtXML"
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
			name: "complete event",
			e:    event.New().Prefix("test").Sub("testing").Level(event.Level_warn).Message("null").Metadata(event.Field{"a": true}).CallStack(true).Build(),
		},
	}

	var verify = func(idx int, test test) {
		f := new(FmtPB)

		b, err := f.Format(test.e)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- marshalling error: %s -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		e, err := event.Decode(b)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- decoding error: %s -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		if e.GetTime().AsTime() != test.e.GetTime().AsTime() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- time mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.e.GetTime().AsTime(),
				e.GetTime().AsTime(),
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
				test.e.GetPrefix(),
				e.GetPrefix(),
				test.name,
			)
			return
		}
		if e.GetSub() != test.e.GetSub() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- sub-prefix mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.e.GetSub(),
				e.GetSub(),
				test.name,
			)
			return
		}
		if e.GetLevel().String() != test.e.GetLevel().String() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- level mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.e.GetLevel().String(),
				e.GetLevel().String(),
				test.name,
			)
			return
		}
		if e.GetMsg() != test.e.GetMsg() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- message mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.e.GetMsg(),
				e.GetMsg(),
				test.name,
			)
			return
		}

		m := e.GetMeta().AsMap()
		if tm := test.e.GetMeta().AsMap(); !reflect.DeepEqual(m, tm) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- metadata mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				tm,
				m,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
