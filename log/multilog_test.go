package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"testing"
)

const prefix string = "multilog-test"

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
	New(prefix+"-0", JSONFormat, mockBufs[0]),
	New(prefix+"-1", JSONFormat, mockBufs[1]),
	New(prefix+"-2", JSONFormat, mockBufs[2]),
	New(prefix+"-3", JSONFormat, mockBufs[3]),
	New(prefix+"-4", JSONFormat, mockBufs[4]),
	New(prefix+"-5", JSONFormat, mockBufs[5]),
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
			test.prefix = append(test.prefix, prefix+"-"+strconv.Itoa(b))
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
					msg := NewMessage().Level(
						mockLogLevelsOK[a],
					).Prefix(
						mockPrefixes[b],
					).Message(
						testAllMessages[c],
					).Metadata(
						testAllObjects[d],
					).Build()

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
							msg: NewMessage().Level(
								mockLogLevelsOK[b],
							).Prefix(
								mockPrefixes[c],
							).Message(
								testAllMessages[d],
							).Metadata(
								testAllObjects[e],
							).Build(),
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
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).SetOuts(...io.Writer) -- retrieved unexpected metadata object: wanted %s ; got %s",
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
							"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).SetOuts(...io.Writer) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
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
						"#%v -- FAILED -- [MultiLogger] MultiLogger(...LoggerI[%v]).SetOuts(...io.Writer) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
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

}

func TestMultiLoggerSetPrefix(t *testing.T) {

}

func TestMultiLoggerFields(t *testing.T) {

}

func TestMultiLoggerPrint(t *testing.T) {

}

func TestMultiLoggerPrintln(t *testing.T) {

}

func TestMultiLoggerPrintf(t *testing.T) {

}

func TestMultiLoggerLog(t *testing.T) {

}

func TestMultiLoggerPanic(t *testing.T) {

}

func TestMultiLoggerPanicln(t *testing.T) {

}

func TestMultiLoggerPanicf(t *testing.T) {

}

func TestMultiLoggerFatal(t *testing.T) {

}

func TestMultiLoggerFatalln(t *testing.T) {

}

func TestMultiLoggerFatalf(t *testing.T) {

}

func TestMultiLoggerError(t *testing.T) {

}

func TestMultiLoggerErrorln(t *testing.T) {

}

func TestMultiLoggerErrorf(t *testing.T) {

}

func TestMultiLoggerWarn(t *testing.T) {

}

func TestMultiLoggerWarnln(t *testing.T) {

}

func TestMultiLoggerWarnf(t *testing.T) {

}

func TestMultiLoggerInfo(t *testing.T) {

}

func TestMultiLoggerInfoln(t *testing.T) {

}

func TestMultiLoggerInfof(t *testing.T) {

}

func TestMultiLoggerDebug(t *testing.T) {

}

func TestMultiLoggerDebugln(t *testing.T) {

}

func TestMultiLoggerDebugf(t *testing.T) {

}

func TestMultiLoggerTrace(t *testing.T) {

}

func TestMultiLoggerTraceln(t *testing.T) {

}

func TestMultiLoggerTracef(t *testing.T) {

}
