package channel

import (
	"testing"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

func TestNew(t *testing.T) {
	module := "LogCh"
	funcname := "New()"

	_ = module
	_ = funcname

	type test struct {
		name string
		l    log.Logger
	}

	var tests = []test{
		{
			name: "simple channeled logger",
			l:    log.New(log.NilConfig),
		},
	}

	var init = func(test test) ChanneledLogger {
		return New(test.l)
	}
	var verify = func(idx int, test test) {
		cl := init(test)
		cl.Log(event.New().Message("null").Build())
		cl.Log()
		cl.Close()

		cl = init(test)
		ch, done := cl.Channels()
		ch <- event.New().Message("null").Build()
		done <- struct{}{}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
