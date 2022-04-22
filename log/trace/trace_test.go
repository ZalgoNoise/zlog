package trace

import (
	"fmt"
	"strconv"
	"testing"
)

func TestNew(t *testing.T) {
	var tests = []struct {
		all bool
	}{
		{
			all: true,
		},
		{
			all: false,
		},
	}

	for id, test := range tests {
		stack := New(test.all)

		if len(stack) == 0 {
			t.Errorf(
				"#%v -- FAILED -- trace.New(%v) -- empty map",
				id,
				test.all,
			)
		}

		t.Logf(
			"#%v -- PASSED -- trace.NEW(%v)",
			id,
			test.all,
		)

	}
}

func TestGetCallStack(t *testing.T) {
	type test struct {
		stack *stacktrace
		all   bool
		want  []byte
	}

	var tests = []test{
		{
			stack: newCallStack(),
			all:   true,
			want:  []byte("goroutine"),
		},
		{
			stack: newCallStack(),
			all:   false,
			want:  []byte("goroutine"),
		},
	}

	var verify = func(id int, test test, stack *stacktrace) {
		if len(stack.buf) <= 0 {
			t.Errorf(
				"#%v -- FAILED -- stacktrace.getCallStack(%v) -- empty buffer error: buffer is %v bytes in length",
				id,
				test.all,
				len(stack.buf),
			)
			return
		}

		header := stack.buf[:9]
		for idx, c := range header {
			if c != test.want[idx] {
				t.Errorf(
					"#%v -- FAILED -- stacktrace.getCallStack(%v) -- character mismatch error: got %s ; wanted %s",
					id,
					test.all,
					string(test.want),
					string(header),
				)
				return
			}
		}

		t.Logf(
			"#%v -- PASSED -- stacktrace.getCallStack(%v) -- %s",
			id,
			test.all,
			string(stack.buf),
		)
	}

	for id, test := range tests {
		stack := test.stack.getCallStack(test.all)

		verify(id, test, stack)
	}

}

func TestSplitCallStack(t *testing.T) {
	type test struct {
		stack    *stacktrace
		all      bool
		minLines int
		want     []byte
	}

	var tests = []test{
		{
			stack:    newCallStack().getCallStack(true),
			all:      true,
			want:     []byte("goroutine"),
			minLines: 21,
		},
		{
			stack:    newCallStack().getCallStack(false),
			all:      false,
			want:     []byte("goroutine"),
			minLines: 9,
		},
		{
			stack:    newCallStack(),
			all:      false,
			want:     []byte("goroutine"),
			minLines: 9,
		},
	}

	var verify = func(id int, test test, stack *stacktrace) {
		if len(stack.buf) <= 0 {
			t.Errorf(
				"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack() -- empty buffer error: buffer is %v bytes in length",
				id,
				test.all,
				len(stack.buf),
			)
			return
		}

		if len(stack.split) < test.minLines {
			t.Errorf(
				"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack() -- buffer splitting issue: wanted %v lines ; got %v lines",
				id,
				test.all,
				test.minLines,
				len(stack.split),
			)
			return
		}

		header := stack.split[0][:9]
		for idx, c := range header {
			if c != test.want[idx] {
				t.Errorf(
					"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack() -- character mismatch error: got %s ; wanted %s",
					id,
					test.all,
					string(test.want),
					string(header),
				)
				return
			}
		}

		t.Logf(
			"#%v -- PASSED -- stacktrace.getCallStack(%v).splitCallStack() -- %s",
			id,
			test.all,
			string(stack.buf),
		)
	}

	for id, test := range tests {
		stack := test.stack.splitCallStack()

		verify(id, test, stack)
	}
}

func TestParseCallStack(t *testing.T) {
	type test struct {
		stack       *stacktrace
		all         bool
		minRoutines int
		minID       int
		want        []string
	}

	var tests = []test{
		{
			stack:       newCallStack().getCallStack(true).splitCallStack(),
			all:         true,
			minRoutines: 2,
			minID:       1,
			want:        []string{"running", "chan receive"},
		},
		{
			stack:       newCallStack().getCallStack(false).splitCallStack(),
			all:         false,
			minRoutines: 1,
			minID:       1,
			want:        []string{"running"},
		},
		{
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
				"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack() -- empty buffer error: buffer is %v bytes in length",
				id,
				test.all,
				len(stack.buf),
			)
			return
		}

		if len(stack.stacks) < test.minRoutines {
			t.Errorf(
				"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack() -- parsed less goroutines than expected: wanted %v ; got %v",
				id,
				test.all,
				test.minRoutines,
				len(stack.stacks),
			)
			return
		}

		for idx, r := range stack.stacks {
			id, err := strconv.Atoi(r.id)
			if err != nil {
				t.Errorf(
					"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack() -- failed to convert ID with error: %s",
					id,
					test.all,
					err,
				)
				return
			}

			if id < test.minID {
				t.Errorf(
					"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack() -- ID is lower than minimum: wanted > %v ; got %v",
					id,
					test.all,
					test.minID,
					id,
				)
				return
			}

			if r.status != test.want[idx] {
				t.Errorf(
					"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack() -- status mismatch: wanted %s ; got %s",
					id,
					test.all,
					test.want,
					r.status,
				)
				return
			}
		}

		t.Logf(
			"#%v -- PASSED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack() -- %s",
			id,
			test.all,
			string(stack.buf),
		)
	}

	for id, test := range tests {
		stack := test.stack.parseCallStack()

		verify(id, test, stack)
	}
}

func TestMapCallStack(t *testing.T) {
	type test struct {
		stack       *stacktrace
		all         bool
		minRoutines int
		minID       int
		want        []string
	}

	var tests = []test{
		{
			stack:       newCallStack().getCallStack(true).splitCallStack().parseCallStack(),
			all:         true,
			minRoutines: 2,
			minID:       1,
			want:        []string{"running", "chan receive"},
		},
		{
			stack:       newCallStack().getCallStack(false).splitCallStack().parseCallStack(),
			all:         false,
			minRoutines: 1,
			minID:       1,
			want:        []string{"running"},
		},
		{
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
				"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack() -- empty buffer error: buffer is %v bytes in length",
				id,
				test.all,
				len(stack.buf),
			)
			return
		}

		if len(stack.stacks) < test.minRoutines {
			t.Errorf(
				"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack() -- parsed less goroutines than expected: wanted %v ; got %v",
				id,
				test.all,
				test.minRoutines,
				len(stack.stacks),
			)
			return
		}

		for idx, r := range stack.stacks {
			id, err := strconv.Atoi(r.id)
			if err != nil {
				t.Errorf(
					"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack() -- failed to convert ID with error: %s",
					id,
					test.all,
					err,
				)
				return
			}

			if id < test.minID {
				t.Errorf(
					"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack() -- ID is lower than minimum: wanted > %v ; got %v",
					id,
					test.all,
					test.minID,
					id,
				)
				return
			}

			if r.status != test.want[idx] {
				t.Errorf(
					"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack() -- status mismatch: wanted %s ; got %s",
					id,
					test.all,
					test.want,
					r.status,
				)
				return
			}

			callmap := stack.out

			routine, ok := callmap["goroutine-"+r.id]
			if !ok {
				t.Errorf(
					"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack() -- field not found in output object: expected key %s, wasn't found",
					id,
					test.all,
					"goroutine-"+r.id,
				)
				return
			}

			if routine.(map[string]interface{})["id"] != r.id {
				t.Errorf(
					"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack() -- ID field mismatch: expected %s, got %s",
					id,
					test.all,
					r.id,
					routine.(map[string]interface{})["id"],
				)
				return
			}
			if routine.(map[string]interface{})["status"] != test.want[idx] {
				t.Errorf(
					"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack() -- status field mismatch: expected %s, got %s",
					id,
					test.all,
					test.want[idx],
					routine.(map[string]interface{})["status"],
				)
				return
			}

			stackmap, ok := routine.(map[string]interface{})["stack"]

			if !ok {
				t.Errorf(
					"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack() -- failed to retrieve stack map: expected key %s, wasn't found",
					id,
					test.all,
					"stack",
				)
				return
			}

			for _, f := range stackmap.([]map[string]interface{}) {
				if f["method"] == "" {
					t.Errorf(
						"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack() -- stack map method key is empty",
						id,
						test.all,
					)
					return
				}
				if f["reference"] == "" {
					t.Errorf(
						"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack() -- stack map reference key is empty",
						id,
						test.all,
					)
					return
				}
			}

		}

		t.Logf(
			"#%v -- PASSED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack() -- %s",
			id,
			test.all,
			string(stack.buf),
		)
	}

	for id, test := range tests {
		stack := test.stack.mapCallStack()

		verify(id, test, stack)
	}
}

func TestStacktraceToMap(t *testing.T) {
	type test struct {
		stack       *stacktrace
		all         bool
		minRoutines int
		minID       int
		want        []string
	}

	var tests = []test{
		{
			stack:       newCallStack().getCallStack(true).splitCallStack().parseCallStack().mapCallStack(),
			all:         true,
			minRoutines: 2,
			minID:       1,
			want:        []string{"running", "chan receive"},
		},
		{
			stack:       newCallStack().getCallStack(false).splitCallStack().parseCallStack().mapCallStack(),
			all:         false,
			minRoutines: 1,
			minID:       1,
			want:        []string{"running"},
		},
		{
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
					"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack().toMap() -- empty ID field",
					id,
					test.all,
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
					"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack().toMap() -- status mismatch",
					id,
					test.all,
				)
				return
			}

			if len(routine["stack"].([]map[string]interface{})) <= 0 {
				t.Errorf(
					"#%v -- FAILED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack().toMap() -- call stack is unexpectedly empty",
					id,
					test.all,
				)
				return
			}
		}

		t.Logf(
			"#%v -- PASSED -- stacktrace.getCallStack(%v).splitCallStack().parseCallStack().mapCallStack().toMap() -- %s",
			id,
			test.all,
			callmap,
		)
	}

	for id, test := range tests {
		verify(id, test, test.stack)
	}
}
