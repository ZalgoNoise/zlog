package event

import (
	"testing"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

func TestLevelInt(t *testing.T) {
	module := "Level"
	funcname := "Int()"

	type test struct {
		name  string
		level Level
		wants int32
	}

	var tests = []test{
		{
			name:  "trace level",
			level: Level_trace,
			wants: 0,
		},
		{
			name:  "debug level",
			level: Level_debug,
			wants: 1,
		},
		{
			name:  "info level",
			level: Level_info,
			wants: 2,
		},
		{
			name:  "warn level",
			level: Level_warn,
			wants: 3,
		},
		{
			name:  "error level",
			level: Level_error,
			wants: 4,
		},
		{
			name:  "fatal level",
			level: Level_fatal,
			wants: 5,
		},
		{
			name:  "panic level",
			level: Level_panic,
			wants: 9,
		},
	}

	var verify = func(idx int, test test, i int32) {
		if i != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- enum integer value mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				i,
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
		i := test.level.Int()

		verify(idx, test, i)

	}
}

func TestLevelNumber(t *testing.T) {
	module := "Level"
	funcname := "Int()"

	type test struct {
		name  string
		level Level
		wants protoreflect.EnumNumber
	}

	var tests = []test{
		{
			name:  "trace level",
			level: Level_trace,
			wants: 0,
		},
		{
			name:  "debug level",
			level: Level_debug,
			wants: 1,
		},
		{
			name:  "info level",
			level: Level_info,
			wants: 2,
		},
		{
			name:  "warn level",
			level: Level_warn,
			wants: 3,
		},
		{
			name:  "error level",
			level: Level_error,
			wants: 4,
		},
		{
			name:  "fatal level",
			level: Level_fatal,
			wants: 5,
		},
		{
			name:  "panic level",
			level: Level_panic,
			wants: 9,
		},
	}

	var verify = func(idx int, test test, i protoreflect.EnumNumber) {
		if i != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- enum integer value mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				i,
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
		i := test.level.Number()

		verify(idx, test, i)

	}
}

func TestLevelEnum(t *testing.T) {
	module := "Level"
	funcname := "Enum()"

	type test struct {
		name  string
		level Level
		wants int32
	}

	var tests = []test{
		{
			name:  "trace level",
			level: Level_trace,
			wants: 0,
		},
		{
			name:  "debug level",
			level: Level_debug,
			wants: 1,
		},
		{
			name:  "info level",
			level: Level_info,
			wants: 2,
		},
		{
			name:  "warn level",
			level: Level_warn,
			wants: 3,
		},
		{
			name:  "error level",
			level: Level_error,
			wants: 4,
		},
		{
			name:  "fatal level",
			level: Level_fatal,
			wants: 5,
		},
		{
			name:  "panic level",
			level: Level_panic,
			wants: 9,
		},
	}

	var verify = func(idx int, test test) {
		enum := test.level.Enum()

		if enum.Int() != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- enum integer value mismatch: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				enum.Int(),
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
