package client

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/zalgonoise/zlog/log/event"
)

var emptyLogFunc logFunc = func(*event.Event) {}

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
			moreThan: time.Millisecond * 2,
		},
		{
			name:     "third iteration",
			attempt:  3,
			lessThan: time.Second * 2,
			moreThan: time.Millisecond * 5,
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

func TestBackoffInit(t *testing.T) {
	module := "Backoff"
	funcname := "NewBackoff() / init()"

	_ = module
	_ = funcname

	type test struct {
		name    string
		builder *gRPCLogClientBuilder
		client  *GRPCLogClient
	}

	var tests = []test{
		{
			name: "register for unary RPC",
			builder: &gRPCLogClientBuilder{
				isUnary: true,
			},
			client: &GRPCLogClient{},
		},
		{
			name:    "register for streaming RPC",
			builder: &gRPCLogClientBuilder{},
			client:  &GRPCLogClient{},
		},
	}

	var verify = func(idx int, test test) {
		b := NewBackoff()

		b.init(test.builder, test.client)
	}

	for idx, test := range tests {
		verify(idx, test)
	}

}

func TestBackoffFunc(t *testing.T) {
	module := "Backoff"
	funcname := "BackoffFunc()"

	_ = module
	_ = funcname

	type test struct {
		name string
		fn   BackoffFunc
	}

	var tests = []test{
		{
			name: "try NoBackoff()",
			fn:   NoBackoff(),
		},
		{
			name: "try BackoffLinear(time.Second)",
			fn:   BackoffLinear(time.Second),
		},
		{
			name: "try BackoffIncremental(time.Second)",
			fn:   BackoffIncremental(time.Second),
		},
		{
			name: "try BackoffExponential()",
			fn:   BackoffExponential(),
		},
	}

	var verify = func(idx int, test test) {
		b := NewBackoff()

		// remove defaults to test setter action
		b.backoffFunc = nil

		b.BackoffFunc(test.fn)

		if b.backoffFunc == nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] backoff function was unset despite the %s call -- action: %s",
				idx,
				module,
				funcname,
				funcname,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestWait(t *testing.T) {
	module := "Backoff"
	funcname := "Wait()"

	_ = module
	_ = funcname

	type test struct {
		name string
		b    *Backoff
		ok   bool
		err  error
	}

	var streamFn streamFunc = func() {}
	var unaryFn logFunc = func(*event.Event) {}

	var tests = []test{
		{
			name: "unlocked, normal unary flow, zero counter",
			b: &Backoff{
				counter: 0,
				max:     time.Minute,
				wait:    time.Second,
				call:    unaryFn,
				msg: []*event.Event{
					event.New().Message("null").Build(), // added message for added coverage on L165
				},
				backoffFunc: BackoffLinear(time.Second),
				locked:      false,
				mu:          sync.Mutex{},
			},
			ok: true,
		},
		{
			name: "unlocked, normal stream flow, zero counter",
			b: &Backoff{
				counter:     0,
				max:         time.Minute,
				wait:        time.Second,
				call:        streamFn,
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Second),
				locked:      false,
				mu:          sync.Mutex{},
			},
			ok: true,
		},
		{
			name: "locked error",
			b: &Backoff{
				counter:     0,
				max:         time.Minute,
				wait:        time.Second,
				call:        unaryFn,
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Second),
				locked:      true,
				mu:          sync.Mutex{},
			},
			err: ErrBackoffLocked,
		},
		{
			name: "zero wait error",
			b: &Backoff{
				counter:     0,
				max:         time.Minute,
				wait:        0,
				call:        unaryFn,
				msg:         []*event.Event{},
				backoffFunc: NoBackoff(),
				locked:      false,
				mu:          sync.Mutex{},
			},
			err: ErrFailedConn,
		},
		{
			name: "invalid call's function type error",
			b: &Backoff{
				counter:     0,
				max:         time.Minute,
				wait:        0,
				call:        func() {},
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Second),
				locked:      false,
				mu:          sync.Mutex{},
			},
			err: ErrInvalidType,
		},
		{
			name: "failed retry error",
			b: &Backoff{
				counter:     0,
				max:         time.Second,
				wait:        time.Minute,
				call:        unaryFn,
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Minute * 2),
				locked:      false,
				mu:          sync.Mutex{},
			},
			err: ErrFailedRetry,
		},
	}

	var verify = func(idx int, test test) {
		fn, err := test.b.Wait()

		if err != nil && test.ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		} else if err != nil && test.err != nil {
			if !errors.Is(err, test.err) {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] error mismatch: wanted %v ; got %v -- action: %s",
					idx,
					module,
					funcname,
					test.err,
					err,
					test.name,
				)
				return
			}
		}

		if fn == nil && test.ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected nil return function -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		} else if fn != nil {
			// added coverage with empty calls
			fn()
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestWaitContext(t *testing.T) {
	module := "Backoff"
	funcname := "WaitContext()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		b     *Backoff
		ok    bool
		err   error
		close bool
	}

	var streamFn streamFunc = func() {}
	var unaryFn logFunc = func(*event.Event) {}

	var tests = []test{
		{
			name: "unlocked, normal unary flow, zero counter",
			b: &Backoff{
				counter: 0,
				max:     time.Minute,
				wait:    time.Second,
				call:    unaryFn,
				msg: []*event.Event{
					event.New().Message("null").Build(), // added message for added coverage on L165
				},
				backoffFunc: BackoffLinear(time.Second),
				locked:      false,
				mu:          sync.Mutex{},
			},
			ok: true,
		},
		{
			name: "unlocked, normal stream flow, zero counter",
			b: &Backoff{
				counter:     0,
				max:         time.Minute,
				wait:        time.Second,
				call:        streamFn,
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Second),
				locked:      false,
				mu:          sync.Mutex{},
			},
			ok: true,
		},
		{
			name: "locked error",
			b: &Backoff{
				counter:     0,
				max:         time.Minute,
				wait:        time.Second,
				call:        unaryFn,
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Second),
				locked:      true,
				mu:          sync.Mutex{},
			},
			err: ErrBackoffLocked,
		},
		{
			name: "zero wait error",
			b: &Backoff{
				counter:     0,
				max:         time.Minute,
				wait:        0,
				call:        unaryFn,
				msg:         []*event.Event{},
				backoffFunc: NoBackoff(),
				locked:      false,
				mu:          sync.Mutex{},
			},
			err: ErrFailedConn,
		},
		{
			name: "invalid call's function type error",
			b: &Backoff{
				counter:     0,
				max:         time.Minute,
				wait:        0,
				call:        func() {},
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Second),
				locked:      false,
				mu:          sync.Mutex{},
			},
			err: ErrInvalidType,
		},
		{
			name: "failed retry error",
			b: &Backoff{
				counter:     0,
				max:         time.Second,
				wait:        time.Minute,
				call:        unaryFn,
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Minute * 2),
				locked:      false,
				mu:          sync.Mutex{},
			},
			err: ErrFailedRetry,
		},
		{
			name: "cancelled context error",
			b: &Backoff{
				counter:     0,
				max:         time.Minute,
				wait:        time.Second,
				call:        streamFn,
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Second),
				locked:      false,
				mu:          sync.Mutex{},
			},
			close: true,
		},
	}

	var verify = func(idx int, test test) {
		ctx := context.Background()
		cctx, cancel := context.WithCancel(ctx)

		if test.close {
			go func() {
				time.Sleep(time.Millisecond)
				cancel()
				return
			}()
		} else {
			defer cancel()
		}

		fn, err := test.b.WaitContext(cctx)

		if err != nil && test.ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		} else if err != nil && test.err != nil {
			if !errors.Is(err, test.err) {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] error mismatch: wanted %v ; got %v -- action: %s",
					idx,
					module,
					funcname,
					test.err,
					err,
					test.name,
				)
				return
			}
		}

		if fn == nil && test.ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected nil return function -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		} else if fn != nil {
			// added coverage with empty calls
			fn()
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
