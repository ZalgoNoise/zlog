package log

import (
	"bytes"
	"encoding/gob"
	"io"
	"os"
	"sync"
)

// Logger interface defines the general behavior of a Logger object
//
// It lists all the methods that a Logger implements in order to print
// timestamped messages to an io.Writer, and additional configuration
// methods to enhance its behavior and application (such as `Prefix()`
// and `Fields()`; and `SetOuts()` or `AddOuts()`)
type Logger interface {
	io.Writer
	Printer

	SetOuts(outs ...io.Writer) Logger
	AddOuts(outs ...io.Writer) Logger
	Prefix(prefix string) Logger
	Sub(sub string) Logger
	Fields(fields map[string]interface{}) Logger
	IsSkipExit() bool
}

var std = New(defaultConfig)
var stdout = os.Stdout

// LoggerBuilder struct describes a builder object for Loggers
//
// The LoggerBuilder object will always be the target for configuration
// settings that are applied when building a Logger, and only after
// all elements are set (with defaults or otherwise) it
// is converted / copied into a Logger
type LoggerBuilder struct {
	out         io.Writer
	prefix      string
	sub         string
	fmt         LogFormatter
	skipExit    bool
	levelFilter int
}

// New function allows creating a basic Logger (implementing the Logger
// interface).
//
// Its input parameters are a list of objects which implement the
// LoggerConfig interface. These parameters are iterated through via a
// `multiConf` object that applies all configurations to the builder.
//
// Defaults are automatically applied to all elements which aren't defined
// in the input configuration.
func New(confs ...LoggerConfig) Logger {
	builder := &LoggerBuilder{}

	// enforce defaults
	defaultConfig.Apply(builder)

	MultiConf(confs...).Apply(builder)

	return &logger{
		out:         builder.out,
		prefix:      builder.prefix,
		sub:         builder.sub,
		fmt:         builder.fmt,
		skipExit:    builder.skipExit,
		levelFilter: builder.levelFilter,
	}
}

// logger struct describes a basic Logger, which is used to print timestamped messages
// to an io.Writer
type logger struct {
	mu          sync.Mutex
	out         io.Writer
	buf         []byte
	prefix      string
	sub         string
	fmt         LogFormatter
	meta        map[string]interface{}
	skipExit    bool
	levelFilter int
}

// SetOuts method will set (replace) the defined io.Writer in the Logger with the list of
// io.Writer set as `outs`.
//
// By default, these input io.Writer will be processed with an io.MultiWriter call to create
// a wrapper for multiple io.Writers
func (l *logger) SetOuts(outs ...io.Writer) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(outs) == 0 {
		l.out = stdout
		return l
	}

	l.out = io.MultiWriter(outs...)

	return l
}

// AddOuts method will add (append) the list of io.Writer set as `outs` to the defined
// ioWriter in the logger
//
// By default, these input io.Writer will be processed with an io.MultiWriter call to create
// a wrapper for multiple io.Writers
func (l *logger) AddOuts(outs ...io.Writer) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	var writers []io.Writer = outs
	writers = append(writers, l.out)
	l.out = io.MultiWriter(writers...)

	return l
}

// Prefix method will set a Logger-scoped (as opposed to message-scoped) prefix string to the logger
//
// Logger-scoped prefix strings can be set in order to avoid calling the `MessageBuilder.Prefix()` method
// repeatedly, and instead doing so via the logger at the beginning of a service or function
//
// A logger-scoped prefix is not cleared with new Log messages, but `MessageBuilder.Prefix()` calls will
// replace it.
func (l *logger) Prefix(prefix string) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	if prefix == "" {
		l.prefix = "log"
		return l
	}

	l.prefix = prefix

	return l
}

// Sub method will set a Logger-scoped (as opposed to message-scoped) sub-prefix string to the logger
//
//
// Logger-scoped sub-prefix strings can be set in order to avoid calling the `MessageBuilder.Sub()` method
// repeatedly, and instead doing so via the logger at the beginning of a service or function
//
// A logger-scoped sub-prefix is not cleared with new Log messages, but `MessageBuilder.Sub()` calls will
// replace it.
func (l *logger) Sub(sub string) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.sub = sub

	return l
}

// Fields method will set Logger-scoped (as opposed to message-scoped) metadata fields to the logger
//
// Logger-scoped metadata can be set in order to avoid calling the `MessageBuilder.Metadata()` method
// repeatedly, and instead doing so via the logger at the beginning of a service or function.
//
// Logger-scoped metadata fields are not cleared with new log messages, but only added to the existing
// metadata map. These can be cleared with a call to `Logger.Fields(nil)`
func (l *logger) Fields(fields map[string]interface{}) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.meta = fields

	return l
}

// IsSkipExit method returns a boolean on whether this logger is set to skip os.Exit(1)
// or panic() calls.
//
// It is used in functions which call these, to first evaluate if those calls should be
// executed or skipped
func (l *logger) IsSkipExit() bool {
	return l.skipExit
}

// Write method implements the io.Writer interface, to allow a logger to be used in a more
// abstract way, simply as a writer.
//
// To allow support for LogMessages, these can be gob-encoded and passed into this function
// by calling its Bytes() method.
//
// A gob-encoded LogMessage can be decoded by a Logger serving as an io.Writer; and correctly
// format the message to be written with all fields it contains.
//
// Otherwise, if a simple slice of bytes is passed, it is considered to be the LogMessage.Msg
// portion, and the remaining fields will default to the Logger's set elements
func (l *logger) Write(p []byte) (n int, err error) {
	// check if it's gob-encoded
	m := &LogMessage{}

	buf := bytes.NewBuffer(p)
	dec := gob.NewDecoder(buf)

	err = dec.Decode(m)

	if err != nil {
		// default to printing message
		return l.Output(NewMessage().
			Level(LLInfo).
			Prefix(l.prefix).
			Sub(l.sub).
			Message(string(p)).
			Metadata(l.meta).
			Build())
	}

	// print gob-encoded message
	return l.Output(m)

}
