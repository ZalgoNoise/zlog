package client

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

var (
	ErrFailedRetry   error = errors.New("failed to connect to server after numerous retries")
	ErrFailedConn    error = errors.New("failed to connect to server")
	ErrInvalidType   error = errors.New("unsupported exponential backoff function type")
	ErrBackoffLocked error = errors.New("operations locked during exponential backoff")
)

const (
	defaultRetryTime   time.Duration = time.Second * 30
	defaultWaitBetween time.Duration = time.Second * 3
)

type streamFunc func()
type logFunc func(*event.Event)
type BackoffFunc func(uint) time.Duration

// Backoff struct defines the elements of a backoff module, which is configured
// by setting a BackoffFunc to define the interval between each attempt.
//
// Backoff will also try to act as a message buffer in case the server connection
// cannot be established -- as it will attempt to flush these records to the server
// as soon as connected.
//
// Also it has a simple lock / unlock switch for concurrent calls to be able to
// verify its state and halt by themselves
type Backoff struct {
	counter     uint
	max         time.Duration
	wait        time.Duration
	call        interface{}
	msg         []*event.Event
	backoffFunc BackoffFunc
	locked      bool
	mu          sync.Mutex
}

// NoBackoff function will return a BackoffFunc that overrides the
// backoff module by setting a zero wait-between duration. This is
// detected as a sign that the module should be overriden.
func NoBackoff() BackoffFunc {
	return func(attempt uint) time.Duration {
		return 0
	}
}

// BackoffLinear function will return a BackoffFunc that calculates
// a linear backoff according to the input duration. If the input
// duration is 0, then the default wait-between time is set.
func BackoffLinear(t time.Duration) BackoffFunc {
	if t == 0 {
		t = defaultWaitBetween
	}

	return func(attempt uint) time.Duration {
		return t
	}
}

// BackoffIncremental function will return a BackoffFunc that calculates
// exponential backoff according to a scalar method
func BackoffIncremental(scalar time.Duration) BackoffFunc {
	return func(attempt uint) time.Duration {
		return scalar * time.Duration((1<<attempt)>>1)
	}
}

// BackoffExponential function will return a BackoffFunc that calculates
// exponential backoff according to its standard
//
// Notes on exponential backoff: https://en.wikipedia.org/wiki/Exponential_backoff
func BackoffExponential() BackoffFunc {
	return func(attempt uint) time.Duration {
		return time.Millisecond * time.Duration(
			int64(math.Pow(2, float64(attempt)))+rand.New(
				rand.NewSource(time.Now().UnixNano())).Int63n(1000),
		)
	}
}

// NewBackoff function initializes a simple exponential backoff module with
// an exponential backoff with set default retry time of 30 seconds
func NewBackoff() *Backoff {
	b := &Backoff{
		max:         defaultRetryTime,
		backoffFunc: BackoffExponential(),
		counter:     0,
	}
	return b
}

func (b *Backoff) init(cb *gRPCLogClientBuilder, c *GRPCLogClient) *Backoff {
	if cb.isUnary {
		// register log() function in the backoff module
		var logF logFunc = c.log
		b.Register(logF)
	} else {
		// register stream() function in the backoff module
		var streamF streamFunc = c.stream
		b.Register(streamF)
	}
	return b
}

func (b *Backoff) BackoffFunc(f BackoffFunc) {
	b.backoffFunc = f
}

// Wait method will wait for the currently set wait time, if the module is unlocked.
//
// After waiting, it returns a func() to call (depending on what it is handling),
// and and an error.
//
// If the waiting time is grater than the deadline set, it will return with an
// ErrFailedRetry
func (b *Backoff) Wait() (func(), error) {

	if b.locked {
		return nil, ErrBackoffLocked
	}

	ok := b.TryLock()
	if !ok {
		return nil, ErrBackoffLocked
	}
	defer b.Unlock()

	b.counter = b.counter + 1
	b.wait = b.backoffFunc(b.counter)

	// exit early
	if b.wait == 0 {
		return nil, ErrFailedConn
	}

	if b.wait <= b.max {

		timer := time.NewTimer(b.wait)
		select {
		case <-timer.C:
		}

		switch v := b.call.(type) {
		case streamFunc:
			return func() {
				v()
			}, nil
		case logFunc:
			list := b.msg
			f := func() {
				for _, msg := range list {
					v(msg)
				}
			}
			return f, nil
		default:
			return nil, ErrInvalidType
		}
	}

	return nil, ErrFailedRetry

}

// WaitContext method will wait for the currently set wait time, if the module is
// unlocked. It also takes in a context, for which it will listen to its Done() signal.
//
// After waiting, it returns a func() to call (depending on what it is handling),
// and and an error.
//
// If the waiting time is grater than the deadline set, it will return with an
// ErrFailedRetry
func (b *Backoff) WaitContext(ctx context.Context) (func(), error) {

	if b.locked {
		return nil, ErrBackoffLocked
	}
	if b.wait <= b.max {
		ok := b.TryLock()
		if !ok {
			return nil, ErrBackoffLocked
		}
		defer b.Unlock()

		timer := time.NewTimer(b.wait)

		var err error

		select {
		case <-ctx.Done():
			timer.Stop()
			err = ctx.Err()
		case <-timer.C:
		}

		switch v := b.call.(type) {
		case streamFunc:
			return func() {
				v()
			}, err
		case logFunc:
			list := b.msg

			f := func() {
				for _, msg := range list {
					v(msg)
				}
			}
			return f, err
		default:
			return nil, ErrInvalidType
		}
	}

	return nil, ErrFailedRetry

}

// Register method will take in a function with the same signature as a stream() function
// and the error channel of the gRPC Log Client; and returns a pointer to itself for method chaining
func (b *Backoff) Register(call interface{}) {

	switch call.(type) {
	case logFunc:
		b.call = call.(logFunc)
	case streamFunc:
		b.call = call.(streamFunc)
	default:
	}
	return
}

// UnaryBackoffHandler method is the Unary gRPC Log Client's standard backoff flow, which
// is used when exchanging unary requests and responses with a gRPC server.
func (b *Backoff) UnaryBackoffHandler(err error, logger log.Logger) error {
	// retry with backoff
	logger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("retry").Metadata(event.Field{
		"error":      err,
		"iterations": b.Counter(),
		"maxWait":    b.Max(),
		"curWait":    b.Current(),
	}).Message("retrying connection").Build())

	// backoff locked -- skip retry until unlocked
	if b.IsLocked() {
		return ErrBackoffLocked
	} else {
		// backoff unlocked -- increment timer and wait
		// the Wait() method returns a registered func() to execute
		// and an error in case the backoff reaches its deadline
		call, err := b.Wait()

		// handle backoff deadline errors
		if err != nil && err != ErrBackoffLocked {
			return err
		}

		// execute registered call and propagate cancelation across
		// involved goroutines
		go call()
		return ErrFailedConn
	}
}

// StreamBackoffHandler method is the Stream gRPC Log Client's standard backoff flow, which
// is used when setting up a stream and when receiving an error from the gRPC Log Server
func (b *Backoff) StreamBackoffHandler(
	errCh chan error,
	cancel context.CancelFunc,
	logger log.Logger,
	done chan struct{},
) {
	// increment timer and wait
	// the Wait() method returns a registered func() to execute
	// and an error in case the backoff reaches its deadline
	call, err := b.Wait()

	// handle backoff deadline errors by closing the stream
	if err != nil {
		logger.Log(event.New().Level(event.Level_fatal).Prefix("gRPC").Sub("stream").Metadata(event.Field{
			"error":      err.Error(),
			"numRetries": b.Counter(),
		}).Message("closing stream after too many failed attempts to reconnect").Build())

		done <- struct{}{}
		errCh <- ErrFailedRetry
		cancel()
		return
	}

	// otherwise the stream will be recreated
	logger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("stream").Metadata(event.Field{
		"error":      err,
		"iterations": b.Counter(),
		"maxWait":    b.Max(),
		"curWait":    b.Current(),
	}).Message("retrying connection").Build())

	go call()
	return
}

// Time method will set the ExpBackoff's deadline, and returns a pointer to
// itself for chaining
func (b *Backoff) Time(t time.Duration) {
	b.max = t
	return
}

// AddMessage method will append a new message to the exponential backoff's
// message queue
func (b *Backoff) AddMessage(msg *event.Event) {
	b.msg = append(b.msg, msg)
	return
}

// Counter method will return the current amount of retries since the connection
// failed to be established
func (b *Backoff) Counter() uint {
	return b.counter
}

// Max method will return the ExpBackoff's deadline, in a string format
func (b *Backoff) Max() string {
	return b.max.String()
}

// Current method will return the current ExpBackoff's wait time, in a string format
func (b *Backoff) Current() string {
	return b.wait.String()
}

// Lock method will set the ExpBackoff's locked element to true, preventing future calls
// from proceeding.
func (b *Backoff) Lock() {
	if b.locked {
		return
	}

	b.mu.Lock()
	b.locked = true
}

// Unlock method will set the ExpBackoff's locked element to false, allowing future calls
// to proceed.
func (b *Backoff) Unlock() {
	if !b.locked {
		return
	}

	b.mu.Unlock()
	b.locked = false
}

// TryLock method is similar to the mutex.TryLock() one; to allow checking for a locked status
// as the method also tries to lock the mutex.
//
// It is used to get a status from the mutex as the call is made.
func (b *Backoff) TryLock() bool {
	if b.locked {
		return true
	}

	b.locked = b.mu.TryLock()
	return b.locked
}

// IsLocked method will return the ExpBackoff's locked status
func (b *Backoff) IsLocked() bool {
	return b.locked
}
