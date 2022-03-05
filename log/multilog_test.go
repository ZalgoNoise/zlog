package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

var mockLoggers = []LoggerI{
	New(WithPrefix(mockMultiPrefixes[0]), JSONFormat, WithOut(mockBufs[0])),
	New(WithPrefix(mockMultiPrefixes[1]), JSONFormat, WithOut(mockBufs[1])),
	New(WithPrefix(mockMultiPrefixes[2]), JSONFormat, WithOut(mockBufs[2])),
	New(WithPrefix(mockMultiPrefixes[3]), JSONFormat, WithOut(mockBufs[3])),
	New(WithPrefix(mockMultiPrefixes[4]), JSONFormat, WithOut(mockBufs[4])),
	New(WithPrefix(mockMultiPrefixes[5]), JSONFormat, WithOut(mockBufs[5])),
}

var mockMultiLogger = struct {
	log LoggerI
	buf []*bytes.Buffer
}{
	log: MultiLogger(mockLoggers...),
	buf: mockBufs,
}

func TestNewMultiLogger(t *testing.T) {
	type test struct {
		input  []LoggerI
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Info(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg,
					err,
				)
				return
			}

			if logEntry.Msg != test.msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Info(%s) -- message mismatch: wanted %s ; got %s",
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Info(%s) -- log level mismatch: wanted %s ; got %s",
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Info(%s) -- log prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg,
					test.prefix,
					logEntry.Prefix,
				)
				return
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Info(%s) -- %s",
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Output(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.Msg,
					err,
				)
				return
			}

			if logEntry.Msg != test.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Output(%s) -- message mismatch: wanted %s ; got %s",
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Output(%s) -- log level mismatch: wanted %s ; got %s",
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Output(%s) -- log prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.Msg,
					test.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Metadata == nil && test.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Output(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.Msg,
					test.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Output(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.Msg,
					test.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Output(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
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
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Output(%s) -- metadata length mismatch -- wanted %v, got %v",
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
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Output(%s) -- %s",
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
	// present to satisfy the LoggerI
	var simpleMultiLogger = struct {
		log LoggerI
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).SetOuts(...io.Writer) -- unmarshal error: %s",
					id,
					bufID,
					err,
				)

				buf.Reset()
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).SetOuts(...io.Writer) -- message mismatch: wanted %s ; got %s",
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).SetOuts(...io.Writer) -- log level mismatch: wanted %s ; got %s",
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).SetOuts(...io.Writer) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				buf.Reset()
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).SetOuts(...io.Writer) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				buf.Reset()
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).SetOuts(...io.Writer) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				buf.Reset()
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).SetOuts(...io.Writer) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
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
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).SetOuts(...io.Writer) -- metadata length mismatch -- wanted %v, got %v",
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
				"#%v -- PASSED TARGET TEST -- [MultiLogger] MultiLogger(...LoggerI[%v]).SetOuts(...io.Writer) -- %s",
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).SetOuts(...io.Writer) -- replaced writer was still written to -- expected empty, got %s",
					id,
					bufID,
					buf.String(),
				)
				buf.Reset()
				return
			}

			pass[1] = true

			t.Logf(
				"#%v -- PASSED SOURCE TEST -- [MultiLogger] MultiLogger(...LoggerI[%v]).SetOuts(...io.Writer) -- %s",
				id,
				bufID,
				buf.String(),
			)
		}

		if pass[0] && pass[1] {
			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI).SetOuts(...io.Writer)",
				id,
			)
			return
		} else if !pass[0] {
			t.Errorf(
				"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI).SetOuts(...io.Writer) -- failed target buffer tests",
				id,
			)
			return
		} else if !pass[1] {
			t.Errorf(
				"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI).SetOuts(...io.Writer) -- failed source buffer tests",
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).AddOuts(...io.Writer) -- unmarshal error: %s",
					id,
					bufID,
					err,
				)

				buf.Reset()
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).AddOuts(...io.Writer) -- message mismatch: wanted %s ; got %s",
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).AddOuts(...io.Writer) -- log level mismatch: wanted %s ; got %s",
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).AddOuts(...io.Writer) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				buf.Reset()
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).AddOuts(...io.Writer) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				buf.Reset()
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).AddOuts(...io.Writer) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				buf.Reset()
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).AddOuts(...io.Writer) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
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
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).AddOuts(...io.Writer) -- metadata length mismatch -- wanted %v, got %v",
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
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).AddOuts(...io.Writer) -- %s",
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
		log LoggerI
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
				var logs []LoggerI
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Prefix(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.prefix,
					err,
				)
				return
			}

			if logEntry.Prefix != test.prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Prefix(%s) -- prefix mismatch: wanted %s ; got %s",
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Prefix(%s) -- log level mismatch: wanted %s ; got %s",
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Prefix(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.prefix,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Prefix(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.prefix,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Prefix(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.prefix,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Prefix(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
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
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Prefix(%s) -- metadata length mismatch -- wanted %v, got %v",
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
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Prefix(%s) -- %s",
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

func TestMultiLoggerFields(t *testing.T) {
	type ml struct {
		log LoggerI
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
				var logs []LoggerI
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fields(map[string]interface{}) -- unmarshal error: %s",
					id,
					bufID,
					err,
				)
				return
			}

			if logEntry.Prefix != test.prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fields(map[string]interface{}) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLInfo.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fields(map[string]interface{}) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					LLInfo.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fields(map[string]interface{}) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fields(map[string]interface{}) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fields(map[string]interface{}) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fields(map[string]interface{}) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
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
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fields(map[string]interface{}) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fields(map[string]interface{}) -- %s",
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

func TestMultiLoggerPrint(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Print(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Print(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLInfo.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Print(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLInfo.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Print(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Print(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Print(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Print(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Print(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Print(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Print(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerPrintln(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Println(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Println(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLInfo.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Println(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLInfo.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Println(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Println(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Println(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Println(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Println(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Println(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Println(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerPrintf(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		format string
		v      []interface{}
		msg    *LogMessage
		ml     ml
	}

	var testAllObjects []map[string]interface{}
	testAllObjects = append(testAllObjects, testObjects...)
	testAllObjects = append(testAllObjects, testEmptyObjects...)

	var tests []test

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(mockMessages[b]).
						Metadata(testAllObjects[c]).
						Build(),
					format: "%s",
					v:      []interface{}{mockMessages[b]},
				}

				tests = append(tests, obj)
			}
		}
	}

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockFmtMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
						Metadata(testAllObjects[c]).
						Build(),
					format: mockFmtMessages[b].format,
					v:      mockFmtMessages[b].v,
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Printf(%s, %s) -- unmarshal error: %s",
					id,
					bufID,
					test.format,
					test.v,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Printf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLInfo.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Printf(%s, %s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					LLInfo.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Printf(%s, %s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Printf(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Printf(%s, %s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Printf(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.format,
							test.v,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Printf(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.format,
						test.v,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Printf(%s, %s) -- %s",
				id,
				bufID,
				test.format,
				test.v,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Printf(test.format, test.v...)

		verify(id, test)

	}
}

func TestMultiLoggerLog(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
	}

	var testAllMessages []string
	testAllMessages = append(testAllMessages, mockMessages...)
	for _, fmtMsg := range mockFmtMessages {
		testAllMessages = append(testAllMessages, fmt.Sprintf(fmtMsg.format, fmtMsg.v...))
	}

	var tests []test

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testObjects); c++ {
				for d := 0; d < len(mockLogLevelsOK); d++ {

					// skip LLFatal -- os.Exit(1)
					if mockLogLevelsOK[d] == LLFatal {
						continue
					}

					var bufs []*bytes.Buffer
					var logs []LoggerI
					for e := 0; e < len(mockMultiPrefixes); e++ {
						buf := &bytes.Buffer{}
						bufs = append(bufs, buf)
						logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
					}
					mlogger := MultiLogger(logs...)

					obj := test{
						ml: ml{
							log: mlogger,
							buf: bufs,
						},
						msg: NewMessage().
							Level(mockLogLevelsOK[d]).
							Prefix(mockPrefixes[a]).
							Message(testAllMessages[b]).
							Metadata(testObjects[c]).
							Build(),
					}

					tests = append(tests, obj)
				}
			}
		}
	}

	for a := 0; a < len(mockEmptyPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testEmptyObjects); c++ {
				for d := 0; d < len(mockLogLevelsNOK); d++ {

					var bufs []*bytes.Buffer
					var logs []LoggerI
					for e := 0; e < len(mockMultiPrefixes); e++ {
						buf := &bytes.Buffer{}
						bufs = append(bufs, buf)
						logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
					}
					mlogger := MultiLogger(logs...)

					obj := test{
						ml: ml{
							log: mlogger,
							buf: bufs,
						},
						msg: NewMessage().
							Level(mockLogLevelsNOK[d]).
							Prefix(mockEmptyPrefixes[a]).
							Message(testAllMessages[b]).
							Metadata(testEmptyObjects[c]).
							Build(),
					}

					tests = append(tests, obj)
				}
			}
		}
	}

	var verify = func(id int, test test) {
		defer func() {
			for _, b := range test.ml.buf {
				b.Reset()
			}
		}()

		r := recover()
		if r != nil {
			if test.msg.Level != LLPanic.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI).Log(%s) -- unexpected panic: %s",
					id,
					test.msg.Msg,
					r,
				)
				return
			}

			if r != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI).Log(%s) -- invalid panic message: wanted %s ; got %s",
					id,
					test.msg.Msg,
					test.msg.Msg,
					r,
				)
				return
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI).Log(%s) [panic] -- %s",
				id,
				test.msg.Msg,
				r,
			)
			return
		}

		for bufID, buf := range test.ml.buf {
			logEntry := &LogMessage{}

			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Log(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Log(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != test.msg.Level {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Log(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Level,
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Log(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Log(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Log(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Log(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Log(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Log(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		if test.msg.Level == LLPanic.String() {
			defer verify(id, test)
		}

		test.ml.log.Log(test.msg)

		if test.msg.Level != LLPanic.String() {
			verify(id, test)
		}

	}
}

func TestMultiLoggerPanic(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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

		r := recover()
		if r == nil {
			t.Errorf(
				"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI).Panic(%s) -- test did not panic",
				id,
				test.msg.Msg,
			)
			return
		}

		if r != test.msg.Msg {
			t.Errorf(
				"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI).Panic(%s) -- panic message %s does not match input %s",
				id,
				test.msg.Msg,
				r,
				test.msg.Msg,
			)
		}

		for bufID, buf := range test.ml.buf {

			logEntry := &LogMessage{}

			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panic(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panic(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLPanic.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panic(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLInfo.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panic(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panic(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panic(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panic(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panic(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panic(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)

		defer verify(id, test)
		test.ml.log.Panic(test.msg.Msg)

	}
}

func TestMultiLoggerPanicln(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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

		r := recover()
		if r == nil {
			t.Errorf(
				"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI).Panicln(%s) -- test did not panic",
				id,
				test.msg.Msg,
			)
			return
		}

		if r != test.msg.Msg+"\n" {
			t.Errorf(
				"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI).Panicln(%s) -- panic message %s does not match input %s",
				id,
				test.msg.Msg,
				r,
				test.msg.Msg,
			)
		}

		for bufID, buf := range test.ml.buf {

			logEntry := &LogMessage{}

			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicln(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicln(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLPanic.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicln(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLInfo.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicln(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicln(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicln(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicln(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicln(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicln(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)

		defer verify(id, test)
		test.ml.log.Panicln(test.msg.Msg)

	}
}

func TestMultiLoggerPanicf(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		format string
		v      []interface{}
		msg    *LogMessage
		ml     ml
	}

	var testAllObjects []map[string]interface{}
	testAllObjects = append(testAllObjects, testObjects...)
	testAllObjects = append(testAllObjects, testEmptyObjects...)

	var tests []test

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(mockMessages[b]).
						Metadata(testAllObjects[c]).
						Build(),
					format: "%s",
					v:      []interface{}{mockMessages[b]},
				}

				tests = append(tests, obj)
			}
		}
	}
	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockFmtMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
						Metadata(testAllObjects[c]).
						Build(),
					format: mockFmtMessages[b].format,
					v:      mockFmtMessages[b].v,
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

		r := recover()
		if r == nil {
			t.Errorf(
				"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI).Panicf(%s, %s) -- test did not panic",
				id,
				test.format,
				test.v,
			)
			return
		}

		if r != test.msg.Msg {
			t.Errorf(
				"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI).Panicf(%s, %s) -- panic message %s does not match input %s",
				id,
				test.format,
				test.v,
				r,
				test.msg.Msg,
			)
		}

		for bufID, buf := range test.ml.buf {

			logEntry := &LogMessage{}

			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicf(%s, %s) -- unmarshal error: %s",
					id,
					bufID,
					test.format,
					test.v,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLPanic.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicf(%s, %s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					LLInfo.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicf(%s, %s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicf(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicf(%s, %s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicf(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.format,
							test.v,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicf(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.format,
						test.v,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Panicf(%s, %s) -- %s",
				id,
				bufID,
				test.format,
				test.v,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)

		defer verify(id, test)
		test.ml.log.Panicf(test.format, test.v...)

	}
}

func TestMultiLoggerFatal(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf), SkipExitCfg))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatal(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatal(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLFatal.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatal(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLFatal.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatal(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatal(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatal(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatal(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatal(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatal(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Fatal(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerFatalln(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf), SkipExitCfg))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalln(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalln(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLFatal.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalln(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLFatal.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalln(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalln(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalln(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalln(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalln(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalln(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Fatalln(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerFatalf(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		format string
		v      []interface{}
		msg    *LogMessage
		ml     ml
	}

	var testAllObjects []map[string]interface{}
	testAllObjects = append(testAllObjects, testObjects...)
	testAllObjects = append(testAllObjects, testEmptyObjects...)

	var tests []test

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf), SkipExitCfg))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(mockMessages[b]).
						Metadata(testAllObjects[c]).
						Build(),
					format: "%s",
					v:      []interface{}{mockMessages[b]},
				}

				tests = append(tests, obj)
			}
		}
	}

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockFmtMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf), SkipExitCfg))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
						Metadata(testAllObjects[c]).
						Build(),
					format: mockFmtMessages[b].format,
					v:      mockFmtMessages[b].v,
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalf(%s, %s) -- unmarshal error: %s",
					id,
					bufID,
					test.format,
					test.v,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLFatal.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalf(%s, %s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					LLFatal.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalf(%s, %s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalf(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalf(%s, %s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalf(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.format,
							test.v,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalf(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.format,
						test.v,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Fatalf(%s, %s) -- %s",
				id,
				bufID,
				test.format,
				test.v,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Fatalf(test.format, test.v...)

		verify(id, test)

	}
}

func TestMultiLoggerError(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Error(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Error(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLError.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Error(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLError.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Error(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Error(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Error(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Error(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Error(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Error(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Error(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerErrorln(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorln(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorln(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLError.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorln(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLError.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorln(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorln(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorln(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorln(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorln(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorln(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Errorln(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerErrorf(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		format string
		v      []interface{}
		msg    *LogMessage
		ml     ml
	}

	var testAllObjects []map[string]interface{}
	testAllObjects = append(testAllObjects, testObjects...)
	testAllObjects = append(testAllObjects, testEmptyObjects...)

	var tests []test

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(mockMessages[b]).
						Metadata(testAllObjects[c]).
						Build(),
					format: "%s",
					v:      []interface{}{mockMessages[b]},
				}

				tests = append(tests, obj)
			}
		}
	}

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockFmtMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
						Metadata(testAllObjects[c]).
						Build(),
					format: mockFmtMessages[b].format,
					v:      mockFmtMessages[b].v,
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorf(%s, %s) -- unmarshal error: %s",
					id,
					bufID,
					test.format,
					test.v,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLError.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorf(%s, %s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					LLError.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorf(%s, %s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorf(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorf(%s, %s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorf(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.format,
							test.v,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorf(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.format,
						test.v,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Errorf(%s, %s) -- %s",
				id,
				bufID,
				test.format,
				test.v,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Errorf(test.format, test.v...)

		verify(id, test)

	}
}

func TestMultiLoggerWarn(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warn(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warn(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLWarn.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warn(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLWarn.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warn(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warn(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warn(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warn(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warn(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warn(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Warn(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerWarnln(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnln(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnln(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLWarn.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnln(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLWarn.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnln(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnln(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnln(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnln(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnln(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnln(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Warnln(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerWarnf(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		format string
		v      []interface{}
		msg    *LogMessage
		ml     ml
	}

	var testAllObjects []map[string]interface{}
	testAllObjects = append(testAllObjects, testObjects...)
	testAllObjects = append(testAllObjects, testEmptyObjects...)

	var tests []test

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(mockMessages[b]).
						Metadata(testAllObjects[c]).
						Build(),
					format: "%s",
					v:      []interface{}{mockMessages[b]},
				}

				tests = append(tests, obj)
			}
		}
	}

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockFmtMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
						Metadata(testAllObjects[c]).
						Build(),
					format: mockFmtMessages[b].format,
					v:      mockFmtMessages[b].v,
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnf(%s, %s) -- unmarshal error: %s",
					id,
					bufID,
					test.format,
					test.v,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLWarn.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnf(%s, %s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					LLWarn.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnf(%s, %s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnf(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnf(%s, %s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnf(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.format,
							test.v,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnf(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.format,
						test.v,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Warnf(%s, %s) -- %s",
				id,
				bufID,
				test.format,
				test.v,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Warnf(test.format, test.v...)

		verify(id, test)

	}
}

func TestMultiLoggerInfo(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Info(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Info(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLInfo.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Info(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLInfo.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Info(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Info(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Info(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Info(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Info(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Info(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Info(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerInfoln(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infoln(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infoln(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLInfo.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infoln(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLInfo.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infoln(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infoln(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infoln(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infoln(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infoln(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infoln(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Infoln(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerInfof(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		format string
		v      []interface{}
		msg    *LogMessage
		ml     ml
	}

	var testAllObjects []map[string]interface{}
	testAllObjects = append(testAllObjects, testObjects...)
	testAllObjects = append(testAllObjects, testEmptyObjects...)

	var tests []test

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(mockMessages[b]).
						Metadata(testAllObjects[c]).
						Build(),
					format: "%s",
					v:      []interface{}{mockMessages[b]},
				}

				tests = append(tests, obj)
			}
		}
	}

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockFmtMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
						Metadata(testAllObjects[c]).
						Build(),
					format: mockFmtMessages[b].format,
					v:      mockFmtMessages[b].v,
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infof(%s, %s) -- unmarshal error: %s",
					id,
					bufID,
					test.format,
					test.v,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infof(%s, %s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLInfo.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infof(%s, %s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					LLInfo.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infof(%s, %s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infof(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infof(%s, %s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infof(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.format,
							test.v,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infof(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.format,
						test.v,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Infof(%s, %s) -- %s",
				id,
				bufID,
				test.format,
				test.v,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Infof(test.format, test.v...)

		verify(id, test)

	}
}

func TestMultiLoggerDebug(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debug(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debug(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLDebug.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debug(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLDebug.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debug(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debug(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debug(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debug(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debug(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debug(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Debug(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerDebugln(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugln(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugln(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLDebug.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugln(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLDebug.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugln(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugln(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugln(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugln(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugln(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugln(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Debugln(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerDebugf(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		format string
		v      []interface{}
		msg    *LogMessage
		ml     ml
	}

	var testAllObjects []map[string]interface{}
	testAllObjects = append(testAllObjects, testObjects...)
	testAllObjects = append(testAllObjects, testEmptyObjects...)

	var tests []test

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(mockMessages[b]).
						Metadata(testAllObjects[c]).
						Build(),
					format: "%s",
					v:      []interface{}{mockMessages[b]},
				}

				tests = append(tests, obj)
			}
		}
	}

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockFmtMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
						Metadata(testAllObjects[c]).
						Build(),
					format: mockFmtMessages[b].format,
					v:      mockFmtMessages[b].v,
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugf(%s, %s) -- unmarshal error: %s",
					id,
					bufID,
					test.format,
					test.v,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLDebug.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugf(%s, %s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					LLDebug.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugf(%s, %s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugf(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugf(%s, %s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugf(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.format,
							test.v,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugf(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.format,
						test.v,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Debugf(%s, %s) -- %s",
				id,
				bufID,
				test.format,
				test.v,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Debugf(test.format, test.v...)

		verify(id, test)

	}
}

func TestMultiLoggerTrace(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Trace(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Trace(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLTrace.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Trace(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLTrace.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Trace(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Trace(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Trace(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Trace(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Trace(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Trace(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Trace(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerTraceln(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		msg *LogMessage
		ml  ml
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

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(testAllMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Traceln(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Traceln(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLTrace.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Traceln(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					LLTrace.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Traceln(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Traceln(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Traceln(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Traceln(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Traceln(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Traceln(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Traceln(test.msg.Msg)

		verify(id, test)

	}
}

func TestMultiLoggerTracef(t *testing.T) {
	type ml struct {
		log LoggerI
		buf []*bytes.Buffer
	}

	type test struct {
		format string
		v      []interface{}
		msg    *LogMessage
		ml     ml
	}

	var testAllObjects []map[string]interface{}
	testAllObjects = append(testAllObjects, testObjects...)
	testAllObjects = append(testAllObjects, testEmptyObjects...)

	var tests []test

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(mockMessages[b]).
						Metadata(testAllObjects[c]).
						Build(),
					format: "%s",
					v:      []interface{}{mockMessages[b]},
				}

				tests = append(tests, obj)
			}
		}
	}

	for a := 0; a < len(mockPrefixes); a++ {
		for b := 0; b < len(mockFmtMessages); b++ {
			for c := 0; c < len(testAllObjects); c++ {

				var bufs []*bytes.Buffer
				var logs []LoggerI
				for e := 0; e < len(mockMultiPrefixes); e++ {
					buf := &bytes.Buffer{}
					bufs = append(bufs, buf)
					logs = append(logs, New(WithPrefix(mockMultiPrefixes[e]), JSONFormat, WithOut(buf)))
				}
				mlogger := MultiLogger(logs...)

				obj := test{
					ml: ml{
						log: mlogger,
						buf: bufs,
					},
					msg: NewMessage().
						Prefix(mockPrefixes[a]).
						Message(fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...)).
						Metadata(testAllObjects[c]).
						Build(),
					format: mockFmtMessages[b].format,
					v:      mockFmtMessages[b].v,
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
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Tracef(%s, %s) -- unmarshal error: %s",
					id,
					bufID,
					test.format,
					test.v,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Tracef(%s, %s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != LLTrace.String() {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Tracef(%s, %s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					LLTrace.String(),
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Tracef(%s, %s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Tracef(%s, %s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Tracef(%s, %s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.format,
					test.v,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Tracef(%s, %s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.format,
							test.v,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Tracef(%s, %s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.format,
						test.v,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [MultiLogger] MultiLogger(...LoggerI[%v]).Tracef(%s, %s) -- %s",
				id,
				bufID,
				test.format,
				test.v,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		for _, b := range test.ml.buf {
			b.Reset()
		}

		test.ml.log.Prefix(test.msg.Prefix).Fields(test.msg.Metadata)
		test.ml.log.Tracef(test.format, test.v...)

		verify(id, test)

	}
}
