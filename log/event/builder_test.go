package event

import (
	"reflect"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	module := "EventBuilder"
	funcname := "New()"

	type test struct {
		name   string
		time   time.Time
		prefix string
		sub    string
		level  string
		msg    string
	}

	var tests = []test{
		{
			name:   "new event builder -- check defaults",
			time:   time.Time{},
			prefix: "log",
			sub:    "",
			level:  "info",
			msg:    "",
		},
	}

	var verify = func(idx int, test test, e *EventBuilder) {

		if *e.prefix != test.prefix {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] default prefix mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.prefix,
				*e.prefix,
				test.name,
			)
			return
		}

		if *e.sub != test.sub {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] default sub-prefix mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.sub,
				*e.sub,
				test.name,
			)
			return
		}

		level := *e.level

		if level.String() != test.level {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] default level mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.level,
				level.String(),
				test.name,
			)
			return
		}

		if e.msg != test.msg {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] default message mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.msg,
				e.msg,
				test.name,
			)
			return
		}

		if *e.metadata != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] default metadata is not nil: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				nil,
				*e.metadata,
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
		e := New()

		verify(idx, test, e)
	}

}

func FuzzPrefix(f *testing.F) {
	module := "EventBuilder"
	funcname := "Prefix()"

	f.Add("test-prefix")
	f.Fuzz(func(t *testing.T, a string) {
		e := New().Prefix(a).Message("null").Build()

		if e.GetPrefix() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed prefix mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				e.GetPrefix(),
			)

		}
	})
}

func FuzzSub(f *testing.F) {
	module := "EventBuilder"
	funcname := "Sub()"

	f.Add("test-prefix")
	f.Fuzz(func(t *testing.T, a string) {
		e := New().Sub(a).Message("null").Build()

		if e.GetSub() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed sub-prefix mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				e.GetSub(),
			)
		}
	})
}

func FuzzMessage(f *testing.F) {
	module := "EventBuilder"
	funcname := "Message()"

	f.Add("test-prefix")
	f.Fuzz(func(t *testing.T, a string) {
		e := New().Message(a).Build()

		if e.GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message mismatch: wanted %s ; got %s",
				module,
				funcname,
				a,
				e.GetMsg(),
			)
		}
	})
}

func TestLevel(t *testing.T) {
	module := "EventBuilder"
	funcname := "Level()"

	type test struct {
		name  string
		level Level
		wants string
	}

	var tests = []test{
		{
			name:  "default level",
			level: Default_Event_Level,
			wants: "info",
		},
		{
			name:  "trace level",
			level: Level(0),
			wants: "trace",
		},
		{
			name:  "debug level",
			level: Level(1),
			wants: "debug",
		},
		{
			name:  "info level",
			level: Level(2),
			wants: "info",
		},
		{
			name:  "warning level",
			level: Level(3),
			wants: "warn",
		},
		{
			name:  "error level",
			level: Level(4),
			wants: "error",
		},
		{
			name:  "fatal level",
			level: Level(5),
			wants: "fatal",
		},
		{
			name:  "panic level",
			level: Level(9),
			wants: "panic",
		},
		{
			name:  "invalid level",
			level: Level(6),
			wants: "6",
		},
		{
			name:  "invalid level",
			level: Level(7),
			wants: "7",
		},
		{
			name:  "invalid level",
			level: Level(8),
			wants: "8",
		},
	}

	var verify = func(idx int, test test, e *Event) {
		if e.GetLevel().String() != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] level mismatch: wanted %s ; got %s -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				e.GetLevel().String(),
				test.name,
			)
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
		var e *Event

		if test.level == Default_Event_Level {
			e = New().Message("null").Build()
		} else {
			e = New().Level(test.level).Message("null").Build()
		}

		verify(idx, test, e)
	}
}

func TestMetadata(t *testing.T) {
	module := "EventBuilder"
	funcname := "Metadata()"

	type test struct {
		name  string
		base  map[string]interface{}
		meta  map[string]interface{}
		wants map[string]interface{}
	}

	var tests = []test{
		{
			name:  "no metadata",
			base:  nil,
			meta:  map[string]interface{}{},
			wants: map[string]interface{}{},
		},
		{
			name:  "nil metadata",
			base:  nil,
			meta:  nil,
			wants: map[string]interface{}{},
		},
		{
			name: "add metadata",
			base: nil,
			meta: map[string]interface{}{
				"d": float64(0), "e": float64(1), "f": float64(3),
			},
			wants: map[string]interface{}{
				"d": float64(0), "e": float64(1), "f": float64(3),
			},
		},
		{
			name: "append metadata",
			base: map[string]interface{}{
				"a": float64(0), "b": float64(1), "c": float64(3),
			},
			meta: map[string]interface{}{
				"d": float64(0), "e": float64(1), "f": float64(3),
			},
			wants: map[string]interface{}{
				"a": float64(0), "b": float64(1), "c": float64(3),
				"d": float64(0), "e": float64(1), "f": float64(3),
			},
		},
		{
			name: "merge metadata",
			base: map[string]interface{}{
				"a": float64(0), "b": float64(1), "c": float64(3),
			},
			meta: map[string]interface{}{
				"a": float64(999), "b": float64(1000), "c": float64(1001),
			},
			wants: map[string]interface{}{
				"a": float64(999), "b": float64(1000), "c": float64(1001),
			},
		},
	}

	var verify = func(idx int, test test, e *Event) {

		if len(e.GetMeta().AsMap()) != len(test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] metadata length mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				len(test.wants),
				len(e.GetMeta().AsMap()),
				test.name,
			)
		}

		for k, v := range e.GetMeta().AsMap() {
			if test.wants[k].(float64) != v.(float64) {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] metadata mismatch: [key %s] wanted %v ; got %v -- action: %s",
					idx,
					module,
					funcname,
					k,
					test.wants[k],
					e.GetMeta().AsMap()[k],
					test.name,
				)
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
		m := New().Message("null")

		if test.base != nil {
			m.Metadata(test.base)
		}

		e := m.Metadata(test.meta).Build()

		verify(idx, test, e)
	}
}

func TestCallStack(t *testing.T) {
	module := "EventBuilder"
	funcname := "CallStack()"

	type test struct {
		name string
		all  bool
	}

	var tests = []test{
		{
			name: "build callstack, all: true",
			all:  true,
		},
		{
			name: "build callstack, all: false",
			all:  false,
		},
	}

	var verify = func(idx int, test test, e *Event) {
		c, ok := e.GetMeta().AsMap()["callstack"]

		if !ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] event metadata doesn't contain a callstack key -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		if c == nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] event metadata is null -- action: %s",
				idx,
				module,
				funcname,
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
		e := New().CallStack(test.all).Build()

		verify(idx, test, e)

	}
}

func TestBuild(t *testing.T) {
	module := "EventBuilder"
	funcname := "Build()"

	type test struct {
		name   string
		base   *EventBuilder
		prefix string
		sub    string
		level  string
		msg    string
		meta   []byte
	}

	var tests = []test{
		{
			name:   "force initialize on build; no input values",
			base:   new(EventBuilder),
			prefix: "log",
			sub:    "",
			level:  "info",
			msg:    "",
			meta:   []byte("{}"),
		},
	}

	var verify = func(idx int, test test, e *Event) {
		if e.GetPrefix() != test.prefix {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] prefix mismatch: wanted %s ; got %s -- action: %s",
				idx,
				module,
				funcname,
				test.prefix,
				e.GetPrefix(),
				test.name,
			)
			return
		}

		if e.GetSub() != test.sub {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] sub-prefix mismatch: wanted %s ; got %s -- action: %s",
				idx,
				module,
				funcname,
				test.sub,
				e.GetSub(),
				test.name,
			)
			return
		}

		if e.GetLevel().String() != test.level {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] level mismatch: wanted %s ; got %s -- action: %s",
				idx,
				module,
				funcname,
				test.level,
				e.GetLevel().String(),
				test.name,
			)
			return
		}

		if e.GetMsg() != test.msg {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] message mismatch: wanted %s ; got %s -- action: %s",
				idx,
				module,
				funcname,
				test.msg,
				e.GetMsg(),
				test.name,
			)
			return
		}

		b, err := e.GetMeta().MarshalJSON()

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] JSON Marshalling error: %s -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		if !reflect.DeepEqual(b, test.meta) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] metadata mismatch: wanted %s ; got %s -- action: %s",
				idx,
				module,
				funcname,
				string(test.meta),
				string(b),
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
		e := test.base.Build()

		verify(idx, test, e)
	}
}
