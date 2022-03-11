package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestTextFmtFormat(t *testing.T) {
	type test struct {
		msg *LogMessage
		rgx *regexp.Regexp
	}

	var testAllMessages []string
	testAllMessages = append(testAllMessages, mockMessages...)
	for _, fmtMsg := range mockFmtMessages {
		testAllMessages = append(testAllMessages, fmt.Sprintf(fmtMsg.format, fmtMsg.v...))
	}

	var tests []test

	for a := 0; a < len(mockLogLevelsOK); a++ {
		for b := 0; b < len(mockPrefixes); b++ {
			for c := 0; c < len(testAllMessages); c++ {

				// skip os.Exit(1) and panic() events
				if mockLogLevelsOK[a] == LLFatal || mockLogLevelsOK[a] == LLPanic {
					continue
				}

				obj := test{
					msg: NewMessage().
						Level(mockLogLevelsOK[a]).
						Prefix(mockPrefixes[b]).
						Message(testAllMessages[c]).
						Build(),
					rgx: regexp.MustCompile(fmt.Sprintf(
						`^\[.*\]\s*\[%s\]\s*\[%s\]\s*%s`,
						mockLogLevelsOK[a].String(),
						mockPrefixes[b],
						strings.Replace(strings.Replace(testAllMessages[c], "[", `\[`, -1), "]", `\]`, -1),
					)),
				}

				tests = append(tests, obj)

			}
		}
	}

	var verify = func(id int, test test, b []byte) {
		if len(b) == 0 {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- empty buffer error",
				id,
			)
			return
		}

		if !test.rgx.Match(b) {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- log message mismatch, expected output to match regex %s -- %s",
				id,
				test.rgx,
				string(b),
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [TextFormat] Format(*LogMessage) -- %s",
			id,
			*test.msg,
		)

	}

	for id, test := range tests {
		txt := TextFormat

		b, err := txt.Format(test.msg)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- failed to format message: %s",
				id,
				err,
			)
		}
		verify(id, test, b)
	}

}

func TestTextFmtFmtMetadata(t *testing.T) {
	type test struct {
		obj map[string]interface{}
		rgx *regexp.Regexp
	}

	// [ simple-test = 0 ; passing = true ; tool = "zlog" ]
	// [ simpler-test = "yes" ]
	// [ cascaded-test = true ; metadata = [ nest-level = 1 ; data = "this is inner-level content" ] ]
	// [ objList = [ [ test = true ] ; [ another = true ] ; [ third = "yes" ] ; [ fourth = "ok" ] ] ; small = [ [ a = 1 ] ; [ b = 2 ] ; [ c = 3 ] ] ]
	// [ values = [ a = 1 ; b = 2 ; c = 3 ] ]
	// [ a-map = [ a = 1 ] ; b-map = [ b = 2 ] ]
	// [ a = "one" ; b = "two" ; c = "three" ; d = "four" ]
	var testSimpleObjects = []map[string]interface{}{
		{
			"simple-test": 0,
			"passing":     true,
			"tool":        "zlog",
		},
		{
			"simpler-test": "yes",
		},
		{
			"cascaded-test": true,
			"metadata": map[string]interface{}{
				"nest-level": 1,
				"data":       "this is inner-level content",
			},
		},
		{
			"objList": []map[string]interface{}{
				{
					"test": true,
				},
				{
					"another": true,
				},
				{
					"third": "yes",
				},
				{
					"fourth": "ok",
				},
			},
			"small": []map[string]interface{}{
				{"a": 1}, {"b": 2}, {"c": 3},
			},
		},
		{
			"values": map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
			},
		},
		{
			"a-map": map[string]interface{}{
				"a": 1,
			},
			"b-map": map[string]interface{}{
				"b": 2,
			},
		},
		{
			"a": "one",
			"b": "two",
			"c": "three",
			"d": "four",
		},
	}

	var rgxSimpleObjects = []*regexp.Regexp{
		regexp.MustCompile(`\[ ((simple-test = 0)|(passing = true)|(tool = "zlog")) ; ((simple-test = 0)|(passing = true)|(tool = "zlog")) ; ((simple-test = 0)|(passing = true)|(tool = "zlog")) \]`),
		regexp.MustCompile(`\[ simpler-test = "yes" \]`),
		regexp.MustCompile(`\[ ((cascaded-test = true)|(metadata = \[ ((nest-level = 1)|(data = "this is inner-level content")) ; ((nest-level = 1)|(data = "this is inner-level content")) \])) ; ((cascaded-test = true)|(metadata = \[ ((nest-level = 1)|(data = "this is inner-level content")) ; ((nest-level = 1)|(data = "this is inner-level content")) \])) \]`),
		regexp.MustCompile(`\[ ((objList = \[ ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) ; ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) ; ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) ; ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) \])|(small = \[ ((\[ a = 1 \])|(\[ b = 2 \])|(\[ c = 3 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])|(\[ c = 3 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])|(\[ c = 3 \])) \])) ; ((objList = \[ ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) ; ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) ; ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) ; ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) \])|(small = \[ ((\[ a = 1 \])|(\[ b = 2 \])|(\[ c = 3 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])|(\[ c = 3 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])|(\[ c = 3 \])) \])) \]`),
		regexp.MustCompile(`\[ values = \[ ((a = 1)|(b = 2)|(c = 3)) ; ((a = 1)|(b = 2)|(c = 3)) ; ((a = 1)|(b = 2)|(c = 3)) \] \]`),
		regexp.MustCompile(`\[ ((a-map = \[ a = 1 \])|(b-map = \[ b = 2 \])) ; ((a-map = \[ a = 1 \])|(b-map = \[ b = 2 \])) \]`),
		regexp.MustCompile(`\[ ((a = "one")|(b = "two")|(c = "three")|(d = "four")) ; ((a = "one")|(b = "two")|(c = "three")|(d = "four")) ; ((a = "one")|(b = "two")|(c = "three")|(d = "four")) ; ((a = "one")|(b = "two")|(c = "three")|(d = "four")) \]`),
	}

	// [ a-map = [ b = 2 ; a = 1 ] ; b-map = [ a = 1 ; b = 2 ] ]
	// [ objList = [ [ a = 1 ] ; [ b = 2 ] ] ; same = [ [ a = 1 ] ; [ b = 2 ] ] ]
	var testFieldObjects = []Field{
		{
			"a-map": Field{
				"a": 1,
				"b": 2,
			},
			"b-map": Field{
				"a": 1,
				"b": 2,
			},
		},
		{
			"objList": []Field{
				{
					"a": 1,
				},
				{
					"b": 2,
				},
			},
			"same": []Field{
				{
					"a": 1,
				},
				{
					"b": 2,
				},
			},
		},
	}

	var rgxFieldObjects = []*regexp.Regexp{
		regexp.MustCompile(`\[ ((a-map = \[ ((a = 1)|(b = 2)) ; ((a = 1)|(b = 2)) \])|(b-map = \[ ((a = 1)|(b = 2)) ; ((a = 1)|(b = 2)) \])) ; ((a-map = \[ ((a = 1)|(b = 2)) ; ((a = 1)|(b = 2)) \])|(b-map = \[ ((a = 1)|(b = 2)) ; ((a = 1)|(b = 2)) \])) \]`),
		regexp.MustCompile(`\[ ((objList = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])|(same = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])) ; ((objList = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])|(same = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])) \]`),
	}

	var tests []test

	for a := 0; a < len(testSimpleObjects); a++ {
		obj := test{
			obj: testSimpleObjects[a],
			rgx: rgxSimpleObjects[a],
		}
		tests = append(tests, obj)
	}

	for a := 0; a < len(testFieldObjects); a++ {
		obj := test{
			obj: testFieldObjects[a],
			rgx: rgxFieldObjects[a],
		}
		tests = append(tests, obj)
	}

	var verify = func(id int, test test, result string) {
		if !test.rgx.MatchString(result) {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] fmtMetadata(map[string]interface{}) -- log message mismatch, expected output to match regex %s -- %s",
				id,
				test.rgx,
				result,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [TextFormat] fmtMetadata(map[string]interface{}) -- %s",
			id,
			result,
		)

	}

	for id, test := range tests {
		txt := &TextFmt{}

		result := txt.fmtMetadata(test.obj)

		verify(id, test, result)

	}

}

func TestJSONFmtFormat(t *testing.T) {
	type test struct {
		msg *LogMessage
	}

	var testAllMessages []string
	testAllMessages = append(testAllMessages, mockMessages...)
	for _, fmtMsg := range mockFmtMessages {
		testAllMessages = append(testAllMessages, fmt.Sprintf(fmtMsg.format, fmtMsg.v...))
	}

	var tests []test

	for a := 0; a < len(mockLogLevelsOK); a++ {
		for b := 0; b < len(mockPrefixes); b++ {
			for c := 0; c < len(testAllMessages); c++ {

				// skip os.Exit(1) and panic() events
				if mockLogLevelsOK[a] == LLFatal || mockLogLevelsOK[a] == LLPanic {
					continue
				}

				obj := test{
					msg: NewMessage().
						Level(mockLogLevelsOK[a]).
						Prefix(mockPrefixes[b]).
						Message(testAllMessages[c]).
						Build(),
				}

				tests = append(tests, obj)

			}
		}
	}

	var verify = func(id int, test test, b []byte) {
		if len(b) == 0 {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- empty buffer error",
				id,
			)
			return
		}

		logEntry := &LogMessage{}

		if err := json.Unmarshal(b, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- unmarshal error: %s",
				id,
				err,
			)
			return
		}
		if logEntry.Msg != test.msg.Msg {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != test.msg.Level {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- log level mismatch: wanted %s ; got %s",
				id,
				LLInfo.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != test.msg.Prefix {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- log prefix mismatch: wanted %s ; got %s",
				id,
				test.msg.Prefix,
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [TextFormat] Format(*LogMessage) -- %s",
			id,
			*test.msg,
		)

	}

	for id, test := range tests {
		jsn := JSONFormat

		b, err := jsn.Format(test.msg)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- failed to format message: %s",
				id,
				err,
			)
		}
		verify(id, test, b)
	}
}

func TestNewTextFormatter(t *testing.T) {

	type test struct {
		desc string
		msg  *LogMessage
		fmt  *TextFmt
		rgx  *regexp.Regexp
	}

	var msg = NewMessage().Prefix("formatter-tests").Level(LLInfo).Message("test content").Build()
	var msgSub = NewMessage().Prefix("formatter-tests").Sub("fmt").Level(LLInfo).Message("test content").Build()

	tests := []test{
		{
			desc: "default",
			msg:  msg,
			fmt:  NewTextFormat().Build(),
			rgx:  regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "time: set RFC3339Nano",
			msg:  msg,
			fmt:  NewTextFormat().Time(LTRFC3339Nano).Build(),
			rgx:  regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "time: set RFC3339",
			msg:  msg,
			fmt:  NewTextFormat().Time(LTRFC3339).Build(),
			rgx:  regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\+\d{2}:\d{2}]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "time: set RFC822Z",
			msg:  msg,
			fmt:  NewTextFormat().Time(LTRFC822Z).Build(),
			rgx:  regexp.MustCompile(`^\[\d{2}\s(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s\d{2}\s\d{2}:\d{2}\s\+\d{4}\]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "time: set RubyDate",
			msg:  msg,
			fmt:  NewTextFormat().Time(LTRubyDate).Build(),
			rgx:  regexp.MustCompile(`^\[(Mon|Tue|Wed|Thu|Fri|Sat|Sun)\s(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s\d{2}\s\d{2}:\d{2}:\d{2}\s\+\d{4}\s\d{4}\]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "time: set UnixNano",
			msg:  msg,
			fmt:  NewTextFormat().Time(LTUnixNano).Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "time: set UnixMilli",
			msg:  msg,
			fmt:  NewTextFormat().Time(LTUnixMilli).Build(),
			rgx:  regexp.MustCompile(`^\[\d{13}\]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "time: set UnixMicro",
			msg:  msg,
			fmt:  NewTextFormat().Time(LTUnixMicro).Build(),
			rgx:  regexp.MustCompile(`^\[\d{16}\]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "level first",
			msg:  msg,
			fmt:  NewTextFormat().Time(LTUnixNano).LevelFirst().Build(),
			rgx:  regexp.MustCompile(`^\[info\]\s*\[\d{10}\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "double space",
			msg:  msg,
			fmt:  NewTextFormat().Time(LTUnixNano).DoubleSpace().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "color",
			msg:  msg,
			fmt:  NewTextFormat().Time(LTUnixNano).Color().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[(.*)info(.*)\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "upper",
			msg:  msg,
			fmt:  NewTextFormat().Time(LTUnixNano).Upper().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[INFO\]\s*\[FORMATTER-TESTS\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- default",
			msg:  msgSub,
			fmt:  NewTextFormat().Build(),
			rgx:  regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- time: set RFC3339Nano",
			msg:  msgSub,
			fmt:  NewTextFormat().Time(LTRFC3339Nano).Build(),
			rgx:  regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- time: set RFC3339",
			msg:  msgSub,
			fmt:  NewTextFormat().Time(LTRFC3339).Build(),
			rgx:  regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\+\d{2}:\d{2}]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- time: set RFC822Z",
			msg:  msgSub,
			fmt:  NewTextFormat().Time(LTRFC822Z).Build(),
			rgx:  regexp.MustCompile(`^\[\d{2}\s(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s\d{2}\s\d{2}:\d{2}\s\+\d{4}\]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- time: set RubyDate",
			msg:  msgSub,
			fmt:  NewTextFormat().Time(LTRubyDate).Build(),
			rgx:  regexp.MustCompile(`^\[(Mon|Tue|Wed|Thu|Fri|Sat|Sun)\s(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s\d{2}\s\d{2}:\d{2}:\d{2}\s\+\d{4}\s\d{4}\]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- time: set UnixNano",
			msg:  msgSub,
			fmt:  NewTextFormat().Time(LTUnixNano).Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- time: set UnixMilli",
			msg:  msgSub,
			fmt:  NewTextFormat().Time(LTUnixMilli).Build(),
			rgx:  regexp.MustCompile(`^\[\d{13}\]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- time: set UnixMicro",
			msg:  msgSub,
			fmt:  NewTextFormat().Time(LTUnixMicro).Build(),
			rgx:  regexp.MustCompile(`^\[\d{16}\]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- level first",
			msg:  msgSub,
			fmt:  NewTextFormat().Time(LTUnixNano).LevelFirst().Build(),
			rgx:  regexp.MustCompile(`^\[info\]\s*\[\d{10}\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- double space",
			msg:  msgSub,
			fmt:  NewTextFormat().Time(LTUnixNano).DoubleSpace().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- color",
			msg:  msgSub,
			fmt:  NewTextFormat().Time(LTUnixNano).Color().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[(.*)info(.*)\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- upper",
			msg:  msgSub,
			fmt:  NewTextFormat().Time(LTUnixNano).Upper().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[INFO\]\s*\[FORMATTER-TESTS\]\s*\[FMT\]\s*test content`),
		},
	}

	var verify = func(id int, test test, buf []byte) {
		if !test.rgx.MatchString(string(buf)) {
			t.Errorf(
				"#%v -- FAILED -- [NewTextFormat.Build()] Format(*LogMessage) -- %s -- mismatch: wanted %s ; got %s",
				id,
				test.desc,
				test.rgx,
				buf,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [NewTextFormat.Build()] Format(*LogMessage) -- %s -- %s",
			id,
			test.desc,
			string(buf),
		)
	}

	// run same tests at least 10x so that all random mapping occurrences are
	// verified (because of separators and square brackets)
	for i := 0; i < 10; i++ {
		for id, test := range tests {
			buf, err := test.fmt.Format(test.msg)

			if err != nil {
				t.Errorf(
					"#%v -- FAILED -- [NewTextFormat.Build()] Format(*LogMessage) -- failed to format message: %s",
					id,
					err,
				)
				break
			}
			verify(id, test, buf)
		}
	}

	// test logger config implementation
	buf := &bytes.Buffer{}

	for id, test := range tests {
		buf.Reset()
		txt := New(WithOut(buf), test.fmt)
		txt.Log(test.msg)
		verify(id, test, buf.Bytes())
	}

}

func TestCSVFmtFormat(t *testing.T) {
	type test struct {
		msg *LogMessage
		rgx *regexp.Regexp
	}

	var tests = []test{
		{
			msg: NewMessage().Level(LLTrace).Prefix("one").Message("two").Metadata(Field{"a": 1}).Build(),
			rgx: regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+\+\d{2}:\d{2},trace,one,two,\[ a = 1 \]`),
		},
		{
			msg: NewMessage().Level(LLTrace).Prefix("one").Sub("two").Message("three").Metadata(Field{"a": 1}).Build(),
			rgx: regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+\+\d{2}:\d{2},trace,one,two,three,\[ a = 1 \]`),
		},
		{
			msg: NewMessage().Level(LLTrace).Prefix("one").Message("two").Metadata(Field{"a": 1, "b": []Field{{"a": 1}, {"b": 2}}}).Build(),
			rgx: regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+\+\d{2}:\d{2},trace,one,two,\[ ((a = 1)|(b = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])) ; ((a = 1)|(b = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])) \]`),
		},
		{
			msg: NewMessage().Level(LLTrace).Prefix("one").Sub("two").Message("three").Metadata(Field{"a": 1, "b": []Field{{"a": 1}, {"b": 2}}}).Build(),
			rgx: regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+\+\d{2}:\d{2},trace,one,two,three,\[ ((a = 1)|(b = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])) ; ((a = 1)|(b = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])) \]`),
		},
		{
			msg: NewMessage().Level(LLTrace).Prefix("one").Message("two").Metadata(Field{"a": "one", "b": []Field{{"a": "one"}, {"b": "one"}}}).Build(),
			rgx: regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+\+\d{2}:\d{2},trace,one,two,"\[ ((a = ""one"")|(b = \[ ((\[ a = ""one"" \])|(\[ b = ""one"" \])) ; ((\[ a = ""one"" \])|(\[ b = ""one"" \])) \])) ; ((a = ""one"")|(b = \[ ((\[ a = ""one"" \])|(\[ b = ""one"" \])) ; ((\[ a = ""one"" \])|(\[ b = ""one"" \])) \])) \] "`),
		},
		{
			msg: NewMessage().Level(LLTrace).Prefix("one").Sub("two").Message("three").Metadata(Field{"a": "one", "b": []Field{{"a": "one"}, {"b": "one"}}}).Build(),
			rgx: regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+\+\d{2}:\d{2},trace,one,two,three,"\[ ((a = ""one"")|(b = \[ ((\[ a = ""one"" \])|(\[ b = ""one"" \])) ; ((\[ a = ""one"" \])|(\[ b = ""one"" \])) \])) ; ((a = ""one"")|(b = \[ ((\[ a = ""one"" \])|(\[ b = ""one"" \])) ; ((\[ a = ""one"" \])|(\[ b = ""one"" \])) \])) \]`),
		},
	}

	var verify = func(id int, test test, b []byte) {
		if len(b) == 0 {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- empty buffer error",
				id,
			)
			return
		}

		if !test.rgx.Match(b) {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- log message mismatch, expected output to match regex %s -- %s",
				id,
				test.rgx,
				string(b),
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [TextFormat] Format(*LogMessage) -- %s",
			id,
			*test.msg,
		)

	}

	for id, test := range tests {
		csv := CSVFormat

		b, err := csv.Format(test.msg)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- failed to format message: %s",
				id,
				err,
			)
		}
		verify(id, test, b)
	}

	// test logger config implementation
	buf := &bytes.Buffer{}
	csv := New(WithOut(buf), CSVFormat)

	for id, test := range tests {
		buf.Reset()
		csv.Log(test.msg)
		verify(id, test, buf.Bytes())
	}

}

func TestXMLFmtFormat(t *testing.T) {
	type test struct {
		msg *LogMessage
		rgx *regexp.Regexp
	}

	var tests = []test{
		{
			msg: NewMessage().Level(LLTrace).Prefix("one").Message("two\n").Metadata(Field{"a": 1}).Build(),
			rgx: regexp.MustCompile(`<logMessage><timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}<\/timestamp><service>one<\/service><level>trace<\/level><message>two<\/message><metadata><key>a<\/key><value>1<\/value><\/metadata><\/logMessage>`),
		},
		{
			msg: NewMessage().Level(LLTrace).Prefix("one").Sub("two").Message("three").Metadata(Field{"a": 1}).Build(),
			rgx: regexp.MustCompile(`<logMessage><timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}<\/timestamp><service>one<\/service><module>two<\/module><level>trace<\/level><message>three<\/message><metadata><key>a<\/key><value>1<\/value><\/metadata><\/logMessage>`),
		},
		{
			msg: NewMessage().Level(LLTrace).Prefix("one").Message("two").Metadata(Field{"a": 1, "b": []Field{{"a": 1}, {"b": 2}}}).Build(),
			rgx: regexp.MustCompile(`<logMessage><timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}<\/timestamp><service>one<\/service><level>trace<\/level><message>two<\/message>((<metadata><key>b<\/key>((<value><key>a<\/key><value>1<\/value><\/value>)|(<value><key>b<\/key><value>2<\/value><\/value>)){2}<\/metadata>)|(<metadata><key>a<\/key><value>1<\/value><\/metadata>)){2}<\/logMessage>`),
		},
		{
			msg: NewMessage().Level(LLTrace).Prefix("one").Sub("two").Message("three").Metadata(Field{"a": 1, "b": []Field{{"a": 1}, {"b": 2}}}).Build(),
			rgx: regexp.MustCompile(`<logMessage><timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}<\/timestamp><service>one<\/service><module>two<\/module><level>trace<\/level><message>three<\/message>((<metadata><key>a<\/key><value>1<\/value><\/metadata>)|(<metadata><key>b<\/key>((<value><key>a<\/key><value>1<\/value><\/value>)|(<value><key>b<\/key><value>2<\/value><\/value>)){2}<\/metadata>)){2}<\/logMessage>`),
		},
		{
			msg: NewMessage().Level(LLTrace).Prefix("one").Message("two").Metadata(Field{"a": "one", "b": []Field{{"a": "one"}, {"b": "one"}}}).Build(),
			rgx: regexp.MustCompile(`<logMessage><timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}<\/timestamp><service>one<\/service><level>trace<\/level><message>two<\/message>((<metadata><key>a<\/key><value>one<\/value><\/metadata>)|(<metadata><key>b<\/key>((<value><key>a<\/key><value>one<\/value><\/value>)|(<value><key>b<\/key><value>one<\/value><\/value>)){2}<\/metadata>)){2}<\/logMessage>`),
		},
		{
			msg: NewMessage().Level(LLTrace).Prefix("one").Sub("two").Message("three").Metadata(Field{"a": "one", "b": []Field{{"a": "one"}, {"b": "one"}}}).Build(),
			rgx: regexp.MustCompile(`<logMessage><timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}<\/timestamp><service>one<\/service><module>two<\/module><level>trace<\/level><message>three<\/message>((<metadata><key>b<\/key>((<value><key>a<\/key><value>one<\/value><\/value>)|(<value><key>b<\/key><value>one<\/value><\/value>)){2}<\/metadata>)|(<metadata><key>a<\/key><value>one<\/value><\/metadata>)){2}<\/logMessage>`),
		},
	}

	var verify = func(id int, test test, b []byte) {
		if len(b) == 0 {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- empty buffer error",
				id,
			)
			return
		}

		if !test.rgx.Match(b) {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- log message mismatch, expected output to match regex %s -- %s",
				id,
				test.rgx,
				string(b),
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [TextFormat] Format(*LogMessage) -- %s",
			id,
			*test.msg,
		)

	}

	for id, test := range tests {
		xml := XMLFormat

		b, err := xml.Format(test.msg)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- failed to format message: %s",
				id,
				err,
			)
		}
		verify(id, test, b)
	}

	// test logger config implementation
	buf := &bytes.Buffer{}
	xml := New(WithOut(buf), XMLFormat)

	for id, test := range tests {
		buf.Reset()
		xml.Log(test.msg)
		verify(id, test, buf.Bytes())
	}

}
