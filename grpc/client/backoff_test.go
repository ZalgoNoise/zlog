package client

import (
	"reflect"
	"testing"
	"time"
)

func TestNoBackoff(t *testing.T) {
	module := "Backoff"
	funcname := "NoBackoff()"

	_ = module
	_ = funcname

	type test struct {
		name    string
		attempt uint
		wants   time.Duration
	}

	var tests = []test{
		{
			name:    "first iteration",
			attempt: 1,
			wants:   0,
		},
		{
			name:    "second iteration",
			attempt: 2,
			wants:   0,
		},
		{
			name:    "third iteration",
			attempt: 3,
			wants:   0,
		},
		{
			name:    "tenth iteration",
			attempt: 10,
			wants:   0,
		},
	}

	var verify = func(idx int, test test) {
		f := NoBackoff()

		duration := f(test.attempt)

		if !reflect.DeepEqual(test.wants, duration) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				duration,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestBackoffLinear(t *testing.T) {
	module := "Backoff"
	funcname := "BackoffLinear()"

	_ = module
	_ = funcname

	type test struct {
		name    string
		input   time.Duration
		attempt uint
		wants   time.Duration
	}

	var tests = []test{
		{
			name:    "zero input; default",
			input:   0,
			attempt: 1,
			wants:   defaultWaitBetween,
		},
		{
			name:    "first iteration",
			input:   time.Second,
			attempt: 1,
			wants:   time.Second,
		},
		{
			name:    "second iteration",
			input:   time.Second,
			attempt: 2,
			wants:   time.Second,
		},
		{
			name:    "third iteration",
			input:   time.Second,
			attempt: 3,
			wants:   time.Second,
		},
		{
			name:    "tenth iteration",
			input:   time.Second,
			attempt: 10,
			wants:   time.Second,
		},
	}

	var verify = func(idx int, test test) {
		f := BackoffLinear(test.input)

		duration := f(test.attempt)

		if !reflect.DeepEqual(test.wants, duration) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				duration,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestBackoffIncremental(t *testing.T) {
	module := "Backoff"
	funcname := "BackoffIncremental()"

	_ = module
	_ = funcname

	type test struct {
		name    string
		input   time.Duration
		attempt uint
		wants   time.Duration
	}

	var tests = []test{
		{
			name:    "first iteration",
			input:   time.Second,
			attempt: 1,
			wants:   time.Second,
		},
		{
			name:    "second iteration",
			input:   time.Second,
			attempt: 2,
			wants:   time.Second * 2,
		},
		{
			name:    "third iteration",
			input:   time.Second,
			attempt: 3,
			wants:   time.Second * 4,
		},
		{
			name:    "tenth iteration",
			input:   time.Second,
			attempt: 10,
			wants:   time.Second * 512,
		},
	}

	var verify = func(idx int, test test) {
		f := BackoffIncremental(test.input)

		duration := f(test.attempt)

		if !reflect.DeepEqual(test.wants, duration) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				duration,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestBackoffExponential(t *testing.T) {
	module := "Backoff"
	funcname := "BackoffExponential()"

	_ = module
	_ = funcname

	type test struct {
		name     string
		attempt  uint
		moreThan time.Duration
		lessThan time.Duration
	}

	var tests = []test{
		{
			name:     "first iteration",
			attempt:  1,
			lessThan: time.Second,
			moreThan: time.Millisecond,
		},
		{
			name:     "second iteration",
			attempt:  2,
			lessThan: time.Second * 2,
			moreThan: time.Millisecond * 5,
		},
		{
			name:     "third iteration",
			attempt:  3,
			lessThan: time.Second * 2,
			moreThan: time.Millisecond * 20,
		},
		{
			name:     "tenth iteration",
			attempt:  10,
			lessThan: time.Second * 3,
			moreThan: time.Second,
		},
	}

	var verify = func(idx int, test test) {
		f := BackoffExponential()

		duration := f(test.attempt)

		if duration > test.lessThan || duration < test.moreThan {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted between %v and %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.moreThan,
				test.lessThan,
				duration,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
