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

func NewBackoff() *ExpBackoff {
	b := &ExpBackoff{}
	b.Time(defaultRetryTime)
	return b
}

func (b *ExpBackoff) Counter() int {
	return int(b.counter)
}
func (b *ExpBackoff) Max() string {
	return b.max.String()
}
func (b *ExpBackoff) Current() string {
	return b.wait.String()
}

func (b *ExpBackoff) WithDone(done *chan struct{}) *ExpBackoff {
	b.exit = done
	return b
}

func (b *ExpBackoff) Time(t time.Duration) *ExpBackoff {
	b.max = t
	return b
}

func (b *ExpBackoff) RegisterStream(call streamFunc, errCh chan error) *ExpBackoff {
	b.call = call
	b.errCh = errCh
	return b
}

func (b *ExpBackoff) RegisterLog(call logFunc, errCh chan error) *ExpBackoff {
	b.call = call
	b.errCh = errCh
	return b
}

func (b *ExpBackoff) AddMessage(msg *log.LogMessage) *ExpBackoff {
	b.msg = append(b.msg, msg)
	return b
}

// func (b *ExpBackoff) flush() {
// 	b.msg = []*log.LogMessage{}
// }

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

func (b *ExpBackoff) Lock() {
	b.locked = true
}

func (b *ExpBackoff) Unlock() {
	b.locked = false
}

func (b *ExpBackoff) IsLocked() bool {
	return b.locked
}

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
