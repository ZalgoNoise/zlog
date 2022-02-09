package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

var mockBuffer = &bytes.Buffer{}
var mockLogger = struct {
	logger LoggerI
	buf    *bytes.Buffer
}{
	logger: New("test-message", JSONFormat, mockBuffer),
	buf:    mockBuffer,
}

var mockLogLevelsOK = []LogLevel{
	LogLevel(0),
	LogLevel(1),
	LogLevel(2),
	LogLevel(3),
	LogLevel(4),
	LogLevel(5),
	LogLevel(9),
}

var mockLogLevelsNOK = []LogLevel{
	LogLevel(6),
	LogLevel(7),
	LogLevel(8),
	LogLevel(10),
	LogLevel(-1),
	LogLevel(200),
	LogLevel(500),
}

var mockMessages = []string{
	"message test #1",
	"message test #2",
	"message test #3",
	"message test #4",
	"message test #5",
	"mock message",
	"{ logger text in brackets }",
}

var mockFmtMessages = []struct {
	format string
	v      []interface{}
}{
	{
		format: "mockLogLevelsOK length: %v",
		v: []interface{}{
			len(mockLogLevelsOK),
		},
	},
	{
		format: "'Hello world!' in a list: %s",
		v: []interface{}{
			[]rune{'H', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd', '!'},
		},
	},
	{
		format: "seven times three = %v",
		v: []interface{}{
			21,
		},
	},
}

func TestLogLevelString(t *testing.T) {
	type test struct {
		input LogLevel
		ok    string
		pass  bool
	}

	var passingTests []test

	for k, v := range logTypeVals {
		passingTests = append(passingTests, test{
			input: k,
			ok:    v,
			pass:  true,
		})
	}

	var failingTests = []test{
		{
			input: LogLevel(6),
			ok:    "",
			pass:  false,
		},
		{
			input: LogLevel(7),
			ok:    "",
			pass:  false,
		},
		{
			input: LogLevel(8),
			ok:    "",
			pass:  false,
		},
		{
			input: LogLevel(10),
			ok:    "",
			pass:  false,
		},
	}

	var allTests []test
	allTests = append(allTests, passingTests...)
	allTests = append(allTests, failingTests...)

	for id, test := range allTests {
		result := test.input.String()

		if result == "" && test.pass {
			t.Errorf(
				"#%v [LogLevel] LogLevel(%v).String() -- unexpected reference, got %s",
				id,
				int(test.input),
				result,
			)
		}

		if result != test.ok && !test.pass {
			t.Errorf(
				"#%v [LogLevel] LogLevel(%v).String() -- expected %s, got %s",
				id,
				int(test.input),
				test.ok,
				result,
			)
		} else {
			t.Logf(
				"#%v -- TESTED -- [LogLevel] LogLevel(%v).String() = %s",
				id,
				int(test.input),
				result,
			)
		}
	}
}

func TestLoggerOutput(t *testing.T) {
	type test struct {
		level     LogLevel
		msg       string
		wantLevel string
		wantMsg   string
		ok        bool
	}

	var tests []test

	for a := 0; a < len(mockLogLevelsOK); a++ {
		for b := 0; b < len(mockMessages); b++ {
			tests = append(tests, test{
				level:     mockLogLevelsOK[a],
				msg:       mockMessages[b],
				wantLevel: mockLogLevelsOK[a].String(),
				wantMsg:   mockMessages[b],
				ok:        true,
			})
		}
		for c := 0; c < len(mockFmtMessages); c++ {
			tests = append(tests, test{
				level:     mockLogLevelsOK[a],
				msg:       fmt.Sprintf(mockFmtMessages[c].format, mockFmtMessages[c].v...),
				wantLevel: mockLogLevelsOK[a].String(),
				wantMsg:   fmt.Sprintf(mockFmtMessages[c].format, mockFmtMessages[c].v...),
				ok:        true,
			})
		}
	}
	for d := 0; d < len(mockLogLevelsNOK); d++ {
		for e := 0; e < len(mockMessages); e++ {
			tests = append(tests, test{
				level:     mockLogLevelsNOK[d],
				msg:       mockMessages[e],
				wantLevel: mockLogLevelsNOK[d].String(),
				wantMsg:   mockMessages[e],
				ok:        false,
			})
		}
		for f := 0; f < len(mockFmtMessages); f++ {
			tests = append(tests, test{
				level:     mockLogLevelsNOK[d],
				msg:       fmt.Sprintf(mockFmtMessages[f].format, mockFmtMessages[f].v...),
				wantLevel: mockLogLevelsNOK[d].String(),
				wantMsg:   fmt.Sprintf(mockFmtMessages[f].format, mockFmtMessages[f].v...),
				ok:        true,
			})
		}
	}

	for id, test := range tests {

		logEntry := &LogMessage{}

		mockLogger.logger.Output(test.level, test.msg)

		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v [LoggerMessage] Output(%s, %s) -- unmarshal error: %s",
				id,
				test.level.String(),
				test.msg,
				err,
			)
		}

		if logEntry.Level != test.wantLevel {
			t.Errorf(
				"#%v [LoggerMessage] Output(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.level.String(),
				test.msg,
				test.wantLevel,
				logEntry.Level,
			)
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v [LoggerMessage] Output(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.level.String(),
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
		}

		t.Logf(
			"#%v -- TESTED -- [LoggerMessage] Output(%s, %s) : %s",
			id,
			test.level.String(),
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}
}

func TestLoggerPrint(t *testing.T) {
	type test struct {
		msg     string
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		tests = append(tests, test{
			msg:     mockMessages[a],
			wantMsg: mockMessages[a],
			ok:      true,
		})
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		tests = append(tests, test{
			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		})
	}

	for id, test := range tests {

		logEntry := &LogMessage{}

		mockLogger.logger.Print(test.msg)

		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v [LoggerMessage] Print(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v [LoggerMessage] Print(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
		}

		t.Logf(
			"#%v -- TESTED -- [LoggerMessage] Print(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}
}

func TestLoggerPrintln(t *testing.T) {
	type test struct {
		msg     string
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		tests = append(tests, test{
			msg:     mockMessages[a],
			wantMsg: mockMessages[a],
			ok:      true,
		})
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		tests = append(tests, test{
			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		})
	}

	for id, test := range tests {

		logEntry := &LogMessage{}

		mockLogger.logger.Println(test.msg)

		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v [LoggerMessage] Println(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v [LoggerMessage] Println(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
		}

		t.Logf(
			"#%v -- TESTED -- [LoggerMessage] Println(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}
}

func TestLoggerPrintf(t *testing.T) {
	type test struct {
		format  string
		v       []interface{}
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockFmtMessages); a++ {
		tests = append(tests, test{
			format:  mockFmtMessages[a].format,
			v:       mockFmtMessages[a].v,
			wantMsg: fmt.Sprintf(mockFmtMessages[a].format, mockFmtMessages[a].v...),
			ok:      true,
		})
	}

	for id, test := range tests {

		logEntry := &LogMessage{}

		mockLogger.logger.Printf(test.format, test.v...)

		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v [LoggerMessage] Printf(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v [LoggerMessage] Printf(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
		}

		t.Logf(
			"#%v -- TESTED -- [LoggerMessage] Printf(%s, %s) : %s",
			id,
			test.format,
			test.v,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}
}

func TestLoggerLog(t *testing.T) {

}

func TestLoggerLogln(t *testing.T) {

}

func TestLoggerLogf(t *testing.T) {

}

func TestLoggerPanic(t *testing.T) {

}

func TestLoggerPanicln(t *testing.T) {

}

func TestLoggerPanicf(t *testing.T) {

}

func TestLoggerFatal(t *testing.T) {

}

func TestLoggerFatalln(t *testing.T) {

}

func TestLoggerFatalf(t *testing.T) {

}

func TestLoggerError(t *testing.T) {

}

func TestLoggerErrorln(t *testing.T) {

}

func TestLoggerErrorf(t *testing.T) {

}

func TestLoggerWarn(t *testing.T) {

}

func TestLoggerWarnln(t *testing.T) {

}

func TestLoggerWarnf(t *testing.T) {

}

func TestLoggerInfo(t *testing.T) {

}

func TestLoggerInfoln(t *testing.T) {

}

func TestLoggerInfof(t *testing.T) {

}

func TestLoggerDebug(t *testing.T) {

}

func TestLoggerDebugln(t *testing.T) {

}

func TestLoggerDebugf(t *testing.T) {

}

func TestLoggerTrace(t *testing.T) {

}

func TestLoggerTraceln(t *testing.T) {

}

func TestLoggerTracef(t *testing.T) {

}

func TestPrint(t *testing.T) {

}

func TestPrintln(t *testing.T) {

}

func TestPrintf(t *testing.T) {

}

func TestLog(t *testing.T) {

}

func TestLogln(t *testing.T) {

}

func TestLogf(t *testing.T) {

}

func TestPanic(t *testing.T) {

}

func TestPanicln(t *testing.T) {

}

func TestPanicf(t *testing.T) {

}

func TestFatal(t *testing.T) {

}

func TestFatalln(t *testing.T) {

}

func TestFatalf(t *testing.T) {

}

func TestError(t *testing.T) {

}

func TestErrorln(t *testing.T) {

}

func TestErrorf(t *testing.T) {

}

func TestWarn(t *testing.T) {

}

func TestWarnln(t *testing.T) {

}

func TestWarnf(t *testing.T) {

}

func TestInfo(t *testing.T) {

}

func TestInfoln(t *testing.T) {

}

func TestInfof(t *testing.T) {

}

func TestDebug(t *testing.T) {

}

func TestDebugln(t *testing.T) {

}

func TestDebugf(t *testing.T) {

}

func TestTrace(t *testing.T) {

}

func TestTraceln(t *testing.T) {

}

func TestTracef(t *testing.T) {

}
