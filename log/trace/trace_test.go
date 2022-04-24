package trace

import (
	"fmt"
	"strconv"
	"testing"
)

func TestNew(t *testing.T) {
	module := "Trace"
	funcname := "New()"

	var tests = []struct {
		name string
		all  bool
	}{
		{
			name: "getting callstack: all = true",
			all:  true,
		},
		{
			name: "getting callstack: all = false",
			all:  false,
		},
	}

	for id, test := range tests {
		stack := New(test.all)

		if len(stack) == 0 {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- empty map -- action: %s",
				id,
				module,
				funcname,
				test.name,
			)
		}

		t.Logf(
			"#%v -- PASSED -- [%s] [%s] -- action: %s",
			id,
			module,
			funcname,
			test.name,
		)

	}
}

func TestGetCallStack(t *testing.T) {
	module := "Trace"
	funcname := "getCallstack()"

	type test struct {
		name  string
		stack *stacktrace
		all   bool
		want  []byte
	}

	var tests = []test{
		{
			name:  "getting callstack: all = true",
			stack: newCallStack(),
			all:   true,
			want:  []byte("goroutine"),
		},
		{
			name:  "getting callstack: all = false",
			stack: newCallStack(),
			all:   false,
			want:  []byte("goroutine"),
		},
	}

	var verify = func(id int, test test, stack *stacktrace) {
		if len(stack.buf) <= 0 {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- empty buffer error: buffer is %v bytes in length -- action: %s",
				id,
				module,
				funcname,
				len(stack.buf),
				test.name,
			)
			return
		}

		header := stack.buf[:9]
		for idx, c := range header {
			if c != test.want[idx] {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- character mismatch error: got %s ; wanted %s -- action: %s",
					id,
					module,
					funcname,
					string(test.want),
					string(header),
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
		stack := test.stack.getCallStack(test.all)

		verify(id, test, stack)
	}

}

func TestSplitCallStack(t *testing.T) {
	module := "Trace"
	funcname := "splitCallStack()"

	type test struct {
		name     string
		stack    *stacktrace
		all      bool
		minLines int
		want     []byte
	}

	var tests = []test{
		{
			name:     "splitting callstack: all = true",
			stack:    newCallStack().getCallStack(true),
			all:      true,
			want:     []byte("goroutine"),
			minLines: 21,
		},
		{
			name:     "splitting callstack: all = false",
			stack:    newCallStack().getCallStack(false),
			all:      false,
			want:     []byte("goroutine"),
			minLines: 9,
		},
		{
			name:     "splitting callstack: defaults",
			stack:    newCallStack(),
			all:      false,
			want:     []byte("goroutine"),
			minLines: 9,
		},
	}

	var verify = func(id int, test test, stack *stacktrace) {
		if len(stack.buf) <= 0 {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- empty buffer error: buffer is %v bytes in length -- action: %s",
				id,
				module,
				funcname,
				len(stack.buf),
				test.name,
			)
			return
		}

		if len(stack.split) < test.minLines {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- buffer splitting issue: wanted %v lines ; got %v lines -- action: %s",
				id,
				module,
				funcname,
				test.minLines,
				len(stack.split),
				test.name,
			)
			return
		}

		header := stack.split[0][:9]
		for idx, c := range header {
			if c != test.want[idx] {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- character mismatch error: got %s ; wanted %s -- action: %s",
					id,
					module,
					funcname,
					string(test.want),
					string(header),
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
		stack := test.stack.splitCallStack()

		verify(id, test, stack)
	}
}

func TestParseCallStack(t *testing.T) {
	module := "Trace"
	funcname := "parseCallStack()"

	type test struct {
		name        string
		stack       *stacktrace
		all         bool
		minRoutines int
		minID       int
		want        []string
	}

	var tests = []test{
		{
			name:        "parsing callstack: all = true",
			stack:       newCallStack().getCallStack(true).splitCallStack(),
			all:         true,
			minRoutines: 2,
			minID:       1,
			want:        []string{"running", "chan receive"},
		},
		{
			name:        "parsing callstack: all = false",
			stack:       newCallStack().getCallStack(false).splitCallStack(),
			all:         false,
			minRoutines: 1,
			minID:       1,
			want:        []string{"running"},
		},
		{
			name:        "parsing callstack: defaults",
			stack:       newCallStack(),
			all:         false,
			minRoutines: 1,
			minID:       1,
			want:        []string{"running"},
		},
	}

	var verify = func(id int, test test, stack *stacktrace) {
		if len(stack.buf) <= 0 {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- empty buffer error: buffer is %v bytes in length -- action: %s",
				id,
				module,
				funcname,
				len(stack.buf),
				test.name,
			)
			return
		}

		if len(stack.stacks) < test.minRoutines {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- parsed less goroutines than expected: wanted %v ; got %v -- aciton: %s",
				id,
				module,
				funcname,
				test.minRoutines,
				len(stack.stacks),
				test.name,
			)
			return
		}

		for idx, r := range stack.stacks {
			id, err := strconv.Atoi(r.id)
			if err != nil {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- failed to convert ID with error: %s -- action: %s",
					id,
					module,
					funcname,
					err,
					test.name,
				)
				return
			}

			if id < test.minID {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- ID is lower than minimum: wanted > %v ; got %v -- action: %s",
					id,
					module,
					funcname,
					test.minID,
					id,
					test.name,
				)
				return
			}

			if r.status != test.want[idx] {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- status mismatch: wanted %s ; got %s -- action: %s",
					id,
					module,
					funcname,
					test.want,
					r.status,
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
		stack := test.stack.parseCallStack()

		verify(id, test, stack)
	}
}

func TestMapCallStack(t *testing.T) {
	module := "Trace"
	funcname := "mapCallStack()"

	type test struct {
		name        string
		stack       *stacktrace
		all         bool
		minRoutines int
		minID       int
		want        []string
	}

	var tests = []test{
		{
			name:        "mapping callstack: all = true",
			stack:       newCallStack().getCallStack(true).splitCallStack().parseCallStack(),
			all:         true,
			minRoutines: 2,
			minID:       1,
			want:        []string{"running", "chan receive"},
		},
		{
			name:        "mapping callstack: all = false",
			stack:       newCallStack().getCallStack(false).splitCallStack().parseCallStack(),
			all:         false,
			minRoutines: 1,
			minID:       1,
			want:        []string{"running"},
		},
		{
			name:        "mapping callstack: defaults",
			stack:       newCallStack(),
			all:         false,
			minRoutines: 1,
			minID:       1,
			want:        []string{"running"},
		},
	}

	var verify = func(id int, test test, stack *stacktrace) {
		if len(stack.buf) <= 0 {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- empty buffer error: buffer is %v bytes in length -- action: %s",
				id,
				module,
				funcname,
				len(stack.buf),
				test.name,
			)
			return
		}

		if len(stack.stacks) < test.minRoutines {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- parsed less goroutines than expected: wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				test.minRoutines,
				len(stack.stacks),
				test.name,
			)
			return
		}

		for idx, r := range stack.stacks {
			id, err := strconv.Atoi(r.id)
			if err != nil {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- failed to convert ID with error: %s -- action: %s",
					id,
					module,
					funcname,
					err,
					test.name,
				)
				return
			}

			if id < test.minID {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- ID is lower than minimum: wanted > %v ; got %v -- action: %s",
					id,
					module,
					funcname,
					test.minID,
					id,
					test.name,
				)
				return
			}

			if r.status != test.want[idx] {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- status mismatch: wanted %s ; got %s -- action: %s",
					id,
					module,
					funcname,
					test.want,
					r.status,
					test.name,
				)
				return
			}

			callmap := stack.out

			routine, ok := callmap["goroutine-"+r.id]
			if !ok {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- field not found in output object: expected key %s, wasn't found -- action: %s",
					id,
					module,
					funcname,
					"goroutine-"+r.id,
					test.name,
				)
				return
			}

			if routine.(map[string]interface{})["id"] != r.id {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- ID field mismatch: expected %s, got %s -- action: %s",
					id,
					module,
					funcname,
					r.id,
					routine.(map[string]interface{})["id"],
					test.name,
				)
				return
			}
			if routine.(map[string]interface{})["status"] != test.want[idx] {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- status field mismatch: expected %s, got %s -- action: %s",
					id,
					module,
					funcname,
					test.want[idx],
					routine.(map[string]interface{})["status"],
					test.name,
				)
				return
			}

			stackmap, ok := routine.(map[string]interface{})["stack"]

			if !ok {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- failed to retrieve stack map: expected key %s, wasn't found -- action: %s",
					id,
					module,
					funcname,
					"stack",
					test.name,
				)
				return
			}

			for _, f := range stackmap.([]map[string]interface{}) {
				if f["method"] == "" {
					t.Errorf(
						"#%v -- FAILED -- [%s] [%s] -- stack map method key is empty -- action: %s",
						id,
						module,
						funcname,
						test.name,
					)
					return
				}
				if f["reference"] == "" {
					t.Errorf(
						"#%v -- FAILED -- [%s] [%s] -- stack map reference key is empty -- action: %s",
						id,
						module,
						funcname,
						test.name,
					)
					return
				}
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
		stack := test.stack.mapCallStack()

		verify(id, test, stack)
	}
}

func TestStacktraceAsMap(t *testing.T) {
	module := "Trace"
	funcname := "asMap()"

	type test struct {
		name        string
		stack       *stacktrace
		all         bool
		minRoutines int
		minID       int
		want        []string
	}

	var tests = []test{
		{
			name:        "converting callstack map: all = true",
			stack:       newCallStack().getCallStack(true).splitCallStack().parseCallStack().mapCallStack(),
			all:         true,
			minRoutines: 2,
			minID:       1,
			want:        []string{"running", "chan receive"},
		},
		{
			name:        "converting callstack map: all = false",
			stack:       newCallStack().getCallStack(false).splitCallStack().parseCallStack().mapCallStack(),
			all:         false,
			minRoutines: 1,
			minID:       1,
			want:        []string{"running"},
		},
		{
			name:        "converting callstack map: defaults",
			stack:       newCallStack(),
			all:         false,
			minRoutines: 1,
			minID:       1,
			want:        []string{"running"},
		},
	}

	var verify = func(id int, test test, stack *stacktrace) {
		callmap := stack.asMap()

		for _, v := range callmap {
			routine := v.(map[string]interface{})
			fmt.Println(routine["id"])

			if routine["id"] == nil || routine["id"] == "" {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- empty ID field -- action: %s",
					id,
					module,
					funcname,
					test.name,
				)
				return
			}

			var match bool
			for _, s := range test.want {
				if routine["status"] == s {
					match = true
					break
				}
			}

			if !match {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- status mismatch -- action: %s",
					id,
					module,
					funcname,
					test.name,
				)
				return
			}

			if len(routine["stack"].([]map[string]interface{})) <= 0 {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] -- call stack is unexpectedly empty -- action: %s",
					id,
					module,
					funcname,
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
		verify(id, test, test.stack)
	}
}
