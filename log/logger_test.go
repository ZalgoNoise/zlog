package log

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"testing"
)

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

	for id, test := range allTests {
		result := test.input.String()

		if result == "" && test.pass {
			t.Errorf(
				"#%v [LogLevel] LogLevel(%v).String() -- unexpected reference, got %s",
				id,
				int(test.input),
				result,
			)
		}

		if result != test.ok && !test.pass {
			t.Errorf(
				"#%v [LogLevel] LogLevel(%v).String() -- expected %s, got %s",
				id,
				int(test.input),
				test.ok,
				result,
			)
		} else {
			t.Logf(
				"#%v -- TESTED -- [LogLevel] LogLevel(%v).String() = %s",
				id,
				int(test.input),
				result,
			)
		}
	}
}

func TestNewSingleWriterLogger(t *testing.T) {
	regxStr := `^\[.*\]\s*\[info\]\s*\[test-new-logger\]\s*test content\s*$`
	regx := regexp.MustCompile(regxStr)

	prefix := "test-new-logger"
	format := TextFormat
	msg := "test content"
	var buf bytes.Buffer

	logger := New(prefix, format, &buf)

	logger.Log(LLInfo, msg)

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

	logger := New(prefix, format, &buf1, &buf2, &buf3)

	logger.Log(LLInfo, msg)

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
	regxStr := `^\[.*\]\s*\[info\]\s*\[test-new-logger\]\s*test content\s*$`
	regx := regexp.MustCompile(regxStr)

	prefix := "test-new-logger"
	format := TextFormat
	msg := "test content"

	out := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := New(prefix, format)
	logger.Log(LLInfo, msg)

	// https://stackoverflow.com/questions/10473800
	// copy the output in a separate goroutine so printing can't block indefinitely
	outCh := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outCh <- buf.String()
	}()

	w.Close()
	os.Stdout = out
	result := <-outCh

	if !regx.MatchString(result) {
		t.Errorf(
			"#%v [Logger] [default-writer] New(%s,%s).Info(%s) = %s ; expected %s",
			0,
			prefix,
			"TextFormat",
			msg,
			result,
			regxStr,
		)
	}

	t.Logf(
		"#%v -- TESTED -- [Logger] [multi-writer] New(%s,%s).Info(%s) = %s",
		0,
		prefix,
		"TextFormat",
		msg,
		result,
	)

}

func TestLoggerSetOuts(t *testing.T) {
	type test struct {
		prefix string
		format LogFormatter
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
		logger := New(test.prefix, test.format)
		logger.SetOuts(test.outs...)
		logger.Log(LLInfo, msg)

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
		format LogFormatter
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
		logger := New(test.prefix, test.format, test.outs[0])
		logger.AddOuts(test.outs[1:]...)
		logger.Log(LLInfo, msg)

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

func TestLoggerSetPrefix(t *testing.T) {
	type test struct {
		prefix string
		format LogFormatter
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
		logger := New("old", test.format, test.outs...)
		logger.SetPrefix(test.prefix)
		logger.Log(LLInfo, msg)

		for _, buf := range test.bufs {
			if !regx.MatchString(buf.String()) {
				t.Errorf(
					"#%v [Logger] SetPrefix().Info(%s) -- message regex mismatch: %s",
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
					"#%v [Logger] SetPrefix().Info(%s) -- unexpected prefix -- wanted %s",
					id,
					msg,
					test.prefix,
				)
			}

			t.Logf(
				"#%v -- TESTED -- [Logger] SetPrefix().Info(%s) -- finding prefix %s",
				id,
				msg,
				test.prefix,
			)
		}
	}

}

func TestLoggerFields(t *testing.T) {
}
