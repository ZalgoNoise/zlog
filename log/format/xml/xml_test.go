package xml

import (
	"regexp"
	"testing"

	"github.com/zalgonoise/zlog/log/event"
)

func TestFormat(t *testing.T) {
	module := "FmtXML"
	funcname := "Format()"

	_ = funcname
	_ = module

	type test struct {
		name  string
		e     *event.Event
		regex string
	}

	var tests = []test{
		{
			name:  "simple event",
			e:     event.New().Message("null\n").Build(),
			regex: `<entry><timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z<\/timestamp><service>log<\/service><level>info<\/level><message>null<\/message><\/entry>`,
		},
	}

	var verify = func(idx int, test test) {
		f := new(FmtXML)

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

		r := regexp.MustCompile(test.regex)

		if !r.MatchString(string(b)) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- output mismatch error: expression: %s ; didn't match: %s -- action: %s",
				idx,
				module,
				funcname,
				test.regex,
				string(b),
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func match(want, got interface{}) bool {
	switch value := got.(type) {
	case []Field:
		w := want.([]Field)
		for idx, f := range value {
			if f.Key != w[idx].Key {
				return false
			}
			if !match(f.Val, w[idx].Val) {
				return false
			}
		}
		return true
	// case field:
	default:
		if value == want {
			return true
		}
	}
	return false
}

func TestMappify(t *testing.T) {
	module := "FmtXML"
	funcname := "Mappify()"

	type test struct {
		name string
		data map[string]interface{}
		obj  []Field
	}

	var tests = []test{
		{
			name: "simple obj",
			data: map[string]interface{}{
				"data": "object",
			},
			obj: []Field{
				{
					Key: "data",
					Val: "object",
				},
			},
		},
		{
			name: "with map",
			data: map[string]interface{}{
				"data": map[string]interface{}{
					"a": 1,
				},
			},
			obj: []Field{
				{
					Key: "data",
					Val: []Field{
						{
							Key: "a",
							Val: 1,
						},
					},
				},
			},
		},
		{
			name: "with Field",
			data: event.Field{
				"data": event.Field{
					"a": 1,
				},
			},
			obj: []Field{
				{
					Key: "data",
					Val: []Field{
						{
							Key: "a",
							Val: 1,
						},
					},
				},
			},
		},
		{
			name: "with slice of maps",
			data: map[string]interface{}{
				"data": []map[string]interface{}{
					{"a": 1}, {"b": 2}, {"c": 3},
				},
			},
			obj: []Field{
				{
					Key: "data",
					Val: []Field{
						{Key: "a", Val: 1},
						{Key: "b", Val: 2},
						{Key: "c", Val: 3},
					},
				},
			},
		},
		{
			name: "with slice of Fields",
			data: event.Field{
				"data": []event.Field{
					{"a": 1}, {"b": 2}, {"c": 3},
				},
			},
			obj: []Field{
				{
					Key: "data",
					Val: []Field{
						{Key: "a", Val: 1},
						{Key: "b", Val: 2},
						{Key: "c", Val: 3},
					},
				},
			},
		},
	}

	var verify = func(id int, test test) {

		fields := Mappify(test.data)

		if len(fields) != len(test.obj) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- object len %v does not match expected len %v -- action: %s",
				id,
				module,
				funcname,
				len(fields),
				len(test.obj),
				test.name,
			)
			return
		}

		for i := 0; i < len(fields); i++ {
			if fields[i].Key != test.obj[i].Key {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- object key mismatch: wanted %s ; got %s -- action: %s",
					id,
					module,
					funcname,
					test.obj[i].Key,
					fields[i].Key,
					test.name,
				)
				return
			}

			ok := match(fields[i].Val, test.obj[i].Val)
			if !ok {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- object value mismatch: wanted %s ; got %s -- action: %s",
					id,
					module,
					funcname,
					test.obj[i].Val,
					fields[i].Val,
					test.name,
				)
				return
			}
		}

		t.Logf(
			"#%v -- PASSED -- [%s] [%s] -- action: %s",
			id,
			module,
			funcname,
			test.name,
		)

	}

	for id, test := range tests {
		verify(id, test)
	}

}
