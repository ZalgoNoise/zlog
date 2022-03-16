package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
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
	// setup a single logger in a multilogger for these tests
	//
	// since setting / adding outs to a multilogger affects
	// all configured loggers, for testing purposes this is done with
	// a multilogger containing only one logger.
	//
	// the multilogger should be only a wrapper for using
	// several loggers and configs, and these methods are only
	// present to satisfy the Logger
	var simpleMultiLogger = struct {
		log Logger
		buf *bytes.Buffer
	}{
		log: MultiLogger(mockLoggers[0]),
		buf: mockBufs[0],
	}

	type test struct {
		msg *LogMessage
		out []io.Writer
		buf []*bytes.Buffer
	}

	newBuffers := []*bytes.Buffer{
		{},
		{},
		{},
		{},
		{},
		{},
	}

	newWriters := []io.Writer{
		newBuffers[0],
		newBuffers[1],
		newBuffers[2],
		newBuffers[3],
		newBuffers[4],
		newBuffers[5],
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

	for a := 1; a <= len(newBuffers); a++ {

		for b := 0; b < len(mockLogLevelsOK); b++ {

			for c := 0; c < len(mockPrefixes); c++ {

				for d := 0; d < len(testAllMessages); d++ {

					for e := 0; e < len(testAllObjects); e++ {

						tests = append(tests, test{
							msg: NewMessage().
								Level(mockLogLevelsOK[b]).
								Prefix(mockPrefixes[c]).
								Message(testAllMessages[d]).
								Metadata(testAllObjects[e]).
								Build(),
							buf: newBuffers[:a],
							out: newWriters[:a],
						})
					}
				}
			}
		}
	}

	var verify = func(id int, test test) {
		defer func() {
			for _, b := range test.buf {
				b.Reset()
			}
			for _, b := range mockBufs {
				b.Reset()
			}
		}()

		pass := [2]bool{
			false,
			false,
		}

		for bufID, buf := range test.buf {

			logEntry := &LogMessage{}

			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).SetOuts(...io.Writer) -- unmarshal error: %s",
					id,
					bufID,
					err,
				)

				buf.Reset()
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).SetOuts(...io.Writer) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					logEntry.Msg,
				)
				buf.Reset()
				return
			}

			if logEntry.Level != test.msg.Level {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).SetOuts(...io.Writer) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Level,
					logEntry.Level,
				)
				buf.Reset()
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).SetOuts(...io.Writer) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				buf.Reset()
				return
			}

			if len(logEntry.Metadata) != len(test.msg.Metadata) {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).SetOuts(...io.Writer) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				buf.Reset()
				return
			}

			if len(logEntry.Metadata) > 0 && len(test.msg.Metadata) > 0 {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).SetOuts(...io.Writer) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							k,
							k,
						)
						buf.Reset()
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).SetOuts(...io.Writer) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					buf.Reset()
					return
				}
			}

			pass[0] = true

			t.Logf(
				"#%v -- PASSED TARGET TEST -- [MultiLogger] MultiLogger(...Logger[%v]).SetOuts(...io.Writer) -- %s",
				id,
				bufID,
				buf.String(),
			)
		}

		for bufID, buf := range mockBufs {
			// SetOuts() will override original writers,
			// these should be found empty

			if buf.Len() > 0 {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).SetOuts(...io.Writer) -- replaced writer was still written to -- expected empty, got %s",
					id,
					bufID,
					buf.String(),
				)
				buf.Reset()
				return
			}

			pass[1] = true

			t.Logf(
				"#%v -- PASSED SOURCE TEST -- [MultiLogger] MultiLogger(...Logger[%v]).SetOuts(...io.Writer) -- %s",
				id,
				bufID,
				buf.String(),
			)
		}

		if pass[0] && pass[1] {
			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger).SetOuts(...io.Writer)",
				id,
			)
			return
		} else if !pass[0] {
			t.Errorf(
				"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger).SetOuts(...io.Writer) -- failed target buffer tests",
				id,
			)
			return
		} else if !pass[1] {
			t.Errorf(
				"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger).SetOuts(...io.Writer) -- failed source buffer tests",
				id,
			)
			return
		}

	}

	for id, test := range tests {
		for _, b := range test.buf {
			b.Reset()
		}
		for _, b := range mockBufs {
			b.Reset()
		}

		simpleMultiLogger.log.SetOuts(test.out...)
		simpleMultiLogger.log.Output(test.msg)

		verify(id, test)
	}

}

func TestMultiLoggerAddOuts(t *testing.T) {
	type test struct {
		msg *LogMessage
		buf []*bytes.Buffer
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

	for a := 1; a <= 6; a++ {

		for b := 0; b < len(mockLogLevelsOK); b++ {

			for c := 0; c < len(mockPrefixes); c++ {

				for d := 0; d < len(testAllMessages); d++ {

					for e := 0; e < len(testAllObjects); e++ {
						var bufs []*bytes.Buffer
						for f := 0; f <= a; f++ {
							bufs = append(bufs, &bytes.Buffer{})
						}

						tests = append(tests, test{
							msg: NewMessage().
								Level(mockLogLevelsOK[b]).
								Prefix(mockPrefixes[c]).
								Message(testAllMessages[d]).
								Metadata(testAllObjects[e]).
								Build(),
							buf: bufs,
						})
					}
				}
			}
		}
	}

	var verify = func(id int, test test) {
		defer func() {
			for _, b := range test.buf {
				b.Reset()
			}
			for _, b := range mockBufs {
				b.Reset()
			}
		}()

		for bufID, buf := range test.buf {

			logEntry := &LogMessage{}

			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).AddOuts(...io.Writer) -- unmarshal error: %s",
					id,
					bufID,
					err,
				)

				buf.Reset()
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).AddOuts(...io.Writer) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					logEntry.Msg,
				)
				buf.Reset()
				return
			}

			if logEntry.Level != test.msg.Level {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).AddOuts(...io.Writer) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Level,
					logEntry.Level,
				)
				buf.Reset()
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).AddOuts(...io.Writer) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				buf.Reset()
				return
			}

			if len(logEntry.Metadata) == 0 && len(test.msg.Metadata) > 0 {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).AddOuts(...io.Writer) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				buf.Reset()
				return
			} else if len(logEntry.Metadata) > 0 && len(test.msg.Metadata) == 0 {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).AddOuts(...io.Writer) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				buf.Reset()
				return
			}

			if len(logEntry.Metadata) > 0 && len(test.msg.Metadata) > 0 {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).AddOuts(...io.Writer) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							k,
							k,
						)
						buf.Reset()
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).AddOuts(...io.Writer) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					buf.Reset()
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).AddOuts(...io.Writer) -- %s",
				id,
				bufID,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.buf {
			b.Reset()
		}
		for _, b := range mockBufs {
			b.Reset()
		}

		multi := MultiLogger(
			New(
				WithPrefix("test-logger"),
				JSONFormat,
				WithOut(test.buf[0]),
			),
		)

		if len(test.buf) > 1 {
			for i := 1; i < len(test.buf); i++ {
				multi.AddOuts(test.buf[i])
			}
		}

		multi.Output(test.msg)

		verify(id, test)
	}
}

func TestMultiLoggerPrefix(t *testing.T) {

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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Prefix(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.prefix,
					err,
				)
				return
			}

			if logEntry.Prefix != test.prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Prefix(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.prefix,
					test.prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLInfo.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Prefix(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.prefix,
					LLInfo.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Prefix(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.prefix,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if len(logEntry.Metadata) != len(test.msg.Metadata) {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Prefix(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.prefix,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if len(logEntry.Metadata) > 0 && len(test.msg.Metadata) > 0 {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Prefix(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.prefix,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...Logger[%v]).Prefix(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.prefix,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...Logger[%v]).Prefix(%s) -- %s",
				id,
				bufID,
				test.prefix,
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
