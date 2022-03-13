package log

import (
	"fmt"
	"os"
)

type Printer interface {
	Output(m *LogMessage) (n int, err error)
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

// Output method is similar to a Logger.Output() method, however the multiLogger will
// range through all of its configured loggers and execute the same Output() method call
// on each of them
func (l *multiLogger) Output(m *LogMessage) (n int, err error) {
	for _, logger := range l.loggers {
		n, err := logger.Output(m)
		if err != nil {
			return n, err
		}
	}
	return n, err
}

// Print method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern
// across all configured Loggers
//
// It applies LogLevel Info
func (l *multiLogger) Print(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Print(v...)
	}
}

// Println method (similar to fmt.Println) will print a message using an fmt.Sprintln(v...) pattern
// across all configured Loggers
//
// It applies LogLevel Info
func (l *multiLogger) Println(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Println(v...)
	}
}

// Printf method (similar to fmt.Printf) will print a message using an fmt.Sprintf(format, v...) pattern
// across all configured Loggers
//
// It applies LogLevel Info
func (l *multiLogger) Printf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Printf(format, v...)
	}
}

// Log method will take in a pointer to a LogMessage, and write it to each Logger's io.Writer
// without returning an error message.
//
// While the resulting error message of running `Logger.Output()` is simply ignored, this is done
// as a blind-write for this Logger. Since this package also supports creation (and maintainance) of
// Logfiles, this is assumed to be safe.
func (l *multiLogger) Log(m *LogMessage) {
	for _, logger := range l.loggers {
		logger.Log(m)
	}
}

// Panic method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Panic.
//
// This method will end calling `panic()` with the LogMessage's message content
func (l *multiLogger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)

	for _, logger := range l.loggers {
		logger.Output(
			NewMessage().
				Level(LLPanic).
				Message(s).
				Build(),
		)
	}

	if !l.IsSkipExit() {
		panic(s)
	}
}

// Panicln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Panic.
//
// This method will end calling `panic()` with the LogMessage's message content
func (l *multiLogger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)

	for _, logger := range l.loggers {
		logger.Output(NewMessage().Level(LLPanic).Message(s).Build())
	}

	if !l.IsSkipExit() {
		panic(s)
	}
}

// Panicf method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern
// across all configured Loggers, while automatically applying LogLevel Panic.
//
// This method will end calling `panic()` with the LogMessage's message content
func (l *multiLogger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)

	for _, logger := range l.loggers {
		logger.Output(NewMessage().Level(LLPanic).Message(s).Build())
	}

	if !l.IsSkipExit() {
		panic(s)
	}
}

// Fatal method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Fatal.
//
// This method will end calling `os.Exit(1)`
func (l *multiLogger) Fatal(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Output(NewMessage().Level(LLFatal).Message(fmt.Sprint(v...)).Build())
	}

	if !l.IsSkipExit() {
		os.Exit(1)
	}
}

// Fatalln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Fatal.
//
// This method will end calling `os.Exit(1)`
func (l *multiLogger) Fatalln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Output(NewMessage().Level(LLFatal).Message(fmt.Sprintln(v...)).Build())
	}

	if !l.IsSkipExit() {
		os.Exit(1)
	}
}

// Fatalf method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern
// across all configured Loggers, while automatically applying LogLevel Fatal.
//
// This method will end calling `os.Exit(1)`
func (l *multiLogger) Fatalf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Output(NewMessage().Level(LLFatal).Message(fmt.Sprintf(format, v...)).Build())
	}

	if !l.IsSkipExit() {
		os.Exit(1)
	}
}

// Error method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Error.
func (l *multiLogger) Error(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Error(v...)
	}
}

// Errorln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Error.
func (l *multiLogger) Errorln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Errorln(v...)
	}
}

// Errorf method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern
// across all configured Loggers, while automatically applying LogLevel Error.
func (l *multiLogger) Errorf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Errorf(format, v...)
	}
}

// Warn method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Warn.
func (l *multiLogger) Warn(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Warn(v...)
	}
}

// Warnln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Warn.
func (l *multiLogger) Warnln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Warnln(v...)
	}
}

// Warnf method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern
// across all configured Loggers, while automatically applying LogLevel Warn.
func (l *multiLogger) Warnf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Warnf(format, v...)
	}
}

// Info method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Info.
func (l *multiLogger) Info(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Info(v...)
	}

}

// Infoln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Info.
func (l *multiLogger) Infoln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Infoln(v...)
	}
}

// Infof method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern
// across all configured Loggers, while automatically applying LogLevel Info.
func (l *multiLogger) Infof(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Infof(format, v...)
	}
}

// Debug method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Debug.
func (l *multiLogger) Debug(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Debug(v...)
	}
}

// Debugln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Debug.
func (l *multiLogger) Debugln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Debugln(v...)
	}
}

// Debugf method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern
// across all configured Loggers, while automatically applying LogLevel Debug.
func (l *multiLogger) Debugf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Debugf(format, v...)
	}
}

// Trace method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Trace.
func (l *multiLogger) Trace(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Trace(v...)
	}
}

// Traceln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Trace.
func (l *multiLogger) Traceln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Traceln(v...)
	}
}

// Tracef method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern
// across all configured Loggers, while automatically applying LogLevel Trace.
func (l *multiLogger) Tracef(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Tracef(format, v...)
	}
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
