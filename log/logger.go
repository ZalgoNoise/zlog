package log

import (
	"io"
	"os"
	"sync"
)

var std = New("log", &TextFmt{}, os.Stdout)

type Logger struct {
	mu     sync.Mutex
	out    io.Writer
	buf    []byte
	prefix string
	fmt    LogFormatter
	meta   map[string]interface{}
}

func New(prefix string, format LogFormatter, outs ...io.Writer) *Logger {
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

func (l *Logger) SetPrefix(prefix string) LoggerI {
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
