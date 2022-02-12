package log

import (
	"bytes"
	"encoding/json"
	"strconv"
	"testing"
)

const prefix string = "multilog-test"

const msg string = "multilogger test message"

var mockBufs = []*bytes.Buffer{
	&bytes.Buffer{},
	&bytes.Buffer{},
	&bytes.Buffer{},
	&bytes.Buffer{},
	&bytes.Buffer{},
	&bytes.Buffer{},
}

var mockLoggers = []LoggerI{
	New(prefix+"-0", JSONFormat, mockBufs[0]),
	New(prefix+"-1", JSONFormat, mockBufs[1]),
	New(prefix+"-2", JSONFormat, mockBufs[2]),
	New(prefix+"-3", JSONFormat, mockBufs[3]),
	New(prefix+"-4", JSONFormat, mockBufs[4]),
	New(prefix+"-5", JSONFormat, mockBufs[5]),
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
