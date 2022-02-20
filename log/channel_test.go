package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"testing"
)

func TestNewLogCh(t *testing.T) {
	type log struct {
		logger LoggerI
		buf    []*bytes.Buffer
	}
	type test struct {
		log
		msg *LogMessage
	}

	var testAllObjects []map[string]interface{}
	testAllObjects = append(testAllObjects, testObjects...)
	testAllObjects = append(testAllObjects, testEmptyObjects...)

	var testAllMessages []string
	testAllMessages = append(testAllMessages, mockMessages...)
	for _, fmtMsg := range mockFmtMessages {
		testAllMessages = append(testAllMessages, fmt.Sprintf(fmtMsg.format, fmtMsg.v...))
	}

	var maxBufs = 6

	var tests []test

	for a := 1; a <= maxBufs; a++ {
		for b := 0; b < len(mockPrefixes); b++ {
			for c := 0; c < len(testAllMessages); c++ {
				for d := 0; d < len(testAllObjects); d++ {
					for e := 0; e < len(mockLogLevelsOK); e++ {

						// skip LLFatal -- os.Exit(1)
						if mockLogLevelsOK[e] == LLFatal || mockLogLevelsOK[e] == LLPanic {
							continue
						}

						var bufs []*bytes.Buffer
						var w []io.Writer

						for f := 1; f <= a; f++ {
							bufs = append(bufs, &bytes.Buffer{})
							w = append(w, bufs[f-1])
						}

						l := New(mockEmptyPrefixes[0], JSONFormat, w...)

						obj := test{
							log: log{
								logger: l,
								buf:    bufs,
							},
							msg: NewMessage().
								Prefix(mockPrefixes[b]).
								Message(testAllMessages[c]).
								Metadata(testAllObjects[d]).
								Level(mockLogLevelsOK[e]).
								Build(),
						}

						tests = append(tests, obj)
					}

				}
			}
		}
	}

	var verify = func(id int, test test) {
		defer func() {
			for _, b := range test.log.buf {
				b.Reset()
			}
		}()

		for bufID, buf := range test.log.buf {
			logEntry := &LogMessage{}

			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
				t.Errorf(
					"#%v -- FAILED -- [ChLogger] [Buffer #%v] Log(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [ChLogger] [Buffer #%v] Log(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != test.msg.Level {
				t.Errorf(
					"#%v -- FAILED -- [ChLogger] [Buffer #%v] Log(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Level,
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [ChLogger] [Buffer #%v] Log(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [ChLogger] [Buffer #%v] Log(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [ChLogger] [Buffer #%v] Log(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [ChLogger] [Buffer #%v] Log(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [ChLogger] [Buffer #%v] Log(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [ChLogger] [Buffer #%v] Log(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		logCh, chLogger := NewLogCh(test.log.logger)

		defer verify(id, test)

		go chLogger()

		logCh <- test.msg

	}
}

func TestNewLogChMultiLogger(t *testing.T) {
	type log struct {
		logger LoggerI
		buf    []*bytes.Buffer
	}
	type test struct {
		log
		msg *LogMessage
	}

	var testAllObjects []map[string]interface{}
	testAllObjects = append(testAllObjects, testObjects...)
	testAllObjects = append(testAllObjects, testEmptyObjects...)

	var testAllMessages []string
	testAllMessages = append(testAllMessages, mockMessages...)
	for _, fmtMsg := range mockFmtMessages {
		testAllMessages = append(testAllMessages, fmt.Sprintf(fmtMsg.format, fmtMsg.v...))
	}

	var maxBufs = 6

	var tests []test

	for a := 1; a <= maxBufs; a++ {
		for b := 0; b < len(mockPrefixes); b++ {
			for c := 0; c < len(testAllMessages); c++ {
				for d := 0; d < len(testAllObjects); d++ {
					for e := 0; e < len(mockLogLevelsOK); e++ {

						// skip LLFatal -- os.Exit(1)
						if mockLogLevelsOK[e] == LLFatal || mockLogLevelsOK[e] == LLPanic {
							continue
						}

						var bufs []*bytes.Buffer
						var logs []LoggerI

						for f := 1; f <= a; f++ {

							var w []io.Writer
							for g := 1; g <= a; g++ {
								newBuf := &bytes.Buffer{}
								bufs = append(bufs, newBuf)
								w = append(w, newBuf)
							}

							l := New(mockEmptyPrefixes[0], JSONFormat, w...)
							logs = append(logs, l)
						}

						ml := MultiLogger(logs...)

						obj := test{
							log: log{
								logger: ml,
								buf:    bufs,
							},
							msg: NewMessage().
								Prefix(mockPrefixes[b]).
								Message(testAllMessages[c]).
								Metadata(testAllObjects[d]).
								Level(mockLogLevelsOK[e]).
								Build(),
						}

						tests = append(tests, obj)
					}

				}
			}
		}
	}

	var verify = func(id int, test test) {
		defer func() {
			for _, b := range test.log.buf {
				b.Reset()
			}
		}()

		for bufID, buf := range test.log.buf {
			logEntry := &LogMessage{}

			if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
				t.Errorf(
					"#%v -- FAILED -- [ChLogger] [MultiLogger] [Buffer #%v] Log(%s) -- unmarshal error: %s",
					id,
					bufID,
					test.msg.Msg,
					err,
				)
				return
			}

			if logEntry.Prefix != test.msg.Prefix {
				t.Errorf(
					"#%v -- FAILED -- [ChLogger] [MultiLogger] [Buffer #%v] Log(%s) -- prefix mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Prefix,
					logEntry.Prefix,
				)
				return
			}

			if logEntry.Level != test.msg.Level {
				t.Errorf(
					"#%v -- FAILED -- [ChLogger] [MultiLogger] [Buffer #%v] Log(%s) -- log level mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Level,
					logEntry.Level,
				)
				return
			}

			if logEntry.Msg != test.msg.Msg {
				t.Errorf(
					"#%v -- FAILED -- [ChLogger] [MultiLogger] [Buffer #%v] Log(%s) -- message mismatch: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Msg,
					logEntry.Msg,
				)
				return
			}

			if logEntry.Metadata == nil && test.msg.Metadata != nil {
				t.Errorf(
					"#%v -- FAILED -- [ChLogger] [MultiLogger] [Buffer #%v] Log(%s) -- retrieved empty metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			} else if logEntry.Metadata != nil && test.msg.Metadata == nil {
				t.Errorf(
					"#%v -- FAILED -- [ChLogger] [MultiLogger] [Buffer #%v] Log(%s) -- retrieved unexpected metadata object: wanted %s ; got %s",
					id,
					bufID,
					test.msg.Msg,
					test.msg.Metadata,
					logEntry.Metadata,
				)
				return
			}

			if logEntry.Metadata != nil && test.msg.Metadata != nil {
				for k, v := range logEntry.Metadata {
					if v != nil && test.msg.Metadata[k] == nil {
						t.Errorf(
							"#%v -- FAILED -- [ChLogger] [MultiLogger] [Buffer #%v] Log(%s) -- metadata mismatch: key %s contains data ; original message's key %s doesn't",
							id,
							bufID,
							test.msg.Msg,
							k,
							k,
						)
						return
					}
				}

				if len(logEntry.Metadata) != len(test.msg.Metadata) {
					t.Errorf(
						"#%v -- FAILED -- [ChLogger] [MultiLogger] [Buffer #%v] Log(%s) -- metadata length mismatch -- wanted %v, got %v",
						id,
						bufID,
						test.msg.Msg,
						len(test.msg.Metadata),
						len(logEntry.Metadata),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [ChLogger] [MultiLogger] [Buffer #%v] Log(%s) -- %s",
				id,
				bufID,
				test.msg.Msg,
				buf.String(),
			)
		}

	}

	for id, test := range tests {
		logCh, chLogger := NewLogCh(test.log.logger)

		defer verify(id, test)

		go chLogger()

		logCh <- test.msg

	}
}

func TestNewLogChMultiEntry(t *testing.T) {
	type log struct {
		logger LoggerI
		buf    []*bytes.Buffer
	}
	type test struct {
		log
		msg []*LogMessage
		rgx []string
	}

	var prefix = "multientry-test"
	var msgs = []string{
		"test log #0",
		"test log #1",
		"test log #2",
		"test log #3",
		"test log #4",
		"test log #5",
		"test log #6",
		"test log #7",
		"test log #8",
		"test log #9",
		"test log #10",
	}

	regxStr := `^\[.*\]\s*\[info\]\s*\[multientry-test\]\s*test log #`

	var regxList = []string{
		regxStr + `0`,
		regxStr + `1`,
		regxStr + `2`,
		regxStr + `3`,
		regxStr + `4`,
		regxStr + `5`,
		regxStr + `6`,
		regxStr + `7`,
		regxStr + `8`,
		regxStr + `9`,
		regxStr + `10`,
	}
	var maxBufs = 6

	var tests []test

	for a := 1; a <= maxBufs; a++ {

		var bufs []*bytes.Buffer
		var logs []LoggerI
		var msgObj []*LogMessage
		var rxList []string

		for b := 1; b <= a; b++ {

			var w = []io.Writer{}
			for c := 1; c <= a; c++ {
				newBuf := &bytes.Buffer{}
				bufs = append(bufs, newBuf)
				w = append(w, newBuf)
			}

			l := New(prefix, TextFormat, w...)
			logs = append(logs, l)
		}

		ml := MultiLogger(logs...)

		for d := 0; d < len(msgs); d++ {
			obj := NewMessage().
				Prefix(prefix).
				Message(msgs[d]).
				Level(LLInfo).
				Build()

			msgObj = append(msgObj, obj)
			rxList = append(rxList, regxList[d])
		}

		obj := test{
			log: log{
				logger: ml,
				buf:    bufs,
			},
			msg: msgObj,
			rgx: regxList,
		}

		tests = append(tests, obj)

	}

	var verify = func(id int, test test) {

		for bufID, buf := range test.log.buf {
			if buf.Len() == 0 {
				t.Logf("Buf #%v has 0-length", bufID)
			}

			var lines [][]byte
			var line []byte
			for _, b := range buf.Bytes() {
				if b != 10 {
					line = append(line, b)
					continue
				}
				lines = append(lines, line)
				line = []byte{}
			}

			for idx, line := range lines {
				rgx := regexp.MustCompile(test.rgx[idx])

				if !rgx.MatchString(string(line)) {
					t.Errorf(
						"#%v -- FAILED -- [ChLogger] [MultiEntry] [Buffer #%v] Log() x%v -- message race error: messages did not arrive in the right order\n\n%s",
						id,
						bufID,
						len(lines),
						buf.String(),
					)
					return
				}
			}

			t.Logf(
				"#%v -- PASSED -- [ChLogger] [MultiEntry] [Buffer #%v] Log() x%v -- \n\n%s",
				id,
				bufID,
				len(lines),
				buf.String(),
			)
		}
	}

	for id, test := range tests {
		defer func() {
			for _, b := range test.log.buf {
				b.Reset()
			}
		}()

		logCh, chLogger := NewLogCh(test.log.logger)

		go chLogger()

		for _, msg := range test.msg {
			logCh <- msg
		}

		verify(id, test)

	}
}
