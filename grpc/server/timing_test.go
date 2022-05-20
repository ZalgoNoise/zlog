package server

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/zalgonoise/zlog/grpc/client"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	jsonfmt "github.com/zalgonoise/zlog/log/format/json"
)

func TestUnaryServerTiming(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "UnaryServerTiming()"

	_ = module
	_ = funcname

	type test struct {
		name string
		s    *GRPCLogServer
	}

	var buf = []*bytes.Buffer{{}, {}, {}}

	var tests = []test{
		{
			name: "unary server timing test",
			s: New(
				WithLogger(log.New(log.WithOut(buf[0]), log.SkipExit)),
				WithServiceLogger(log.New(log.WithOut(buf[1]), log.CfgTextLevelFirst, log.SkipExit, log.CfgFormatJSON)),
				WithAddr("127.0.0.1:9099"),
				WithGRPCOpts(),
				WithTiming(),
			),
		},
	}

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	var stop = func(test test) {
		test.s.Stop()
		return
	}

	var initClient = func() (client.GRPCLogger, chan error) {
		return client.New(
			client.WithLogger(log.New(log.WithOut(buf[2]), log.SkipExit)),
			client.WithAddr("127.0.0.1:9099"),
			client.UnaryRPC(),
			client.WithGRPCOpts(),
		)
	}

	var bufferFilter = func(in []byte) (events []*event.Event, err error) {
		// split lines
		var line [][]byte
		var buf []byte

		for _, b := range in {
			if b == 10 {
				if len(buf) > 0 {
					copy := buf
					line = append(line, copy)
					buf = []byte{}
				}
				continue
			}
			buf = append(buf, b)
		}

		if len(buf) > 0 {
			copy := buf
			line = append(line, copy)
			buf = []byte{}
		}

		for _, l := range line {
			e, err := jsonfmt.Decode(l)
			if err != nil {
				return nil, err
			}
			events = append(events, e)
		}

		return
	}

	var verifyServiceLogger = func(
		idx int,
		test test,
		c client.GRPCLogger,
		done chan struct{},
	) {
		c.Info("null")
		time.Sleep(time.Millisecond * 1000)

		var err error

		events, err := bufferFilter(buf[1].Bytes())

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] error decoding events: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		for _, e := range events {

			if e.GetSub() != "timer" {
				continue
			}

			meta := e.GetMeta().AsMap()

			mTime, ok := meta["time"]

			if !ok {
				err = fmt.Errorf(
					"time metadata isn't available: %v",
					meta,
				)
				test.s.ErrCh <- err
				return
			}

			if mTime == "" {
				err = fmt.Errorf(
					"time value isn't set: %v",
					mTime,
				)
				test.s.ErrCh <- err
				return
			}
		}

		done <- struct{}{}
	}

	var verify = func(idx int, test test) {
		defer reset()
		defer stop(test)

		var done = make(chan struct{})

		go test.s.Serve()
		time.Sleep(time.Second)

		c, clientErr := initClient()

		go verifyServiceLogger(idx, test, c, done)

		for {
			select {
			case err := <-clientErr:
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unexpected client error: %v -- action: %s",
					idx,
					module,
					funcname,
					err,
					test.name,
				)
				return

			case err := <-test.s.ErrCh:
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unexpected server error: %v -- action: %s",
					idx,
					module,
					funcname,
					err,
					test.name,
				)
				return
			case <-done:
				return
			}
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestStreamServerTiming(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "StreamServerTiming()"

	_ = module
	_ = funcname

	type test struct {
		name string
		s    *GRPCLogServer
	}

	var buf = []*bytes.Buffer{{}, {}, {}}

	var tests = []test{
		{
			name: "unary server timing test",
			s: New(
				WithLogger(log.New(log.WithOut(buf[0]), log.SkipExit)),
				WithServiceLogger(log.New(log.WithOut(buf[1]), log.CfgTextLevelFirst, log.SkipExit, log.CfgFormatJSON)),
				WithAddr("127.0.0.1:9099"),
				WithGRPCOpts(),
				WithTiming(),
			),
		},
	}

	var reset = func() {
		for _, b := range buf {
			b.Reset()
		}
	}

	var stop = func(test test) {
		test.s.Stop()
		return
	}

	var initClient = func() (client.GRPCLogger, chan error) {
		return client.New(
			client.WithLogger(log.New(log.WithOut(buf[2]), log.SkipExit)),
			client.WithAddr("127.0.0.1:9099"),
			client.StreamRPC(),
			client.WithGRPCOpts(),
		)
	}

	var bufferFilter = func(in []byte) (events []*event.Event, err error) {
		// split lines
		var line [][]byte
		var buf []byte

		for _, b := range in {
			if b == 10 {
				if len(buf) > 0 {
					copy := buf
					line = append(line, copy)
					buf = []byte{}
				}
				continue
			}
			buf = append(buf, b)
		}

		if len(buf) > 0 {
			copy := buf
			line = append(line, copy)
			buf = []byte{}
		}

		for _, l := range line {
			e, err := jsonfmt.Decode(l)
			if err != nil {
				return nil, err
			}
			events = append(events, e)
		}

		return
	}

	var verifyServiceLogger = func(
		idx int,
		test test,
		c client.GRPCLogger,
		done chan struct{},
	) {
		c.Info("null")
		time.Sleep(time.Millisecond * 1000)

		var err error

		events, err := bufferFilter(buf[1].Bytes())

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] error decoding events: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		for _, e := range events {

			if e.GetSub() != "timer" {
				continue
			}

			meta := e.GetMeta().AsMap()

			mTime, ok := meta["time"]

			if !ok {
				err = fmt.Errorf(
					"time metadata isn't available: %v",
					meta,
				)
				test.s.ErrCh <- err
				return
			}

			if mTime == "" {
				err = fmt.Errorf(
					"time value isn't set: %v",
					mTime,
				)
				test.s.ErrCh <- err
				return
			}
		}

		done <- struct{}{}
	}

	var verify = func(idx int, test test) {
		defer reset()
		defer stop(test)

		var done = make(chan struct{})

		go test.s.Serve()
		time.Sleep(time.Second)

		c, clientErr := initClient()

		go verifyServiceLogger(idx, test, c, done)

		for {
			select {
			case err := <-clientErr:
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unexpected client error: %v -- action: %s",
					idx,
					module,
					funcname,
					err,
					test.name,
				)
				return

			case err := <-test.s.ErrCh:
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] unexpected server error: %v -- action: %s",
					idx,
					module,
					funcname,
					err,
					test.name,
				)
				return
			case <-done:
				return
			}
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}
