package log

import (
	"fmt"
	"os"
	"time"
)

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

func (ll LogLevel) String() string {
	return logTypeVals[ll]
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

type LogMessage struct {
	Time     string                 `json:"timestamp,omitempty"`
	Prefix   string                 `json:"service,omitempty"`
	Level    string                 `json:"level,omitempty"`
	Msg      string                 `json:"message,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type MessageBuilder struct {
	time     string
	prefix   string
	level    string
	msg      string
	metadata map[string]interface{}
}

func NewMessage() *MessageBuilder {
	return &MessageBuilder{}
}

func (b *MessageBuilder) Prefix(p string) *MessageBuilder {
	b.prefix = p
	return b
}

func (b *MessageBuilder) Message(m string) *MessageBuilder {
	b.msg = m
	return b
}

func (b *MessageBuilder) Level(l LogLevel) *MessageBuilder {
	b.level = l.String()
	return b
}

func (b *MessageBuilder) Metadata(m map[string]interface{}) *MessageBuilder {
	b.metadata = m
	return b
}

func (b *MessageBuilder) Build() *LogMessage {
	now := time.Now()

	b.time = now.Format(time.RFC3339Nano)

	if b.prefix == "" {
		b.prefix = "log"
	}

	if b.level == "" {
		b.level = LLInfo.String()
	}

	return &LogMessage{
		Time:     b.time,
		Prefix:   b.prefix,
		Level:    b.level,
		Msg:      b.msg,
		Metadata: b.metadata,
	}
}

// func (l *Logger) Output(level LogLevel, msg string) error {

// 	l.mu.Lock()
// 	defer l.mu.Unlock()

// 	// build message
// 	log := NewMessage().Level(level).Prefix(l.prefix).Message(msg).Metadata(l.meta).Build()

// 	// clear metadata
// 	l.meta = nil

// 	// format message
// 	buf, err := l.fmt.Format(log)

// 	if err != nil {
// 		return err
// 	}

// 	l.buf = buf

// 	// write message to outs
// 	_, err = l.out.Write(l.buf)

// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (l *Logger) Output(m *LogMessage) error {

	l.mu.Lock()
	defer l.mu.Unlock()

	// use logger prefix if default
	if m.Prefix == "log" && l.prefix != m.Prefix {
		m.Prefix = l.prefix
	}

	// clear metadata
	if m.Metadata == nil && l.meta != nil {
		m.Metadata = l.meta
	}
	l.meta = nil

	// format message
	buf, err := l.fmt.Format(m)

	if err != nil {
		return err
	}

	l.buf = buf

	// write message to outs
	_, err = l.out.Write(l.buf)

	if err != nil {
		return err
	}
	return nil
}

// print methods

func (l *Logger) Print(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLInfo).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

func (l *Logger) Println(v ...interface{}) {

	// build message
	log := NewMessage().Level(LLInfo).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

func (l *Logger) Printf(format string, v ...interface{}) {
	// build message
	log := NewMessage().Level(LLInfo).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// log methods

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

	if m.Level == LLPanic.String() {
		panic(s)
	} else if m.Level == LLFatal.String() {
		os.Exit(1)
	}
}

// panic methods

func (l *Logger) Panic(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLPanic).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)

	panic(log.Msg)
}

func (l *Logger) Panicln(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLPanic).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)

	panic(log.Msg)

}

func (l *Logger) Panicf(format string, v ...interface{}) {
	// build message
	log := NewMessage().Level(LLPanic).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)

	panic(log.Msg)

}

// fatal methods

func (l *Logger) Fatal(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLFatal).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)

	os.Exit(1)
}

func (l *Logger) Fatalln(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLFatal).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)

	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	// build message
	log := NewMessage().Level(LLFatal).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)

	os.Exit(1)
}

// error methods

func (l *Logger) Error(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLError).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

func (l *Logger) Errorln(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLError).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	// build message
	log := NewMessage().Level(LLError).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// warn methods

func (l *Logger) Warn(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLWarn).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)

}

func (l *Logger) Warnln(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLWarn).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

func (l *Logger) Warnf(format string, v ...interface{}) {

	// build message
	log := NewMessage().Level(LLWarn).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// info methods

func (l *Logger) Info(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLInfo).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

func (l *Logger) Infoln(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLInfo).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	// build message
	log := NewMessage().Level(LLInfo).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// debug methods

func (l *Logger) Debug(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLDebug).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

func (l *Logger) Debugln(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLDebug).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	// build message
	log := NewMessage().Level(LLDebug).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// trace methods

func (l *Logger) Trace(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLTrace).Prefix(l.prefix).Message(
		fmt.Sprint(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

func (l *Logger) Traceln(v ...interface{}) {
	// build message
	log := NewMessage().Level(LLTrace).Prefix(l.prefix).Message(
		fmt.Sprintln(v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

func (l *Logger) Tracef(format string, v ...interface{}) {
	// build message
	log := NewMessage().Level(LLTrace).Prefix(l.prefix).Message(
		fmt.Sprintf(format, v...),
	).Metadata(l.meta).Build()

	l.Output(log)
}

// print functions

func Print(v ...interface{}) {
	std.Print(v...)
}

func Println(v ...interface{}) {
	std.Println(v...)
}

func Printf(format string, v ...interface{}) {
	std.Printf(format, v...)
}

// log methods

func Log(m *LogMessage) {
	std.Log(m)
}

// panic functions

func Panic(v ...interface{}) {
	std.Panic(v...)
}

func Panicln(v ...interface{}) {
	std.Panicln(v...)
}

func Panicf(format string, v ...interface{}) {
	std.Panicf(format, v...)
}

// fatal functions

func Fatal(v ...interface{}) {
	std.Fatal(v...)
}

func Fatalln(v ...interface{}) {
	std.Fatalln(v...)
}

func Fatalf(format string, v ...interface{}) {
	std.Fatalf(format, v...)
}

// error functions

func Error(v ...interface{}) {
	std.Error(v...)
}

func Errorln(v ...interface{}) {
	std.Errorln(v...)
}

func Errorf(format string, v ...interface{}) {
	std.Errorf(format, v...)
}

// warn functions

func Warn(v ...interface{}) {
	std.Warn(v...)
}

func Warnln(v ...interface{}) {
	std.Warnln(v...)
}

func Warnf(format string, v ...interface{}) {
	std.Warnf(format, v...)
}

// info functions

func Info(v ...interface{}) {
	std.Info(v...)
}

func Infoln(v ...interface{}) {
	std.Infoln(v...)
}

func Infof(format string, v ...interface{}) {
	std.Infof(format, v...)
}

// debug functions

func Debug(v ...interface{}) {
	std.Debug(v...)
}

func Debugln(v ...interface{}) {
	std.Debugln(v...)
}

func Debugf(format string, v ...interface{}) {
	std.Debugf(format, v...)
}

// trace functions

func Trace(v ...interface{}) {
	std.Trace(v...)
}

func Traceln(v ...interface{}) {
	std.Traceln(v...)
}

func Tracef(format string, v ...interface{}) {
	std.Tracef(format, v...)
}
