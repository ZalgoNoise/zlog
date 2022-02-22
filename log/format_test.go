package log

import (
	"fmt"
	"regexp"
	"testing"
)

func TestTextFmtFormat(t *testing.T) {
	type test struct {
		msg *LogMessage
		rgx *regexp.Regexp
	}

	// var txtLogger LoggerI = New("genesis", TextFormat, mockBuffer)

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
						`^\[.*\]\s*\[%s\]\s*\[%s\]\s*%s\s*$`,
						mockLogLevelsOK[a].String(),
						mockPrefixes[b],
						testAllMessages[c],
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
		}

		// logEntry := &LogMessage{}

		// if err := json.Unmarshal(b, logEntry); err != nil {
		// 	t.Errorf(
		// 		"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- unmarshal error: %s",
		// 		id,
		// 		err,
		// 	)
		// 	return
		// }
		// if logEntry.Msg != test.msg.Msg {
		// 	t.Errorf(
		// 		"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- message mismatch: wanted %s ; got %s",
		// 		id,
		// 		test.msg,
		// 		logEntry.Msg,
		// 	)
		// 	return
		// }

		// if logEntry.Level != test.msg.Level {
		// 	t.Errorf(
		// 		"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- log level mismatch: wanted %s ; got %s",
		// 		id,
		// 		LLInfo.String(),
		// 		logEntry.Level,
		// 	)
		// 	return
		// }

		// if logEntry.Prefix != test.msg.Prefix {
		// 	t.Errorf(
		// 		"#%v -- FAILED -- [TextFormat] Format(*LogMessage) -- log prefix mismatch: wanted %s ; got %s",
		// 		id,
		// 		test.msg.Prefix,
		// 		logEntry.Prefix,
		// 	)
		// 	return
		// }

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

func TestTextFmtFmtMetadata(t *testing.T) {}

func TestJSONFmtFormat(t *testing.T) {}
