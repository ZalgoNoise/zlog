package log

import (
	"encoding/json"
	"fmt"
	"io"
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

// var formatTypeKeys = map[string]int{
// 	"text": 0,
// 	"json": 1,
// }

// var formatTypeVals = map[int]string{
// 	0: "text",
// 	1: "json",
// }

// TODO: consider an interface / struct combo instead of func maps
var formatTypeFunc = map[int]func(log *LogMessage) ([]byte, error){
	0: func(log *LogMessage) ([]byte, error) {
		var buf []byte
		message := fmt.Sprintf(
			"[%s]\t[%s] [%s]\t%s",
			log.Time,
			log.Prefix,
			log.Level,
			log.Msg,
		)

		buf = []byte(message)
		return buf, nil
	},
	1: func(log *LogMessage) ([]byte, error) {
		var buf []byte
		// remove trailing newline on JSON format
		if log.Msg[len(log.Msg)-1] == 10 {
			log.Msg = log.Msg[:len(log.Msg)-1]
		}

		data, err := json.Marshal(log)
		if err != nil {
			return nil, err
		}
		buf = data
		return buf, nil
	},
}

type Logger struct {
	out    io.Writer
	buf    []byte
	prefix string
	fmt    int
}

type LogMessage struct {
	Time   string `json:"timestamp,omitempty"`
	Prefix string `json:"service,omitempty"`
	Level  string `json:"level,omitempty"`
	Msg    string `json:"message,omitempty"`
	// Metadata  map[string]interface{} `json:"metadata,omitemtpy"`
}

func New(prefix string, format int, outs ...io.Writer) *Logger {
	var out io.Writer

	if len(outs) == 0 {
		out = os.Stdout
	} else if len(outs) > 0 {
		out = io.MultiWriter(outs...)
	}

	return &Logger{
		out:    out,
		buf:    []byte{},
		prefix: prefix,
		fmt:    format,
	}
}

func (l *Logger) Output(level int, msg string) error {
	now := time.Now()

	log := &LogMessage{
		Time:   now.Format(time.RFC3339Nano),
		Prefix: l.prefix,
		Level:  logTypeVals[level],
		Msg:    msg,
	}

	buf, err := formatMessage(l.fmt, log)

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

func formatMessage(format int, log *LogMessage) ([]byte, error) {
	return formatTypeFunc[format](log)
}

// output setter methods

func (l *Logger) SetOuts(outs ...io.Writer) {
	l.out = io.MultiWriter(outs...)
}

func (l *Logger) AddOuts(outs ...io.Writer) {
	var writers []io.Writer = outs
	writers = append(writers, l.out)
	l.out = io.MultiWriter(writers...)
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
