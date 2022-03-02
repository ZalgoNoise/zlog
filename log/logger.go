package log

import (
	"io"
	"sync"
)

// LoggerI interface defines the general behavior of a Logger object
//
// It lists all the methods that a Logger implements in order to print
// timestamped messages to an io.Writer, and additional configuration
// methods to enhance its behavior and application (such as `Prefix()`
// and `Fields()`; and `SetOuts()` or `AddOuts()`)
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

var std = New(defaultConfig)

// LoggerBuilder struct describes a builder object for Loggers
//
// The LoggerBuilder object will always be the target for configuration
// settings that are applied when building a Logger, and only after
// all elements are set (with defaults or otherwise) it
// is converted / copied into a Logger
type LoggerBuilder struct {
	out    io.Writer
	prefix string
	fmt    LogFormatter
}

// New function allows creating a basic Logger (implementing the LoggerI
// interface).
//
// Its input parameters are a list of objects which implement the
// LoggerConfig interface. These parameters are iterated through via a
// `multiConf` object that applies all configurations to the builder.
//
// Defaults are automatically applied to all elements which aren't defined
// in the input configuration.
func New(confs ...LoggerConfig) LoggerI {
	builder := &LoggerBuilder{}

	MultiConf(confs...).Apply(builder)

	if builder.out == nil {
		StdOutCfg.Apply(builder)
	}

	if builder.fmt == nil {
		TextCfg.Apply(builder)
	}

	if builder.prefix == "" {
		DefPrefixCfg.Apply(builder)
	}

	return &Logger{
		out:    builder.out,
		prefix: builder.prefix,
		fmt:    builder.fmt,
	}
}

// Logger struct describes a basic Logger, which is used to print timestamped messages
// to an io.Writer
type Logger struct {
	mu     sync.Mutex
	out    io.Writer
	buf    []byte
	prefix string
	fmt    LogFormatter
	meta   map[string]interface{}
}

// SetOuts method will set (replace) the defined io.Writer in the Logger with the list of
// io.Writer set as `outs`.
//
// By default, these input io.Writer will be processed with an io.MultiWriter call to create
// a wrapper for multiple io.Writers
func (l *Logger) SetOuts(outs ...io.Writer) LoggerI {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.out = io.MultiWriter(outs...)

	return l
}

// AddOuts method will add (append) the list of io.Writer set as `outs` to the defined
// ioWriter in the logger
//
// By default, these input io.Writer will be processed with an io.MultiWriter call to create
// a wrapper for multiple io.Writers
func (l *Logger) AddOuts(outs ...io.Writer) LoggerI {
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
func (l *Logger) Prefix(prefix string) LoggerI {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.prefix = prefix

	return l
}

// Fields method will set Logger-scoped (as opposed to message-scoped) metadata fields to the logger
//
// Logger-scoped metadata can be set in order to avoid calling the `MessageBuilder.Metadata()` method
// repeatedly, and instead doing so via the logger at the beginning of a service or function.
//
// Logger-scoped metadata fields are not cleared with new log messages, but only added to the existing
// metadata map. These can be cleared with a call to `Logger.Fields(nil)`
func (l *Logger) Fields(fields map[string]interface{}) LoggerI {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.meta = fields

	return l
}
