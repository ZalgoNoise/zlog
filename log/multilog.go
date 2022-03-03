package log

import (
	"fmt"
	"io"
	"os"
)

type multiLogger struct {
	loggers []LoggerI
}

// MultiLogger function is a wrapper for multiple LoggerI
//
// Similar to how io.MultiWriter() is implemented, this function generates a single
// LoggerI that targets a set of configured LoggerI.
//
// As such, a single LoggerI can have multiple Loggers with different configurations and
// output files, while still registering the same log message.
func MultiLogger(loggers ...LoggerI) LoggerI {
	allLoggers := make([]LoggerI, 0, len(loggers))
	allLoggers = append(allLoggers, loggers...)

	return &multiLogger{allLoggers}
}

// Output method is similar to a Logger.Output() method, however the multiLogger will
// range through all of its configured loggers and execute the same Output() method call
// on each of them
func (l *multiLogger) Output(m *LogMessage) error {
	for _, logger := range l.loggers {
		err := logger.Output(m)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetOuts method is similar to a Logger.SetOuts() method, however the multiLogger will
// range through all of its configured loggers and execute the same SetOuts() method call
// on each of them
//
// This method has been created to comply with the LoggerI interface -- but it's worth underlining
// that it will overwrite all the loggers' outs. This may not be exactly the best action
// considering if there are different formats or more than one logger, it will result in
// different types of messages and / or repeated ones, respectively.
func (l *multiLogger) SetOuts(outs ...io.Writer) LoggerI {
	for _, logger := range l.loggers {
		logger.SetOuts(outs...)
	}

	return l
}

// AddOuts method is similar to a Logger.AddOuts() method, however the multiLogger will
// range through all of its configured loggers and execute the same AddOuts() method call
// on each of them
//
// This method has been created to comply with the LoggerI interface -- but it's worth underlining
// that it will add the same io.Writer to all the loggers' outs. This may not be exactly
// the best action considering if there are different formats or more than one logger, it will
// result in different types of messages and / or repeated ones, respectively.
func (l *multiLogger) AddOuts(outs ...io.Writer) LoggerI {
	for _, logger := range l.loggers {
		logger.AddOuts(outs...)
	}

	return l
}

// Prefix method is similar to a Logger.Prefix() method, however the multiLogger will
// range through all of its configured loggers and execute the same Prefix() method call
// on each of them -- applying the input prefix string as each Logger's prefix.
func (l *multiLogger) Prefix(prefix string) LoggerI {
	for _, logger := range l.loggers {
		logger.Prefix(prefix)
	}

	return l
}

// Fields method is similar to a Logger.Fields() method, however the multiLogger will
// range through all of its configured loggers and execute the same Fields() method call
// on each of them -- applying the input Metadata map as the Logger's metadata.
func (l *multiLogger) Fields(fields map[string]interface{}) LoggerI {
	for _, logger := range l.loggers {
		logger.Fields(fields)
	}
	return l
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

	panic(s)
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

	panic(s)
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

	panic(s)
}

// Fatal method (similar to fmt.Print) will print a message using an fmt.Sprint(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Fatal.
//
// This method will end calling `os.Exit(1)`
func (l *multiLogger) Fatal(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Output(NewMessage().Level(LLFatal).Message(fmt.Sprint(v...)).Build())
	}
	os.Exit(1)
}

// Fatalln method (similar to fmt.Print) will print a message using an fmt.Sprintln(v...) pattern
// across all configured Loggers, while automatically applying LogLevel Fatal.
//
// This method will end calling `os.Exit(1)`
func (l *multiLogger) Fatalln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Output(NewMessage().Level(LLFatal).Message(fmt.Sprintln(v...)).Build())
	}
	os.Exit(1)
}

// Fatalf method (similar to fmt.Print) will print a message using an fmt.Sprintf(format, v...) pattern
// across all configured Loggers, while automatically applying LogLevel Fatal.
//
// This method will end calling `os.Exit(1)`
func (l *multiLogger) Fatalf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Output(NewMessage().Level(LLFatal).Message(fmt.Sprintf(format, v...)).Build())
	}
	os.Exit(1)
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
