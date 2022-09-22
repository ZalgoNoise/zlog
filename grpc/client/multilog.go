package client

import (
	"fmt"
	"io"
	"os"

	"github.com/zalgonoise/zlog/grpc/address"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

type multiLogger struct {
	loggers []GRPCLogger
}

func (ml *multiLogger) addLoggers(l ...GRPCLogger) {
	ml.loggers = make([]GRPCLogger, 0, len(l))
	for _, logger := range l {
		ml.addLogger(logger)
	}
}

func (ml *multiLogger) addLogger(l GRPCLogger) {
	if l == nil {
		return
	}

	if iml, ok := l.(*multiLogger); ok {
		for _, logger := range iml.loggers {
			ml.addLogger(logger)
		}
		return
	}

	ml.loggers = append(ml.loggers, l)
}

func (ml *multiLogger) build() GRPCLogger {
	if len(ml.loggers) == 0 {
		return nil
	}

	if len(ml.loggers) == 1 {
		return ml.loggers[0]
	}

	return ml
}

// MultiLogger function is a wrapper for multiple GRPCLogger
//
// Similar to how io.MultiWriter() is implemented, this function generates a single
// GRPCLogger that targets a set of configured GRPCLogger.
//
// As such, a single GRPCLogger can have multiple GRPCLoggers with different configurations
// and output addresses, while still registering the same log message.
func MultiLogger(loggers ...GRPCLogger) GRPCLogger {

	if len(loggers) == 0 {
		return nil
	}

	if len(loggers) == 1 {
		return loggers[0]
	}

	ml := new(multiLogger)

	ml.addLoggers(loggers...)
	return ml.build()
}

func (l *multiLogger) SetOuts(outs ...io.Writer) log.Logger {
	var o []io.Writer

	for _, remote := range outs {
		if remote == nil {
			continue
		}

		if r, ok := remote.(*address.ConnAddr); !ok {
			continue
		} else {
			o = append(o, r)
		}
	}

	for _, logger := range l.loggers {
		r := logger.SetOuts(o...)

		if r == nil {
			return nil
		}
	}

	return l
}

func (l *multiLogger) AddOuts(outs ...io.Writer) log.Logger {
	var o []io.Writer

	for _, remote := range outs {
		if remote == nil {
			continue
		}

		if r, ok := remote.(*address.ConnAddr); !ok {
			continue
		} else {
			o = append(o, r)
		}
	}

	for _, logger := range l.loggers {
		r := logger.AddOuts(o...)

		if r == nil {
			return nil
		}
	}

	return l
}

func (l *multiLogger) Prefix(prefix string) log.Logger {
	for _, logger := range l.loggers {
		logger.Prefix(prefix)
	}
	return l
}

func (l *multiLogger) Sub(sub string) log.Logger {
	for _, logger := range l.loggers {
		logger.Sub(sub)
	}
	return l
}

func (l *multiLogger) Fields(fields map[string]interface{}) log.Logger {
	for _, logger := range l.loggers {
		logger.Fields(fields)
	}
	return l
}

func (l *multiLogger) IsSkipExit() bool {
	for _, logger := range l.loggers {
		ok := logger.IsSkipExit()
		if !ok {
			return false
		}
	}
	return true
}

func (l *multiLogger) Write(p []byte) (n int, err error) {

	var errs []error

	for idx, logger := range l.loggers {
		n, err = logger.Write(p)

		if err != nil {
			errs = append(errs, err)
		}

		if n == 0 {
			errs = append(errs, fmt.Errorf("logger with index %v wrote %v bytes", idx, n))
		}
	}

	if len(errs) > 0 {
		if len(errs) == 1 {
			return -1, errs[0]
		}

		var err error

		for _, e := range errs {
			if err == nil {
				err = e
			} else {
				err = fmt.Errorf("%w ; %v", err, e)
			}
		}

		return -1, fmt.Errorf("multiple errors when writing message: %w", err)
	}

	return n, nil
}

func (l *multiLogger) Close() {
	for _, logger := range l.loggers {
		logger.Close()
	}
}

func (l *multiLogger) Channels() (chan *event.Event, chan struct{}) {
	logCh := make(chan *event.Event)
	done := make(chan struct{})

	logChSet := make([]chan *event.Event, len(l.loggers))
	doneSet := make([]chan struct{}, len(l.loggers))

	for _, logger := range l.loggers {
		localLog, localDone := logger.Channels()

		logChSet = append(logChSet, localLog)
		doneSet = append(doneSet, localDone)
	}

	go func(logCh chan *event.Event, logChSet []chan *event.Event) {
		for msg := range logCh {
			for _, ch := range logChSet {
				ch <- msg
			}
		}
	}(logCh, logChSet)

	go func(done chan struct{}, doneSet []chan struct{}) {
		for range done {
			for _, ch := range doneSet {
				ch <- struct{}{}
			}
		}
	}(done, doneSet)

	return logCh, done
}

// Output method is similar to a Logger.Output() method, however the multiLogger will
// range through all of its configured loggers and execute the same Output() method call
// on each of them
func (l *multiLogger) Output(m *event.Event) (n int, err error) {
	var firstn int

	for _, logger := range l.loggers {
		n, err := logger.Output(m)

		if firstn == 0 {
			firstn = n
		}

		if err != nil {
			return firstn, err
		}
	}

	return firstn, err
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

// Log method will take in a pointer to a event.Event, and write it to each Logger's io.Writer
// without returning an error message.
//
// While the resulting error message of running `Logger.Output()` is simply ignored, this is done
// as a blind-write for this Logger.
func (l *multiLogger) Log(m ...*event.Event) {
	for _, logger := range l.loggers {
		logger.Log(m...)
	}
}

// Panic method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Panic.
//
// This method will end calling `panic()` with the event.Event's message content, if the logger is not set to
// skip exit calls.
func (l *multiLogger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)

	for _, logger := range l.loggers {
		_, _ = logger.Output( // deliberately ignore error in this method call
			event.New().
				Level(event.Level_panic).
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
// This method will end calling `panic()` with the event.Event's message content, if the logger is not set to
// skip exit calls.
func (l *multiLogger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)

	for _, logger := range l.loggers {
		_, _ = logger.Output(event.New().Level(event.Level_panic).Message(s).Build()) // deliberately ignore error in this method call
	}

	if !l.IsSkipExit() {
		panic(s)
	}
}

// Panicf method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern
// across all configured Loggers, while automatically applying LogLevel Panic.
//
// This method will end calling `panic()` with the event.Event's message content, if the logger is not set to
// skip exit calls.
func (l *multiLogger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)

	for _, logger := range l.loggers {
		_, _ = logger.Output(event.New().Level(event.Level_panic).Message(s).Build()) // deliberately ignore error in this method call
	}

	if !l.IsSkipExit() {
		panic(s)
	}
}

// Fatal method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Fatal.
//
// This method will end calling `os.Exit(1)`, if the logger is not set to skip exit calls.
func (l *multiLogger) Fatal(v ...interface{}) {
	for _, logger := range l.loggers {
		_, _ = logger.Output(event.New().Level(event.Level_fatal).Message(fmt.Sprint(v...)).Build()) // deliberately ignore error in this method call
	}

	if !l.IsSkipExit() {
		os.Exit(1)
	}
}

// Fatalln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Fatal.
//
// This method will end calling `os.Exit(1)`, if the logger is not set to skip exit calls.
func (l *multiLogger) Fatalln(v ...interface{}) {
	for _, logger := range l.loggers {
		_, _ = logger.Output(event.New().Level(event.Level_fatal).Message(fmt.Sprintln(v...)).Build()) // deliberately ignore error in this method call
	}

	if !l.IsSkipExit() {
		os.Exit(1)
	}
}

// Fatalf method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern
// across all configured Loggers, while automatically applying LogLevel Fatal.
//
// This method will end calling `os.Exit(1)`, if the logger is not set to skip exit calls.
func (l *multiLogger) Fatalf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		_, _ = logger.Output(event.New().Level(event.Level_fatal).Message(fmt.Sprintf(format, v...)).Build()) // deliberately ignore error in this method call
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
