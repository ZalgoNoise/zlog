// Package log will contain logic to implement loggers, which write messages
// to writers with modular configuration. Loggers are defined by the Logger
// interface, to provide both abstraction and extensibility of loggers.
//
// This package covers the native features of the Logger interface, and its objects'
// configuration and API. Web solutions like the gRPC client and server are found
// in their designated packages.
//
// The log package implements the following interfaces:
//
//     type Logger interface {
//         io.Writer
//         Printer
//
//         SetOuts(outs ...io.Writer) Logger
//         AddOuts(outs ...io.Writer) Logger
//         Prefix(prefix string) Logger
//         Sub(sub string) Logger
//         Fields(fields map[string]interface{}) Logger
//         IsSkipExit() bool
//     }
//
//     type Printer interface {
//         Output(m *event.Event) (n int, err error)
//         Log(m ...*event.Event)
//
//         Print(v ...interface{})
//         Println(v ...interface{})
//         Printf(format string, v ...interface{})
//
//         Panic(v ...interface{})
//         Panicln(v ...interface{})
//         Panicf(format string, v ...interface{})
//
//         Fatal(v ...interface{})
//         Fatalln(v ...interface{})
//         Fatalf(format string, v ...interface{})
//
//         Error(v ...interface{})
//         Errorln(v ...interface{})
//         Errorf(format string, v ...interface{})
//
//         Warn(v ...interface{})
//         Warnln(v ...interface{})
//         Warnf(format string, v ...interface{})
//
//         Info(v ...interface{})
//         Infoln(v ...interface{})
//         Infof(format string, v ...interface{})
//
//         Debug(v ...interface{})
//         Debugln(v ...interface{})
//         Debugf(format string, v ...interface{})
//
//         Trace(v ...interface{})
//         Traceln(v ...interface{})
//         Tracef(format string, v ...interface{})
//     }
//
//     type ChanneledLogger interface {
//         Log(msg ...*event.Event)
//         Close()
//         Channels() (logCh chan *event.Event, done chan struct{})
//     }
//
// The remaining interfaces found in this package (LoggerConfig interface, LogFormatter interface)
// are merely for configuring the logger, so these will be mostly useful for extensibility / new features
//
//     type LoggerConfig interface {
//         Apply(lb *LoggerBuilder)
//     }
//
//     type LogFormatter interface {
//         Format(log *LogMessage) (buf []byte, err error)
//         LoggerConfig
//     }
//
package log

import (
	"io"
	"os"
	"sync"

	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/store"
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

var std = New(DefaultConfig)
var stdout = os.Stderr

// LoggerBuilder struct describes a builder object for Loggers
//
// The LoggerBuilder object will always be the target for configuration
// settings that are applied when building a Logger, and only after
// all elements are set (with defaults or otherwise) it
// is converted / copied into a Logger
type LoggerBuilder struct {
	Out         io.Writer
	Prefix      string
	Sub         string
	Fmt         LogFormatter
	SkipExit    bool
	LevelFilter int32
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
	DefaultConfig.Apply(builder)

	MultiConf(confs...).Apply(builder)

	if builder.Out == store.EmptyWriter && builder.SkipExit {
		return &nilLogger{}
	}

	return &logger{
		out:         builder.Out,
		prefix:      builder.Prefix,
		sub:         builder.Sub,
		fmt:         builder.Fmt,
		skipExit:    builder.SkipExit,
		levelFilter: builder.LevelFilter,
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
	levelFilter int32
}

// SetOuts method will set (replace) the defined io.Writer in the Logger with the list of
// io.Writer set as `outs`.
//
// By default, these input io.Writer will be processed with an io.MultiWriter call to create
// a wrapper for multiple io.Writers
func (l *logger) SetOuts(outs ...io.Writer) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	var newouts []io.Writer

	for i := 0; i < len(outs); i++ {
		if outs[i] != nil {
			newouts = append(newouts, outs[i])
		}
	}

	if len(newouts) == 0 {
		l.out = stdout
		return l
	}

	l.out = io.MultiWriter(newouts...)

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

	var newouts []io.Writer

	for i := 0; i < len(outs); i++ {
		if outs[i] != nil {
			newouts = append(newouts, outs[i])
		}
	}

	if len(newouts) == 0 {
		return l
	}

	newouts = append(newouts, l.out)

	l.out = io.MultiWriter(newouts...)

	return l
}

// Prefix method will set a Logger-scoped (as opposed to message-scoped) prefix string to the logger
//
// Logger-scoped prefix strings can be set in order to avoid calling the `event.New().Prefix()` method
// repeatedly, and instead doing so via the logger at the beginning of a service or function
//
// A logger-scoped prefix is not cleared with new Log events, but `event.New().Prefix()` calls will
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
// Logger-scoped sub-prefix strings can be set in order to avoid calling the `event.New().Sub()` method
// repeatedly, and instead doing so via the logger at the beginning of a service or function
//
// A logger-scoped sub-prefix is not cleared with new Log events, but `event.New().Sub()` calls will
// replace it.
func (l *logger) Sub(sub string) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.sub = sub

	return l
}

// Fields method will set Logger-scoped (as opposed to message-scoped) metadata fields to the logger
//
// Logger-scoped metadata can be set in order to avoid calling the `event.New().Metadata()` method
// repeatedly, and instead doing so via the logger at the beginning of a service or function.
//
// Logger-scoped metadata fields are not cleared with new log events, but only added to the existing
// metadata map. These can be cleared with a call to `Logger.Fields(nil)`
func (l *logger) Fields(fields map[string]interface{}) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	if fields == nil {
		l.meta = map[string]interface{}{}
		return l
	}

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
// As the event.Event struct is using protocol buffers, this method will check if the message
// is properly encoded. If it is, it can be decoded by a Logger serving as an io.Writer; and correctly
// format the message to be written with all fields it contains.
//
// Otherwise, if a simple slice of bytes is passed, it is considered to be the event.Event.Msg
// portion, and the remaining fields will default to the Logger's set elements
func (l *logger) Write(p []byte) (n int, err error) {
	// decode bytes
	m, err := event.Decode(p)

	// default to printing message as if it was a byte slice payload for the log event body
	if err != nil {
		return l.Output(event.New().
			Level(event.Default_Event_Level).
			Prefix(l.prefix).
			Sub(l.sub).
			Message(string(p)).
			Metadata(l.meta).
			Build())
	}

	// print message
	return l.Output(m)

}

// nilLogger struct describes an empty Logger, set as a separate type
// mostly for prototyping or testing
type nilLogger struct{}

func (l *nilLogger) Write(p []byte) (n int, err error)           { return 1, nil }
func (l *nilLogger) SetOuts(outs ...io.Writer) Logger            { return l }
func (l *nilLogger) AddOuts(outs ...io.Writer) Logger            { return l }
func (l *nilLogger) Prefix(prefix string) Logger                 { return l }
func (l *nilLogger) Sub(sub string) Logger                       { return l }
func (l *nilLogger) Fields(fields map[string]interface{}) Logger { return l }
func (l *nilLogger) IsSkipExit() bool                            { return true }
func (l *nilLogger) Output(m *event.Event) (n int, err error)    { return 1, nil }
func (l *nilLogger) Log(m ...*event.Event)                       {}
func (l *nilLogger) Print(v ...interface{})                      {}
func (l *nilLogger) Println(v ...interface{})                    {}
func (l *nilLogger) Printf(format string, v ...interface{})      {}
func (l *nilLogger) Panic(v ...interface{})                      {}
func (l *nilLogger) Panicln(v ...interface{})                    {}
func (l *nilLogger) Panicf(format string, v ...interface{})      {}
func (l *nilLogger) Fatal(v ...interface{})                      {}
func (l *nilLogger) Fatalln(v ...interface{})                    {}
func (l *nilLogger) Fatalf(format string, v ...interface{})      {}
func (l *nilLogger) Error(v ...interface{})                      {}
func (l *nilLogger) Errorln(v ...interface{})                    {}
func (l *nilLogger) Errorf(format string, v ...interface{})      {}
func (l *nilLogger) Warn(v ...interface{})                       {}
func (l *nilLogger) Warnln(v ...interface{})                     {}
func (l *nilLogger) Warnf(format string, v ...interface{})       {}
func (l *nilLogger) Info(v ...interface{})                       {}
func (l *nilLogger) Infoln(v ...interface{})                     {}
func (l *nilLogger) Infof(format string, v ...interface{})       {}
func (l *nilLogger) Debug(v ...interface{})                      {}
func (l *nilLogger) Debugln(v ...interface{})                    {}
func (l *nilLogger) Debugf(format string, v ...interface{})      {}
func (l *nilLogger) Trace(v ...interface{})                      {}
func (l *nilLogger) Traceln(v ...interface{})                    {}
func (l *nilLogger) Tracef(format string, v ...interface{})      {}
