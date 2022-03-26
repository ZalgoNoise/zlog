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

	var verify = func(id int, logw, w io.Writer, action string) {

		if !reflect.DeepEqual(logw, w) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] writer mismatch: wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				w,
				logw,
				action,
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

		verify(id, logw, test.wants, test.name)

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

	var verify = func(id int, logw, w io.Writer, action string) {

		if !reflect.DeepEqual(logw, w) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] writer mismatch: wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				w,
				logw,
				action,
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

		verify(id, logw, test.wants, test.name)

		// reset
		tlogger.SetOuts(mockBufs[5])

	}
}

func TestLoggerPrefix(t *testing.T) {
	module := "Logger"
	funcname := "Prefix()"

	tlogger := New(
		WithPrefix("test-new-logger"),
		WithOut(mockBufs[0]),
		TextFormat,
	)

	var tests = []struct {
		name  string
		input string
		wants string
	}{
		{
			name:  "switch logger prefixes",
			input: "logger-prefix",
			wants: "logger-prefix",
		},
		{
			name:  "switch logger prefixes",
			input: "logger-test",
			wants: "logger-test",
		},
		{
			name:  "switch logger prefixes",
			input: "logger-new",
			wants: "logger-new",
		},
		{
			name:  "switch to defaults",
			input: "",
			wants: "log",
		},
	}

	var verify = func(id int, p, wants, action string) {
		if p != wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] prefix mismatch: wanted %s ; got %s -- action: %s",
				id,
				module,
				funcname,
				p,
				wants,
				action,
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
		tlogger.Prefix(test.input)

		p := tlogger.(*logger).prefix

		verify(id, p, test.wants, test.name)

	}

}

func TestLoggerSub(t *testing.T) {
	module := "Logger"
	funcname := "Sub()"

	tlogger := New(
		WithPrefix("test-new-logger"),
		WithOut(mockBufs[0]),
		TextFormat,
	)

	var tests = []struct {
		name  string
		input string
		wants string
	}{
		{
			name:  "switch logger sub-prefixes",
			input: "logger-subprefix",
			wants: "logger-subprefix",
		},
		{
			name:  "switch logger sub-prefixes",
			input: "logger-test",
			wants: "logger-test",
		},
		{
			name:  "switch logger sub-prefixes",
			input: "logger-new",
			wants: "logger-new",
		},
		{
			name:  "switch to defaults",
			input: "",
			wants: "",
		},
	}

	var verify = func(id int, s, wants, action string) {
		if s != wants {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] subprefix mismatch: wanted %s ; got %s -- action: %s",
				id,
				module,
				funcname,
				s,
				wants,
				action,
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
		tlogger.Sub("")

		tlogger.Sub(test.input)

		s := tlogger.(*logger).sub

		verify(id, s, test.wants, test.name)

		tlogger.Sub("")
	}

}

func TestLoggerFields(t *testing.T) {
	module := "Logger"
	funcname := "Fields()"

	tlogger := New(
		WithPrefix("test-new-logger"),
		WithOut(mockBufs[0]),
		TextFormat,
	)

	var tests = []struct {
		name  string
		input map[string]interface{}
		wants map[string]interface{}
	}{
		{
			name:  "switch logger metadata",
			input: testObjects[0],
			wants: testObjects[0],
		},
		{
			name:  "switch logger metadata",
			input: testObjects[1],
			wants: testObjects[1],
		},
		{
			name:  "switch logger metadata",
			input: testObjects[2],
			wants: testObjects[2],
		},
		{
			name:  "switch logger metadata",
			input: testObjects[3],
			wants: testObjects[3],
		},
		{
			name:  "switch to defaults",
			input: map[string]interface{}{},
			wants: map[string]interface{}{},
		},
		{
			name:  "nil input check",
			input: nil,
			wants: map[string]interface{}{},
		},
	}

	var verify = func(id int, m, wants map[string]interface{}, action string) {
		if len(m) != len(wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] metadata length mismatch: wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				len(m),
				len(wants),
				action,
			)
			return
		}

		// empty content expected, exit successfully
		if len(m) == 0 && len(wants) == 0 {
			t.Logf(
				"#%v -- PASSED -- [%s] [%s]",
				id,
				module,
				funcname,
			)
			return
		}

		if !reflect.DeepEqual(m, wants) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] metadata content mismatch: wanted %v ; got %v -- action: %s",
				id,
				module,
				funcname,
				m,
				wants,
				action,
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
		// reset
		tlogger.Fields(nil)

		tlogger.Fields(test.input)

		m := tlogger.(*logger).meta

		verify(id, m, test.wants, test.name)

		// reset
		tlogger.Fields(nil)
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

// func (l *nilLogger) Write(p []byte) (n int, err error)           { return 0, nil }
func TestNilLoggerWrite(t *testing.T) {
	module := "NilLogger"
	funcname := "Write()"

	nillog := New(NilConfig)

	var tests = []struct {
		input []byte
		n     int
		err   error
	}{
		{
			input: []byte("abc"),
			n:     0,
			err:   nil,
		},
		{
			input: []byte("123"),
			n:     0,
			err:   nil,
		},
		{
			input: []byte("!@#"),
			n:     0,
			err:   nil,
		},
	}

	for id, test := range tests {
		n, err := nillog.Write(test.input)

		if n != test.n {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] bytes written mismatch: wanted %v ; got %v",
				id,
				module,
				funcname,
				test.n,
				n,
			)
			return

		}

		if err != test.err {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] returning error mismatch: wanted %v ; got %v",
				id,
				module,
				funcname,
				test.err,
				err,
			)
			return
		}
	}
}

// func (l *nilLogger) SetOuts(outs ...io.Writer) Logger            { return l }
func TestNilLoggerSetOuts(t *testing.T) {
	// module := "NilLogger"
	// funcname := "Write()"

	nillog := New(NilConfig)

	var tests = []struct {
		input []io.Writer
	}{
		{
			input: []io.Writer{mockBufs[0]},
		},
		{
			input: []io.Writer{mockBufs[0], mockBufs[1]},
		},
		{
			input: []io.Writer{},
		},
	}

	for _, test := range tests {
		nillog.SetOuts(test.input...)
	}
}

// func (l *nilLogger) AddOuts(outs ...io.Writer) Logger            { return l }
func TestNilLoggerAddOuts(t *testing.T) {
	// module := "NilLogger"
	// funcname := "Write()"

	nillog := New(NilConfig)

	var tests = []struct {
		input []io.Writer
	}{
		{
			input: []io.Writer{mockBufs[0]},
		},
		{
			input: []io.Writer{mockBufs[0], mockBufs[1]},
		},
		{
			input: []io.Writer{},
		},
	}

	for _, test := range tests {
		nillog.AddOuts(test.input...)
	}
}

// func (l *nilLogger) Prefix(prefix string) Logger                 { return l }
func TestNilLoggerPrefix(t *testing.T) {
	// module := "NilLogger"
	// funcname := "Write()"

	nillog := New(NilConfig)

	var tests = []struct {
		input string
	}{
		{
			input: "test",
		},
		{
			input: "for",
		},
		{
			input: "nothing",
		},
	}

	for _, test := range tests {
		nillog.Prefix(test.input)
	}
}

// func (l *nilLogger) Sub(sub string) Logger                       { return l }
func TestNilLoggerSub(t *testing.T) {
	// module := "NilLogger"
	// funcname := "Write()"

	nillog := New(NilConfig)

	var tests = []struct {
		input string
	}{
		{
			input: "test",
		},
		{
			input: "for",
		},
		{
			input: "nothing",
		},
	}

	for _, test := range tests {
		nillog.Sub(test.input)
	}
}

// func (l *nilLogger) Fields(fields map[string]interface{}) Logger { return l }
func TestNilLoggerFields(t *testing.T) {
	// module := "NilLogger"
	// funcname := "Write()"

	nillog := New(NilConfig)

	var tests = []struct {
		input map[string]interface{}
	}{
		{
			input: map[string]interface{}{"a": 0},
		},
		{
			input: map[string]interface{}{"a": "b"},
		},
		{
			input: map[string]interface{}{},
		},
	}

	for _, test := range tests {
		nillog.Fields(test.input)
	}
}

// func (l *nilLogger) IsSkipExit() bool                            { return true }
func TestNilLoggerIsSkipExit(t *testing.T) {
	module := "NilLogger"
	funcname := "IsSkipExit()"

	nillog := New(NilConfig)

	if !nillog.IsSkipExit() {
		t.Errorf(
			"-- FAILED -- [%s] [%s] nilLogger should be set as skipping exit",
			module,
			funcname,
		)
	}
}
