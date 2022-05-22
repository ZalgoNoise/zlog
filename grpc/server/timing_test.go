package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/zalgonoise/zlog/grpc/client"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	jsonfmt "github.com/zalgonoise/zlog/log/format/json"
	pb "github.com/zalgonoise/zlog/proto/service"
	"google.golang.org/grpc/metadata"
)

var testErrUnexpected = errors.New("unexpected error")

type testServerStream struct{}

func (testServerStream) SetHeader(metadata.MD) error {
	return nil
}

func (testServerStream) SendHeader(metadata.MD) error {
	return nil
}

func (testServerStream) SetTrailer(metadata.MD) {}

func (testServerStream) Context() context.Context {
	return context.Background()
}

func (testServerStream) SendMsg(m interface{}) error {
	msg, ok := m.(*pb.LogResponse)

	if !ok {
		return testErrUnexpected
	}

	if !msg.GetOk() {
		return testErrUnexpected
	}
	return nil
}

func (testServerStream) RecvMsg(m interface{}) error {
	msg, ok := m.(*event.Event)

	if !ok {
		return testErrUnexpected
	}

	if msg.GetLevel().String() == "error" {
		return testErrUnexpected
	}
	return nil
}

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

func TestSendMsg(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "timingStream.SendMsg()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *timingStream
		m    *pb.LogResponse
		ok   bool
	}

	var buf = new(bytes.Buffer)
	var bytesResponse = []int32{
		203,
		1008,
	}
	var errResponse = []string{
		"",
		testErrUnexpected.Error(),
	}

	var tests = []test{
		{
			name: "working test",
			t: &timingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
				method: "testLog",
			},
			m: &pb.LogResponse{
				Ok:    true,
				ReqID: "123",
				Bytes: &bytesResponse[0],
				Err:   &errResponse[0],
			},
			ok: true,
		},
		{
			name: "errored test",
			t: &timingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
				method: "testLog",
			},
			m: &pb.LogResponse{
				Ok:    false,
				ReqID: "123",
				Bytes: &bytesResponse[1],
				Err:   &errResponse[1],
			},
		},
	}

	var verify = func(idx int, test test) {
		err := test.t.SendMsg(test.m)

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
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestRecvMsg(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "timingStream.RecvMsg()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *timingStream
		m    *event.Event
		ok   bool
	}

	var buf = new(bytes.Buffer)

	var tests = []test{
		{
			name: "working test",
			t: &timingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
				method: "testLog",
			},
			m:  event.New().Message("null").Build(),
			ok: true,
		},
		{
			name: "errored test",
			t: &timingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
				method: "testLog",
			},
			m: event.New().Level(event.Level_error).Message("null").Build(),
		},
	}

	var verify = func(idx int, test test) {
		err := test.t.RecvMsg(test.m)

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
		}
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestSetHeader(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "timingStream.SetHeader()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *timingStream
		m    metadata.MD
	}

	var buf = new(bytes.Buffer)

	var tests = []test{
		{
			name: "working test",
			t: &timingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
				method: "testLog",
			},
			m: metadata.New(map[string]string{
				"test": "header",
			}),
		},
	}

	var verify = func(idx int, test test) {
		err := test.t.SetHeader(test.m)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
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

func TestSendHeader(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "timingStream.SendHeader()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *timingStream
		m    metadata.MD
	}

	var buf = new(bytes.Buffer)

	var tests = []test{
		{
			name: "working test",
			t: &timingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
				method: "testLog",
			},
			m: metadata.New(map[string]string{
				"test": "header",
			}),
		},
	}

	var verify = func(idx int, test test) {
		err := test.t.SendHeader(test.m)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
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

func TestSetTrailer(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "timingStream.SendHeader()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *timingStream
		m    metadata.MD
	}

	var buf = new(bytes.Buffer)

	var tests = []test{
		{
			name: "working test",
			t: &timingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
				method: "testLog",
			},
			m: metadata.New(map[string]string{
				"test": "trailer",
			}),
		},
	}

	var verify = func(idx int, test test) {
		test.t.SetTrailer(test.m)
	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestContext(t *testing.T) {
	module := "LogServer Interceptors"
	funcname := "timingStream.Context()"

	_ = module
	_ = funcname

	type test struct {
		name string
		t    *timingStream
	}

	var buf = new(bytes.Buffer)

	var tests = []test{
		{
			name: "working test",
			t: &timingStream{
				stream: testServerStream{},
				logger: log.New(log.WithOut(buf), log.SkipExit),
				method: "testLog",
			},
		},
	}

	var verify = func(idx int, test test) {
		ctx := test.t.Context()

		if ctx == nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected nil context -- action: %s",
				idx,
				module,
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
