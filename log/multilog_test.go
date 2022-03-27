package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/grpc/address"
	"github.com/zalgonoise/zlog/store"
)

var mockMultiPrefixes = []string{
	"multilog-test-01",
	"multilog-test-02",
	"multilog-test-03",
	"multilog-test-04",
	"multilog-test-05",
	"multilog-test-06",
}

const msg string = "multilogger test message"

// add 6 mock buffers
var mockBufs = []*bytes.Buffer{
	{},
	{},
	{},
	{},
	{},
	{},
}

var mockLoggers = []Logger{
	New(WithPrefix(mockMultiPrefixes[0]), JSONFormat, WithOut(mockBufs[0])),
	New(WithPrefix(mockMultiPrefixes[1]), JSONFormat, WithOut(mockBufs[1])),
	New(WithPrefix(mockMultiPrefixes[2]), JSONFormat, WithOut(mockBufs[2])),
	New(WithPrefix(mockMultiPrefixes[3]), JSONFormat, WithOut(mockBufs[3])),
	New(WithPrefix(mockMultiPrefixes[4]), JSONFormat, WithOut(mockBufs[4])),
	New(WithPrefix(mockMultiPrefixes[5]), JSONFormat, WithOut(mockBufs[5])),
}

var mockMultiLogger = struct {
	log Logger
	buf []*bytes.Buffer
}{
	log: MultiLogger(mockLoggers...),
	buf: mockBufs,
}

func TestNewMultiLogger(t *testing.T) {
	type test struct {
		input  []Logger
		bufs   []*bytes.Buffer
		prefix []string
		msg    string
	}

	var tests []test

	for a := 0; a < len(mockLoggers); a++ {

		var test = test{}
		test.msg = msg
		for b := 0; b <= a; b++ {
			test.input = append(test.input, mockLoggers[b])
			test.bufs = append(test.bufs, mockBufs[b])
			test.prefix = append(test.prefix, mockMultiPrefixes[b])
		}
		tests = append(tests, test)

	}

	var verify = func(id int, test test) {
		for bufID, buf := range test.bufs {
			logEntry := &LogMessage{}

			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg,
					err,
				)
				return
			}

			if logEntry.Msg != test.msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg,
					test.msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Level != LLInfo.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg,
					LLInfo.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Prefix != test.prefix[bufID] {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- log prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg,
					test.prefix,
					logEntry.Prefix,
				)
				return
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Info(%s) -- %s",
				id,
				bufID,
				test.msg,
				test.msg,
			)
		}

	}

	for id, test := range tests {
		for _, buf := range test.bufs {
			buf.Reset()
		}

		ml := MultiLogger(test.input...)
		ml.Info(test.msg)

		verify(id, test)
	}
}

func TestMultiLoggerOutput(t *testing.T) {

	var testAllObjects []map[string]interface{}
	testAllObjects = append(testAllObjects, testObjects...)
	testAllObjects = append(testAllObjects, testEmptyObjects...)

	var testAllMessages []string
	testAllMessages = append(testAllMessages, mockMessages...)
	for _, fmtMsg := range mockFmtMessages {
		testAllMessages = append(testAllMessages, fmt.Sprintf(fmtMsg.format, fmtMsg.v...))
	}

	var tests []*LogMessage

	for a := 0; a < len(mockLogLevelsOK); a++ {

		for b := 0; b < len(mockPrefixes); b++ {

			for c := 0; c < len(testAllMessages); c++ {

				for d := 0; d < len(testAllObjects); d++ {
					msg := NewMessage().
						Level(mockLogLevelsOK[a]).
						Prefix(mockPrefixes[b]).
						Message(testAllMessages[c]).
						Metadata(testAllObjects[d]).
						Build()

					tests = append(tests, msg)
				}
			}
		}
	}

	var verify = func(id int, test *LogMessage) {
		for bufID, buf := range mockMultiLogger.buf {
			logEntry := &LogMessage{}

			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.Msg,
					err,
				)
				return
			}

			if logEntry.Msg != test.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.Msg,
					test.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Level != test.Level {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.Msg,
					test.Level,
					logEntry.Level,
				)
				return
			}

			if logEntry.Prefix != test.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- log prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.Msg,
					test.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if len(logEntry.Metadata) == 0 && len(test.Metadata) > 0 {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.Msg,
					test.Metadata,
					logEntry.Metadata,
				)
				return
			} else if len(logEntry.Metadata) > 0 && len(test.Metadata) == 0 {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.Msg,
					test.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if len(logEntry.Metadata) > 0 && len(test.Metadata) > 0 {
				for k, v := range logEntry.Metadata {
					if v != nil && test.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.Msg,
						len(logEntry.Metadata),
						len(test.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Output(%s) -- %s",
				id,
				bufID,
				test.Msg,
				buf.String(),
			)
		}

	}

	for id, msg := range tests {
		for _, buf := range mockMultiLogger.buf {
			buf.Reset()
		}

		mockMultiLogger.log.Output(msg)

		verify(id, msg)
	}

}

func TestMultiLoggerSetOuts(t *testing.T) {
	module := "MultiLogger"
	funcname := "SetOuts()"

	t1logger := New(
		WithPrefix("test-new-logger"),
		TextFormat,
		WithOut(mockBufs[5]),
	)
	t2logger := New(
		WithPrefix("test-new-logger-2"),
		TextFormat,
		WithOut(mockBufs[4]),
	)
	innerML := MultiLogger(t2logger)

	nilLogger := New(EmptyConfig)

	ml := MultiLogger(t1logger, innerML, nilLogger)

	type test struct {
		name  string
		input []io.Writer
		wants io.Writer
	}

	var tests = []test{
		{
			name:  "switching to buffer #0",
			input: []io.Writer{mockBufs[0]},
			wants: io.MultiWriter(mockBufs[0]),
		},
		{
			name:  "switching to multi-buffer #0",
			input: []io.Writer{mockBufs[0], mockBufs[1], mockBufs[3]},
			wants: io.MultiWriter(mockBufs[0], mockBufs[1], mockBufs[3]),
		},
		{
			name:  "ConnAddr flow test",
			input: []io.Writer{mockBufs[0], &address.ConnAddr{}},
			wants: io.MultiWriter(mockBufs[0]),
		},
		{
			name:  "switching to default writer with zero arguments",
			input: nil,
			wants: os.Stderr,
		},
		{
			name:  "switching to default writer with nil writers",
			input: []io.Writer{nil, nil, nil},
			wants: os.Stderr,
		},
		{
			name:  "ensure the empty writer works",
			input: []io.Writer{store.EmptyWriter},
			wants: io.MultiWriter(store.EmptyWriter),
		},
	}

	var verify = func(id int, logw, w io.Writer, action string) {
		if !reflect.DeepEqual(logw, w) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] writer mismatch: wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				w,
				logw,
				action,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [%s] [%s]",
			id,
			module,
			funcname,
		)
	}

	for id, test := range tests {
		if test.input != nil {
			ml.SetOuts(test.input...)
		} else {
			ml.SetOuts()
		}

		for _, l := range ml.(*multiLogger).loggers {
			if _, ok := l.(*logger); ok {
				logw := l.(*logger).out
				verify(id, logw, test.wants, test.name)
			}
		}

	}

}

func TestMultiLoggerAddOuts(t *testing.T) {

	module := "MultiLogger"
	funcname := "AddOuts()"

	t1logger := New(
		WithPrefix("test-new-logger"),
		TextFormat,
		WithOut(mockBufs[5]),
	)
	t2logger := New(
		WithPrefix("test-new-logger-2"),
		TextFormat,
		WithOut(mockBufs[4]),
	)
	innerML := MultiLogger(t2logger)

	nilLogger := New(EmptyConfig)

	ml := MultiLogger(t1logger, innerML, nilLogger)

	type test struct {
		name  string
		input []io.Writer
		wants io.Writer
	}

	var tests = []test{
		{
			name:  "adding buffer #0",
			input: []io.Writer{mockBufs[0]},
			wants: io.MultiWriter(mockBufs[0], mockBufs[5]),
		},
		{
			name:  "adding multi-buffer #0",
			input: []io.Writer{mockBufs[0], mockBufs[1], mockBufs[3]},
			wants: io.MultiWriter(mockBufs[0], mockBufs[1], mockBufs[3], mockBufs[5]),
		},
		{
			name:  "ConnAddr flow test",
			input: []io.Writer{mockBufs[0], &address.ConnAddr{}},
			wants: io.MultiWriter(mockBufs[0], mockBufs[5]),
		},
		{
			name:  "adding default writer with zero arguments",
			input: nil,
			wants: io.MultiWriter(mockBufs[5]),
		},
		{
			name:  "adding default writer with nil writers",
			input: []io.Writer{nil, nil, nil},
			wants: io.MultiWriter(mockBufs[5]),
		},
		{
			name:  "ensure the empty writer works",
			input: []io.Writer{store.EmptyWriter},
			wants: io.MultiWriter(store.EmptyWriter, mockBufs[5]),
		},
	}

	var verify = func(id int, logw, w io.Writer, action string) {
		if !reflect.DeepEqual(logw, w) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] writer mismatch: wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				w,
				logw,
				action,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [%s] [%s]",
			id,
			module,
			funcname,
		)
	}

	for id, test := range tests {
		if test.input != nil {
			ml.AddOuts(test.input...)
		} else {
			ml.AddOuts()
		}

		for _, l := range ml.(*multiLogger).loggers {
			if _, ok := l.(*logger); ok {
				logw := l.(*logger).out
				verify(id, logw, test.wants, test.name)
			}
		}
		// reset
		ml.SetOuts(mockBufs[5])

	}
}

func TestMultiLoggerSub(t *testing.T) {

	type ml struct {
		log Logger
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
		sub string
	}

	var newSubPrefixes = []string{
		"Prefix()",
		"new prefix",
		"awesome service",
		"alert!!",
		"@whatever",
		"01101001101",
	}

	var testAllObjects []map[string]interface{}
	testAllObjects = append(testAllObjects, testObjects...)
	testAllObjects = append(testAllObjects, testEmptyObjects...)

	var testAllMessages []string
	testAllMessages = append(testAllMessages, mockMessages...)
	for _, fmtMsg := range mockFmtMessages {
		testAllMessages = append(testAllMessages, fmt.Sprintf(fmtMsg.format, fmtMsg.v...))
	}

	var tests []test

	for a := 0; a < len(newSubPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []Logger
				for d := 0; d < len(newSubPrefixes); d++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix("log"), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					sub: newSubPrefixes[a],
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Message(testAllMessages[b]).
						Metadata(testAllObjects[c]).
						Build(),
				}

				tests = append(tests, obj)
			}

		}

	}

	var verify = func(id int, test test) {
		defer func() {
			for _, b := range test.ml.buf {
				b.Reset()
			}
		}()

		for bufID, buf := range test.ml.buf {
			logEntry := &LogMessage{}

			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.sub,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.sub,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Sub != test.sub {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- sub-prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.sub,
					test.sub,
					logEntry.Sub,
				)
				return
			}

			if logEntry.Level != LLInfo.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.sub,
					LLInfo.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.sub,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if len(logEntry.Metadata) == 0 && len(test.msg.Metadata) > 0 {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.sub,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if len(logEntry.Metadata) > 0 && len(test.msg.Metadata) == 0 {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.sub,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if len(logEntry.Metadata) > 0 && len(test.msg.Metadata) > 0 {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.sub,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.sub,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Sub(%s) -- %s",
				id,
				bufID,
				test.sub,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}
		test.ml.log.Sub(test.sub).Fields(test.msg.Metadata)
		test.ml.log.Info(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerFields(t *testing.T) {
	type ml struct {
		log Logger
		buf []*bytes.Buffer
	}

	type test struct {
		msg    *LogMessage
		ml     ml
		prefix string
	}

	var newPrefixes = []string{
		"Prefix()",
		"new prefix",
		"awesome service",
		"alert!!",
		"@whatever",
		"01101001101",
	}

	var testAllObjects []map[string]interface{}
	testAllObjects = append(testAllObjects, testObjects...)
	testAllObjects = append(testAllObjects, testEmptyObjects...)

	var testAllMessages []string
	testAllMessages = append(testAllMessages, mockMessages...)
	for _, fmtMsg := range mockFmtMessages {
		testAllMessages = append(testAllMessages, fmt.Sprintf(fmtMsg.format, fmtMsg.v...))
	}

	var tests []test

	for a := 0; a < len(newPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []Logger
				for d := 0; d < len(mockMultiPrefixes); d++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[d]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					prefix: newPrefixes[a],
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Message(testAllMessages[b]).
						Metadata(testAllObjects[c]).
						Build(),
				}

				tests = append(tests, obj)
			}

		}

	}

	var verify = func(id int, test test) {
		defer func() {
			for _, b := range test.ml.buf {
				b.Reset()
			}
		}()

		for bufID, buf := range test.ml.buf {
			logEntry := &LogMessage{}

			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- unmarshal error: %s",
					id,
					bufID,
					err,
				)
				return
			}

			if logEntry.Prefix != test.prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLInfo.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					LLInfo.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if len(logEntry.Metadata) == 0 && len(test.msg.Metadata) > 0 {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if len(logEntry.Metadata) > 0 && len(test.msg.Metadata) == 0 {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if len(logEntry.Metadata) > 0 && len(test.msg.Metadata) > 0 {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Fields(map[string]interface{}) -- %s",
				id,
				bufID,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}
		test.ml.log.Prefix(test.prefix).Fields(test.msg.Metadata)
		test.ml.log.Info(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerWrite(t *testing.T) {
	type test struct {
		msg  []byte
		want LogMessage
	}

	var tests = []test{
		{
			msg: NewMessage().Level(LLInfo).Prefix("test").Sub("tester").Message("write test").Build().Bytes(),
			want: LogMessage{
				Prefix: "test",
				Sub:    "tester",
				Level:  LLInfo.String(),
				Msg:    "write test",
			},
		},
		{
			msg: []byte("hello world"),
			want: LogMessage{
				Prefix: "log",
				Sub:    "",
				Level:  LLInfo.String(),
				Msg:    "hello world",
			},
		},
	}

	bufs := []*bytes.Buffer{{}, {}, {}}

	logger := MultiLogger(
		New(JSONFormat, WithOut(bufs[0])),
		New(JSONFormat, WithOut(bufs[1])),
		New(JSONFormat, WithOut(bufs[2])),
	)
	var verify = func(id int, test test) {

		for bid, buffer := range bufs {

			buf := buffer.Bytes()

			if len(buf) <= 0 {
				t.Errorf(
					"#%v [Logger] -- FAILED -- [MultiLogger] Write([]byte) [buffer #%v] -- empty buffer error: %v bytes written",
					id,
					bid,
					len(buf),
				)
				return
			}

			logEntry := &LogMessage{}

			err := json.Unmarshal(buf, logEntry)
			if err != nil {
				t.Errorf(
					"#%v [Logger] -- FAILED -- [MultiLogger] Write([]byte) [buffer #%v] -- JSON decoding error: %s ; buf: %s",
					id,
					bid,
					err,
					string(buf),
				)
				return
			}

			if logEntry.Prefix != test.want.Prefix {
				t.Errorf(
					"#%v [Logger] -- FAILED -- [MultiLogger] Write([]byte) [buffer #%v] -- prefix mismatch: wanted %s ; got %s",
					id,
					bid,
					logEntry.Prefix,
					test.want.Prefix,
				)
				return
			}

			if logEntry.Sub != test.want.Sub {
				t.Errorf(
					"#%v [Logger] -- FAILED -- [MultiLogger] Write([]byte) [buffer #%v] -- sub-prefix mismatch: wanted %s ; got %s",
					id,
					bid,
					logEntry.Sub,
					test.want.Sub,
				)
				return
			}

			if logEntry.Level != test.want.Level {
				t.Errorf(
					"#%v [Logger] -- FAILED -- [MultiLogger] Write([]byte) [buffer #%v] -- log level mismatch: wanted %s ; got %s",
					id,
					bid,
					logEntry.Level,
					test.want.Level,
				)
				return
			}

			if logEntry.Msg != test.want.Msg {
				t.Errorf(
					"#%v [Logger] -- FAILED -- [MultiLogger] Write([]byte) [buffer #%v] -- message mismatch: wanted %s ; got %s",
					id,
					bid,
					logEntry.Msg,
					test.want.Msg,
				)
				return
			}

			t.Logf(
				"#%v [Logger] -- PASSED -- [MultiLogger] Write([]byte) [buffer #%v]",
				id,
				bid,
			)
		}

	}

	for id, test := range tests {
		for _, b := range bufs {
			b.Reset()
		}
		n, err := logger.Write(test.msg)

		if err != nil {
			t.Errorf(
				"#%v [Logger] -- FAILED -- Write([]byte) -- write error: %s",
				id,
				err,
			)
		}

		if n <= 0 {
			t.Errorf(
				"#%v [Logger] -- FAILED -- Write([]byte) -- no bytes written: %v",
				id,
				n,
			)
		}

		verify(id, test)

		for _, b := range bufs {
			b.Reset()
		}
	}

	// failing tests:
	tmpf, err := os.Create(`tmp.log`)
	if err != nil {
		t.Errorf(
			"#%v [Logger] -- FAILED -- Write([]byte) -- failed to create temp file: %s",
			0,
			err,
		)
	}
	tmpf.Close()
	defer os.RemoveAll(`tmp.log`)

	closedBuf, err := os.OpenFile(`tmp.log`, os.O_RDONLY, 0o000)
	if err != nil {
		t.Errorf(
			"#%v [Logger] -- FAILED -- Write([]byte) -- failed to open temp file: %s",
			0,
			err,
		)
	}
	logger.SetOuts(closedBuf)

	for id, test := range tests {
		n, err := logger.Write(test.msg)
		if err == nil && n <= 0 {
			t.Errorf(
				"#%v [Logger] -- FAILED -- Write([]byte) -- write succeeded when it shouldn't",
				id,
			)
		}
		t.Logf(
			"#%v [Logger] -- PASSED -- Write([]byte) -- write failed as expected: error: %s ; bytes written: %v",
			id,
			err,
			n,
		)
	}

}
