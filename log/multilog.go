package log

import (
	"errors"
	"fmt"
	"io"
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

// Sub method is similar to a Logger.Sub() method, however the multiLogger will
// range through all of its configured loggers and execute the same Sub() method call
// on each of them -- applying the input sub-prefix string as each Logger's sub-prefix.
func (l *multiLogger) Sub(sub string) LoggerI {
	for _, logger := range l.loggers {
		logger.Sub(sub)
	}
	return l
}

// IsSkipExit method is similar to a Logger.IsSkipExit() method, however the multiLogger will
// range through all of its configured loggers and execute the same IsSkipExit() method call
// on each of them -- the first (if any) Logger which lists having this option set to true
// will cause an immediate return of this value, otherwise it will return as false
func (l *multiLogger) IsSkipExit() bool {
	for _, logger := range l.loggers {
		ok := logger.IsSkipExit()
		if ok {
			return ok // true
		}
	}
	return false
}

// IsSkipExit method is similar to a Logger.IsSkipExit() method, however the multiLogger will
// range through all of its configured loggers and execute the same IsSkipExit() method call
// on each of them -- ensuring that no errors are found through all writes.
//
// If errors are found, they are concatenated and returned as a single error. The reasoning for
// this decision is because the io.Writer interface returns a single error. However, blocking
// the whole operation if one writer fails seems less approachable
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
		return -1, errors.New(fmt.Sprint("failed to write with errors: ", errs))
	}

	return n, nil
}
