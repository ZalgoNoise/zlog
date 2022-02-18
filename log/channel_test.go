package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
