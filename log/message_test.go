package log

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/zalgonoise/zlog/log/event"
	pb "github.com/zalgonoise/zlog/proto/message"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMessageBuilder(t *testing.T) {
	type data struct {
		level  event.LogLevel
		prefix string
		msg    string
		meta   map[string]interface{}
	}

	type test struct {
		input  data
		wants  *event.Event
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
			continue // skip event.LLFatal, or os.Exit(1)
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
						wants: &event.Event{
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
			continue // skip event.LLFatal, or os.Exit(1)
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
						wants: &event.Event{
							Level:    event.LLInfo.String(),
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

	var verify = func(id int, test test, msg *event.EventBuilder) {
		r := recover()

		if r != nil {
			if test.wants.Level != event.LLPanic.String() {
				t.Errorf(
					"#%v -- FAILED -- [MessageBuilder] event.New().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- unexpected panic: %s",
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
					"#%v -- FAILED -- [MessageBuilder] event.New().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- panic message doesn't match: %s with input %s",
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
				"#%v -- PASSED -- [MessageBuilder] event.New().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- %s",
				id,
				test.input.level.String(),
				test.input.prefix,
				test.input.msg,
				test.input.meta,
				mockLogger.buf.String(),
			)
			return
		}

		logEntry := msg.Build()

		if logEntry.Level != test.wants.Level {
			t.Errorf(
				"#%v -- FAILED -- [MessageBuilder] event.New().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- log level mismatch -- wanted %s, got %s",
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
				"#%v -- FAILED -- [MessageBuilder] event.New().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- prefix mismatch -- wanted %s, got %s",
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
				"#%v -- FAILED -- [MessageBuilder] event.New().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- message mismatch -- wanted %s, got %s",
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

		if len(logEntry.Metadata) == 0 && len(test.wants.Metadata) > 0 {
			t.Errorf(
				"#%v -- FAILED -- [MessageBuilder] event.New().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- retrieved empty metadata object: wanted %s ; got %s",
				id,
				test.input.level.String(),
				test.input.prefix,
				test.input.msg,
				test.input.meta,
				test.wants.Metadata,
				logEntry.Metadata,
			)
			return
		} else if len(logEntry.Metadata) > 0 && len(test.wants.Metadata) == 0 {
			t.Errorf(
				"#%v -- FAILED -- [MessageBuilder] event.New().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- retrieved unexpected metadata object: wanted %s ; got %s",
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

		if len(logEntry.Metadata) > 0 && len(test.wants.Metadata) > 0 {
			for k, v := range logEntry.Metadata {
				if v != nil && test.wants.Metadata[k] == nil {
					t.Errorf(
						"#%v -- FAILED -- [MessageBuilder] event.New().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
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
					"#%v -- FAILED -- [MessageBuilder] event.New().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- metadata length mismatch -- wanted %v, got %v",
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
			"#%v -- PASSED -- [MessageBuilder] event.New().Level(%s).Prefix(%s).Message(%s).Metadata(%s).Build() Log(msg) -- %s",
			id,
			test.input.level.String(),
			test.input.prefix,
			test.input.msg,
			test.input.meta,
			mockLogger.buf.String(),
		)

		mockLogger.buf.Reset()
	}

	// test metadata appendage
	mockLogger.buf.Reset()
	msg := event.New().
		Prefix("pref").
		Message("hi").
		Metadata(map[string]interface{}{"a": 1}).
		Metadata(event.Field{"b": 2})

	metatest := test{
		input: data{
			level:  event.LLInfo,
			prefix: "pref",
			msg:    "hi",
			meta: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
		},
		wants: &event.Event{
			Level:  event.LLInfo.String(),
			Prefix: "pref",
			Msg:    "hi",
			Metadata: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
		},
		panics: false,
	}

	verify(0, metatest, msg)

	for id, test := range tests {
		mockLogger.buf.Reset()

		msg := event.New().Level(test.input.level).Prefix(test.input.prefix).Message(test.input.msg).Metadata(test.input.meta)

		verify(id, test, msg)

	}

}

func TestMessageBuilderCallStack(t *testing.T) {
	type test struct {
		msg *event.EventBuilder
		all bool
		ok  bool
	}
	var tests = []test{
		{
			msg: event.New().Level(event.LLInfo).Prefix("test").Message("message"),
			all: true,
			ok:  true,
		},
		{
			msg: event.New().Level(event.LLInfo).Prefix("test").Message("message"),
			all: false,
			ok:  true,
		},
		{
			msg: event.New().Level(event.LLInfo).Prefix("test").Message("message"),
			all: false,
			ok:  false,
		},
		{
			msg: event.New().Level(event.LLInfo).Prefix("test").Message("message").Metadata(event.Field{"a": 1}),
			all: true,
			ok:  true,
		},
		{
			msg: event.New().Level(event.LLInfo).Prefix("test").Message("message").Metadata(event.Field{"callstack": 1}),
			all: true,
			ok:  true,
		},
	}

	var verify = func(id int, test test, msg *event.Event) {
		if !test.ok {
			if len(msg.Metadata) > 0 {
				t.Errorf(
					"#%v -- FAILED -- [MessageBuilder] event.New().CallStack().Build() -- callstack present expected otherwise",
					id,
				)
				return
			}
			t.Logf(
				"#%v -- PASSED -- [MessageBuilder] event.New().Build() -- no CallStack() call",
				id,
			)
			return

		}

		if test.ok && (msg.Metadata == nil || len(msg.Metadata) <= 0) {
			t.Errorf(
				"#%v -- FAILED -- [MessageBuilder] event.New().CallStack().Build() -- metadata object is emtpy",
				id,
			)
			return
		}

		v, ok := msg.Metadata["callstack"]

		if ok != test.ok {
			t.Errorf(
				"#%v -- FAILED -- [MessageBuilder] event.New().CallStack().Build() -- callstack absent when expected otherwise",
				id,
			)
			return
		}

		field := v.(map[string]interface{})

		for k, val := range field {
			routine := val.(map[string]interface{})

			if routine["id"] == nil || routine["id"] == "" {
				t.Errorf(
					"#%v -- FAILED -- [MessageBuilder] event.New().CallStack().Build() -- empty ID field in key %s",
					id,
					k,
				)
				return
			}

			if routine["status"] == nil || routine["status"] == "" {
				t.Errorf(
					"#%v -- FAILED -- [MessageBuilder] event.New().CallStack().Build() -- empty status field in key %s",
					id,
					k,
				)
				return
			}

			for idx, s := range routine["stack"].([]map[string]interface{}) {
				if s["method"] == nil || s["method"] == "" {
					t.Errorf(
						"#%v -- FAILED -- [MessageBuilder] event.New().CallStack().Build() -- empty method field in key %s.stack[%v]",
						id,
						k,
						idx,
					)
					return
				}

				if s["reference"] == nil || s["reference"] == "" {
					t.Errorf(
						"#%v -- FAILED -- [MessageBuilder] event.New().CallStack().Build() -- empty reference field in key %s.stack[%v]",
						id,
						k,
						idx,
					)
					return
				}
			}
		}
		t.Logf(
			"#%v -- PASSED -- [MessageBuilder] event.New().CallStack().Build()",
			id,
		)

	}

	for id, test := range tests {
		var msg *event.Event

		if !test.ok {
			msg = test.msg.Build()
		} else {
			msg = test.msg.CallStack(test.all).Build()
		}

		verify(id, test, msg)
	}

}

func TestLogLevelString(t *testing.T) {
	type test struct {
		input event.LogLevel
		ok    string
		pass  bool
	}

	var passingTests []test

	for k, v := range event.LogTypeVals {
		passingTests = append(passingTests, test{
			input: k,
			ok:    v,
			pass:  true,
		})
	}

	var failingTests = []test{
		{
			input: event.LogLevel(6),
			ok:    "info",
			pass:  false,
		},
		{
			input: event.LogLevel(7),
			ok:    "info",
			pass:  false,
		},
		{
			input: event.LogLevel(8),
			ok:    "info",
			pass:  false,
		},
		{
			input: event.LogLevel(10),
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
				"#%v -- FAILED -- [event.LogLevel] event.LogLevel(%v).String() -- unexpected reference, got %s",
				id,
				int(test.input),
				result,
			)
			return
		}

		if test.pass && result != test.ok {
			t.Errorf(
				"#%v -- FAILED -- [event.LogLevel] event.LogLevel(%v).String() -- expected %s, got %s",
				id,
				int(test.input),
				test.ok,
				result,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [event.LogLevel] event.LogLevel(%v).String() = %s",
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
		input event.LogLevel
		ok    int
		pass  bool
	}

	var passingTests = []test{
		{
			input: event.LogLevel(0),
			ok:    0,
			pass:  true,
		}, {
			input: event.LogLevel(1),
			ok:    1,
			pass:  true,
		}, {
			input: event.LogLevel(2),
			ok:    2,
			pass:  true,
		}, {
			input: event.LogLevel(3),
			ok:    3,
			pass:  true,
		}, {
			input: event.LogLevel(4),
			ok:    4,
			pass:  true,
		}, {
			input: event.LogLevel(5),
			ok:    5,
			pass:  true,
		}, {
			input: event.LogLevel(9),
			ok:    9,
			pass:  true,
		},
	}

	var failingTests = []test{
		{
			input: event.LogLevel(6),
			ok:    6,
			pass:  false,
		},
		{
			input: event.LogLevel(7),
			ok:    7,
			pass:  false,
		},
		{
			input: event.LogLevel(8),
			ok:    8,
			pass:  false,
		},
		{
			input: event.LogLevel(10),
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
				"#%v -- FAILED -- [event.LogLevel] event.LogLevel(%v).Int() --  wanted %v, got %v",
				id,
				int(test.input),
				test.ok,
				result,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [event.LogLevel] event.LogLevel(%v).Int() = %v",
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
		level     event.LogLevel
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

	var verify = func(id int, test test, logEntry *event.Event) {
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

		logEntry := &event.Event{}
		mockLogger.buf.Reset()

		logMessage := event.New().Level(test.level).Message(test.msg).Build()

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

func TestMessageToProto(t *testing.T) {
	module := "Message"
	funcname := "ToProto()"

	type test struct {
		name       string
		input      *event.Event
		wantLevel  string
		wantPrefix string
		wantSub    string
		wantMsg    string
		wantMeta   []byte
	}

	var tests = []test{
		{
			name:       "simple message",
			input:      event.New().Message("hello world").Build(),
			wantLevel:  "INFO",
			wantPrefix: "log",
			wantSub:    "",
			wantMsg:    "hello world",
			wantMeta:   []byte("{}"),
		},
		{
			name:       "complete message",
			input:      event.New().Level(event.LLWarn).Prefix("proto").Sub("conv").Message("hello world").Build(),
			wantLevel:  "WARNING",
			wantPrefix: "proto",
			wantSub:    "conv",
			wantMsg:    "hello world",
			wantMeta:   []byte("{}"),
		},
		{
			name:       "complete message w/meta",
			input:      event.New().Level(event.LLWarn).Prefix("proto").Sub("conv").Message("hello world").Metadata(event.Field{"a": 0}).Build(),
			wantLevel:  "WARNING",
			wantPrefix: "proto",
			wantSub:    "conv",
			wantMsg:    "hello world",
			wantMeta:   []byte(`{"a":0}`),
		},
	}

	var verify = func(id int, test test, pb *pb.MessageRequest) {
		if pb.GetLevel().String() != test.wantLevel {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] level mismatch: wanted %s ; got %s",
				id,
				module,
				funcname,
				test.wantLevel,
				pb.GetLevel().String(),
			)
			return
		}

		if pb.GetPrefix() != test.wantPrefix {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] prefix mismatch: wanted %s ; got %s",
				id,
				module,
				funcname,
				test.wantPrefix,
				pb.GetPrefix(),
			)
			return
		}

		if pb.GetSub() != test.wantSub {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] prefix mismatch: wanted %s ; got %s",
				id,
				module,
				funcname,
				test.wantSub,
				pb.GetSub(),
			)
			return
		}

		if pb.GetMsg() != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] message mismatch: wanted %s ; got %s",
				id,
				module,
				funcname,
				test.wantMsg,
				pb.GetMsg(),
			)
			return
		}

		meta, err := pb.GetMeta().MarshalJSON()
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] failed to convert metadata to bytes: %s",
				id,
				module,
				funcname,
				err,
			)
			return
		}

		if !reflect.DeepEqual(meta, test.wantMeta) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] metadata mismatch: wanted %s ; got %s",
				id,
				module,
				funcname,
				string(test.wantMeta),
				string(meta),
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
		proto, err := test.input.ToProto()
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] error during conversion: %s",
				id,
				module,
				funcname,
				err,
			)
			continue
		}

		verify(id, test, proto)
	}

}

func TestMessageProto(t *testing.T) {
	module := "Message"
	funcname := "Proto()"

	type test struct {
		name       string
		input      *event.Event
		wantLevel  string
		wantPrefix string
		wantSub    string
		wantMsg    string
		wantMeta   []byte
	}

	var tests = []test{
		{
			name:       "simple message",
			input:      event.New().Message("hello world").Build(),
			wantLevel:  "INFO",
			wantPrefix: "log",
			wantSub:    "",
			wantMsg:    "hello world",
			wantMeta:   []byte("{}"),
		},
		{
			name:       "complete message",
			input:      event.New().Level(event.LLWarn).Prefix("proto").Sub("conv").Message("hello world").Build(),
			wantLevel:  "WARNING",
			wantPrefix: "proto",
			wantSub:    "conv",
			wantMsg:    "hello world",
			wantMeta:   []byte("{}"),
		},
		{
			name:       "complete message w/meta",
			input:      event.New().Level(event.LLWarn).Prefix("proto").Sub("conv").Message("hello world").Metadata(event.Field{"a": 0}).Build(),
			wantLevel:  "WARNING",
			wantPrefix: "proto",
			wantSub:    "conv",
			wantMsg:    "hello world",
			wantMeta:   []byte(`{"a":0}`),
		},
	}

	var verify = func(id int, test test, pb *pb.MessageRequest) {
		if pb.GetLevel().String() != test.wantLevel {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] level mismatch: wanted %s ; got %s",
				id,
				module,
				funcname,
				test.wantLevel,
				pb.GetLevel().String(),
			)
			return
		}

		if pb.GetPrefix() != test.wantPrefix {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] prefix mismatch: wanted %s ; got %s",
				id,
				module,
				funcname,
				test.wantPrefix,
				pb.GetPrefix(),
			)
			return
		}

		if pb.GetSub() != test.wantSub {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] prefix mismatch: wanted %s ; got %s",
				id,
				module,
				funcname,
				test.wantSub,
				pb.GetSub(),
			)
			return
		}

		if pb.GetMsg() != test.wantMsg {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] message mismatch: wanted %s ; got %s",
				id,
				module,
				funcname,
				test.wantMsg,
				pb.GetMsg(),
			)
			return
		}

		meta, err := pb.GetMeta().MarshalJSON()
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] failed to convert metadata to bytes: %s",
				id,
				module,
				funcname,
				err,
			)
			return
		}

		if !reflect.DeepEqual(meta, test.wantMeta) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] metadata mismatch: wanted %s ; got %s",
				id,
				module,
				funcname,
				string(test.wantMeta),
				string(meta),
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
		proto := test.input.Proto()

		verify(id, test, proto)
	}

}

func TestMessageFromProto(t *testing.T) {
	module := "Message"
	funcname := "FromProto()"

	type test struct {
		name        string
		want        *event.Event
		inputLevel  int32
		inputPrefix string
		inputSub    string
		inputMsg    string
		inputMeta   []byte
	}

	var tests = []test{
		{
			name:        "simple message",
			want:        event.New().Message("hello world").Build(),
			inputLevel:  2,
			inputPrefix: "log",
			inputSub:    "",
			inputMsg:    "hello world",
			inputMeta:   []byte("{}"),
		},
		{
			name:        "complete message",
			want:        event.New().Level(event.LLWarn).Prefix("proto").Sub("conv").Message("hello world").Build(),
			inputLevel:  3,
			inputPrefix: "proto",
			inputSub:    "conv",
			inputMsg:    "hello world",
			inputMeta:   []byte("{}"),
		},
		{
			name:        "complete message w/meta",
			want:        event.New().Level(event.LLWarn).Prefix("proto").Sub("conv").Message("hello world").Metadata(event.Field{"a": 0}).Build(),
			inputLevel:  3,
			inputPrefix: "proto",
			inputSub:    "conv",
			inputMsg:    "hello world",
			inputMeta:   []byte(`{"a":0}`),
		},
		{
			name:        "all nil values",
			want:        event.New().Message("hi").Build(),
			inputLevel:  -1,
			inputPrefix: "",
			inputSub:    "",
			inputMsg:    "hi",
			inputMeta:   []byte(`{}`),
		},
	}

	var verify = func(id int, test test, msg *event.Event) {

		if msg.Level != test.want.Level {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] level mismatch: wanted %s ; got %s",
				id,
				module,
				funcname,
				test.want.Level,
				msg.Level,
			)
			return
		}

		if msg.Prefix != test.want.Prefix {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] prefix mismatch: wanted %s ; got %s",
				id,
				module,
				funcname,
				test.want.Prefix,
				msg.Prefix,
			)
			return
		}

		if msg.Sub != test.want.Sub {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] prefix mismatch: wanted %s ; got %s",
				id,
				module,
				funcname,
				test.want.Sub,
				msg.Sub,
			)
			return
		}

		if msg.Msg != test.want.Msg {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] message mismatch: wanted %s ; got %s",
				id,
				module,
				funcname,
				test.want.Msg,
				msg.Msg,
			)
			return
		}

		if len(msg.Metadata) == 0 && len(test.want.Metadata) == 0 {
			t.Logf(
				"#%v -- PASSED -- [%s] [%s]",
				id,
				module,
				funcname,
			)
			return
		}

		for k, v := range msg.Metadata {
			if wantV, ok := test.want.Metadata[k]; !ok {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] metadata mismatch: key %s isn't originally set",
					id,
					module,
					funcname,
					k,
				)
				return

			} else if v == nil && wantV != nil {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] metadata mismatch: resulting object's %s value was nil, when it shouldn't be",
					id,
					module,
					funcname,
					k,
				)
				return

			} else if v != nil && wantV == nil {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] metadata mismatch: resulting object's %s value wasn't nil, when it should be",
					id,
					module,
					funcname,
					k,
				)
			}
		}

		t.Logf(
			"#%v -- PASSED -- [%s] [%s]",
			id,
			module,
			funcname,
		)

	}

	for id, test := range tests {
		var proto *pb.MessageRequest

		if test.inputLevel < 0 {
			faketime := timestamppb.Timestamp{
				Seconds: 0,
			}

			proto = &pb.MessageRequest{
				Time:   &faketime,
				Level:  nil,
				Prefix: nil,
				Sub:    nil,
				Msg:    test.inputMsg,
				Meta:   nil,
			}

		} else {
			level := pb.Level(test.inputLevel)

			meta, err := event.EncodeProto(test.inputMeta)
			if err != nil {
				t.Errorf(
					"#%v -- FAILED -- [%s] [%s] error during conversion: %s",
					id,
					module,
					funcname,
					err,
				)
				continue
			}

			proto = &pb.MessageRequest{
				Level:  &level,
				Prefix: &test.inputPrefix,
				Sub:    &test.inputSub,
				Msg:    test.inputMsg,
				Meta:   meta,
			}

		}

		msg := event.New().FromProto(proto).Build()

		verify(id, test, msg)
	}

}

func TestEncodeProtoErr(t *testing.T) {
	module := "Message"
	funcname := "encodeProto()"

	var tests = []struct {
		input []byte
		ok    bool
	}{
		{
			input: []byte(`{}`),
			ok:    true,
		},
		{
			input: []byte(`{"a":0}`),
			ok:    true,
		},
		{
			input: []byte(`{"a":"b"}`),
			ok:    true,
		},
		{
			input: []byte(`{"a":{"a":"b"}}`),
			ok:    true,
		},
		{
			input: []byte(`{"a":"b`),
			ok:    false,
		},
	}

	for id, test := range tests {
		_, err := event.EncodeProto(test.input)
		if err != nil && test.ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected conversion error: %s",
				id,
				module,
				funcname,
				err,
			)
		}
		t.Logf(
			"#%v -- PASSED -- [%s] [%s]",
			id,
			module,
			funcname,
		)
	}
}
