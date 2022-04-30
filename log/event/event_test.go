package event

import (
	"reflect"
	"regexp"
	"testing"
)

func TestMarshalUnmarshal(t *testing.T) {
	module := "Event"
	funcname := "Marshal()/Unmarshal()"

	type test struct {
		name string
		e    *Event
	}

	var tests = []test{
		{
			name: "basic event marshalling",
			e:    New().Message("null").Build(),
		},
		{
			name: "complete event marshalling",
			e:    New().Level(Level_warn).Prefix("test").Sub("testing").Message("null").Metadata(Field{"a": true}).Build(),
		},
		{
			name: "complex event marshalling",
			e: New().Level(Level_warn).Prefix("test").Sub("testing").Message("null").CallStack(true).Metadata(Field{
				"a": []Field{
					{"b": 0, "c": 1, "d": 2},
					{"e": 0, "f": 1, "g": 2},
				},
			}).Build(),
		},
	}

	var verify = func(idx int, test test) {
		b, err := test.e.Marshal()
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

		e := new(Event)

		err = e.Unmarshal(b)

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

		if e.GetTime().AsTime() != test.e.GetTime().AsTime() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- time mismatch: wanted %v ; got %v -- action: %s",
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
				"#%v -- FAILED -- [%s] [%s] -- prefix mismatch: wanted %v ; got %v -- action: %s",
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
				"#%v -- FAILED -- [%s] [%s] -- sub-prefix mismatch: wanted %v ; got %v -- action: %s",
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
				"#%v -- FAILED -- [%s] [%s] -- level mismatch: wanted %v ; got %v -- action: %s",
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
				"#%v -- FAILED -- [%s] [%s] -- message mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.e.GetMsg(),
				e.GetMsg(),
				test.name,
			)
			return
		}

		if !reflect.DeepEqual(test.e.GetMeta().AsMap(), e.GetMeta().AsMap()) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- metadata mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.e.GetMeta().AsMap(),
				e.GetMeta().AsMap(),
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

func TestEncodeDecodeMethod(t *testing.T) {
	module := "Event"
	funcname := "Encode()/Decode()"

	type test struct {
		name string
		e    *Event
	}

	var tests = []test{
		{
			name: "basic event encoding",
			e:    New().Message("null").Build(),
		},
		{
			name: "complete event encoding",
			e:    New().Level(Level_warn).Prefix("test").Sub("testing").Message("null").Metadata(Field{"a": true}).Build(),
		},
		{
			name: "complex event encoding",
			e: New().Level(Level_warn).Prefix("test").Sub("testing").Message("null").CallStack(true).Metadata(Field{
				"a": []Field{
					{"b": 0, "c": 1, "d": 2},
					{"e": 0, "f": 1, "g": 2},
				},
			}).Build(),
		},
	}

	var verify = func(idx int, test test) {
		b := test.e.Encode()

		e := new(Event)

		e.Decode(b)

		if e.GetTime().AsTime() != test.e.GetTime().AsTime() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- time mismatch: wanted %v ; got %v -- action: %s",
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
				"#%v -- FAILED -- [%s] [%s] -- prefix mismatch: wanted %v ; got %v -- action: %s",
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
				"#%v -- FAILED -- [%s] [%s] -- sub-prefix mismatch: wanted %v ; got %v -- action: %s",
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
				"#%v -- FAILED -- [%s] [%s] -- level mismatch: wanted %v ; got %v -- action: %s",
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
				"#%v -- FAILED -- [%s] [%s] -- message mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.e.GetMsg(),
				e.GetMsg(),
				test.name,
			)
			return
		}

		if !reflect.DeepEqual(test.e.GetMeta().AsMap(), e.GetMeta().AsMap()) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- metadata mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.e.GetMeta().AsMap(),
				e.GetMeta().AsMap(),
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

func TestEncodeDecodeFunction(t *testing.T) {
	module := "Event"
	funcname := "Encode()/Decode()"

	type test struct {
		name string
		e    *Event
	}

	var tests = []test{
		{
			name: "basic event encoding",
			e:    New().Message("null").Build(),
		},
		{
			name: "complete event encoding",
			e:    New().Level(Level_warn).Prefix("test").Sub("testing").Message("null").Metadata(Field{"a": true}).Build(),
		},
		{
			name: "complex event encoding",
			e: New().Level(Level_warn).Prefix("test").Sub("testing").Message("null").CallStack(true).Metadata(Field{
				"a": []Field{
					{"b": 0, "c": 1, "d": 2},
					{"e": 0, "f": 1, "g": 2},
				},
			}).Build(),
		},
	}

	var verify = func(idx int, test test) {
		b, err := Encode(test.e)
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

		if e.GetTime().AsTime() != test.e.GetTime().AsTime() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- time mismatch: wanted %v ; got %v -- action: %s",
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
				"#%v -- FAILED -- [%s] [%s] -- prefix mismatch: wanted %v ; got %v -- action: %s",
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
				"#%v -- FAILED -- [%s] [%s] -- sub-prefix mismatch: wanted %v ; got %v -- action: %s",
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
				"#%v -- FAILED -- [%s] [%s] -- level mismatch: wanted %v ; got %v -- action: %s",
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
				"#%v -- FAILED -- [%s] [%s] -- message mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.e.GetMsg(),
				e.GetMsg(),
				test.name,
			)
			return
		}

		if !reflect.DeepEqual(test.e.GetMeta().AsMap(), e.GetMeta().AsMap()) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- metadata mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.e.GetMeta().AsMap(),
				e.GetMeta().AsMap(),
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

func TestString(t *testing.T) {
	module := "Event"
	funcname := "String()"

	type test struct {
		name  string
		e     *Event
		wants string
	}

	var tests = []test{
		{
			name:  "basic event encoding",
			e:     New().Message("null").Build(),
			wants: `((time:\{seconds:\d{10}\s+nanos:\d+\})|(prefix:"log")|(sub:"")|(level:info)|(msg:"null")|(\s+)){9}`,
		},
		{
			name:  "complete event encoding",
			e:     New().Level(Level_warn).Prefix("test").Sub("testing").Message("null").Metadata(Field{"a": true}).Build(),
			wants: `((time:\{seconds:\d{10}\s+nanos:\d+\})|(prefix:"test")|(sub:"testing")|(level:warn)|(msg:"null")|(meta:{fields:{key:"a"\s+value:{bool_value:true}}})|(\s+)){11}`,
		},
	}

	var verify = func(idx int, test test) {
		r := regexp.MustCompile(test.wants)
		if !r.MatchString(test.e.String()) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- regexp mismatch: expression %s ; didn't match %s -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				test.e.String(),
				test.name,
			)
			return
		}

		// execute ProtoMessage() method for coverage
		test.e.ProtoMessage()
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestProtoDefaults(t *testing.T) {
	module := "Event"
	funcname := "GetXxx()"

	type test struct {
		name string
		e    *Event
	}

	var tests = []test{
		{
			name: "test defaults",
			e:    new(Event),
		},
	}

	var verify = func(idx int, test test) {
		if test.e.GetTime() != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- GetTime(): wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				nil,
				test.e.GetTime(),
				test.name,
			)
			return
		}
		if test.e.GetPrefix() != Default_Event_Prefix {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- GetPrefix(): wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				Default_Event_Prefix,
				test.e.GetPrefix(),
				test.name,
			)
			return
		}
		if test.e.GetSub() != "" {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- GetSub(): wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				`""`,
				test.e.GetSub(),
				test.name,
			)
			return
		}
		if test.e.GetLevel() != Default_Event_Level {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- GetLevel(): wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				Default_Event_Level,
				test.e.GetLevel(),
				test.name,
			)
			return
		}
		if test.e.GetMsg() != "" {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- GetMsg(): wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				`""`,
				test.e.GetMsg(),
				test.name,
			)
			return
		}
		if test.e.GetMeta() != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- GetMeta(): wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				nil,
				test.e.GetMeta(),
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
