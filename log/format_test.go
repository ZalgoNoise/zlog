package log

import (
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
	}

	var rgxSimpleObjects = []*regexp.Regexp{
		regexp.MustCompile(`\[ ((simple-test = 0)|(passing = true)|(tool = "zlog")) ; ((simple-test = 0)|(passing = true)|(tool = "zlog")) ; ((simple-test = 0)|(passing = true)|(tool = "zlog")) \]`),
		regexp.MustCompile(`\[ simpler-test = "yes" \]`),
		regexp.MustCompile(`\[ ((cascaded-test = true)|(metadata = \[ ((nest-level = 1)|(data = "this is inner-level content")) ; ((nest-level = 1)|(data = "this is inner-level content")) \])) ; ((cascaded-test = true)|(metadata = \[ ((nest-level = 1)|(data = "this is inner-level content")) ; ((nest-level = 1)|(data = "this is inner-level content")) \])) \]`),
	}

	var tests []test

	for a := 0; a < len(testSimpleObjects); a++ {
		obj := test{
			obj: testSimpleObjects[a],
			rgx: rgxSimpleObjects[a],
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
