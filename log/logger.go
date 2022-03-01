package log

import (
	"io"
	"os"
	"sync"
)

type LoggerI interface {
	Output(m *LogMessage) error
	SetOuts(outs ...io.Writer) LoggerI
	AddOuts(outs ...io.Writer) LoggerI
	Prefix(prefix string) LoggerI
	Fields(fields map[string]interface{}) LoggerI

	Log(m *LogMessage)

	Print(v ...interface{})
	Println(v ...interface{})
	Printf(format string, v ...interface{})

	Panic(v ...interface{})
	Panicln(v ...interface{})
	Panicf(format string, v ...interface{})

	Fatal(v ...interface{})
	Fatalln(v ...interface{})
	Fatalf(format string, v ...interface{})

	Error(v ...interface{})
	Errorln(v ...interface{})
	Errorf(format string, v ...interface{})

	Warn(v ...interface{})
	Warnln(v ...interface{})
	Warnf(format string, v ...interface{})

	Info(v ...interface{})
	Infoln(v ...interface{})
	Infof(format string, v ...interface{})

	Debug(v ...interface{})
	Debugln(v ...interface{})
	Debugf(format string, v ...interface{})

	Trace(v ...interface{})
	Traceln(v ...interface{})
	Tracef(format string, v ...interface{})
}

var std = New("log", TextFormat, os.Stdout)

type LoggerBuilder struct {
	out    io.Writer
	prefix string
	fmt    LogFormatter
}

func NewLogger(confs ...LoggerConfig) LoggerI {
	builder := &LoggerBuilder{}

	if len(confs) == 0 {
		DefaultConfig.Apply(builder)
	} else {
		for _, conf := range confs {
			conf.Apply(builder)
		}
	}

	return &Logger{
		out:    builder.out,
		prefix: builder.prefix,
		fmt:    builder.fmt,
	}
}

type Logger struct {
	mu     sync.Mutex
	out    io.Writer
	buf    []byte
	prefix string
	fmt    LogFormatter
	meta   map[string]interface{}
}

func New(prefix string, format LogFormatter, outs ...io.Writer) LoggerI {
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

// output setter methods

func (l *Logger) SetOuts(outs ...io.Writer) LoggerI {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.out = io.MultiWriter(outs...)

	return l
}

func (l *Logger) AddOuts(outs ...io.Writer) LoggerI {
	l.mu.Lock()
	defer l.mu.Unlock()

	var writers []io.Writer = outs
	writers = append(writers, l.out)
	l.out = io.MultiWriter(writers...)

	return l
}

// prefix setter methods

func (l *Logger) Prefix(prefix string) LoggerI {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.prefix = prefix

	return l
}

// metadata methods

func (l *Logger) Fields(fields map[string]interface{}) LoggerI {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.meta = fields

	return l
}
