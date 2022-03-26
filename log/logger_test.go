package log

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/zalgonoise/zlog/store"
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
	logger := New(WithOut(os.Stdout), FormatText)

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
	module := "Logger"
	funcname := "SetOuts()"

	tlogger := New(
		WithPrefix("test-new-logger"),
		TextFormat,
	)

	var tests = []struct {
		name  string
		input []io.Writer
		wants io.Writer
	}{
		{
			name:  "switching to buffer #0",
			input: []io.Writer{mockBufs[0]},
			wants: io.MultiWriter(mockBufs[0]),
		},
		{
			name:  "switching to os.Stdout",
			input: []io.Writer{os.Stdout},
			wants: io.MultiWriter(os.Stdout),
		},
		{
			name:  "switching to multi-buffer #0",
			input: []io.Writer{mockBufs[0], mockBufs[1], mockBufs[3]},
			wants: io.MultiWriter(mockBufs[0], mockBufs[1], mockBufs[3]),
		},
		{
			name:  "switching to default writer with zero arguments",
			input: nil,
			wants: os.Stdout,
		},
		{
			name:  "switching to default writer with nil writers",
			input: []io.Writer{nil, nil, nil},
			wants: os.Stdout,
		},
		{
			name:  "ensure the empty writer works",
			input: []io.Writer{store.EmptyWriter},
			wants: io.MultiWriter(store.EmptyWriter),
		},
	}

	var verify = func(id int, logw, w io.Writer) {

		if !reflect.DeepEqual(logw, w) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] writer mismatch: wanted %v ; got %v",
				id,
				module,
				funcname,
				w,
				logw,
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
		if test.input != nil {
			tlogger.SetOuts(test.input...)
		} else {
			tlogger.SetOuts()
		}

		logw := tlogger.(*logger).out

		verify(id, logw, test.wants)

	}
}

func TestLoggerAddOuts(t *testing.T) {
	module := "Logger"
	funcname := "AddOuts()"

	tlogger := New(
		WithPrefix("test-new-logger"),
		WithOut(mockBufs[5]),
		TextFormat,
	)

	var tests = []struct {
		name  string
		input []io.Writer
		wants io.Writer
	}{
		{
			name:  "adding buffer #0",
			input: []io.Writer{mockBufs[0]},
			wants: io.MultiWriter(mockBufs[0], mockBufs[5]),
		},
		{
			name:  "adding os.Stdout",
			input: []io.Writer{os.Stdout},
			wants: io.MultiWriter(os.Stdout, mockBufs[5]),
		},
		{
			name:  "adding multi-buffer #0",
			input: []io.Writer{mockBufs[0], mockBufs[1], mockBufs[3]},
			wants: io.MultiWriter(mockBufs[0], mockBufs[1], mockBufs[3], mockBufs[5]),
		},
		{
			name:  "adding default writer with zero arguments",
			input: nil,
			wants: io.MultiWriter(mockBufs[5]),
		},
		{
			name:  "adding default writer with nil writers",
			input: []io.Writer{nil, nil, nil},
			wants: io.MultiWriter(mockBufs[5]),
		},
		{
			name:  "ensure the empty writer works",
			input: []io.Writer{store.EmptyWriter},
			wants: io.MultiWriter(store.EmptyWriter, mockBufs[5]),
		},
	}

	var verify = func(id int, logw, w io.Writer) {

		if !reflect.DeepEqual(logw, w) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] writer mismatch: wanted %v ; got %v",
				id,
				module,
				funcname,
				w,
				logw,
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

		if test.input != nil {
			tlogger.AddOuts(test.input...)
		} else {
			tlogger.AddOuts()
		}

		logw := tlogger.(*logger).out

		verify(id, logw, test.wants)

		// reset
		tlogger.SetOuts(mockBufs[5])

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

func TestLoggerWrite(t *testing.T) {
	type test struct {
		msg  []byte
		want LogMessage
	}

	var tests = []test{
		{
			msg: NewMessage().Level(LLInfo).Prefix("test").Sub("tester").Message("write test").Build().Bytes(),
			want: LogMessage{
				Prefix: "test",
				Sub:    "tester",
				Level:  LLInfo.String(),
				Msg:    "write test",
			},
		},
		{
			msg: []byte("hello world"),
			want: LogMessage{
				Prefix: "log",
				Sub:    "",
				Level:  LLInfo.String(),
				Msg:    "hello world",
			},
		},
	}

	logger := New(FormatJSON, WithOut(mockBuffer))

	var verify = func(id int, test test, buf []byte) {

		logEntry := &LogMessage{}

		err := json.Unmarshal(buf, logEntry)
		if err != nil {
			t.Errorf(
				"#%v [Logger] -- FAILED -- Write([]byte) -- JSON decoding error: %s",
				id,
				err,
			)
			return
		}

		if logEntry.Prefix != test.want.Prefix {
			t.Errorf(
				"#%v [Logger] -- FAILED -- Write([]byte) -- prefix mismatch: wanted %s ; got %s",
				id,
				logEntry.Prefix,
				test.want.Prefix,
			)
			return
		}

		if logEntry.Sub != test.want.Sub {
			t.Errorf(
				"#%v [Logger] -- FAILED -- Write([]byte) -- sub-prefix mismatch: wanted %s ; got %s",
				id,
				logEntry.Sub,
				test.want.Sub,
			)
			return
		}

		if logEntry.Level != test.want.Level {
			t.Errorf(
				"#%v [Logger] -- FAILED -- Write([]byte) -- log level mismatch: wanted %s ; got %s",
				id,
				logEntry.Level,
				test.want.Level,
			)
			return
		}

		if logEntry.Msg != test.want.Msg {
			t.Errorf(
				"#%v [Logger] -- FAILED -- Write([]byte) -- message mismatch: wanted %s ; got %s",
				id,
				logEntry.Msg,
				test.want.Msg,
			)
			return
		}

		t.Logf(
			"#%v [Logger] -- PASSED -- Write([]byte)",
			id,
		)

	}

	for id, test := range tests {
		mockBuffer.Reset()
		n, err := logger.Write(test.msg)

		if err != nil {
			t.Errorf(
				"#%v [Logger] -- FAILED -- Write([]byte) -- write error: %s",
				id,
				err,
			)
		}

		if n <= 0 {
			t.Errorf(
				"#%v [Logger] -- FAILED -- Write([]byte) -- no bytes written: %v",
				id,
				n,
			)
		}

		verify(id, test, mockBuffer.Bytes())
	}
}
