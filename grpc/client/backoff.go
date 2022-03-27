package client

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/zalgonoise/zlog/log"
)

var (
	ErrFailedRetry   error = errors.New("failed to connect to server after numerous retries")
	ErrFailedConn    error = errors.New("failed to connect to server")
	ErrInvalidType   error = errors.New("unsupported exponential backoff function type")
	ErrBackoffLocked error = errors.New("operations locked during exponential backoff")
)

const defaultRetryTime time.Duration = time.Second * 300

type streamFunc func(chan error)
type logFunc func(*log.LogMessage, chan error)

// ExpBackoff struct defines the elements of an Exponential Backoff module,
// which is configured by setting a time.Duration deadline and by registering
// a (concurrent) function, named call.
//
// ExpBackoff will also try to act as a message buffer in case the server connection
// cannot be established -- as it will attempt to flush these records to the server
// as soon as connected.
//
// Also it has a simple lock / unlock switch for concurrent calls to be able to
// verify its state and halt by themselves
//
// The ExpBackoff object is initialized with a package-scope so it can be
// referenced by any function
//
// Notes on exponential backoff: https://en.wikipedia.org/wiki/Exponential_backoff
//
//
type ExpBackoff struct {
	counter float64
	max     time.Duration
	wait    time.Duration
	call    interface{}
	errCh   chan error
	exit    *chan struct{}
	msg     []*log.LogMessage
	locked  bool
}

// NewBackoff function initializes a simple exponential backoff module with
// a set default retry time of 300 seconds
func NewBackoff() *ExpBackoff {
	b := &ExpBackoff{}
	b.Time(defaultRetryTime)
	return b
}

// Increment method will increase the wait time exponentially, on each iteration.
//
// It's chained with a Wait() call right after.
func (b *ExpBackoff) Increment() *ExpBackoff {
	if b.locked {
		return b
	}
	b.counter = b.counter + 1
	b.wait = time.Millisecond * time.Duration(
		int64(math.Pow(2, b.counter))+rand.New(
			rand.NewSource(time.Now().UnixNano())).Int63n(1000),
	)
	return b
}

// Wait method will wait for the currently set wait time, if the module is unlocked.
//
// After waiting, it returns a func() to call (depending on what it is handling),
// and and an error.
//
// If the waiting time is grater than the deadline set, it will return with an
// ErrFailedRetry
func (b *ExpBackoff) Wait() (func(), error) {
	if b.locked {
		return nil, ErrBackoffLocked
	}
	if b.wait <= b.max {
		b.Lock()
		defer b.Unlock()
		time.Sleep(b.wait)
		switch v := b.call.(type) {
		case streamFunc:
			return func() {
				v(b.errCh)
			}, nil
		case logFunc:
			list := b.msg
			f := func() {
				for _, msg := range list {
					v(msg, b.errCh)
				}
			}
			return f, nil
		default:
			return nil, ErrInvalidType
		}
	}

	return nil, ErrFailedRetry

}

// RegisterStream method will take in a function with the same signature as a stream() function
// and the error channel of the gRPC Log Client; and returns a pointer to itself for method chaining
func (b *ExpBackoff) RegisterStream(call streamFunc, errCh chan error) *ExpBackoff {
	b.call = call
	b.errCh = errCh
	return b
}

// RegisterLog method will take in a function with the same signature as a log() function
// and the error channel of the gRPC Log Client; and returns a pointer to itself for method chaining
func (b *ExpBackoff) RegisterLog(call logFunc, errCh chan error) *ExpBackoff {
	b.call = call
	b.errCh = errCh
	return b
}

// WithDone method will register a gRPC Log Client's done channel, and returns a pointer to
// itself for chaining
func (b *ExpBackoff) WithDone(done *chan struct{}) *ExpBackoff {
	b.exit = done
	return b
}

// Time method will set the ExpBackoff's deadline, and returns a pointer to
// itself for chaining
func (b *ExpBackoff) Time(t time.Duration) *ExpBackoff {
	b.max = t
	return b
}

// AddMessage method will append a new message to the exponential backoff's
// message queue
func (b *ExpBackoff) AddMessage(msg *log.LogMessage) *ExpBackoff {
	b.msg = append(b.msg, msg)
	return b
}

// Counter method will return the current amount of retries since the connection
// failed to be established
func (b *ExpBackoff) Counter() int {
	return int(b.counter)
}

// Max method will return the ExpBackoff's deadline, in a string format
func (b *ExpBackoff) Max() string {
	return b.max.String()
}

// Current method will return the current ExpBackoff's wait time, in a string format
func (b *ExpBackoff) Current() string {
	return b.wait.String()
}

// Lock method will set the ExpBackoff's locked element to true, preventing future calls
// from proceeding.
func (b *ExpBackoff) Lock() {
	b.locked = true
}

// Unlock method will set the ExpBackoff's locked element to false, allowing future calls
// to proceed.
func (b *ExpBackoff) Unlock() {
	b.locked = false
}

// IsLocked method will return the ExpBackoff's locked status
func (b *ExpBackoff) IsLocked() bool {
	return b.locked
}
