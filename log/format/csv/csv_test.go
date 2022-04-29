package csv

import (
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/log/event"
)

func TestNew(t *testing.T) {
	module := "FmtCSV"
	funcname := "New()"

	type test struct {
		name string
		unix bool
		json bool
	}

	var tests = []test{
		{
			name: "CSV defaults",
		},
		{
			name: "with Unix time",
			unix: true,
		},
		{
			name: "with JSON metadata",
			json: true,
		},
		{
			name: "with all opts",
			unix: true,
			json: true,
		},
	}

	var init = func(test test) *FmtCSV {
		f := New()

		if test.unix {
			f.Unix()
		}

		if test.json {
			f.JSON()
		}

		return f.Build()
	}

	var verify = func(idx int, test test) {
		f := init(test)

		if f.UnixTime != test.unix {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] mismatching configs for unix time: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.unix,
				f.UnixTime,
				test.name,
			)
			return
		}

		if f.JsonMeta != test.json {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] mismatching configs for JSON meta: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.json,
				f.JsonMeta,
				test.name,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [%s] [%s] -- action: %s",
			idx,
			module,
			funcname,
			test.name,
		)

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestFormat(t *testing.T) {
	module := "FmtCSV"
	funcname := "Format()"

	type test struct {
		name string
		unix bool
		json bool
	}

	var e *event.Event = event.New().
		Prefix("test").
		Sub("testing").
		Level(event.Level_warn).
		Message("null").
		Metadata(event.Field{"a": true}).
		Build()

	var tests = []test{
		{
			name: "CSV defaults",
		},
		{
			name: "with Unix time",
			unix: true,
		},
		{
			name: "with JSON metadata",
			json: true,
		},
		{
			name: "with all opts",
			unix: true,
			json: true,
		},
	}

	var init = func(test test) *FmtCSV {
		f := New()

		if test.unix {
			f.Unix()
		}

		if test.json {
			f.JSON()
		}

		return f.Build()
	}

	var verify = func(idx int, test test) {
		f := init(test)

		b, err := f.Format(e)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] formatting error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		if test.json {
			new, err := Decode(b)

			if err != nil {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] decoding error: %v -- action: %s",
					idx,
					module,
					funcname,
					err,
					test.name,
				)
				return
			}

			if !reflect.DeepEqual(*new, *e) {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
					idx,
					module,
					funcname,
					*e,
					*new,
					test.name,
				)
				return
			}

		}

		t.Logf(
			"#%v -- PASSED -- [%s] [%s] -- action: %s",
			idx,
			module,
			funcname,
			test.name,
		)

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
