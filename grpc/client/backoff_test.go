package client

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

var emptyLogFunc logFunc = func(*event.Event) {}

const maxWait time.Duration = time.Millisecond * 50

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
				max:     time.Second,
				wait:    time.Millisecond,
				call:    unaryFn,
				msg: []*event.Event{
					event.New().Message("null").Build(), // added message for added coverage on L165
				},
				backoffFunc: BackoffLinear(time.Millisecond),
				locked:      false,
				mu:          sync.Mutex{},
			},
			ok: true,
		},
		{
			name: "unlocked, normal stream flow, zero counter",
			b: &Backoff{
				counter:     0,
				max:         time.Second,
				wait:        time.Millisecond,
				call:        streamFn,
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Millisecond),
				locked:      false,
				mu:          sync.Mutex{},
			},
			ok: true,
		},
		{
			name: "locked error",
			b: &Backoff{
				counter:     0,
				max:         time.Second,
				wait:        time.Millisecond,
				call:        unaryFn,
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Millisecond),
				locked:      true,
				mu:          sync.Mutex{},
			},
			err: ErrBackoffLocked,
		},
		{
			name: "zero wait error",
			b: &Backoff{
				counter:     0,
				max:         time.Second,
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
				max:         time.Second,
				wait:        0,
				call:        func() {},
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Millisecond),
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
				wait:        time.Millisecond,
				call:        unaryFn,
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Second * 2),
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
				max:     time.Second,
				wait:    time.Millisecond,
				call:    unaryFn,
				msg: []*event.Event{
					event.New().Message("null").Build(), // added message for added coverage on L165
				},
				backoffFunc: BackoffLinear(time.Millisecond),
				locked:      false,
				mu:          sync.Mutex{},
			},
			ok: true,
		},
		{
			name: "unlocked, normal stream flow, zero counter",
			b: &Backoff{
				counter:     0,
				max:         time.Second,
				wait:        time.Millisecond,
				call:        streamFn,
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Millisecond),
				locked:      false,
				mu:          sync.Mutex{},
			},
			ok: true,
		},
		{
			name: "locked error",
			b: &Backoff{
				counter:     0,
				max:         time.Second,
				wait:        time.Millisecond,
				call:        unaryFn,
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Millisecond),
				locked:      true,
				mu:          sync.Mutex{},
			},
			err: ErrBackoffLocked,
		},
		{
			name: "zero wait error",
			b: &Backoff{
				counter:     0,
				max:         time.Second,
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
				max:         time.Second,
				wait:        0,
				call:        func() {},
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Millisecond),
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
				wait:        time.Millisecond,
				call:        unaryFn,
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Second * 2),
				locked:      false,
				mu:          sync.Mutex{},
			},
			err: ErrFailedRetry,
		},
		{
			name: "cancelled context error",
			b: &Backoff{
				counter:     0,
				max:         time.Second,
				wait:        time.Millisecond,
				call:        streamFn,
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Millisecond),
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

func TestRegister(t *testing.T) {
	module := "Backoff"
	funcname := "Register()"

	_ = module
	_ = funcname

	type test struct {
		name    string
		call    interface{}
		isUnary bool
		ok      bool
	}

	backoff := &Backoff{
		counter:     0,
		max:         time.Minute,
		wait:        time.Second,
		call:        nil,
		msg:         []*event.Event{},
		backoffFunc: BackoffLinear(time.Second),
		locked:      false,
		mu:          sync.Mutex{},
	}

	var streamFn streamFunc = func() {}
	var unaryFn logFunc = func(*event.Event) {}

	var tests = []test{
		{
			name:    "unary function",
			call:    unaryFn,
			isUnary: true,
			ok:      true,
		},
		{
			name: "stream function",
			call: streamFn,
			ok:   true,
		},
		{
			name: "invalid function",
			call: func() {},
		},
	}

	var verify = func(idx int, test test) {

		new := backoff
		new.call = nil

		if test.isUnary && test.ok {
			new.Register(test.call.(logFunc))

			if new.call == nil {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unary call wasn't set as expected -- action: %s",
					idx,
					module,
					funcname,
					test.name,
				)
				return
			}

		} else if !test.isUnary && test.ok {
			new.Register(test.call.(streamFunc))

			if new.call == nil {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] stream call wasn't set as expected -- action: %s",
					idx,
					module,
					funcname,
					test.name,
				)
				return
			}
		} else {
			// no valid input function, call element should remain nil
			if new.call != nil {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] backoff function call should be nil -- action: %s",
					idx,
					module,
					funcname,
					test.name,
				)
				return
			}
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestUnaryBackoffHandler(t *testing.T) {
	module := "Backoff"
	funcname := "UnaryBackoffHandler()"

	_ = module
	_ = funcname

	type test struct {
		name   string
		b      *Backoff
		err    error
		logger log.Logger
		wants  error
	}

	var unaryFn logFunc = func(*event.Event) {}
	var nilL log.Logger = log.New(log.NilConfig)
	var err error = errors.New("test error")

	var tests = []test{
		{
			name: "valid flow w/ error",
			b: &Backoff{
				counter: 0,
				max:     time.Minute,
				wait:    time.Second,
				call:    unaryFn,
				msg: []*event.Event{
					event.New().Message("null").Build(), // added message for added coverage on L165
				},
				backoffFunc: BackoffLinear(time.Millisecond),
				locked:      false,
				mu:          sync.Mutex{},
			},
			err:    err,
			logger: nilL,
			wants:  ErrFailedConn,
		},
		{
			name: "valid flow w/ locked backoff",
			b: &Backoff{
				counter:     0,
				max:         time.Minute,
				wait:        time.Second,
				call:        nil,
				msg:         []*event.Event{},
				backoffFunc: NoBackoff(),
				locked:      true,
				mu:          sync.Mutex{},
			},
			err:    err,
			logger: nilL,
			wants:  ErrBackoffLocked,
		},
		{
			name: "invalid flow w/ error",
			b: &Backoff{
				counter:     0,
				max:         time.Minute,
				wait:        time.Second,
				call:        nil,
				msg:         []*event.Event{},
				backoffFunc: NoBackoff(),
				locked:      false,
				mu:          sync.Mutex{},
			},
			err:    err,
			logger: nilL,
			wants:  ErrFailedConn,
		},
	}

	var verify = func(idx int, test test) {
		err := test.b.UnaryBackoffHandler(test.err, test.logger)

		if !errors.Is(err, test.wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted error %v ; got error %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				err,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestStreamBackoffHandler(t *testing.T) {
	module := "Backoff"
	funcname := "UnaryBackoffHandler()"

	_ = module
	_ = funcname

	type test struct {
		name   string
		b      *Backoff
		logger log.Logger
		wants  error
	}

	type testRequest struct {
		ctx    context.Context
		cancel context.CancelFunc
		errCh  chan error
		done   chan struct{}
	}

	var streamFn streamFunc = func() {}
	var nilL log.Logger = log.New(log.NilConfig)

	var tests = []test{
		{
			name: "valid flow w/o error",
			b: &Backoff{
				counter:     0,
				max:         time.Minute,
				wait:        time.Second,
				call:        streamFn,
				msg:         []*event.Event{},
				backoffFunc: BackoffLinear(time.Millisecond),
				locked:      false,
				mu:          sync.Mutex{},
			},
			logger: nilL,
		},
		{
			name: "valid flow w/o error",
			b: &Backoff{
				counter:     0,
				max:         time.Minute,
				wait:        time.Second,
				call:        streamFn,
				msg:         []*event.Event{},
				backoffFunc: NoBackoff(),
				locked:      true,
				mu:          sync.Mutex{},
			},
			logger: nilL,
			wants:  ErrBackoffLocked,
		},
	}

	var init = func() *testRequest {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)

		var errCh chan error
		var done chan struct{}

		return &testRequest{
			ctx:    ctx,
			cancel: cancel,
			errCh:  errCh,
			done:   done,
		}
	}

	var verify = func(idx int, test test) {
		var err error

		req := init()

		go test.b.StreamBackoffHandler(
			req.errCh,
			req.cancel,
			test.logger,
			req.done,
		)

		for {
			select {
			case <-time.After(maxWait):
				// call executed, no signals received
				return
			case <-req.done:
				// ignore done signal, focus on error channel
				continue
			case err = <-req.errCh:
				if !errors.Is(err, test.wants) {
					t.Errorf(
						"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted error %v ; got error %v -- action: %s",
						idx,
						module,
						funcname,
						test.wants,
						err,
						test.name,
					)
					return
				}
				return
			}
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func FuzzBackoffTime(f *testing.F) {
	module := "Backoff"
	funcname := "Time()"
	action := "fuzz testing *Backoff.Time(t time.Duration)"

	var fuzzInput = []int{0, 1, 5, 30}

	for _, i := range fuzzInput {
		f.Add(i)
	}

	f.Fuzz(func(t *testing.T, a int) {

		b := &Backoff{
			counter:     0,
			max:         0,
			wait:        time.Second,
			call:        nil,
			msg:         []*event.Event{},
			backoffFunc: NoBackoff(),
			locked:      false,
			mu:          sync.Mutex{},
		}

		var duration = time.Duration(int64(a))

		b.Time(duration)

		if !reflect.DeepEqual(duration, b.max) {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed time mismatch: wanted %v ; got %v -- action: %s",
				module,
				funcname,
				a,
				duration,
				action,
			)
			return
		}
	})
}

func FuzzBackoffAddMessage(f *testing.F) {
	module := "Backoff"
	funcname := "AddMessage()"
	action := "fuzz testing *Backoff.AddMessage(msg *event.Event); checking a fuzzed message body"

	var fuzzInput = []string{"null", "test", "fuzz"}

	for _, i := range fuzzInput {
		f.Add(i)
	}

	f.Fuzz(func(t *testing.T, a string) {

		b := &Backoff{
			counter:     0,
			max:         0,
			wait:        time.Second,
			call:        nil,
			msg:         []*event.Event{},
			backoffFunc: NoBackoff(),
			locked:      false,
			mu:          sync.Mutex{},
		}

		var msg = event.New().Message(a).Build()

		b.AddMessage(msg)

		if b.msg[0].GetMsg() != a {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed message content mismatch: wanted %s ; got %s -- action: %s",
				module,
				funcname,
				a,
				b.msg[0].GetMsg(),
				action,
			)
			return
		}
	})
}

func FuzzBackoffCounter(f *testing.F) {
	module := "Backoff"
	funcname := "Counter()"
	action := "fuzz testing *Backoff.Counter() uint"

	var fuzzInput = []uint{0, 1, 5, 30}

	for _, i := range fuzzInput {
		f.Add(i)
	}

	f.Fuzz(func(t *testing.T, a uint) {

		b := &Backoff{
			counter:     a,
			max:         0,
			wait:        time.Second,
			call:        nil,
			msg:         []*event.Event{},
			backoffFunc: NoBackoff(),
			locked:      false,
			mu:          sync.Mutex{},
		}

		if !reflect.DeepEqual(b.Counter(), a) {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed counter mismatch: wanted %v ; got %v -- action: %s",
				module,
				funcname,
				a,
				b.Counter(),
				action,
			)
			return
		}
	})
}

func FuzzBackoffMax(f *testing.F) {
	module := "Backoff"
	funcname := "Max()"
	action := "fuzz testing *Backoff.Max() string"

	var fuzzInput = []int{0, 1, 5, 30}

	for _, i := range fuzzInput {
		f.Add(i)
	}

	f.Fuzz(func(t *testing.T, a int) {
		var max = time.Duration(int64(a))

		b := &Backoff{
			counter:     0,
			max:         max,
			wait:        time.Second,
			call:        nil,
			msg:         []*event.Event{},
			backoffFunc: NoBackoff(),
			locked:      false,
			mu:          sync.Mutex{},
		}

		if b.Max() != max.String() {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed max time mismatch: wanted %v ; got %v -- action: %s",
				module,
				funcname,
				max.String(),
				b.Max(),
				action,
			)
			return
		}
	})
}

func FuzzBackoffCurrent(f *testing.F) {
	module := "Backoff"
	funcname := "Current()"
	action := "fuzz testing *Backoff.Current() string"

	var fuzzInput = []int{0, 1, 5, 30}

	for _, i := range fuzzInput {
		f.Add(i)
	}

	f.Fuzz(func(t *testing.T, a int) {
		var cur = time.Duration(int64(a))

		b := &Backoff{
			counter:     0,
			max:         0,
			wait:        cur,
			call:        nil,
			msg:         []*event.Event{},
			backoffFunc: NoBackoff(),
			locked:      false,
			mu:          sync.Mutex{},
		}

		if b.Current() != cur.String() {
			t.Errorf(
				"FAILED -- [%s] [%s] fuzzed max time mismatch: wanted %v ; got %v -- action: %s",
				module,
				funcname,
				cur.String(),
				b.Current(),
				action,
			)
			return
		}
	})
}

func TestBackoffLock(t *testing.T) {
	module := "Backoff"
	funcname := "Lock()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		b     *Backoff
		init  bool
		wants bool
	}

	mu := new(sync.Mutex)
	mu.Lock()

	var tests = []test{
		{
			name: "from unlocked backoff",
			b: &Backoff{
				locked: false,
				mu:     sync.Mutex{},
			},
			wants: true,
		},
		{
			name: "from locked backoff",
			b: &Backoff{
				locked: true,
				mu:     *mu,
			},
			wants: true,
		},
	}

	var verify = func(idx int, test test) {
		test.b.Lock()

		if test.b.locked != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				test.b.locked,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestBackoffUnlock(t *testing.T) {
	module := "Backoff"
	funcname := "Unlock()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		b     *Backoff
		wants bool
	}

	mu := new(sync.Mutex)
	mu.Lock()

	var tests = []test{
		{
			name: "from unlocked backoff",
			b: &Backoff{
				locked: false,
				mu:     sync.Mutex{},
			},
			wants: false,
		},
		{
			name: "from locked backoff",
			b: &Backoff{
				locked: true,
				mu:     *mu,
			},
			wants: false,
		},
	}

	var verify = func(idx int, test test) {
		test.b.Unlock()

		if test.b.locked != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				test.b.locked,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestBackoffTryLock(t *testing.T) {
	module := "Backoff"
	funcname := "TryLock()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		b     *Backoff
		wants bool
	}

	mu := new(sync.Mutex)
	mu.Lock()

	var tests = []test{
		{
			name: "from unlocked backoff",
			b: &Backoff{
				locked: false,
				mu:     sync.Mutex{},
			},
			wants: true,
		},
		{
			name: "from locked backoff",
			b: &Backoff{
				locked: true,
				mu:     *mu,
			},
			wants: true,
		},
	}

	var verify = func(idx int, test test) {
		res := test.b.TryLock()

		if res != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				res,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestBackoffIsLocked(t *testing.T) {
	module := "Backoff"
	funcname := "IsLocked()"

	_ = module
	_ = funcname

	type test struct {
		name  string
		b     *Backoff
		wants bool
	}

	mu := new(sync.Mutex)
	mu.Lock()

	var tests = []test{
		{
			name: "from unlocked backoff",
			b: &Backoff{
				locked: false,
			},
			wants: false,
		},
		{
			name: "from locked backoff",
			b: &Backoff{
				locked: true,
			},
			wants: true,
		},
	}

	var verify = func(idx int, test test) {
		res := test.b.IsLocked()

		if res != test.wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] output mismatch error: wanted %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				res,
				test.name,
			)
			return
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
