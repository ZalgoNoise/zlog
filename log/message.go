package log

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"time"
)

// LogLevel type describes a numeric value for a log level with priority increasing in
// relation to its value
//
// LogLevel also implements the Stringer interface, used to convey this log level in a message
type LogLevel int

const (
	LLTrace LogLevel = iota
	LLDebug
	LLInfo
	LLWarn
	LLError
	LLFatal
	_
	_
	_
	LLPanic
)

// String method is defined for LogLevel objects to implement the Stringer interface
//
// It returns the string to which this log level is mapped to, in `logTypeVals`
func (ll LogLevel) String() string {
	return logTypeVals[ll]
}

// Int method returns a LogLevel's value as an integer, to be used for comparison with
// input log level filters
func (ll LogLevel) Int() int {
	return int(ll)
}

var logTypeVals = map[LogLevel]string{
	0: "trace",
	1: "debug",
	2: "info",
	3: "warn",
	4: "error",
	5: "fatal",
	9: "panic",
}

var logTypeKeys = map[string]int{
	"trace": 0,
	"debug": 1,
	"info":  2,
	"warn":  3,
	"error": 4,
	"fatal": 5,
	"panic": 9,
}

// Field type is a generic type to build LogMessage Metadata
type Field map[string]interface{}

// ToMap method returns the Field in it's (raw) string-interface{} map format
func (f Field) ToMap() map[string]interface{} {
	return f
}

type field struct {
	Key string      `xml:"key,omitempty"`
	Val interface{} `xml:"value,omitempty"`
}

func mappify(data map[string]interface{}) []field {
	var fields []field

	for k, v := range data {
		switch value := v.(type) {
		case []map[string]interface{}:
			f := []field{}

			for _, im := range value {
				ifield := field{}
				for ik, iv := range im {
					ifield.Key = ik
					ifield.Val = iv
				}

				f = append(f, ifield)
			}

			fields = append(fields, field{
				Key: k,
				Val: f,
			})
		case []Field:
			f := []field{}

			for _, im := range value {
				ifield := field{}
				for ik, iv := range im.ToMap() {
					ifield.Key = ik
					ifield.Val = iv
				}

				f = append(f, ifield)
			}

			fields = append(fields, field{
				Key: k,
				Val: f,
			})
		case map[string]interface{}:
			fields = append(fields, field{
				Key: k,
				Val: mappify(value),
			})
		case Field:
			fields = append(fields, field{
				Key: k,
				Val: mappify(value.ToMap()),
			})
		default:
			fields = append(fields, field{
				Key: k,
				Val: value,
			})
		}
	}

	return fields
}

// ToXML method returns the Field in a list of key-value objects,
// compatible with XML marshalling of data objects
func (f Field) ToXML() []field {
	return mappify(f.ToMap())
}

// LogMessage struct describes a Log Message's elements, already in a format that can be
// parsed by a valid formatter.
type LogMessage struct {
	Time     time.Time              `json:"timestamp,omitempty" xml:"timestamp,omitempty"`
	Prefix   string                 `json:"service,omitempty" xml:"service,omitempty"`
	Sub      string                 `json:"module,omitempty" xml:"module,omitempty"`
	Level    string                 `json:"level,omitempty" xml:"level,omitempty"`
	Msg      string                 `json:"message,omitempty" xml:"message,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty" xml:"metadata,omitempty"`
}

func (m *LogMessage) encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	gob.Register(Field{})
	enc := gob.NewEncoder(buf)

	err := enc.Encode(m)

	return buf.Bytes(), err
}

// Bytes method will return a LogMessage as a gob-encoded slice of bytes. It is compatible with
// a Logger's io.Writer implementation, as its Write() method will decode this type of data
func (m *LogMessage) Bytes() []byte {
	// skip error checking
	buf, _ := m.encode()
	return buf
}

// MessageBuilder struct describes the elements in a LogMessage's builder, which will
// be the target of different changes until its `Build()` method is called -- returning
// then a pointer to a LogMessage object
type MessageBuilder struct {
	time     time.Time
	prefix   string
	sub      string
	level    string
	msg      string
	metadata map[string]interface{}
}

// NewMessage function is the initializer of a MessageBuilder. From this call, further
// MessageBuilder methods can be chained since they all return pointers to the same object
func NewMessage() *MessageBuilder {
	return &MessageBuilder{}
}

// Prefix method will set the prefix element in the MessageBuilder with string p, and
// return the builder
func (b *MessageBuilder) Prefix(p string) *MessageBuilder {
	b.prefix = p
	return b
}

// Sub method will set the sub-prefix element in the MessageBuilder with string s, and
// return the builder
func (b *MessageBuilder) Sub(s string) *MessageBuilder {
	b.sub = s
	return b
}

// Message method will set the message element in the MessageBuilder with string m, and
// return the builder
func (b *MessageBuilder) Message(m string) *MessageBuilder {
	b.msg = m
	return b
}

// Level method will set the level element in the MessageBuilder with LogLevel l, and
// return the builder
func (b *MessageBuilder) Level(l LogLevel) *MessageBuilder {
	b.level = l.String()
	return b
}

// Metadata method will set (or add) the metadata element in the MessageBuilder
// with map m, and return the builder
func (b *MessageBuilder) Metadata(m map[string]interface{}) *MessageBuilder {
	if b.metadata == nil {
		b.metadata = m
	} else {
		for k, v := range m {
			b.metadata[k] = v
		}
	}
	return b
}

// CallStack method will grab the current call stack, and add it as a "callstack" object
// in the MessageBuilder's metadata.
func (b *MessageBuilder) CallStack() *MessageBuilder {
	if b.metadata == nil {
		b.metadata = map[string]interface{}{}
	}
	b.metadata["callstack"] = newCallStack().
		getCallStack(false).
		splitCallStack().
		parseCallStack().
		mapCallStack().
		toField()

	return b
}

// Build method will create a new timestamp, review all elements in the `MessageBuilder`,
// apply any defaults to non-defined elements, and return a pointer to a LogMessage
func (b *MessageBuilder) Build() *LogMessage {
	// b.time = now.Format(time.RFC3339Nano)
	b.time = time.Now()

	if b.prefix == "" {
		b.prefix = "log"
	}

	if b.level == "" {
		b.level = LLInfo.String()
	}

	return &LogMessage{
		Time:     b.time,
		Prefix:   b.prefix,
		Sub:      b.sub,
		Level:    b.level,
		Msg:      b.msg,
		Metadata: b.metadata,
	}
}

func (l *Logger) checkDefaults(m *LogMessage) {
	// use logger prefix if default
	// do not clear Logger.prefix
	if m.Prefix == "log" && l.prefix != m.Prefix {
		m.Prefix = l.prefix
	}

	// use logger sub-prefix if default
	// do not clear Logger.sub
	if m.Sub == "" && l.sub != m.Sub {
		m.Sub = l.sub
	}

	// push logger metadata to message
	if m.Metadata == nil && l.meta != nil {
		m.Metadata = l.meta
	} else if m.Metadata != nil && l.meta != nil {
		// add Logger metadata to existing metadata
		for k, v := range l.meta {
			m.Metadata[k] = v
		}
	}
}

// Output method will take in a pointer to a LogMessage, apply defaults to any unset elements
// (or add its metadata to the message), format it -- and lastly to write it in the output io.Writer
//
// The `Output()` method is the the placeholder action to write a generic message to the logger's io.Writer
//
// All printing messages are either applying a `Logger.Log()` action or a `Logger.Output` one; while the former
// is simply calling the latter.
func (l *Logger) Output(m *LogMessage) (n int, err error) {

	if l.levelFilter > logTypeKeys[m.Level] {
		return 0, nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.checkDefaults(m)

	// format message
	buf, err := l.fmt.Format(m)

	if err != nil {
		return -1, err
	}

	l.buf = buf

	// write message to outs
	n, err = l.out.Write(l.buf)

	if err != nil {
		return n, err
	}
	return n, err
}

// Print method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern
//
// It applies LogLevel Info
func (l *Logger) Print(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLInfo).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Println method (similar to fmt.Println) will print a message using an fmt.Sprintln(v...) pattern
//
// It applies LogLevel Info
func (l *Logger) Println(v ...interface{}) {

	// build message
	log := NewMessage().Level(LLInfo).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Printf method (similar to fmt.Printf) will print a message using an fmt.Sprintf(format, v...) pattern
//
// It applies LogLevel Info
func (l *Logger) Printf(format string, v ...interface{}) {
	// build message
	log := NewMessage().Level(LLInfo).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Log method will take in a pointer to a LogMessage, and write it to the Logger's io.Writer
// without returning an error message.
//
// While the resulting error message of running `Logger.Output()` is simply ignored, this is done
// as a blind-write for this Logger. Since this package also supports creation (and maintainance) of
// Logfiles, this is assumed to be safe.
func (l *Logger) Log(m *LogMessage) {
	// replace defaults if logger has them set
	if m.Prefix == "log" && l.prefix != "" {
		m.Prefix = l.prefix
	}

	// replace defaults if logger has them set
	if m.Metadata == nil && l.meta != nil {
		m.Metadata = l.meta
	}

	s := m.Msg
	l.Output(m)

	if !l.IsSkipExit() && m.Level == LLPanic.String() {
		panic(s)
	} else if !l.IsSkipExit() && m.Level == LLFatal.String() {
		os.Exit(1)
	}
}

// Panic method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Panic.
//
// This method will end calling `panic()` with the LogMessage's message content
func (l *Logger) Panic(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLPanic).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)

	if !l.IsSkipExit() {
		panic(log.Msg)
	}
}

// Panicln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern, while
// automatically applying LogLevel Panic.
//
// This method will end calling `panic()` with the LogMessage's message content
func (l *Logger) Panicln(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLPanic).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)

	if !l.IsSkipExit() {
		panic(log.Msg)
	}

}

// Panicf method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern, while
// automatically applying LogLevel Panic.
//
// This method will end calling `panic()` with the LogMessage's message content
func (l *Logger) Panicf(format string, v ...interface{}) {
	// build message
	log := NewMessage().Level(LLPanic).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)

	if !l.IsSkipExit() {
		panic(log.Msg)
	}

}

// Fatal method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Fatal.
//
// This method will end calling `os.Exit(1)`
func (l *Logger) Fatal(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLFatal).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)

	if !l.IsSkipExit() {
		os.Exit(1)
	}
}

// Fatalln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern, while
// automatically applying LogLevel Fatal.
//
// This method will end calling `os.Exit(1)`
func (l *Logger) Fatalln(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLFatal).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)

	if !l.IsSkipExit() {
		os.Exit(1)
	}
}

// Fatalf method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern, while
// automatically applying LogLevel Fatal.
//
// This method will end calling `os.Exit(1)`
func (l *Logger) Fatalf(format string, v ...interface{}) {
	// build message
	log := NewMessage().Level(LLFatal).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)

	if !l.IsSkipExit() {
		os.Exit(1)
	}
}

// Error method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Error.
func (l *Logger) Error(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLError).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Errorln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern, while
// automatically applying LogLevel Error.
func (l *Logger) Errorln(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLError).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Errorf method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern, while
// automatically applying LogLevel Error.
func (l *Logger) Errorf(format string, v ...interface{}) {
	// build message
	log := NewMessage().Level(LLError).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Warn method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Warn.
func (l *Logger) Warn(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLWarn).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)

}

// Warnln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern, while
// automatically applying LogLevel Warn.
func (l *Logger) Warnln(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLWarn).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Warnf method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern, while
// automatically applying LogLevel Warn.
func (l *Logger) Warnf(format string, v ...interface{}) {

	// build message
	log := NewMessage().Level(LLWarn).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Info method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Info.
func (l *Logger) Info(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLInfo).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Infoln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern, while
// automatically applying LogLevel Info.
func (l *Logger) Infoln(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLInfo).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Infof method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern, while
// automatically applying LogLevel Info.
func (l *Logger) Infof(format string, v ...interface{}) {
	// build message
	log := NewMessage().Level(LLInfo).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Debug method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Debug.
func (l *Logger) Debug(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLDebug).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Debugln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern, while
// automatically applying LogLevel Debug.
func (l *Logger) Debugln(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLDebug).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Debugf method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern, while
// automatically applying LogLevel Debug.
func (l *Logger) Debugf(format string, v ...interface{}) {
	// build message
	log := NewMessage().Level(LLDebug).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Trace method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Trace.
func (l *Logger) Trace(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLTrace).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Traceln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern, while
// automatically applying LogLevel Trace.
func (l *Logger) Traceln(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLTrace).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Tracef method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern, while
// automatically applying LogLevel Trace.
func (l *Logger) Tracef(format string, v ...interface{}) {
	// build message
	log := NewMessage().Level(LLTrace).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// Print function (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern
//
// It applies LogLevel Info
func Print(v ...interface{}) {
	std.Print(v...)
}

// Println function (similar to fmt.Println) will print a message using an fmt.Sprintln(v...) pattern
//
// It applies LogLevel Info
func Println(v ...interface{}) {
	std.Println(v...)
}

// Printf function (similar to fmt.Printf) will print a message using an fmt.Sprintf(format, v...) pattern
//
// It applies LogLevel Info
func Printf(format string, v ...interface{}) {
	std.Printf(format, v...)
}

// Log function will take in a pointer to a LogMessage, and write it to the Logger's io.Writer
// without returning an error message.
//
// While the resulting error message of running `Logger.Output()` is simply ignored, this is done
// as a blind-write for this Logger. Since this package also supports creation (and maintainance) of
// Logfiles, this is assumed to be safe.
func Log(m *LogMessage) {
	std.Log(m)
}

// Panic function (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Panic.
//
// This function will end calling `panic()` with the LogMessage's message content
func Panic(v ...interface{}) {
	std.Panic(v...)
}

// Panicln function (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern, while
// automatically applying LogLevel Panic.
//
// This function will end calling `panic()` with the LogMessage's message content
func Panicln(v ...interface{}) {
	std.Panicln(v...)
}

// Panicf function (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern, while
// automatically applying LogLevel Panic.
//
// This function will end calling `panic()` with the LogMessage's message content
func Panicf(format string, v ...interface{}) {
	std.Panicf(format, v...)
}

// Fatal function (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Fatal.
//
// This function will end calling `os.Exit(1)`
func Fatal(v ...interface{}) {
	std.Fatal(v...)
}

// Fatalln function (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern, while
// automatically applying LogLevel Fatal.
//
// This function will end calling `os.Exit(1)`
func Fatalln(v ...interface{}) {
	std.Fatalln(v...)
}

// Fatalf function (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern, while
// automatically applying LogLevel Fatal.
//
// This function will end calling `os.Exit(1)`
func Fatalf(format string, v ...interface{}) {
	std.Fatalf(format, v...)
}

// Error function (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Error.
func Error(v ...interface{}) {
	std.Error(v...)
}

// Errorln function (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern, while
// automatically applying LogLevel Error.
func Errorln(v ...interface{}) {
	std.Errorln(v...)
}

// Errorf function (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern, while
// automatically applying LogLevel Error.
func Errorf(format string, v ...interface{}) {
	std.Errorf(format, v...)
}

// Warn function (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Warn.
func Warn(v ...interface{}) {
	std.Warn(v...)
}

// Warnln function (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern, while
// automatically applying LogLevel Warn.
func Warnln(v ...interface{}) {
	std.Warnln(v...)
}

// Warnf function (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern, while
// automatically applying LogLevel Warn.
func Warnf(format string, v ...interface{}) {
	std.Warnf(format, v...)
}

// Info function (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Info.
func Info(v ...interface{}) {
	std.Info(v...)
}

// Infoln function (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern, while
// automatically applying LogLevel Info.
func Infoln(v ...interface{}) {
	std.Infoln(v...)
}

// Infof function (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern, while
// automatically applying LogLevel Info.
func Infof(format string, v ...interface{}) {
	std.Infof(format, v...)
}

// Debug function (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Debug.
func Debug(v ...interface{}) {
	std.Debug(v...)
}

// Debugln function (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern, while
// automatically applying LogLevel Debug.
func Debugln(v ...interface{}) {
	std.Debugln(v...)
}

// Debugf function (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern, while
// automatically applying LogLevel Debug.
func Debugf(format string, v ...interface{}) {
	std.Debugf(format, v...)
}

// Trace function (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Trace.
func Trace(v ...interface{}) {
	std.Trace(v...)
}

// Traceln function (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern, while
// automatically applying LogLevel Trace.
func Traceln(v ...interface{}) {
	std.Traceln(v...)
}

// Tracef function (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern, while
// automatically applying LogLevel Trace.
func Tracef(format string, v ...interface{}) {
	std.Tracef(format, v...)
}
