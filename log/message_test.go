package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

var mockBuffer = &bytes.Buffer{}
var mockLogger = struct {
	logger LoggerI
	buf    *bytes.Buffer
}{
	logger: New(
		WithPrefix("test-message"),
		JSONFormat,
		WithOut(mockBuffer),
	),
	buf: mockBuffer,
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

var mockPrefixes = []string{
	"test-logger",
	"test-prefix",
	"test-log",
	"test-service",
	"test-module",
	"test-logic",
}

var mockEmptyPrefixes = []string{
	"",
	"",
	"",
	"",
	"",
	"",
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

var testObjects = []map[string]interface{}{
	{
		"testID": 0,
		"desc":   "this is a test with custom metadata",
		"content": map[string]interface{}{
			"nestLevel": 1,
			"data":      "nested object #1",
			"content": map[string]interface{}{
				"nestLevel": 2,
				"data":      "nested object #2",
			},
		},
		"date": time.Now().Format(time.RFC3339),
	}, {
		"testID": 1,
		"desc":   "this is a test with custom metadata",
		"content": map[string]interface{}{
			"nestLevel": 1,
			"data":      "nested object #1",
			"content": map[string]interface{}{
				"nestLevel": 2,
				"data":      "nested object #2",
				"content": map[string]interface{}{
					"nestLevel": 3,
					"data":      "nested object #3",
				},
			},
		},
		"date": time.Now().Format(time.RFC3339),
	}, {
		"testID": 2,
		"desc":   "this is a test with custom metadata",
		"content": map[string]interface{}{
			"nestLevel": 1,
			"data":      "nested object #1",
			"content": map[string]interface{}{
				"nestLevel": 2,
				"data":      "nested object #2",
				"content": map[string]interface{}{
					"nestLevel": 3,
					"data":      "nested object #3",
					"content": map[string]interface{}{
						"nestLevel": 4,
						"data":      "nested object #4",
					},
				},
			},
		},
		"date": time.Now().Format(time.RFC3339),
	}, {
		"testID": 3,
		"desc":   "this is a test with custom metadata",
		"content": map[string]interface{}{
			"nestLevel": 1,
			"data":      "nested object #1",
			"content": map[string]interface{}{
				"nestLevel": 2,
				"data":      "nested object #2",
				"content": map[string]interface{}{
					"nestLevel": 3,
					"data":      "nested object #3",
					"content": map[string]interface{}{
						"nestLevel": 4,
						"data":      "nested object #4",
						"content": map[string]interface{}{
							"nestLevel": 5,
							"data":      "nested object #5",
						},
					},
				},
			},
		},
		"date": time.Now().Format(time.RFC3339),
	},
}

var testEmptyObjects = []map[string]interface{}{
	nil,
	nil,
	nil,
	nil,
}

func match(want, got interface{}) bool {
	switch value := got.(type) {
	case []field:
		w := want.([]field)
		for idx, f := range value {
			if f.Key != w[idx].Key {
				return false
			}
			if !match(f.Val, w[idx].Val) {
				return false
			}
		}
		return true
	// case field:
	default:
		if value == want {
			return true
		}
	}
	return false
}

func TestMappify(t *testing.T) {
	type test struct {
		desc string
		data Field
		obj  []field
	}

	var tests = []test{
		{
			desc: "simple obj",
			data: map[string]interface{}{
				"data": "object",
			},
			obj: []field{
				{
					Key: "data",
					Val: "object",
				},
			},
		},
		{
			desc: "with map",
			data: map[string]interface{}{
				"data": map[string]interface{}{
					"a": 1,
				},
			},
			obj: []field{
				{
					Key: "data",
					Val: []field{
						{
							Key: "a",
							Val: 1,
						},
					},
				},
			},
		},
		{
			desc: "with Field",
			data: Field{
				"data": Field{
					"a": 1,
				},
			},
			obj: []field{
				{
					Key: "data",
					Val: []field{
						{
							Key: "a",
							Val: 1,
						},
					},
				},
			},
		},
		{
			desc: "with slice of maps",
			data: map[string]interface{}{
				"data": []map[string]interface{}{
					{"a": 1}, {"b": 2}, {"c": 3},
				},
			},
			obj: []field{
				{
					Key: "data",
					Val: []field{
						{Key: "a", Val: 1},
						{Key: "b", Val: 2},
						{Key: "c", Val: 3},
					},
				},
			},
		},
		{
			desc: "with slice of Fields",
			data: map[string]interface{}{
				"data": []Field{
					{"a": 1}, {"b": 2}, {"c": 3},
				},
			},
			obj: []field{
				{
					Key: "data",
					Val: []field{
						{Key: "a", Val: 1},
						{Key: "b", Val: 2},
						{Key: "c", Val: 3},
					},
				},
			},
		},
	}

	var verify = func(id int, test test, fields []field) {
		if len(fields) != len(test.obj) {
			t.Errorf(
				"#%v -- FAILED --  mappify(map[string]interface{}) []field -- object len %v does not match expected len %v",
				id,
				len(fields),
				len(test.obj),
			)
			return
		}

		for i := 0; i < len(fields); i++ {
			if fields[i].Key != test.obj[i].Key {
				t.Errorf(
					"#%v -- FAILED --  mappify(map[string]interface{}) []field -- object key mismatch: wanted %s ; got %s",
					id,
					test.obj[i].Key,
					fields[i].Key,
				)
				return
			}

			ok := match(fields[i].Val, test.obj[i].Val)
			if !ok {
				t.Errorf(
					"#%v -- FAILED --  mappify(map[string]interface{}) []field -- object value mismatch: wanted %s ; got %s",
					id,
					test.obj[i].Val,
					fields[i].Val,
				)
				return
			}
		}

		t.Logf(
			"#%v -- PASSED --  mappify(map[string]interface{}) []field",
			id,
		)

	}

	for id, test := range tests {
		fields := mappify(test.data)
		verify(id, test, fields)
	}

	// test implementation
	for id, test := range tests {
		fields := test.data.ToXML()
		verify(id, test, fields)
	}
}

func TestMessageBuilder(t *testing.T) {
	type data struct {
		level  LogLevel
		prefix string
		msg    string
		meta   map[string]interface{}
	}

	type test struct {
		input  data
		wants  *LogMessage
		panics bool
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

	for a := 0; a < len(mockLogLevelsOK); a++ {
		if a == 5 {
			continue // skip LLFatal, or os.Exit(1)
		}
		for b := 0; b < len(mockPrefixes); b++ {

			for c := 0; c < len(testAllMessages); c++ {

				for d := 0; d < len(testAllObjects); d++ {
					t := test{
						input: data{
							level:  mockLogLevelsOK[a],
							prefix: mockPrefixes[b],
							msg:    testAllMessages[c],
							meta:   testAllObjects[d],
						},
						wants: &LogMessage{
							Level:    mockLogLevelsOK[a].String(),
							Prefix:   mockPrefixes[b],
							Msg:      testAllMessages[c],
							Metadata: testAllObjects[d],
						},
					}

					if a == 0 {
						t.panics = true
					}

					tests = append(tests, t)
				}

			}
		}
	}
	for a := 0; a < len(mockLogLevelsNOK); a++ {
		if a == 5 {
			continue // skip LLFatal, or os.Exit(1)
		}
		for b := 0; b < len(mockEmptyPrefixes); b++ {

			for c := 0; c < len(testAllMessages); c++ {

				for d := 0; d < len(testAllObjects); d++ {
					t := test{
						input: data{
							level:  mockLogLevelsNOK[a],
							prefix: mockEmptyPrefixes[b],
							msg:    testAllMessages[c],
							meta:   testAllObjects[d],
						},
						wants: &LogMessage{
							Level:    LLInfo.String(),
							Prefix:   "log",
							Msg:      testAllMessages[c],
							Metadata: testAllObjects[d],
						},
					}

					if a == 0 {
						t.panics = true
					}

					tests = append(tests, t)
				}
			}
		}
	}

	var verify = func(id int, test test, logEntry *LogMessage) {
		r := recover()

		if r != nil {
			if test.wants.Level != LLPanic.String() {
				t.Errorf(
					"#%v -- FAILED -- [MessageBuilder] NewMessage().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- unexpected panic: %s",
					id,
					test.input.level.String(),
					test.input.prefix,
					test.input.msg,
					test.input.meta,
					r,
				)
				return
			}

			if r != test.wants.Msg {
				t.Errorf(
					"#%v -- FAILED -- [MessageBuilder] NewMessage().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- panic message doesn't match: %s with input %s",
					id,
					test.input.level.String(),
					test.input.prefix,
					test.input.msg,
					test.input.meta,
					r,
					test.input.msg,
				)
				return
			}
			t.Logf(
				"#%v -- PASSED -- [MessageBuilder] NewMessage().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- %s",
				id,
				test.input.level.String(),
				test.input.prefix,
				test.input.msg,
				test.input.meta,
				mockLogger.buf.String(),
			)
			return
		}

		if logEntry.Level != test.wants.Level {
			t.Errorf(
				"#%v -- FAILED -- [MessageBuilder] NewMessage().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- log level mismatch -- wanted %s, got %s",
				id,
				test.input.level.String(),
				test.input.prefix,
				test.input.msg,
				test.input.meta,
				test.wants.Level,
				test.input.level.String(),
			)
			return
		}

		if logEntry.Prefix != test.wants.Prefix {
			t.Errorf(
				"#%v -- FAILED -- [MessageBuilder] NewMessage().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- prefix mismatch -- wanted %s, got %s",
				id,
				test.input.level.String(),
				test.input.prefix,
				test.input.msg,
				test.input.meta,
				test.wants.Prefix,
				test.input.prefix,
			)
			return
		}

		if logEntry.Msg != test.wants.Msg {
			t.Errorf(
				"#%v -- FAILED -- [MessageBuilder] NewMessage().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- message mismatch -- wanted %s, got %s",
				id,
				test.input.level.String(),
				test.input.prefix,
				test.input.msg,
				test.input.meta,
				test.wants.Msg,
				test.input.msg,
			)
			return
		}

		if logEntry.Metadata == nil && test.wants.Metadata != nil {
			t.Errorf(
				"#%v -- FAILED -- [MessageBuilder] NewMessage().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- retrieved empty metadata object: wanted %s ; got %s",
				id,
				test.input.level.String(),
				test.input.prefix,
				test.input.msg,
				test.input.meta,
				test.wants.Metadata,
				logEntry.Metadata,
			)
			return
		} else if logEntry.Metadata != nil && test.wants.Metadata == nil {
			t.Errorf(
				"#%v -- FAILED -- [MessageBuilder] NewMessage().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- retrieved unexpected metadata object: wanted %s ; got %s",
				id,
				test.input.level.String(),
				test.input.prefix,
				test.input.msg,
				test.input.meta,
				test.wants.Metadata,
				logEntry.Metadata,
			)
			return
		}

		if logEntry.Metadata != nil && test.wants.Metadata != nil {
			for k, v := range logEntry.Metadata {
				if v != nil && test.wants.Metadata[k] == nil {
					t.Errorf(
						"#%v -- FAILED -- [MessageBuilder] NewMessage().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
						id,
						test.input.level.String(),
						test.input.prefix,
						test.input.msg,
						test.input.meta,
						k,
						k,
					)
					return
				}

			}
			if len(logEntry.Metadata) != len(test.wants.Metadata) {
				t.Errorf(
					"#%v -- FAILED -- [MessageBuilder] NewMessage().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- metadata length mismatch -- wanted %v, got %v",
					id,
					test.input.level.String(),
					test.input.prefix,
					test.input.msg,
					test.input.meta,
					len(test.wants.Metadata),
					len(logEntry.Metadata),
				)
				return
			}
		}

		// test passes
		t.Logf(
			"#%v -- PASSED -- [MessageBuilder] NewMessage().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- %s",
			id,
			test.input.level.String(),
			test.input.prefix,
			test.input.msg,
			test.input.meta,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {
		mockLogger.buf.Reset()

		builtMsg := NewMessage().Level(test.input.level).Prefix(test.input.prefix).Message(test.input.msg).Metadata(test.input.meta).Build()

		verify(id, test, builtMsg)

	}

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
			ok:    "info",
			pass:  false,
		},
		{
			input: LogLevel(7),
			ok:    "info",
			pass:  false,
		},
		{
			input: LogLevel(8),
			ok:    "info",
			pass:  false,
		},
		{
			input: LogLevel(10),
			ok:    "info",
			pass:  false,
		},
	}

	var allTests []test
	allTests = append(allTests, passingTests...)
	allTests = append(allTests, failingTests...)

	var verify = func(id int, test test, result string) {
		if test.pass && result == "" {
			t.Errorf(
				"#%v -- FAILED -- [LogLevel] LogLevel(%v).String() -- unexpected reference, got %s",
				id,
				int(test.input),
				result,
			)
			return
		}

		if test.pass && result != test.ok {
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

func TestLogLevelInt(t *testing.T) {
	type test struct {
		input LogLevel
		ok    int
		pass  bool
	}

	var passingTests = []test{
		{
			input: LogLevel(0),
			ok:    0,
			pass:  true,
		}, {
			input: LogLevel(1),
			ok:    1,
			pass:  true,
		}, {
			input: LogLevel(2),
			ok:    2,
			pass:  true,
		}, {
			input: LogLevel(3),
			ok:    3,
			pass:  true,
		}, {
			input: LogLevel(4),
			ok:    4,
			pass:  true,
		}, {
			input: LogLevel(5),
			ok:    5,
			pass:  true,
		}, {
			input: LogLevel(9),
			ok:    9,
			pass:  true,
		},
	}

	var failingTests = []test{
		{
			input: LogLevel(6),
			ok:    6,
			pass:  false,
		},
		{
			input: LogLevel(7),
			ok:    7,
			pass:  false,
		},
		{
			input: LogLevel(8),
			ok:    8,
			pass:  false,
		},
		{
			input: LogLevel(10),
			ok:    10,
			pass:  false,
		},
	}

	var allTests []test
	allTests = append(allTests, passingTests...)
	allTests = append(allTests, failingTests...)

	var verify = func(id, result int, test test) {
		if test.pass && result != test.ok {
			t.Errorf(
				"#%v -- FAILED -- [LogLevel] LogLevel(%v).Int() --  wanted %v, got %v",
				id,
				int(test.input),
				test.ok,
				result,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LogLevel] LogLevel(%v).Int() = %v",
			id,
			int(test.input),
			result,
		)

	}

	for id, test := range allTests {
		result := test.input.Int()

		verify(id, result, test)

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
				ok:        false,
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

		if test.ok && logEntry.Level != test.wantLevel {
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

		if test.ok && logEntry.Msg != test.wantMsg {
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

		logMessage := NewMessage().Level(test.level).Message(test.msg).Build()

		_, err := mockLogger.logger.Output(logMessage)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Output(%s, %s) -- Output func error: %s",
				id,
				test.level.String(),
				test.msg,
				err,
			)
			return
		}

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
		r := recover()

		if test.level == LLPanic {
			if r == nil {
				t.Errorf(
					"#%v -- FAILED -- LoggerMessage] Log(%s, %s) -- panic did not occur",
					id,
					test.level.String(),
					test.msg,
				)
				return
			}

			if r != test.wantMsg {
				t.Errorf(
					"#%v -- FAILED -- LoggerMessage] Log(%s, %s) -- message mismatch: wanted %s ; got %s",
					id,
					test.level.String(),
					test.msg,
					test.wantMsg,
					r,
				)
				return
			}
			t.Logf(
				"#%v -- PASSED -- LoggerMessage] Log(%s, %s) : %s",
				id,
				test.level.String(),
				test.msg,
				mockLogger.buf.String(),
			)

			mockLogger.buf.Reset()
			return
		}

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

		if test.level == LLPanic {
			defer verify(id, test, logEntry)
		}

		logMessage := NewMessage().Level(test.level).Message(test.msg).Build()

		mockLogger.logger.Log(logMessage)

		if test.level != LLPanic {
			verify(id, test, logEntry)
		}
	}

}

func TestLoggerPanic(t *testing.T) {
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

	var verify = func(id int, test test) {
		r := recover()

		if r == nil {
			t.Errorf(
				"#%v -- FAILED -- LoggerMessage] Panic(%s) -- panic did not occur",
				id,
				test.msg,
			)
			return
		}

		if r != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- LoggerMessage] Panic(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				r,
			)
			return
		}
		t.Logf(
			"#%v -- PASSED -- LoggerMessage] Panic(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		mockLogger.buf.Reset()

		defer verify(id, test)

		mockLogger.logger.Panic(test.msg)

	}
}

func TestLoggerPanicln(t *testing.T) {
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

	var verify = func(id int, test test) {
		r := recover()

		if r == nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Panicln(%s) -- panic did not occur",
				id,
				test.msg,
			)
			return
		}

		if r != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Panicln(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				r,
			)
			return
		}
		t.Logf(
			"#%v -- PASSED -- LoggerMessage] Panicln(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()

	}

	for id, test := range tests {

		mockLogger.buf.Reset()

		defer verify(id, test)

		mockLogger.logger.Panicln(test.msg)

	}
}

func TestLoggerPanicf(t *testing.T) {
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

	var verify = func(id int, test test) {
		r := recover()

		if r == nil {
			t.Errorf(
				"#%v -- FAILED -- LoggerMessage] Panicf(%s, %s) -- panic did not occur",
				id,
				test.format,
				test.v,
			)
			return
		}

		if r != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- LoggerMessage] Panicf(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				r,
			)
			return
		}
		t.Logf(
			"#%v -- PASSED -- LoggerMessage] Panicf(%s, %s) : %s",
			id,
			test.format,
			test.v,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()

	}

	for id, test := range tests {

		mockLogger.buf.Reset()

		defer verify(id, test)

		mockLogger.logger.Panicf(test.format, test.v...)

	}
}

func TestLoggerFatal(t *testing.T) {
	var noExitLogger = struct {
		logger LoggerI
		buf    *bytes.Buffer
	}{
		logger: New(
			WithPrefix("test-message"),
			JSONFormat,
			WithOut(mockBuffer),
			SkipExitCfg,
		),
		buf: mockBuffer,
	}

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
		if err := json.Unmarshal(noExitLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Fatal(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != LLFatal.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Fatal(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLFatal.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Fatal(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Fatal(%s) : %s",
			id,
			test.msg,
			noExitLogger.buf.String(),
		)

		noExitLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		noExitLogger.buf.Reset()

		noExitLogger.logger.Fatal(test.msg)

		verify(id, test, logEntry)
	}
}

func TestLoggerFatalln(t *testing.T) {
	var noExitLogger = struct {
		logger LoggerI
		buf    *bytes.Buffer
	}{
		logger: New(
			WithPrefix("test-message"),
			JSONFormat,
			WithOut(mockBuffer),
			SkipExitCfg,
		),
		buf: mockBuffer,
	}

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
		if err := json.Unmarshal(noExitLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Fatalln(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != LLFatal.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Fatalln(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLFatal.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Fatalln(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Fatalln(%s) : %s",
			id,
			test.msg,
			noExitLogger.buf.String(),
		)

		noExitLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		noExitLogger.buf.Reset()

		noExitLogger.logger.Fatalln(test.msg)

		verify(id, test, logEntry)
	}
}

func TestLoggerFatalf(t *testing.T) {
	var noExitLogger = struct {
		logger LoggerI
		buf    *bytes.Buffer
	}{
		logger: New(
			WithPrefix("test-message"),
			JSONFormat,
			WithOut(mockBuffer),
			SkipExitCfg,
		),
		buf: mockBuffer,
	}

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
		if err := json.Unmarshal(noExitLogger.buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Fatalf(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Level != LLFatal.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Fatalf(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				LLFatal.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Fatalf(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Fatalf(%s, %s) : %s",
			id,
			test.format,
			test.v,
			noExitLogger.buf.String(),
		)

		noExitLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		noExitLogger.buf.Reset()

		noExitLogger.logger.Fatalf(test.format, test.v...)

		verify(id, test, logEntry)
	}
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
				"#%v -- FAILED -- [LoggerMessage] Trace(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != LLTrace.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Trace(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLTrace.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Trace(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Trace(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Trace(test.msg)

		verify(id, test, logEntry)
	}
}

func TestLoggerTraceln(t *testing.T) {
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
				"#%v -- FAILED -- [LoggerMessage] Traceln(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Level != LLTrace.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Traceln(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLTrace.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Traceln(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Traceln(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	for id, test := range tests {

		logEntry := &LogMessage{}
		mockLogger.buf.Reset()

		mockLogger.logger.Traceln(test.msg)

		verify(id, test, logEntry)
	}
}

func TestLoggerTracef(t *testing.T) {
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
				"#%v -- FAILED -- [LoggerMessage] Tracef(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Level != LLTrace.String() {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Tracef(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				LLTrace.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [LoggerMessage] Tracef(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [LoggerMessage] Tracef(%s, %s) : %s",
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

		mockLogger.logger.Tracef(test.format, test.v...)

		verify(id, test, logEntry)
	}
}

func TestPrint(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

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

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Print(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Print(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLInfo.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Print(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLInfo.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Print(%s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.msg,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Print(%s) : %s",
			id,
			test.msg,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Print(test.msg)

		verify(id, test, buf.Bytes())

	}

	std = oldstd

}

func TestPrintln(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

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

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Println(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Println(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLInfo.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Println(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLInfo.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Println(%s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.msg,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Println(%s) : %s",
			id,
			test.msg,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Println(test.msg)

		verify(id, test, buf.Bytes())

	}

	std = oldstd

}

func TestPrintf(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

	type test struct {
		format  string
		v       []interface{}
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		tests = append(tests, test{
			format:  "%s",
			v:       []interface{}{mockMessages[a]},
			wantMsg: mockMessages[a],
			ok:      true,
		})
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		tests = append(tests, test{
			format:  mockFmtMessages[b].format,
			v:       mockFmtMessages[b].v,
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		})
	}

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Printf(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Printf(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLInfo.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Printf(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				LLInfo.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Printf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Printf(%s, %s) : %s",
			id,
			test.format,
			test.v,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Printf(test.format, test.v...)

		verify(id, test, buf.Bytes())

	}

	std = oldstd
}

func TestLog(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

	defer func() {
		if r := recover(); r != nil {
			t.Logf(
				"# -- TEST -- [Log()] Intended panic recovery -- %s",
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

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Log(%s, %s) -- unmarshal error: %s",
				id,
				test.level.String(),
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Log(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.level.String(),
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != test.wantLevel {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Log(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.level.String(),
				test.msg,
				test.wantLevel,
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Log(%s, %s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.level.String(),
				test.msg,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Log(%s, %s) : %s",
			id,
			test.level.String(),
			test.msg,
			string(result),
		)

		buf.Reset()

	}

	defer func() { std = oldstd }()

	for id, test := range tests {

		buf.Reset()

		if test.panics {
			defer verify(id, test, buf.Bytes())
		}

		logMessage := NewMessage().Level(test.level).Message(test.msg).Build()

		Log(logMessage)

		if !test.panics {
			verify(id, test, buf.Bytes())
		}

	}

}

func TestPanic(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

	type test struct {
		msg     string
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		test := test{
			msg:     mockMessages[a],
			wantMsg: mockMessages[a],
			ok:      true,
		}

		tests = append(tests, test)
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		test := test{
			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		}

		tests = append(tests, test)
	}

	var verify = func(id int, test test) {
		r := recover()

		if r == nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Panic(%s) -- panic did not occur",
				id,
				test.msg,
			)
			return
		}

		if r != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Panic(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				r,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Panic(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		buf.Reset()

	}

	defer func() { std = oldstd }()

	for id, test := range tests {

		buf.Reset()

		defer verify(id, test)
		Panic(test.msg)

	}
}

func TestPanicln(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

	type test struct {
		msg     string
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		test := test{
			msg:     mockMessages[a],
			wantMsg: mockMessages[a],
			ok:      true,
		}

		tests = append(tests, test)
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		test := test{
			msg:     fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		}

		tests = append(tests, test)
	}

	var verify = func(id int, test test) {
		r := recover()

		if r == nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Panicln(%s) -- panic did not occur",
				id,
				test.msg,
			)
			return
		}

		if r != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Panicln(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				r,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Panicln(%s) : %s",
			id,
			test.msg,
			mockLogger.buf.String(),
		)

		buf.Reset()

	}

	defer func() { std = oldstd }()

	for id, test := range tests {

		buf.Reset()

		defer verify(id, test)
		Panicln(test.msg)

	}
}

func TestPanicf(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

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

	var verify = func(id int, test test) {
		r := recover()

		if r == nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Panicf(%s, %s) -- panic did not occur",
				id,
				test.format,
				test.v,
			)
			return
		}

		if r != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Panicf(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				r,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Panicf(%s, %s) : %s",
			id,
			test.format,
			test.v,
			mockLogger.buf.String(),
		)

		buf.Reset()

	}

	defer func() { std = oldstd }()

	for id, test := range tests {

		buf.Reset()

		defer verify(id, test)
		Panicf(test.format, test.v...)

	}
}

func TestFatal(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
		SkipExitCfg,
	)

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

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Fatal(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Fatal(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLFatal.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Fatal(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLFatal.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Fatal(%s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.msg,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Fatal(%s) : %s",
			id,
			test.msg,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Fatal(test.msg)

		verify(id, test, buf.Bytes())

	}

	std = oldstd

}

func TestFatalln(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
		SkipExitCfg,
	)

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

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Fatalln(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Fatalln(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLFatal.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Fatalln(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLFatal.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Fatalln(%s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.msg,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Fatalln(%s) : %s",
			id,
			test.msg,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Fatalln(test.msg)

		verify(id, test, buf.Bytes())

	}

	std = oldstd

}

func TestFatalf(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
		SkipExitCfg,
	)

	type test struct {
		format  string
		v       []interface{}
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		tests = append(tests, test{
			format:  "%s",
			v:       []interface{}{mockMessages[a]},
			wantMsg: mockMessages[a],
			ok:      true,
		})
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		tests = append(tests, test{
			format:  mockFmtMessages[b].format,
			v:       mockFmtMessages[b].v,
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		})
	}

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Fatalf(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Fatalf(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLFatal.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Fatalf(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				LLFatal.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Fatalf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Fatalf(%s, %s) : %s",
			id,
			test.format,
			test.v,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Fatalf(test.format, test.v...)

		verify(id, test, buf.Bytes())

	}

	std = oldstd
}

func TestError(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

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

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Error(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Error(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLError.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Error(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLError.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Error(%s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.msg,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Error(%s) : %s",
			id,
			test.msg,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Error(test.msg)

		verify(id, test, buf.Bytes())

	}

	std = oldstd

}

func TestErrorln(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

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

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Errorln(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Errorln(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLError.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Errorln(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLError.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Errorln(%s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.msg,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Errorln(%s) : %s",
			id,
			test.msg,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Errorln(test.msg)

		verify(id, test, buf.Bytes())

	}

	std = oldstd

}

func TestErrorf(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

	type test struct {
		format  string
		v       []interface{}
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		tests = append(tests, test{
			format:  "%s",
			v:       []interface{}{mockMessages[a]},
			wantMsg: mockMessages[a],
			ok:      true,
		})
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		tests = append(tests, test{
			format:  mockFmtMessages[b].format,
			v:       mockFmtMessages[b].v,
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		})
	}

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Errorf(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Errorf(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLError.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Errorf(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				LLError.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Errorf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Errorf(%s, %s) : %s",
			id,
			test.format,
			test.v,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Errorf(test.format, test.v...)

		verify(id, test, buf.Bytes())

	}

	std = oldstd
}

func TestWarn(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

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

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Warn(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Warn(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLWarn.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Warn(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLWarn.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Warn(%s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.msg,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Warn(%s) : %s",
			id,
			test.msg,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Warn(test.msg)

		verify(id, test, buf.Bytes())

	}

	std = oldstd

}

func TestWarnln(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

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

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Warnln(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Warnln(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLWarn.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Warnln(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLWarn.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Warnln(%s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.msg,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Warnln(%s) : %s",
			id,
			test.msg,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Warnln(test.msg)

		verify(id, test, buf.Bytes())

	}

	std = oldstd

}

func TestWarnf(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

	type test struct {
		format  string
		v       []interface{}
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		tests = append(tests, test{
			format:  "%s",
			v:       []interface{}{mockMessages[a]},
			wantMsg: mockMessages[a],
			ok:      true,
		})
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		tests = append(tests, test{
			format:  mockFmtMessages[b].format,
			v:       mockFmtMessages[b].v,
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		})
	}

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Warnf(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Warnf(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLWarn.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Warnf(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				LLWarn.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Warnf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Warnf(%s, %s) : %s",
			id,
			test.format,
			test.v,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Warnf(test.format, test.v...)

		verify(id, test, buf.Bytes())

	}

	std = oldstd
}

func TestInfo(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

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

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Info(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Info(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLInfo.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Info(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLInfo.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Info(%s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.msg,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Info(%s) : %s",
			id,
			test.msg,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Info(test.msg)

		verify(id, test, buf.Bytes())

	}

	std = oldstd

}

func TestInfoln(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

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

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Infoln(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Infoln(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLInfo.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Infoln(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLInfo.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Infoln(%s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.msg,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Infoln(%s) : %s",
			id,
			test.msg,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Infoln(test.msg)

		verify(id, test, buf.Bytes())

	}

	std = oldstd

}

func TestInfof(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

	type test struct {
		format  string
		v       []interface{}
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		tests = append(tests, test{
			format:  "%s",
			v:       []interface{}{mockMessages[a]},
			wantMsg: mockMessages[a],
			ok:      true,
		})
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		tests = append(tests, test{
			format:  mockFmtMessages[b].format,
			v:       mockFmtMessages[b].v,
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		})
	}

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Infof(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Infof(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLInfo.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Infof(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				LLInfo.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Infof(%s, %s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Infof(%s, %s) : %s",
			id,
			test.format,
			test.v,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Infof(test.format, test.v...)

		verify(id, test, buf.Bytes())

	}

	std = oldstd
}

func TestDebug(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

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

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Debug(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Debug(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLDebug.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Debug(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLDebug.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Debug(%s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.msg,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Debug(%s) : %s",
			id,
			test.msg,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Debug(test.msg)

		verify(id, test, buf.Bytes())

	}

	std = oldstd

}

func TestDebugln(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

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

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Debugln(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Debugln(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLDebug.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Debugln(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLDebug.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Debugln(%s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.msg,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Debugln(%s) : %s",
			id,
			test.msg,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Debugln(test.msg)

		verify(id, test, buf.Bytes())

	}

	std = oldstd

}

func TestDebugf(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

	type test struct {
		format  string
		v       []interface{}
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		tests = append(tests, test{
			format:  "%s",
			v:       []interface{}{mockMessages[a]},
			wantMsg: mockMessages[a],
			ok:      true,
		})
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		tests = append(tests, test{
			format:  mockFmtMessages[b].format,
			v:       mockFmtMessages[b].v,
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		})
	}

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Debugf(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Debugf(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLDebug.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Debugf(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				LLDebug.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Debugf(%s, %s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Debugf(%s, %s) : %s",
			id,
			test.format,
			test.v,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Debugf(test.format, test.v...)

		verify(id, test, buf.Bytes())

	}

	std = oldstd
}

func TestTrace(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

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

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Trace(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Trace(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLTrace.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Trace(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLTrace.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Trace(%s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.msg,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Trace(%s) : %s",
			id,
			test.msg,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Trace(test.msg)

		verify(id, test, buf.Bytes())

	}

	std = oldstd

}

func TestTraceln(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

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

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Traceln(%s) -- unmarshal error: %s",
				id,
				test.msg,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Traceln(%s) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLTrace.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Traceln(%s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.msg,
				LLTrace.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Traceln(%s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.msg,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Traceln(%s) : %s",
			id,
			test.msg,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Traceln(test.msg)

		verify(id, test, buf.Bytes())

	}

	std = oldstd

}

func TestTracef(t *testing.T) {
	// std logger override
	oldstd := std
	buf := &bytes.Buffer{}
	std = New(
		WithPrefix("log"),
		JSONFormat,
		WithOut(buf),
	)

	type test struct {
		format  string
		v       []interface{}
		wantMsg string
		ok      bool
	}

	var tests []test

	for a := 0; a < len(mockMessages); a++ {
		tests = append(tests, test{
			format:  "%s",
			v:       []interface{}{mockMessages[a]},
			wantMsg: mockMessages[a],
			ok:      true,
		})
	}
	for b := 0; b < len(mockFmtMessages); b++ {
		tests = append(tests, test{
			format:  mockFmtMessages[b].format,
			v:       mockFmtMessages[b].v,
			wantMsg: fmt.Sprintf(mockFmtMessages[b].format, mockFmtMessages[b].v...),
			ok:      true,
		})
	}

	var verify = func(id int, test test, result []byte) {
		logEntry := &LogMessage{}

		if err := json.Unmarshal(result, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Tracef(%s, %s) -- unmarshal error: %s",
				id,
				test.format,
				test.v,
				err,
			)
			return
		}

		if logEntry.Msg != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Tracef(%s, %s) -- message mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				test.wantMsg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != LLTrace.String() {
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Tracef(%s, %s) -- log level mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				LLTrace.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != "log" { // std logger prefix
			t.Errorf(
				"#%v -- FAILED -- [DefaultLogger] Tracef(%s, %s) -- prefix mismatch: wanted %s ; got %s",
				id,
				test.format,
				test.v,
				"log",
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [DefaultLogger] Tracef(%s, %s) : %s",
			id,
			test.format,
			test.v,
			string(result),
		)

		buf.Reset()

	}

	for id, test := range tests {

		buf.Reset()

		Tracef(test.format, test.v...)

		verify(id, test, buf.Bytes())

	}

	std = oldstd
}
