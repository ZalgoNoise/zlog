package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/zalgonoise/zlog/log/event"
	zbson "github.com/zalgonoise/zlog/log/format/bson"
	"github.com/zalgonoise/zlog/log/format/csv"
	"github.com/zalgonoise/zlog/log/format/gob"
	"github.com/zalgonoise/zlog/log/format/text"
	"go.mongodb.org/mongo-driver/bson"
)

func TestFmtTextFormat(t *testing.T) {
	type test struct {
		msg *event.Event
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
				if mockLogLevelsOK[a] == event.LLFatal || mockLogLevelsOK[a] == event.LLPanic {
					continue
				}

				obj := test{
					msg: event.New().
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
				"#%v -- FAILED -- [TextFormat] Format(*event.Event) -- empty buffer error",
				id,
			)
			return
		}

		if !test.rgx.Match(b) {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*event.Event) -- log message mismatch, expected output to match regex %s -- %s",
				id,
				test.rgx,
				string(b),
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [TextFormat] Format(*event.Event) -- %s",
			id,
			*test.msg,
		)

	}

	for id, test := range tests {
		txt := FormatText

		b, err := txt.Format(test.msg)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*event.Event) -- failed to format message: %s",
				id,
				err,
			)
		}
		verify(id, test, b)
	}

}

func TestFmtTextFmtMetadata(t *testing.T) {

	type mapTest struct {
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
	var mapTests = []mapTest{
		{
			obj: map[string]interface{}{
				"simple-test": 0,
				"passing":     true,
				"tool":        "zlog",
			},
			rgx: regexp.MustCompile(`\[ ((simple-test = 0)|(passing = true)|(tool = "zlog")) ; ((simple-test = 0)|(passing = true)|(tool = "zlog")) ; ((simple-test = 0)|(passing = true)|(tool = "zlog")) \]`),
		},
		{
			obj: map[string]interface{}{
				"simpler-test": "yes",
			},
			rgx: regexp.MustCompile(`\[ simpler-test = "yes" \]`),
		},
		{
			obj: map[string]interface{}{
				"cascaded-test": true,
				"metadata": map[string]interface{}{
					"nest-level": 1,
					"data":       "this is inner-level content",
				},
			},
			rgx: regexp.MustCompile(`\[ ((cascaded-test = true)|(metadata = \[ ((nest-level = 1)|(data = "this is inner-level content")) ; ((nest-level = 1)|(data = "this is inner-level content")) \])) ; ((cascaded-test = true)|(metadata = \[ ((nest-level = 1)|(data = "this is inner-level content")) ; ((nest-level = 1)|(data = "this is inner-level content")) \])) \]`),
		},
		{
			obj: map[string]interface{}{
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
			rgx: regexp.MustCompile(`\[ ((objList = \[ ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) ; ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) ; ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) ; ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) \])|(small = \[ ((\[ a = 1 \])|(\[ b = 2 \])|(\[ c = 3 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])|(\[ c = 3 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])|(\[ c = 3 \])) \])) ; ((objList = \[ ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) ; ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) ; ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) ; ((\[ test = true \])|(\[ another = true \])|(\[ third = "yes" \])|(\[ fourth = "ok" \])) \])|(small = \[ ((\[ a = 1 \])|(\[ b = 2 \])|(\[ c = 3 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])|(\[ c = 3 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])|(\[ c = 3 \])) \])) \]`),
		},
		{
			obj: map[string]interface{}{
				"values": map[string]interface{}{
					"a": 1,
					"b": 2,
					"c": 3,
				},
			},
			rgx: regexp.MustCompile(`\[ values = \[ ((a = 1)|(b = 2)|(c = 3)) ; ((a = 1)|(b = 2)|(c = 3)) ; ((a = 1)|(b = 2)|(c = 3)) \] \]`),
		},
		{
			obj: map[string]interface{}{
				"a-map": map[string]interface{}{
					"a": 1,
				},
				"b-map": map[string]interface{}{
					"b": 2,
				},
			},
			rgx: regexp.MustCompile(`\[ ((a-map = \[ a = 1 \])|(b-map = \[ b = 2 \])) ; ((a-map = \[ a = 1 \])|(b-map = \[ b = 2 \])) \]`),
		},
		{
			obj: map[string]interface{}{
				"a": "one",
				"b": "two",
				"c": "three",
				"d": "four",
			},
			rgx: regexp.MustCompile(`\[ ((a = "one")|(b = "two")|(c = "three")|(d = "four")) ; ((a = "one")|(b = "two")|(c = "three")|(d = "four")) ; ((a = "one")|(b = "two")|(c = "three")|(d = "four")) ; ((a = "one")|(b = "two")|(c = "three")|(d = "four")) \]`),
		},
		{
			obj: map[string]interface{}{},
			rgx: regexp.MustCompile(``),
		},
	}

	type fieldTest struct {
		obj event.Field
		rgx *regexp.Regexp
	}

	// [ a-map = [ b = 2 ; a = 1 ] ; b-map = [ a = 1 ; b = 2 ] ]
	// [ objList = [ [ a = 1 ] ; [ b = 2 ] ] ; same = [ [ a = 1 ] ; [ b = 2 ] ] ]
	var fieldTests = []fieldTest{
		{
			obj: event.Field{
				"a-map": event.Field{
					"a": 1,
					"b": 2,
				},
				"b-map": event.Field{
					"a": 1,
					"b": 2,
				},
			},
			rgx: regexp.MustCompile(`\[ ((a-map = \[ ((a = 1)|(b = 2)) ; ((a = 1)|(b = 2)) \])|(b-map = \[ ((a = 1)|(b = 2)) ; ((a = 1)|(b = 2)) \])) ; ((a-map = \[ ((a = 1)|(b = 2)) ; ((a = 1)|(b = 2)) \])|(b-map = \[ ((a = 1)|(b = 2)) ; ((a = 1)|(b = 2)) \])) \]`),
		},
		{
			obj: event.Field{
				"objList": []event.Field{
					{
						"a": 1,
					},
					{
						"b": 2,
					},
				},
				"same": []event.Field{
					{
						"a": 1,
					},
					{
						"b": 2,
					},
				},
			},
			rgx: regexp.MustCompile(`\[ ((objList = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])|(same = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])) ; ((objList = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])|(same = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])) \]`),
		},
	}

	var verify = func(id int, rgx *regexp.Regexp, result string) {
		if !rgx.MatchString(result) {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] fmtMetadata(map[string]interface{}) -- log message mismatch, expected output to match regex %s -- %s",
				id,
				rgx,
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

	for id, test := range mapTests {
		txt := &text.FmtText{}

		result := txt.FmtMetadata(test.obj)

		verify(id, test.rgx, result)
	}

	for id, test := range fieldTests {
		txt := &text.FmtText{}

		result := txt.FmtMetadata(test.obj)

		verify(id, test.rgx, result)
	}

}

func TestJSONFmtFormat(t *testing.T) {
	type test struct {
		msg *event.Event
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
				if mockLogLevelsOK[a] == event.LLFatal || mockLogLevelsOK[a] == event.LLPanic {
					continue
				}

				obj := test{
					msg: event.New().
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
				"#%v -- FAILED -- [TextFormat] Format(*event.Event) -- empty buffer error",
				id,
			)
			return
		}

		logEntry := &event.Event{}

		if err := json.Unmarshal(b, logEntry); err != nil {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*event.Event) -- unmarshal error: %s",
				id,
				err,
			)
			return
		}
		if logEntry.Msg != test.msg.Msg {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*event.Event) -- message mismatch: wanted %s ; got %s",
				id,
				test.msg,
				logEntry.Msg,
			)
			return
		}

		if logEntry.Level != test.msg.Level {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*event.Event) -- log level mismatch: wanted %s ; got %s",
				id,
				event.LLInfo.String(),
				logEntry.Level,
			)
			return
		}

		if logEntry.Prefix != test.msg.Prefix {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*event.Event) -- log prefix mismatch: wanted %s ; got %s",
				id,
				test.msg.Prefix,
				logEntry.Prefix,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [TextFormat] Format(*event.Event) -- %s",
			id,
			*test.msg,
		)

	}

	for id, test := range tests {
		jsn := FormatJSON

		b, err := jsn.Format(test.msg)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*event.Event) -- failed to format message: %s",
				id,
				err,
			)
		}
		verify(id, test, b)
	}
}

func TestNewTextFormat(t *testing.T) {

	type test struct {
		desc string
		msg  *event.Event
		fmt  *text.FmtText
		rgx  *regexp.Regexp
	}

	var msg = event.New().Prefix("formatter-tests").Level(event.LLInfo).Message("test content").Build()
	var msgSub = event.New().Prefix("formatter-tests").Sub("fmt").Level(event.LLInfo).Message("test content").Build()
	var msgMeta = event.New().Prefix("formatter-tests").Sub("fmt").Level(event.LLInfo).Message("test content").Metadata(event.Field{"a": 0}).Build()

	tests := []test{
		{
			desc: "default",
			msg:  msg,
			fmt:  text.New().Build(),
			rgx:  regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "time: set RFC3339Nano",
			msg:  msg,
			fmt:  text.New().Time(text.LTRFC3339Nano).Build(),
			rgx:  regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "time: set RFC3339",
			msg:  msg,
			fmt:  text.New().Time(text.LTRFC3339).Build(),
			rgx:  regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\+\d{2}:\d{2}]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "time: set RFC822Z",
			msg:  msg,
			fmt:  text.New().Time(text.LTRFC822Z).Build(),
			rgx:  regexp.MustCompile(`^\[\d{2}\s(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s\d{2}\s\d{2}:\d{2}\s\+\d{4}\]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "time: set RubyDate",
			msg:  msg,
			fmt:  text.New().Time(text.LTRubyDate).Build(),
			rgx:  regexp.MustCompile(`^\[(Mon|Tue|Wed|Thu|Fri|Sat|Sun)\s(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s\d{2}\s\d{2}:\d{2}:\d{2}\s\+\d{4}\s\d{4}\]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "time: set UnixNano",
			msg:  msg,
			fmt:  text.New().Time(text.LTUnixNano).Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "time: set UnixMilli",
			msg:  msg,
			fmt:  text.New().Time(text.LTUnixMilli).Build(),
			rgx:  regexp.MustCompile(`^\[\d{13}\]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "time: set UnixMicro",
			msg:  msg,
			fmt:  text.New().Time(text.LTUnixMicro).Build(),
			rgx:  regexp.MustCompile(`^\[\d{16}\]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "level first",
			msg:  msg,
			fmt:  text.New().Time(text.LTUnixNano).LevelFirst().Build(),
			rgx:  regexp.MustCompile(`^\[info\]\s*\[\d{10}\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "level first double-space",
			msg:  msg,
			fmt:  text.New().Time(text.LTUnixNano).LevelFirst().DoubleSpace().Build(),
			rgx:  regexp.MustCompile(`^\[info\]\s*\[\d{10}\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "no level",
			msg:  msg,
			fmt:  text.New().Time(text.LTUnixNano).NoLevel().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "no level: override level-first",
			msg:  msg,
			fmt:  text.New().Time(text.LTUnixNano).LevelFirst().NoLevel().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "no level: override level-first inverse",
			msg:  msg,
			fmt:  text.New().Time(text.LTUnixNano).NoLevel().LevelFirst().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "no level: override color",
			msg:  msg,
			fmt:  text.New().Time(text.LTUnixNano).Color().NoLevel().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "no level: override color inverse",
			msg:  msg,
			fmt:  text.New().Time(text.LTUnixNano).NoLevel().Color().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "no headers",
			msg:  msg,
			fmt:  text.New().Time(text.LTUnixNano).NoHeaders().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[info\]\s*test content`),
		},
		{
			desc: "no level / no headers: override uppercase",
			msg:  msg,
			fmt:  text.New().Time(text.LTUnixNano).NoHeaders().NoLevel().Upper().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*test content`),
		},
		{
			desc: "double space",
			msg:  msg,
			fmt:  text.New().Time(text.LTUnixNano).DoubleSpace().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[info\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "color",
			msg:  msg,
			fmt:  text.New().Time(text.LTUnixNano).Color().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[(.*)info(.*)\]\s*\[formatter-tests\]\s*test content`),
		},
		{
			desc: "upper",
			msg:  msg,
			fmt:  text.New().Time(text.LTUnixNano).Upper().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[INFO\]\s*\[FORMATTER-TESTS\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- default",
			msg:  msgSub,
			fmt:  text.New().Build(),
			rgx:  regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- time: set RFC3339Nano",
			msg:  msgSub,
			fmt:  text.New().Time(text.LTRFC3339Nano).Build(),
			rgx:  regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- time: set RFC3339",
			msg:  msgSub,
			fmt:  text.New().Time(text.LTRFC3339).Build(),
			rgx:  regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\+\d{2}:\d{2}]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- time: set RFC822Z",
			msg:  msgSub,
			fmt:  text.New().Time(text.LTRFC822Z).Build(),
			rgx:  regexp.MustCompile(`^\[\d{2}\s(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s\d{2}\s\d{2}:\d{2}\s\+\d{4}\]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- time: set RubyDate",
			msg:  msgSub,
			fmt:  text.New().Time(text.LTRubyDate).Build(),
			rgx:  regexp.MustCompile(`^\[(Mon|Tue|Wed|Thu|Fri|Sat|Sun)\s(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s\d{2}\s\d{2}:\d{2}:\d{2}\s\+\d{4}\s\d{4}\]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- time: set UnixNano",
			msg:  msgSub,
			fmt:  text.New().Time(text.LTUnixNano).Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- time: set UnixMilli",
			msg:  msgSub,
			fmt:  text.New().Time(text.LTUnixMilli).Build(),
			rgx:  regexp.MustCompile(`^\[\d{13}\]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- time: set UnixMicro",
			msg:  msgSub,
			fmt:  text.New().Time(text.LTUnixMicro).Build(),
			rgx:  regexp.MustCompile(`^\[\d{16}\]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- level first",
			msg:  msgSub,
			fmt:  text.New().Time(text.LTUnixNano).LevelFirst().Build(),
			rgx:  regexp.MustCompile(`^\[info\]\s*\[\d{10}\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- double space",
			msg:  msgSub,
			fmt:  text.New().Time(text.LTUnixNano).DoubleSpace().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- color",
			msg:  msgSub,
			fmt:  text.New().Time(text.LTUnixNano).Color().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[(.*)info(.*)\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content`),
		},
		{
			desc: "w/sub-prefix -- upper",
			msg:  msgSub,
			fmt:  text.New().Time(text.LTUnixNano).Upper().Build(),
			rgx:  regexp.MustCompile(`^\[\d{10}\]\s*\[INFO\]\s*\[FORMATTER-TESTS\]\s*\[FMT\]\s*test content`),
		},
		{
			desc: "w/sub-prefix + metadata",
			msg:  msgMeta,
			fmt:  text.New().Build(),
			rgx:  regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content\s*\[ a = 0 \]`),
		},
		{
			desc: "w/sub-prefix + metadata + double-spaced",
			msg:  msgMeta,
			fmt:  text.New().DoubleSpace().Build(),
			rgx:  regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}]\s*\[info\]\s*\[formatter-tests\]\s*\[fmt\]\s*test content\s*\[ a = 0 \]`),
		},
	}

	var verify = func(id int, test test, buf []byte) {
		if !test.rgx.MatchString(string(buf)) {
			t.Errorf(
				"#%v -- FAILED -- [text.New.Build()] Format(*event.Event) -- %s -- mismatch: wanted %s ; got %s",
				id,
				test.desc,
				test.rgx,
				buf,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [text.New.Build()] Format(*event.Event) -- %s -- %s",
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
					"#%v -- FAILED -- [text.New.Build()] Format(*event.Event) -- failed to format message: %s",
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
		txt := New(WithOut(buf), WithFormat(test.fmt))
		txt.Log(test.msg)
		verify(id, test, buf.Bytes())
	}

}

func TestNewCSVFormat(t *testing.T) {
	module := "Format"
	funcname := "csv.New()"

	type test struct {
		name     string
		unixTime bool
		jsonMeta bool
	}

	var tests = []test{
		{
			name: "default object",
		},
		{
			name:     "set unixTime",
			unixTime: true,
		},
		{
			name:     "set jsonMeta",
			jsonMeta: true,
		},
		{
			name:     "set both options",
			unixTime: true,
			jsonMeta: true,
		},
	}

	var verify = func(id int, test test, fmt *csv.FmtCSV) {
		if fmt.UnixTime != test.unixTime {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- unixTime value mismatch: wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				test.unixTime,
				fmt.UnixTime,
				test.name,
			)
			return
		}
		if fmt.JsonMeta != test.jsonMeta {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- unixTime value mismatch: wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				test.jsonMeta,
				fmt.JsonMeta,
				test.name,
			)
			return
		}
	}

	for id, test := range tests {
		fmt := csv.New()

		if test.unixTime {
			fmt.Unix()
		}

		if test.jsonMeta {
			fmt.JSON()
		}

		csv := fmt.Build()

		verify(id, test, csv)
	}
}

func TestCSVFmtFormat(t *testing.T) {
	module := "FormatCSV"
	funcname := "Format()"

	type test struct {
		name string
		fmt  *csv.FmtCSV
		msg  *event.Event
		rgx  *regexp.Regexp
	}

	var tests = []test{
		{
			name: "default fmt -- simple, trace, no sub",
			fmt:  csv.New().Build(),
			msg:  event.New().Level(event.LLTrace).Prefix("one").Message("two").Metadata(event.Field{"a": 1}).Build(),
			rgx:  regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+\+\d{2}:\d{2},trace,one,,two,\[ a = 1 \]`),
		},
		{
			name: "default fmt -- simple, trace, w/ sub",
			fmt:  csv.New().Build(),
			msg:  event.New().Level(event.LLTrace).Prefix("one").Sub("two").Message("three").Metadata(event.Field{"a": 1}).Build(),
			rgx:  regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+\+\d{2}:\d{2},trace,one,two,three,\[ a = 1 \]`),
		},
		{
			name: "default fmt -- complex meta, trace, no sub",
			fmt:  csv.New().Build(),
			msg:  event.New().Level(event.LLTrace).Prefix("one").Message("two").Metadata(event.Field{"a": 1, "b": []event.Field{{"a": 1}, {"b": 2}}}).Build(),
			rgx:  regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+\+\d{2}:\d{2},trace,one,,two,\[ ((a = 1)|(b = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])) ; ((a = 1)|(b = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])) \]`),
		},
		{
			name: "default fmt -- complex meta, trace, w/ sub",
			fmt:  csv.New().Build(),
			msg:  event.New().Level(event.LLTrace).Prefix("one").Sub("two").Message("three").Metadata(event.Field{"a": 1, "b": []event.Field{{"a": 1}, {"b": 2}}}).Build(),
			rgx:  regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+\+\d{2}:\d{2},trace,one,two,three,\[ ((a = 1)|(b = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])) ; ((a = 1)|(b = \[ ((\[ a = 1 \])|(\[ b = 2 \])) ; ((\[ a = 1 \])|(\[ b = 2 \])) \])) \]`),
		},
		{
			name: "default fmt -- complex meta strings, trace, no sub",
			fmt:  csv.New().Build(),
			msg:  event.New().Level(event.LLTrace).Prefix("one").Message("two").Metadata(event.Field{"a": "one", "b": []event.Field{{"a": "one"}, {"b": "one"}}}).Build(),
			rgx:  regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+\+\d{2}:\d{2},trace,one,,two,"\[ ((a = ""one"")|(b = \[ ((\[ a = ""one"" \])|(\[ b = ""one"" \])) ; ((\[ a = ""one"" \])|(\[ b = ""one"" \])) \])) ; ((a = ""one"")|(b = \[ ((\[ a = ""one"" \])|(\[ b = ""one"" \])) ; ((\[ a = ""one"" \])|(\[ b = ""one"" \])) \])) \] "`),
		},
		{
			name: "default fmt -- complex meta strings, trace, w/ sub",
			fmt:  csv.New().Build(),
			msg:  event.New().Level(event.LLTrace).Prefix("one").Sub("two").Message("three").Metadata(event.Field{"a": "one", "b": []event.Field{{"a": "one"}, {"b": "one"}}}).Build(),
			rgx:  regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d+\+\d{2}:\d{2},trace,one,two,three,"\[ ((a = ""one"")|(b = \[ ((\[ a = ""one"" \])|(\[ b = ""one"" \])) ; ((\[ a = ""one"" \])|(\[ b = ""one"" \])) \])) ; ((a = ""one"")|(b = \[ ((\[ a = ""one"" \])|(\[ b = ""one"" \])) ; ((\[ a = ""one"" \])|(\[ b = ""one"" \])) \])) \]`),
		},
		{
			name: "unixTime fmt -- complex meta strings, trace, w/ sub",
			fmt:  csv.New().Unix().Build(),
			msg:  event.New().Level(event.LLTrace).Prefix("one").Sub("two").Message("three").Metadata(event.Field{"a": "one", "b": []event.Field{{"a": "one"}, {"b": "one"}}}).Build(),
			rgx:  regexp.MustCompile(`\d{10},trace,one,two,three,"\[ ((a = ""one"")|(b = \[ ((\[ a = ""one"" \])|(\[ b = ""one"" \])) ; ((\[ a = ""one"" \])|(\[ b = ""one"" \])) \])) ; ((a = ""one"")|(b = \[ ((\[ a = ""one"" \])|(\[ b = ""one"" \])) ; ((\[ a = ""one"" \])|(\[ b = ""one"" \])) \])) \]`),
		},
		{
			name: "unixTime+jsonMeta fmt -- complex meta strings, trace, w/ sub",
			fmt:  csv.New().Unix().JSON().Build(),
			msg:  event.New().Level(event.LLTrace).Prefix("one").Sub("two").Message("three").Metadata(event.Field{"a": "one", "b": []event.Field{{"a": "one"}, {"b": "one"}}}).Build(),
			rgx:  regexp.MustCompile(`\d+,trace,one,two,three,\"{\"\"a\"\":\"\"one\"\",\"\"b\"\":\[{\"\"a\"\":\"\"one\"\"},{\"\"b\"\":\"\"one\"\"}\]}\"`),
		},
	}

	var verify = func(id int, test test, b []byte) {
		if len(b) == 0 {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- empty buffer error -- action: %s",
				id,
				module,
				funcname,
				test.name,
			)
			return
		}

		if !test.rgx.Match(b) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- log message mismatch, expected output to match regex %s -- %s -- action: %s",
				id,
				module,
				funcname,
				test.rgx,
				string(b),
				test.name,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [%s] [%s] -- %s",
			id,
			module,
			funcname,
			*test.msg,
		)

	}

	for id, test := range tests {
		csv := test.fmt

		b, err := csv.Format(test.msg)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*event.Event) -- failed to format message: %s",
				id,
				err,
			)
		}
		verify(id, test, b)
	}

	// test logger config implementation
	buf := &bytes.Buffer{}

	for id, test := range tests {
		buf.Reset()
		csv := New(WithOut(buf), WithFormat(test.fmt))
		csv.Log(test.msg)
		verify(id, test, buf.Bytes())
	}

}

func TestXMLFmtFormat(t *testing.T) {
	type test struct {
		msg *event.Event
		rgx *regexp.Regexp
	}

	var tests = []test{
		{
			msg: event.New().Level(event.LLTrace).Prefix("one").Message("two\n").Metadata(event.Field{"a": 1}).Build(),
			rgx: regexp.MustCompile(`<logMessage><timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}<\/timestamp><service>one<\/service><level>trace<\/level><message>two<\/message><metadata><key>a<\/key><value>1<\/value><\/metadata><\/logMessage>`),
		},
		{
			msg: event.New().Level(event.LLTrace).Prefix("one").Sub("two").Message("three").Metadata(event.Field{"a": 1}).Build(),
			rgx: regexp.MustCompile(`<logMessage><timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}<\/timestamp><service>one<\/service><module>two<\/module><level>trace<\/level><message>three<\/message><metadata><key>a<\/key><value>1<\/value><\/metadata><\/logMessage>`),
		},
		{
			msg: event.New().Level(event.LLTrace).Prefix("one").Message("two").Metadata(event.Field{"a": 1, "b": []event.Field{{"a": 1}, {"b": 2}}}).Build(),
			rgx: regexp.MustCompile(`<logMessage><timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}<\/timestamp><service>one<\/service><level>trace<\/level><message>two<\/message>((<metadata><key>b<\/key>((<value><key>a<\/key><value>1<\/value><\/value>)|(<value><key>b<\/key><value>2<\/value><\/value>)){2}<\/metadata>)|(<metadata><key>a<\/key><value>1<\/value><\/metadata>)){2}<\/logMessage>`),
		},
		{
			msg: event.New().Level(event.LLTrace).Prefix("one").Sub("two").Message("three").Metadata(event.Field{"a": 1, "b": []event.Field{{"a": 1}, {"b": 2}}}).Build(),
			rgx: regexp.MustCompile(`<logMessage><timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}<\/timestamp><service>one<\/service><module>two<\/module><level>trace<\/level><message>three<\/message>((<metadata><key>a<\/key><value>1<\/value><\/metadata>)|(<metadata><key>b<\/key>((<value><key>a<\/key><value>1<\/value><\/value>)|(<value><key>b<\/key><value>2<\/value><\/value>)){2}<\/metadata>)){2}<\/logMessage>`),
		},
		{
			msg: event.New().Level(event.LLTrace).Prefix("one").Message("two").Metadata(event.Field{"a": "one", "b": []event.Field{{"a": "one"}, {"b": "one"}}}).Build(),
			rgx: regexp.MustCompile(`<logMessage><timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}<\/timestamp><service>one<\/service><level>trace<\/level><message>two<\/message>((<metadata><key>a<\/key><value>one<\/value><\/metadata>)|(<metadata><key>b<\/key>((<value><key>a<\/key><value>one<\/value><\/value>)|(<value><key>b<\/key><value>one<\/value><\/value>)){2}<\/metadata>)){2}<\/logMessage>`),
		},
		{
			msg: event.New().Level(event.LLTrace).Prefix("one").Sub("two").Message("three").Metadata(event.Field{"a": "one", "b": []event.Field{{"a": "one"}, {"b": "one"}}}).Build(),
			rgx: regexp.MustCompile(`<logMessage><timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+\+\d{2}:\d{2}<\/timestamp><service>one<\/service><module>two<\/module><level>trace<\/level><message>three<\/message>((<metadata><key>b<\/key>((<value><key>a<\/key><value>one<\/value><\/value>)|(<value><key>b<\/key><value>one<\/value><\/value>)){2}<\/metadata>)|(<metadata><key>a<\/key><value>one<\/value><\/metadata>)){2}<\/logMessage>`),
		},
	}

	var verify = func(id int, test test, b []byte) {
		if len(b) == 0 {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*event.Event) -- empty buffer error",
				id,
			)
			return
		}

		if !test.rgx.Match(b) {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*event.Event) -- log message mismatch, expected output to match regex %s -- %s",
				id,
				test.rgx,
				string(b),
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [TextFormat] Format(*event.Event) -- %s",
			id,
			*test.msg,
		)

	}

	for id, test := range tests {
		xml := FormatXML

		b, err := xml.Format(test.msg)
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [TextFormat] Format(*event.Event) -- failed to format message: %s",
				id,
				err,
			)
		}
		verify(id, test, b)
	}

	// test logger config implementation
	buf := &bytes.Buffer{}
	xml := New(WithOut(buf), WithFormat(FormatXML))

	for id, test := range tests {
		buf.Reset()
		xml.Log(test.msg)
		verify(id, test, buf.Bytes())
	}

}

func TestGobFmt(t *testing.T) {
	module := "FormatGob"
	funcname := "Format()"
	type test struct {
		name string
		msg  *event.Event
	}

	var tests = []test{
		{
			name: "simple message",
			msg:  event.New().Message("hello world").Build(),
		},
		{
			name: "complete message w/o metadata",
			msg:  event.New().Level(event.LLWarn).Prefix("prefix").Sub("sub").Message("hello complete world").Build(),
		},
		{
			name: "complete message w/ metadata",
			msg: event.New().Level(event.LLWarn).Prefix("prefix").Sub("sub").Message("hello complex world").Metadata(event.Field{
				"a": true,
				"b": 1,
				"c": "data",
				"d": map[string]interface{}{
					"e": "inner",
					"f": []string{
						"g", "h", "i",
					},
				},
			}).Build(),
		},
	}

	g := &gob.FmtGob{}

	var verifyFormat = func(id int, test test) ([]byte, error) {
		b, err := g.Format(test.msg)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- error when formatting message: %s -- action: %s",
				id,
				module,
				funcname,
				err,
				test.name,
			)
			return nil, err
		}
		return b, nil
	}

	var verify = func(id int, test test, b []byte) {

		if b == nil || len(b) == 0 {
			buf, err := verifyFormat(id, test)
			if err != nil {
				return
			}
			b = buf
		}

		new, err := event.New().FromGob(b)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- error when converting gob to message: %s -- action: %s",
				id,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		fmt.Println(msg, test.msg)

		if new.Time.Unix() != test.msg.Time.Unix() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- message time mismatch -- wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				test.msg.Time,
				new.Time,
				test.name,
			)
			return
		}
		if new.Level != test.msg.Level {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- message level mismatch -- wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				test.msg.Level,
				new.Level,
				test.name,
			)
			return
		}
		if new.Prefix != test.msg.Prefix {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- message prefix mismatch -- wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				test.msg.Prefix,
				new.Prefix,
				test.name,
			)
			return
		}
		if new.Sub != test.msg.Sub {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- message sub-prefix mismatch -- wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				test.msg.Sub,
				new.Sub,
				test.name,
			)
			return
		}
		if new.Msg != test.msg.Msg {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- message body mismatch -- wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				test.msg.Msg,
				new.Msg,
				test.name,
			)
			return
		}

		if len(new.Metadata) != len(test.msg.Metadata) {
			return
		}
		for k := range new.Metadata {
			if _, ok := test.msg.Metadata[k]; !ok {
				return
			}
		}
	}

	var buf = &bytes.Buffer{}
	var logGob = New(WithOut(buf), WithFormat(FormatGob), SkipExit)

	for id, test := range tests {
		verify(id, test, nil)
	}

	for id, test := range tests {
		buf.Reset()
		logGob.Log(test.msg)
		verify(id, test, buf.Bytes())
		buf.Reset()
	}

	// ensure FromGob can fail:
	fake := []byte(`{"this":"is","not":"gob"}`)
	_, err := event.New().FromGob(fake)
	if err == nil {
		t.Errorf(
			"#0 -- FAILED -- [%s] [%s] -- FromGob() call with invalid data didn't result in an error",
			module,
			funcname,
		)
		return
	}
}

func TestBSONFmt(t *testing.T) {
	module := "FormatBSON"
	funcname := "Format()"
	type test struct {
		name string
		msg  *event.Event
	}

	var tests = []test{
		{
			name: "simple message",
			msg:  event.New().Message("hello world").Build(),
		},
		{
			name: "complete message w/o metadata",
			msg:  event.New().Level(event.LLWarn).Prefix("prefix").Sub("sub").Message("hello complete world").Build(),
		},
		{
			name: "complete message w/ metadata",
			msg: event.New().Level(event.LLWarn).Prefix("prefix").Sub("sub").Message("hello complex world").Metadata(event.Field{
				"a": true,
				"b": 1,
				"c": "data",
				"d": map[string]interface{}{
					"e": "inner",
					"f": []string{
						"g", "h", "i",
					},
				},
			}).Build(),
		},
	}

	g := &zbson.FmtBSON{}

	var verifyFormat = func(id int, test test) ([]byte, error) {

		b, err := g.Format(test.msg)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- error when formatting message: %s -- action: %s",
				id,
				module,
				funcname,
				err,
				test.name,
			)
			return nil, err
		}
		return b, nil
	}

	var verify = func(id int, test test, b []byte) {

		if b == nil || len(b) == 0 {
			buf, err := verifyFormat(id, test)
			if err != nil {
				return
			}
			b = buf
		}

		var new = &event.Event{}
		err := bson.Unmarshal(b, new)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- error when converting gob to message: %s -- action: %s",
				id,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		fmt.Println(msg, test.msg)

		if new.Time.Unix() != test.msg.Time.Unix() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- message time mismatch -- wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				test.msg.Time,
				new.Time,
				test.name,
			)
			return
		}
		if new.Level != test.msg.Level {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- message level mismatch -- wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				test.msg.Level,
				new.Level,
				test.name,
			)
			return
		}
		if new.Prefix != test.msg.Prefix {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- message prefix mismatch -- wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				test.msg.Prefix,
				new.Prefix,
				test.name,
			)
			return
		}
		if new.Sub != test.msg.Sub {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- message sub-prefix mismatch -- wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				test.msg.Sub,
				new.Sub,
				test.name,
			)
			return
		}
		if new.Msg != test.msg.Msg {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- message body mismatch -- wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				test.msg.Msg,
				new.Msg,
				test.name,
			)
			return
		}

		if len(new.Metadata) != len(test.msg.Metadata) {
			return
		}
		for k := range new.Metadata {
			if _, ok := test.msg.Metadata[k]; !ok {
				return
			}
		}
	}

	var buf = &bytes.Buffer{}
	var logBSON = New(WithOut(buf), WithFormat(FormatBSON), SkipExit)

	for id, test := range tests {
		verify(id, test, nil)
	}

	for id, test := range tests {
		buf.Reset()
		logBSON.Log(test.msg)
		verify(id, test, buf.Bytes())
		buf.Reset()
	}

	buf.Reset()
	logBSON.Infoln(tests[0].msg.Msg)
	verify(0, tests[0], buf.Bytes())
	buf.Reset()
}
