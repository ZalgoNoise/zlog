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
			[]string{"H", "e", "l", "l", "o", " ", "w", "o", "r", "l", "d", "!"},
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

	var verify = func(id int, test test, result string) {
		if result == "" && test.pass {
			t.Errorf(
				"#%v -- FAILED -- [LogLevel] LogLevel(%v).String() -- unexpected reference, got %s",
				id,
				int(test.input),
				result,
			)
			return
		}

		if result != test.ok && !test.pass {
			t.Errorf(
				"#%v -- FAILED -- [LogLevel] LogLevel(%v).String() -- expected %s, got %s",
				id,
				int(test.input),
				test.ok,
				result,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LogLevel] LogLevel(%v).String() = %s",
			id,
			int(test.input),
			result,
		)

	}

	for id, test := range allTests {
		result := test.input.String()

		verify(id, test, result)

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

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Output(%s, %s) -- unmarshal error: %s",
				id,
				test.level.String(),
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != test.wantLevel {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Output(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.level.String(),
				test.msg,
				test.wantLevel,
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Output(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.level.String(),
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Output(%s, %s) : %s",
			id,
			test.level.String(),
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Output(test.level, test.msg)

		verify(id, test, logEntry)

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

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Print(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Print(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Print(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Print(test.msg)

		verify(id, test, logEntry)
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

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Println(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Println(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Println(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Println(test.msg)

		verify(id, test, logEntry)
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

	for a := 0; a < len(mockMessages); a++ {
		test := test{
			format:  "%s",
			v:       []interface{}{mockMessages[a]},
			wantMsg: mockMessages[a],
			ok:      true,
		}

		tests = append(tests, test)
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		test := test{
			format:  mockFmtMessages[b].format,
			v:       mockFmtMessages[b].v,
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		}

		tests = append(tests, test)
	}

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Printf(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Printf(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Printf(%s, %s) : %s",
			id,
			test.format,
			test.v,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Printf(test.format, test.v...)

		verify(id, test, logEntry)
	}
}

func TestLoggerLog(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf(
				"# -- TEST -- [Logger.Log()] Intended panic recovery -- %s",
				r,
			)
		}
	}()

	type test struct {
		level     LogLevel
		msg       string
		wantLevel string
		wantMsg   string
		ok        bool
		panics    bool
	}

	var tests []test

	for a := 0; a < len(mockLogLevelsOK); a++ {
		if a == 5 {
			continue // skip LLFatal, or os.Exit(1)
		}
		for b := 0; b < len(mockMessages); b++ {
			test := test{
				level:     mockLogLevelsOK[a],
				msg:       mockMessages[b],
				wantLevel: mockLogLevelsOK[a].String(),
				wantMsg:   mockMessages[b],
				ok:        true,
				panics:    false,
			}

			if a == 9 {
				test.panics = true
			}

			tests = append(tests, test)
		}
		for c := 0; c < len(mockFmtMessages); c++ {
			test := test{
				level:     mockLogLevelsOK[a],
				msg:       fmt.Sprintf(mockFmtMessages[c].format, mockFmtMessages[c].v...),
				wantLevel: mockLogLevelsOK[a].String(),
				wantMsg:   fmt.Sprintf(mockFmtMessages[c].format, mockFmtMessages[c].v...),
				ok:        true,
				panics:    false,
			}

			if a == 9 {
				test.panics = true
			}

			tests = append(tests, test)
		}
	}
	for d := 0; d < len(mockLogLevelsNOK); d++ {
		if d == 5 {
			continue // skip LLFatal, or os.Exit(1)
		}
		for e := 0; e < len(mockMessages); e++ {
			test := test{
				level:     mockLogLevelsNOK[d],
				msg:       mockMessages[e],
				wantLevel: mockLogLevelsNOK[d].String(),
				wantMsg:   mockMessages[e],
				ok:        false,
				panics:    false,
			}

			if d == 9 {
				test.panics = true
			}

			tests = append(tests, test)
		}
		for f := 0; f < len(mockFmtMessages); f++ {
			test := test{
				level:     mockLogLevelsNOK[d],
				msg:       fmt.Sprintf(mockFmtMessages[f].format, mockFmtMessages[f].v...),
				wantLevel: mockLogLevelsNOK[d].String(),
				wantMsg:   fmt.Sprintf(mockFmtMessages[f].format, mockFmtMessages[f].v...),
				ok:        true,
				panics:    false,
			}

			if d == 9 {
				test.panics = true
			}

			tests = append(tests, test)
		}
	}

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Log(%s, %s) -- unmarshal error: %s",
				id,
				test.level.String(),
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != test.wantLevel {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Log(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.level.String(),
				test.msg,
				test.wantLevel,
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Log(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.level.String(),
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Log(%s, %s) : %s",
			id,
			test.level.String(),
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		if test.level.String() == "panic" {
			defer verify(id, test, logEntry)
		}

		mockLogger.logger.Log(test.level, test.msg)

		if test.level.String() != "panic" {
			verify(id, test, logEntry)
		}
	}

}

func TestLoggerLogln(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf(
				"# -- TEST -- [Logger.Logln()] Intended panic recovery -- %s",
				r,
			)
		}
	}()

	type test struct {
		level     LogLevel
		msg       string
		wantLevel string
		wantMsg   string
		ok        bool
		panics    bool
	}

	var tests []test

	for a := 0; a < len(mockLogLevelsOK); a++ {
		if a == 5 {
			continue // skip LLFatal, or os.Exit(1)
		}
		for b := 0; b < len(mockMessages); b++ {
			test := test{
				level:     mockLogLevelsOK[a],
				msg:       mockMessages[b],
				wantLevel: mockLogLevelsOK[a].String(),
				wantMsg:   mockMessages[b],
				ok:        true,
				panics:    false,
			}

			if a == 9 {
				test.panics = true
			}

			tests = append(tests, test)
		}
		for c := 0; c < len(mockFmtMessages); c++ {
			test := test{
				level:     mockLogLevelsOK[a],
				msg:       fmt.Sprintf(mockFmtMessages[c].format, mockFmtMessages[c].v...),
				wantLevel: mockLogLevelsOK[a].String(),
				wantMsg:   fmt.Sprintf(mockFmtMessages[c].format, mockFmtMessages[c].v...),
				ok:        true,
				panics:    false,
			}

			if a == 9 {
				test.panics = true
			}

			tests = append(tests, test)
		}
	}
	for d := 0; d < len(mockLogLevelsNOK); d++ {
		if d == 5 {
			continue // skip LLFatal, or os.Exit(1)
		}
		for e := 0; e < len(mockMessages); e++ {
			test := test{
				level:     mockLogLevelsNOK[d],
				msg:       mockMessages[e],
				wantLevel: mockLogLevelsNOK[d].String(),
				wantMsg:   mockMessages[e],
				ok:        false,
				panics:    false,
			}

			if d == 9 {
				test.panics = true
			}

			tests = append(tests, test)
		}
		for f := 0; f < len(mockFmtMessages); f++ {
			test := test{
				level:     mockLogLevelsNOK[d],
				msg:       fmt.Sprintf(mockFmtMessages[f].format, mockFmtMessages[f].v...),
				wantLevel: mockLogLevelsNOK[d].String(),
				wantMsg:   fmt.Sprintf(mockFmtMessages[f].format, mockFmtMessages[f].v...),
				ok:        true,
				panics:    false,
			}

			if d == 9 {
				test.panics = true
			}

			tests = append(tests, test)
		}
	}

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Logln(%s, %s) -- unmarshal error: %s",
				id,
				test.level.String(),
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != test.wantLevel {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Logln(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.level.String(),
				test.msg,
				test.wantLevel,
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Logln(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.level.String(),
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Logln(%s, %s) : %s",
			id,
			test.level.String(),
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		if test.level.String() == "panic" {
			defer verify(id, test, logEntry)
		}

		mockLogger.logger.Logln(test.level, test.msg)

		if test.level.String() != "panic" {
			verify(id, test, logEntry)
		}

	}
}

func TestLoggerLogf(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf(
				"# -- TEST -- [Logger.Logf()] Intended panic recovery -- %s",
				r,
			)
		}
	}()

	type test struct {
		level     LogLevel
		format    string
		v         []interface{}
		wantLevel string
		wantMsg   string
		ok        bool
		panics    bool
	}

	var tests []test

	for a := 0; a < len(mockLogLevelsOK); a++ {
		if a == 5 {
			continue // skip LLFatal, or os.Exit(1)
		}

		for b := 0; b < len(mockMessages); b++ {
			test := test{
				level:     mockLogLevelsOK[a],
				format:    "%s",
				v:         []interface{}{mockMessages[b]},
				wantLevel: mockLogLevelsOK[a].String(),
				wantMsg:   mockMessages[b],
				ok:        true,
				panics:    false,
			}

			if a == 9 {
				test.panics = true
			}

			tests = append(tests, test)
		}

		for c := 0; c < len(mockFmtMessages); c++ {
			test := test{
				level:     mockLogLevelsOK[a],
				format:    mockFmtMessages[c].format,
				v:         mockFmtMessages[c].v,
				wantLevel: mockLogLevelsOK[a].String(),
				wantMsg:   fmt.Sprintf(mockFmtMessages[c].format, mockFmtMessages[c].v...),
				ok:        true,
				panics:    false,
			}

			if a == 9 {
				test.panics = true
			}

			tests = append(tests, test)
		}
	}
	for d := 0; d < len(mockLogLevelsNOK); d++ {
		if d == 5 {
			continue // skip LLFatal, or os.Exit(1)
		}

		for e := 0; e < len(mockMessages); e++ {
			test := test{
				level:     mockLogLevelsNOK[d],
				format:    "%s",
				v:         []interface{}{mockMessages[e]},
				wantLevel: mockLogLevelsNOK[d].String(),
				wantMsg:   mockMessages[e],
				ok:        true,
				panics:    false,
			}

			if d == 9 {
				test.panics = true
			}

			tests = append(tests, test)
		}

		for f := 0; f < len(mockFmtMessages); f++ {
			test := test{
				level:     mockLogLevelsNOK[d],
				format:    mockFmtMessages[f].format,
				v:         mockFmtMessages[f].v,
				wantLevel: mockLogLevelsNOK[d].String(),
				wantMsg:   fmt.Sprintf(mockFmtMessages[f].format, mockFmtMessages[f].v...),
				ok:        true,
				panics:    false,
			}

			if d == 9 {
				test.panics = true
			}

			tests = append(tests, test)
		}
	}

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Logf(%s, %s, %s) -- unmarshal error: %s",
				id,
				test.level.String(),
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Level != test.wantLevel {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Logf(%s, %s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.level.String(),
				test.format,
				test.v,
				test.wantLevel,
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Logf(%s, %s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.level.String(),
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Logf(%s, %s, %s) : %s",
			id,
			test.level.String(),
			test.format,
			test.v,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		if test.level.String() == "panic" {
			defer verify(id, test, logEntry)
		}

		mockLogger.logger.Logf(test.level, test.format, test.v...)

		if test.level.String() != "panic" {
			verify(id, test, logEntry)
		}

	}
}

func TestLoggerPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf(
				"# -- TEST -- [Logger.Panic()] Intended panic recovery -- %s",
				r,
			)
		}
	}()

	type test struct {
		msg     string
		wantMsg string
		ok      bool
		panics  bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		test := test{
			msg:     mockMessages[a],
			wantMsg: mockMessages[a],
			ok:      true,
			panics:  true,
		}

		tests = append(tests, test)
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		test := test{
			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
			panics:  true,
		}

		tests = append(tests, test)
	}

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Panic(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != LLPanic.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Panic(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLPanic.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Panic(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Panic(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()

	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		defer verify(id, test, logEntry)

		mockLogger.logger.Panic(test.msg)

	}
}

func TestLoggerPanicln(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf(
				"# -- TEST -- [Logger.Panicln()] Intended panic recovery -- %s",
				r,
			)
		}
	}()

	type test struct {
		msg     string
		wantMsg string
		ok      bool
		panics  bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		test := test{
			msg:     mockMessages[a],
			wantMsg: mockMessages[a],
			ok:      true,
			panics:  true,
		}

		tests = append(tests, test)
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		test := test{
			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
			panics:  true,
		}

		tests = append(tests, test)
	}

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Panicln(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != LLPanic.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Panicln(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLPanic.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Panicln(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Panicln(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()

	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		defer verify(id, test, logEntry)

		mockLogger.logger.Panicln(test.msg)

	}
}

func TestLoggerPanicf(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf(
				"# -- TEST -- [Logger.Panicf()] Intended panic recovery -- %s",
				r,
			)
		}
	}()

	type test struct {
		format  string
		v       []interface{}
		wantMsg string
		ok      bool
		panics  bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		test := test{
			format:  "%s",
			v:       []interface{}{mockMessages[a]},
			wantMsg: mockMessages[a],
			ok:      true,
			panics:  true,
		}

		tests = append(tests, test)
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		test := test{
			format:  mockFmtMessages[b].format,
			v:       mockFmtMessages[b].v,
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
			panics:  true,
		}

		tests = append(tests, test)
	}

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Panicf(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Level != LLPanic.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Panicf(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				LLPanic.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Panicf(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Panicf(%s, %s) : %s",
			id,
			test.format,
			test.v,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()

	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		defer verify(id, test, logEntry)

		mockLogger.logger.Panicf(test.format, test.v...)

	}
}

func TestLoggerFatal(t *testing.T) {
	// testing LLFatal with Logger.Output, since otherwise it will cause program to exit

	t.Logf(
		"#%v -- SKIPPED -- [LoggerMessage] Fatal(v ...interface{}) -- testing LLFatal will cause program to exit (code 1). To test Fatal errors, its logic and execution is explored in Logger.Output() instead.",
		0,
	)
}

func TestLoggerFatalln(t *testing.T) {
	// testing LLFatal with Logger.Output, since otherwise it will cause program to exit

	t.Logf(
		"#%v -- SKIPPED -- [LoggerMessage] Fatalln(v ...interface{}) -- testing LLFatal will cause program to exit (code 1). To test Fatal errors, its logic and execution is explored in Logger.Output() instead.",
		0,
	)
}

func TestLoggerFatalf(t *testing.T) {
	// testing LLFatal with Logger.Output, since otherwise it will cause program to exit

	t.Logf(
		"#%v -- SKIPPED -- [LoggerMessage] Fatalf(format string, v ...interface{}) -- testing LLFatal will cause program to exit (code 1). To test Fatal errors, its logic and execution is explored in Logger.Output() instead.",
		0,
	)
}

func TestLoggerError(t *testing.T) {
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

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Error(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != LLError.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Error(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLError.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Error(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Error(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Error(test.msg)

		verify(id, test, logEntry)
	}
}

func TestLoggerErrorln(t *testing.T) {
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

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Errorln(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != LLError.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Errorln(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLError.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Errorln(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Errorln(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Errorln(test.msg)

		verify(id, test, logEntry)
	}
}

func TestLoggerErrorf(t *testing.T) {
	type test struct {
		format  string
		v       []interface{}
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		test := test{
			format:  "%s",
			v:       []interface{}{mockMessages[a]},
			wantMsg: mockMessages[a],
			ok:      true,
		}

		tests = append(tests, test)
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		test := test{
			format:  mockFmtMessages[b].format,
			v:       mockFmtMessages[b].v,
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		}

		tests = append(tests, test)
	}

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Errorf(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Level != LLError.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Errorf(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				LLError.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Errorf(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Errorf(%s, %s) : %s",
			id,
			test.format,
			test.v,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Errorf(test.format, test.v...)

		verify(id, test, logEntry)
	}
}

func TestLoggerWarn(t *testing.T) {
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

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Warn(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != LLWarn.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Warn(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLWarn.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Warn(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Warn(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Warn(test.msg)

		verify(id, test, logEntry)
	}
}

func TestLoggerWarnln(t *testing.T) {
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

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Warnln(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != LLWarn.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Warnln(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLWarn.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Warnln(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Warnln(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Warnln(test.msg)

		verify(id, test, logEntry)
	}
}

func TestLoggerWarnf(t *testing.T) {
	type test struct {
		format  string
		v       []interface{}
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		test := test{
			format:  "%s",
			v:       []interface{}{mockMessages[a]},
			wantMsg: mockMessages[a],
			ok:      true,
		}

		tests = append(tests, test)
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		test := test{
			format:  mockFmtMessages[b].format,
			v:       mockFmtMessages[b].v,
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		}

		tests = append(tests, test)
	}

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Warnf(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Level != LLWarn.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Warnf(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				LLWarn.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Warnf(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Warnf(%s, %s) : %s",
			id,
			test.format,
			test.v,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Warnf(test.format, test.v...)

		verify(id, test, logEntry)
	}
}

func TestLoggerInfo(t *testing.T) {
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

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Info(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != LLInfo.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Info(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLInfo.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Info(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Info(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Info(test.msg)

		verify(id, test, logEntry)
	}
}

func TestLoggerInfoln(t *testing.T) {
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

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Infoln(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != LLInfo.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Infoln(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLInfo.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Infoln(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Infoln(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Infoln(test.msg)

		verify(id, test, logEntry)
	}
}

func TestLoggerInfof(t *testing.T) {
	type test struct {
		format  string
		v       []interface{}
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		test := test{
			format:  "%s",
			v:       []interface{}{mockMessages[a]},
			wantMsg: mockMessages[a],
			ok:      true,
		}

		tests = append(tests, test)
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		test := test{
			format:  mockFmtMessages[b].format,
			v:       mockFmtMessages[b].v,
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		}

		tests = append(tests, test)
	}

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Infof(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Level != LLInfo.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Infof(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				LLInfo.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Infof(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Infof(%s, %s) : %s",
			id,
			test.format,
			test.v,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Infof(test.format, test.v...)

		verify(id, test, logEntry)
	}
}

func TestLoggerDebug(t *testing.T) {
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

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Debug(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != LLDebug.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Debug(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLDebug.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Debug(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Debug(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Debug(test.msg)

		verify(id, test, logEntry)
	}
}

func TestLoggerDebugln(t *testing.T) {
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

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Debugln(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != LLDebug.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Debugln(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLDebug.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Debugln(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Debugln(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Debugln(test.msg)

		verify(id, test, logEntry)
	}
}

func TestLoggerDebugf(t *testing.T) {
	type test struct {
		format  string
		v       []interface{}
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		test := test{
			format:  "%s",
			v:       []interface{}{mockMessages[a]},
			wantMsg: mockMessages[a],
			ok:      true,
		}

		tests = append(tests, test)
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		test := test{
			format:  mockFmtMessages[b].format,
			v:       mockFmtMessages[b].v,
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		}

		tests = append(tests, test)
	}

	var verify = func(id int, test test, logEntry *LogMessage) {
		if err := json.Unmarshal(mockLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Infof(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Level != LLDebug.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Debugf(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				LLDebug.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Debugf(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Debugf(%s, %s) : %s",
			id,
			test.format,
			test.v,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Debugf(test.format, test.v...)

		verify(id, test, logEntry)
	}
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
