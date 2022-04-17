package xml

import (
	"testing"

	"github.com/zalgonoise/zlog/log/event"
)

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
	type test struct {
		desc string
		data map[string]interface{}
		obj  []Field
	}

	var tests = []test{
		{
			desc: "simple obj",
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
			desc: "with map",
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
			desc: "with Field",
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
			desc: "with slice of maps",
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
			desc: "with slice of Fields",
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

	var verify = func(id int, test test, fields []Field) {
		if len(fields) != len(test.obj) {
			t.Errorf(
				"#%v -- FAILED --  mappify(map[string]interface{}) []field -- object len %v does not match expected len %v",
				id,
				len(fields),
				len(test.obj),
			)
			return
		}

		for i := 0; i < len(fields); i++ {
			if fields[i].Key != test.obj[i].Key {
				t.Errorf(
					"#%v -- FAILED --  mappify(map[string]interface{}) []field -- object key mismatch: wanted %s ; got %s",
					id,
					test.obj[i].Key,
					fields[i].Key,
				)
				return
			}

			ok := match(fields[i].Val, test.obj[i].Val)
			if !ok {
				t.Errorf(
					"#%v -- FAILED --  mappify(map[string]interface{}) []field -- object value mismatch: wanted %s ; got %s",
					id,
					test.obj[i].Val,
					fields[i].Val,
				)
				return
			}
		}

		t.Logf(
			"#%v -- PASSED --  mappify(map[string]interface{}) []field",
			id,
		)

	}

	for id, test := range tests {
		fields := Mappify(test.data)
		verify(id, test, fields)
	}

	// // test implementation
	// for id, test := range tests {
	// 	fields := test.data.ToXML()
	// 	verify(id, test, fields)
	// }
}
