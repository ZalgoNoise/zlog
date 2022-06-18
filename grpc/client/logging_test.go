package client

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/zalgonoise/zlog/grpc/server"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	jsonpb "github.com/zalgonoise/zlog/log/format/json"
)

func TestUnaryClientLogging(t *testing.T) {
	module := "LogClient Interceptors"
	funcname := "UnaryClientLogging()"

	_ = module
	_ = funcname

	type testGRPCLogger struct {
		l GRPCLogger
		e chan error
	}

	type test struct {
		name  string
		cfg   []LogClientConfig
		wants []string
	}

	var buf = []*bytes.Buffer{{}, {}}

	var writers = []log.Logger{
		log.New(log.WithOut(buf[0]), log.SkipExit, log.CfgFormatJSON),
		log.New(log.WithOut(buf[1]), log.SkipExit, log.CfgFormatJSONSkipNewline),
	}

	var mockServer = server.New(
		server.WithAddr("127.0.0.1:9099"),
		server.WithLogger(writers[1]),
	)

	go mockServer.Serve()
	defer mockServer.Stop()

	var tests = []test{
		{
			name: "unary client logging test",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
				WithLoggerV(writers[0]),
				UnaryRPC(),
			},
			wants: []string{
				`setting up Unary gRPC client`,
				`connecting to remote`,
				`dialed the address successfully`,
				`setting up log service with connection`,
				`received a new log message to register`,
				`[send] unary RPC logger -- /logservice.LogService/Log`,
				`[recv] unary RPC logger`,
			},
		},
		{
			name: "unary client logging test w/ timer",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
				WithLoggerV(writers[0]),
				UnaryRPC(),
				WithTiming(),
			},
			wants: []string{
				`setting up Unary gRPC client`,
				`connecting to remote`,
				`dialed the address successfully`,
				`setting up log service with connection`,
				`received a new log message to register`,
				`[send] unary RPC logger -- /logservice.LogService/Log`,
				`[recv] unary RPC logger`,
			},
		},
		{
			name: "stream client logging test",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
				WithLogger(writers[0]),
				StreamRPC(),
			},
			wants: []string{
				`setting up Stream gRPC client`,
				`connecting to remote`,
				`dialed the address successfully`,
				`setting up log service with connection`,
				`setting request ID for long-lived connection`,
			},
		},
		{
			name: "stream client logging test w/ timer",
			cfg: []LogClientConfig{
				WithAddr("127.0.0.1:9099"),
				WithLogger(writers[0]),
				StreamRPC(),
				WithTiming(),
			},
			wants: []string{
				`setting up Stream gRPC client`,
				`connecting to remote`,
				`dialed the address successfully`,
				`setting up log service with connection`,
				`setting request ID for long-lived connection`,
				`[conn] stream RPC -- connection was established`,
				`[send] stream RPC`,
				`[recv] stream RPC`,
			},
		},
	}

	var validateLogs = func(input []string, test test) bool {
		results := make([]bool, 0, len(test.wants))

		for _, line := range input {
			for _, w := range test.wants {
				if line == w {
					results = append(results, true)
					break
				}
			}
		}

		var count int = 0

		for _, r := range results {
			if r {
				count++
			}
		}

		if count+1 >= len(test.wants) {
			return true
		}
		return false

	}

	var bufferFilter = func(in []byte) []string {
		// split lines
		var out []string
		var buf []byte

		for _, b := range in {
			if b == 10 {
				if len(buf) > 0 {
					e, _ := jsonpb.Decode(buf)
					out = append(out, e.GetMsg())
					buf = []byte{}
				}
				continue
			}
			buf = append(buf, b)
		}

		if len(buf) > 0 {
			e, _ := jsonpb.Decode(buf)
			out = append(out, e.GetMsg())
			buf = []byte{}
		}

		return out
	}

	var verifyLoggers = func(idx int, test test, client GRPCLogger, errCh chan error, done chan struct{}) {

		// f := jsonpb.FmtJSON{SkipNewline: true}
		msg := event.New().Message("null").Build()
		// out, err := f.Format(msg)

		n, err := client.Output(msg)
		time.Sleep(time.Millisecond * 50)

		if err != nil {
			errCh <- fmt.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		if n == 0 {
			errCh <- fmt.Errorf(
				"#%v -- FAILED -- [%s] [%s] zero bytes written error -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		b := buf[0].Bytes()
		msgs := bufferFilter(b)

		if !validateLogs(msgs, test) {
			errCh <- fmt.Errorf(
				"#%v -- FAILED -- [%s] [%s] retrieved content does not match expected: wanted messages: %v ; got %v -- action: %s",
				idx,
				module,
				funcname,
				test.wants,
				msgs,
				test.name,
			)
			return
		}

		done <- struct{}{}

	}

	var verify = func(idx int, test test) {
		buf[0].Reset()
		defer buf[0].Reset()

		var done = make(chan struct{})

		client, errCh := New(test.cfg...)
		defer client.Close()

		if client == nil || errCh == nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] client or error channel are unexpectedly nil values -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		go verifyLoggers(idx, test, client, errCh, done)

		for {
			select {
			case err := <-errCh:
				t.Error(err.Error())
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
