package log

import (
	"fmt"
	"os"
	"time"
)

// var logTypeKeys = map[string]int{
// 	"trace": 0,
// 	"debug": 1,
// 	"info":  2,
// 	"warn":  3,
// 	"error": 4,
// 	"fatal": 5,
// 	"panic": 9,
// }

var logTypeVals = map[int]string{
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

func (l *Logger) Output(level int, msg string) error {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	log := &LogMessage{
		Time:     now.Format(time.RFC3339Nano),
		Prefix:   l.prefix,
		Level:    logTypeVals[level],
		Msg:      msg,
		Metadata: l.meta,
	}
	// clear metadata
	l.meta = nil

	buf, err := l.fmt.Format(log)

	if err != nil {
		return err
	}

	l.buf = buf

	_, err = l.out.Write(l.buf)

	if err != nil {
		return err
	}
	return nil
}

// print methods

func (l *Logger) Print(v ...interface{}) {
	l.Output(2, fmt.Sprint(v...))
}

func (l *Logger) Println(v ...interface{}) {
	l.Output(2, fmt.Sprintln(v...))
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf(format, v...))
}

// log methods

func (l *Logger) Log(level int, v ...interface{}) {
	l.Output(level, fmt.Sprint(v...))
}

func (l *Logger) Logln(level int, v ...interface{}) {
	l.Output(level, fmt.Sprintln(v...))
}

func (l *Logger) Logf(level int, format string, v ...interface{}) {
	l.Output(level, fmt.Sprintf(format, v...))
}

// panic methods

func (l *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.Output(9, s)
	panic(s)
}

func (l *Logger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	l.Output(9, s)
	panic(s)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.Output(9, s)
	panic(s)
}

// fatal methods

func (l *Logger) Fatal(v ...interface{}) {
	l.Output(5, fmt.Sprint(v...))
	os.Exit(1)
}

func (l *Logger) Fatalln(v ...interface{}) {
	l.Output(5, fmt.Sprintln(v...))
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Output(5, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// error methods

func (l *Logger) Error(v ...interface{}) {
	l.Output(4, fmt.Sprint(v...))
}

func (l *Logger) Errorln(v ...interface{}) {
	l.Output(4, fmt.Sprintln(v...))
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Output(4, fmt.Sprintf(format, v...))
}

// warn methods

func (l *Logger) Warn(v ...interface{}) {
	l.Output(3, fmt.Sprint(v...))
}

func (l *Logger) Warnln(v ...interface{}) {
	l.Output(3, fmt.Sprintln(v...))
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Output(3, fmt.Sprintf(format, v...))
}

// info methods

func (l *Logger) Info(v ...interface{}) {
	l.Output(2, fmt.Sprint(v...))
}

func (l *Logger) Infoln(v ...interface{}) {
	l.Output(2, fmt.Sprintln(v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf(format, v...))
}

// debug methods

func (l *Logger) Debug(v ...interface{}) {
	l.Output(1, fmt.Sprint(v...))
}

func (l *Logger) Debugln(v ...interface{}) {
	l.Output(1, fmt.Sprintln(v...))
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Output(1, fmt.Sprintf(format, v...))
}

// trace methods

func (l *Logger) Trace(v ...interface{}) {
	l.Output(0, fmt.Sprint(v...))
}

func (l *Logger) Traceln(v ...interface{}) {
	l.Output(0, fmt.Sprintln(v...))
}

func (l *Logger) Tracef(format string, v ...interface{}) {
	l.Output(0, fmt.Sprintf(format, v...))
}

// print functions

func Print(v ...interface{}) {
	std.Output(2, fmt.Sprint(v...))
}

func Println(v ...interface{}) {
	std.Output(2, fmt.Sprintln(v...))
}

func Printf(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf(format, v...))
}

// log methods

func Log(level int, v ...interface{}) {
	std.Output(level, fmt.Sprint(v...))
}

func Logln(level int, v ...interface{}) {
	std.Output(level, fmt.Sprintln(v...))
}

func Logf(level int, format string, v ...interface{}) {
	std.Output(level, fmt.Sprintf(format, v...))
}

// panic functions

func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	std.Output(9, s)
	panic(s)
}

func Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	std.Output(9, s)
	panic(s)
}

func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	std.Output(9, s)
	panic(s)
}

// fatal functions

func Fatal(v ...interface{}) {
	std.Output(5, fmt.Sprint(v...))
	os.Exit(1)
}

func Fatalln(v ...interface{}) {
	std.Output(5, fmt.Sprintln(v...))
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	std.Output(5, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// error functions

func Error(v ...interface{}) {
	std.Output(4, fmt.Sprint(v...))
}

func Errorln(v ...interface{}) {
	std.Output(4, fmt.Sprintln(v...))
}

func Errorf(format string, v ...interface{}) {
	std.Output(4, fmt.Sprintf(format, v...))
}

// warn functions

func Warn(v ...interface{}) {
	std.Output(3, fmt.Sprint(v...))
}

func Warnln(v ...interface{}) {
	std.Output(3, fmt.Sprintln(v...))
}

func Warnf(format string, v ...interface{}) {
	std.Output(3, fmt.Sprintf(format, v...))
}

// info functions

func Info(v ...interface{}) {
	std.Output(2, fmt.Sprint(v...))
}

func Infoln(v ...interface{}) {
	std.Output(2, fmt.Sprintln(v...))
}

func Infof(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf(format, v...))
}

// debug functions

func Debug(v ...interface{}) {
	std.Output(1, fmt.Sprint(v...))
}

func Debugln(v ...interface{}) {
	std.Output(1, fmt.Sprintln(v...))
}

func Debugf(format string, v ...interface{}) {
	std.Output(1, fmt.Sprintf(format, v...))
}

// trace functions

func Trace(v ...interface{}) {
	std.Output(0, fmt.Sprint(v...))
}

func Traceln(v ...interface{}) {
	std.Output(0, fmt.Sprintln(v...))
}

func Tracef(format string, v ...interface{}) {
	std.Output(0, fmt.Sprintf(format, v...))
}
