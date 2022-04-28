package json

import (
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/log/event"
)

func TestFormat(t *testing.T) {
	module := "FmtJSON"
	funcname := "Format()"

	_ = funcname
	_ = module

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
		f := new(FmtJSON)

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

		e, err := Decode(b)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- unmarshalling error: %s -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return

		}

		if !reflect.DeepEqual(*e, *test.e) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				*test.e,
				*e,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestDecodeError(t *testing.T) {
	module := "FmtJSON"
	funcname := "Decode()"

	type test struct {
		name string
		json string
	}

	var tests = []test{
		{
			name: "simple event",
			json: `{"invalid_json":tr`,
		},
	}

	var verify = func(idx int, test test) {

		_, err := Decode([]byte(test.json))

		if err == nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- expected unmarshalling error; got nil -- action: %s",
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
