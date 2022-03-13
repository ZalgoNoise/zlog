package log

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"regexp"
	"testing"
)

func TestTextFormatLogger(t *testing.T) {
	regxStr := `^\[.*\]\s*\[info\]\s*\[test-new-logger\]\s*test content\s*$`
	regx := regexp.MustCompile(regxStr)

	prefix := "test-new-logger"
	format := TextFormat
	msg := "test content"
	var buf bytes.Buffer

	logger := New(
		WithPrefix(prefix),
		format,
		WithOut(&buf),
	)

	logMessage := NewMessage().Level(LLInfo).Message(msg).Build()

	logger.Log(logMessage)

	if !regx.MatchString(buf.String()) {
		t.Errorf(
			"#%v [Logger] [text-fmt] New(%s,%s).Info(%s) = %s ; expected %s",
			0,
			prefix,
			"TextFormat",
			msg,
			buf.String(),
			regxStr,
		)
	}

	t.Logf(
		"#%v -- TESTED -- [Logger] [text-fmt] New(%s,%s).Info(%s) = %s",
		0,
		prefix,
		"TextFormat",
		msg,
		buf.String(),
	)
}

func TestJSONFormatLogger(t *testing.T) {
	prefix := "test-new-logger"
	format := JSONFormat
	msg := "test content"
	buf := &bytes.Buffer{}
	logEntry := &LogMessage{}

	logger := New(
		WithPrefix(prefix),
		format,
		WithOut(buf),
	)

	logMessage := NewMessage().Level(LLInfo).Message(msg).Build()

	logger.Log(logMessage)

	if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
		t.Errorf(
			"#%v [Logger] [json-fmt] New(%s,%s).Info(%s) -- unmarshal error: %s",
			0,
			prefix,
			"JSONFormat",
			msg,
			err,
		)
	}

	if logEntry.Level != LLInfo.String() ||
		logEntry.Prefix != prefix ||
		logEntry.Msg != msg {
		t.Errorf(
			"#%v [Logger] [json-fmt] New(%s,%s).Info(%s) -- data mismatch",
			0,
			prefix,
			"JSONFormat",
			msg,
		)
	}

	t.Logf(
		"#%v -- TESTED -- [Logger] [json-fmt] New(%s,%s).Info(%s)",
		0,
		prefix,
		"JSONFormat",
		msg,
	)
}

func TestNewSingleWriterLogger(t *testing.T) {
	regxStr := `^\[.*\]\s*\[info\]\s*\[test-new-logger\]\s*test content\s*$`
	regx := regexp.MustCompile(regxStr)

	prefix := "test-new-logger"
	format := TextFormat
	msg := "test content"
	var buf bytes.Buffer

	logger := New(
		WithPrefix(prefix),
		format,
		WithOut(&buf),
	)
	logMessage := NewMessage().Level(LLInfo).Message(msg).Build()

	logger.Log(logMessage)

	if !regx.MatchString(buf.String()) {
		t.Errorf(
			"#%v [Logger] New(%s,%s).Info(%s) = %s ; expected %s",
			0,
			prefix,
			"TextFormat",
			msg,
			buf.String(),
			regxStr,
		)
	}

	t.Logf(
		"#%v -- TESTED -- [Logger] New(%s,%s).Info(%s) = %s",
		0,
		prefix,
		"TextFormat",
		msg,
		buf.String(),
	)

}

func TestNewMultiWriterLogger(t *testing.T) {
	regxStr := `^\[.*\]\s*\[info\]\s*\[test-new-logger\]\s*test content\s*$`
	regx := regexp.MustCompile(regxStr)

	prefix := "test-new-logger"
	format := TextFormat
	msg := "test content"

	var buf1 bytes.Buffer
	var buf2 bytes.Buffer
	var buf3 bytes.Buffer

	buffers := []*bytes.Buffer{
		&buf1, &buf2, &buf3,
	}

	logger := New(
		WithPrefix(prefix),
		format,
		WithOut(&buf1, &buf2, &buf3),
	)

	logMessage := NewMessage().Level(LLInfo).Message(msg).Build()

	logger.Log(logMessage)

	for id, buf := range buffers {
		if !regx.MatchString(buf.String()) {
			t.Errorf(
				"#%v [Logger] [multi-writer] New(%s,%s).Info(%s) = %s ; expected %s",
				id,
				prefix,
				"TextFormat",
				msg,
				buf.String(),
				regxStr,
			)
		}
		t.Logf(
			"#%v -- TESTED -- [Logger] [multi-writer] New(%s,%s).Info(%s) = %s",
			id,
			prefix,
			"TextFormat",
			msg,
			buf.String(),
		)
	}

}

func TestNewDefaultWriterLogger(t *testing.T) {
	regxStr := `^\[.*\]\s*\[info\]\s*\[log\]\s*test content\s*$`
	regx := regexp.MustCompile(regxStr)

	msg := "test content"

	out := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// forcing this override of os.Stdout so that we can read from it
	logger := New(WithOut(os.Stdout))

	logMessage := NewMessage().Level(LLInfo).Message(msg).Build()

	// https://stackoverflow.com/questions/10473800
	// copy the output in a separate goroutine so printing can't block indefinitely
	outCh := make(chan string)
	go func() {
		var buf bytes.Buffer
		n, err := io.Copy(&buf, r)
		if err != nil {
			t.Errorf(
				"#%v [Logger] [default-writer] New().Info(%s) -- error when copying piped buffer's data: %s",
				0,
				msg,
				err,
			)
		}

		if n == 0 {
			t.Errorf(
				"#%v [Logger] [default-writer] New().Info(%s) -- piped buffer's data is zero bytes in length",
				0,
				msg,
			)
		}
		outCh <- buf.String()
	}()
	logger.Log(logMessage)

	w.Close()
	result := <-outCh
	os.Stdout = out

	if !regx.MatchString(result) {
		t.Errorf(
			"#%v [Logger] [default-writer] New().Info(%s) = %s ; expected %s",
			0,
			msg,
			result,
			regxStr,
		)
	}

	t.Logf(
		"#%v -- TESTED -- [Logger] [multi-writer] New().Info(%s) = %s",
		0,
		msg,
		result,
	)

}

func TestLoggerSetOuts(t *testing.T) {
	type test struct {
		prefix string
		format LoggerConfig
		outs   []io.Writer
		bufs   []*bytes.Buffer
	}

	var tests []test

	regxStr := `^\[.*\]\s*\[info\]\s*\[test-new-logger\]\s*test content\s*$`
	regx := regexp.MustCompile(regxStr)

	prefix := "test-new-logger"
	format := TextFormat
	msg := "test content"

	for i := 0; i < 5; i++ {
		var writters []io.Writer
		var buffers []*bytes.Buffer

		for b := 0; b <= i; b++ {
			var buf bytes.Buffer
			writters = append(writters, &buf)
			buffers = append(buffers, &buf)
		}

		tests = append(tests, test{
			prefix: prefix,
			format: format,
			outs:   writters,
			bufs:   buffers,
		})
	}

	for _, test := range tests {
		logger := New(
			WithPrefix(test.prefix),
			test.format,
		)
		logger.SetOuts(test.outs...)
		logMessage := NewMessage().Level(LLInfo).Message(msg).Build()

		logger.Log(logMessage)

		for id, buf := range test.bufs {
			if !regx.MatchString(buf.String()) {
				t.Errorf(
					"#%v [Logger] SetOuts().Info(%s) -- message mismatch",
					id,
					msg,
				)
			}

			t.Logf(
				"#%v -- TESTED -- [Logger] SetOuts().Info(%s) over %v buffers",
				id,
				msg,
				len(test.bufs),
			)
		}

	}

}

func TestLoggerAddOuts(t *testing.T) {
	type test struct {
		prefix string
		format LoggerConfig
		outs   []io.Writer
		bufs   []*bytes.Buffer
	}

	var tests []test

	regxStr := `^\[.*\]\s*\[info\]\s*\[test-new-logger\]\s*test content\s*$`
	regx := regexp.MustCompile(regxStr)

	prefix := "test-new-logger"
	format := TextFormat
	msg := "test content"

	for i := 1; i < 5; i++ {
		var writters []io.Writer
		var buffers []*bytes.Buffer

		for b := 0; b <= i; b++ {
			var buf bytes.Buffer
			writters = append(writters, &buf)
			buffers = append(buffers, &buf)
		}

		tests = append(tests, test{
			prefix: prefix,
			format: format,
			outs:   writters,
			bufs:   buffers,
		})
	}

	for _, test := range tests {
		logger := New(
			WithPrefix(test.prefix),
			test.format,
			WithOut(test.outs[0]),
		)
		logger.AddOuts(test.outs[1:]...)
		logMessage := NewMessage().Level(LLInfo).Message(msg).Build()

		logger.Log(logMessage)

		for id, buf := range test.bufs {
			if !regx.MatchString(buf.String()) {
				t.Errorf(
					"#%v [Logger] AddOuts().Info(%s) -- message mismatch",
					id,
					msg,
				)
			}

			t.Logf(
				"#%v -- TESTED -- [Logger] AddOuts().Info(%s) over %v buffers",
				id,
				msg,
				len(test.bufs),
			)
		}

	}
}

func TestLoggerPrefix(t *testing.T) {
	type test struct {
		prefix string
		format LoggerConfig
		outs   []io.Writer
		bufs   []*bytes.Buffer
	}

	var tests []test

	var testPrefixes = []string{
		"logger-prefix",
		"logger-test",
		"logger-new",
		"logger-changed",
		"logger-done",
	}

	regxStr := `^\[.*\]\s*\[info\]\s*\[(.*)\]\s*test content\s*$`
	regx := regexp.MustCompile(regxStr)

	format := TextFormat
	msg := "test content"

	for _, p := range testPrefixes {
		buf := &bytes.Buffer{}
		tests = append(tests, test{
			prefix: p,
			format: format,
			outs:   []io.Writer{buf},
			bufs:   []*bytes.Buffer{buf},
		})
	}

	for id, test := range tests {
		logger := New(
			WithPrefix("old"),
			test.format,
			WithOut(test.outs...),
		)
		logger.Prefix(test.prefix)
		logMessage := NewMessage().Level(LLInfo).Message(msg).Build()

		logger.Log(logMessage)

		for _, buf := range test.bufs {
			if !regx.MatchString(buf.String()) {
				t.Errorf(
					"#%v [Logger] Prefix().Info(%s) -- message regex mismatch: %s",
					id,
					msg,
					regxStr,
				)
			}

			match := regx.FindStringSubmatch(buf.String())

			var ok bool
			for _, v := range match {
				ok = false
				if v == test.prefix {
					ok = true
					break
				}
			}
			if !ok {
				t.Errorf(
					"#%v [Logger] Prefix().Info(%s) -- unexpected prefix -- wanted %s",
					id,
					msg,
					test.prefix,
				)
			}

			t.Logf(
				"#%v -- TESTED -- [Logger] Prefix().Info(%s) -- finding prefix %s",
				id,
				msg,
				test.prefix,
			)
		}
	}

}

func TestLoggerSub(t *testing.T) {
	type test struct {
		sub    string
		format LoggerConfig
		outs   []io.Writer
		bufs   []*bytes.Buffer
	}

	var tests []test

	var testSubPrefixes = []string{
		"logger-prefix",
		"logger-test",
		"logger-new",
		"logger-changed",
		"logger-done",
	}

	regxStr := `^\[.*\]\s*\[info\]\s*\[log\]\s*\[(.*)\]\s*test content\s*$`
	regx := regexp.MustCompile(regxStr)

	format := TextFormat
	msg := "test content"

	for _, s := range testSubPrefixes {
		buf := &bytes.Buffer{}
		tests = append(tests, test{
			sub:    s,
			format: format,
			outs:   []io.Writer{buf},
			bufs:   []*bytes.Buffer{buf},
		})
	}

	for id, test := range tests {
		logger := New(
			WithSub("old"),
			test.format,
			WithOut(test.outs...),
		)
		logger.Sub(test.sub)
		logMessage := NewMessage().Level(LLInfo).Message(msg).Build()

		logger.Log(logMessage)

		for _, buf := range test.bufs {
			if !regx.MatchString(buf.String()) {
				t.Errorf(
					"#%v [Logger] FAILED -- Sub().Info(%s) -- message regex mismatch: %s",
					id,
					msg,
					regxStr,
				)
			}

			match := regx.FindStringSubmatch(buf.String())

			var ok bool
			for _, v := range match {
				ok = false
				if v == test.sub {
					ok = true
					break
				}
			}
			if !ok {
				t.Errorf(
					"#%v [Logger] FAILED -- Sub().Info(%s) -- unexpected subprefix -- got %s ; wanted %s",
					id,
					msg,
					buf.String(),
					test.sub,
				)
			}

			t.Logf(
				"#%v -- TESTED -- [Logger] Prefix().Info(%s) -- finding prefix %s",
				id,
				msg,
				test.sub,
			)
		}
	}

}

func TestLoggerFields(t *testing.T) {

	prefix := "test-new-logger"
	format := JSONFormat
	msg := "test content"

	for id, obj := range testObjects {
		buf := &bytes.Buffer{}
		logEntry := &LogMessage{}

		logger := New(
			WithPrefix(prefix),
			format,
			WithOut(buf),
		)

		logger.Fields(obj).Info(msg)

		if err := json.Unmarshal(buf.Bytes(), logEntry); err != nil {
			t.Errorf(
				"#%v [Logger] [json-fmt] Fields().Info(%s) -- unmarshal error: %s",
				id,
				msg,
				err,
			)
		}

		t.Logf(
			"#%v -- TESTED -- [Logger] [json-fmt] Fields().Info(%s) : %s",
			id,
			msg,
			buf.String(),
		)

	}

}
