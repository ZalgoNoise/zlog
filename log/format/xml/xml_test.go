package xml

import (
	"encoding/json"
	"reflect"
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
		{
			name:  "simple event; metadata with empty map",
			e:     event.New().Message("null\n").Metadata(map[string]interface{}{"empty": map[string]interface{}{}}).Build(),
			regex: `<entry><timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z<\/timestamp><service>log<\/service><level>info<\/level><message>null<\/message><metadata><key>empty<\/key><\/metadata><\/entry>`,
		},
		{
			name:  "complete event",
			e:     event.New().Prefix("test").Sub("testing").Level(event.Level_warn).Message("null").Metadata(event.Field{"a": true}).Build(),
			regex: `<entry><timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z<\/timestamp><service>test<\/service><module>testing<\/module><level>warn<\/level><message>null<\/message><metadata><key>a<\/key><value>true<\/value><\/metadata><\/entry>`,
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

func TestMappify(t *testing.T) {
	module := "FmtXML"
	funcname := "Mappify()"

	type test struct {
		name string
		data map[string]interface{}
	}

	var tests = []test{
		{
			name: "boolean obj",
			data: map[string]interface{}{
				"data": true,
			},
		},
		{
			name: "simple obj",
			data: map[string]interface{}{
				"data": "object",
			},
		},
		{
			name: "with map",
			data: map[string]interface{}{
				"data": map[string]interface{}{
					"a": 1,
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
		},
		{
			name: "with slice of maps",
			data: map[string]interface{}{
				"data": []map[string]interface{}{
					{"a": 1}, {"b": 2}, {"c": 3},
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
		},
	}

	var verify = func(id int, test test) {

		fields := Mappify(test.data)

		conv := mapMetadata(fields)

		tm, _ := json.Marshal(test.data)
		om, _ := json.Marshal(conv)

		if !reflect.DeepEqual(om, tm) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- output mismatch error: wanted %v ; got %v -- raw-want: %s ; raw-got: %s -- action: %s",
				id,
				module,
				funcname,
				test.data,
				conv,
				string(tm),
				string(om),
				test.name,
			)
			return
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
