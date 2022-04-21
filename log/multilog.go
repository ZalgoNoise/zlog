package log

import (
	"fmt"
	"io"

	"github.com/zalgonoise/zlog/grpc/address"
	// "github.com/zalgonoise/zlog/grpc/client"
)

type multiLogger struct {
	loggers []Logger
}

// MultiLogger function is a wrapper for multiple Logger
//
// Similar to how io.MultiWriter() is implemented, this function generates a single
// Logger that targets a set of configured Logger.
//
// As such, a single Logger can have multiple Loggers with different configurations and
// output files, while still registering the same log message.
func MultiLogger(loggers ...Logger) Logger {
	allLoggers := make([]Logger, 0, len(loggers))
	allLoggers = append(allLoggers, loggers...)

	return &multiLogger{allLoggers}
}

// SetOuts method is similar to a Logger.SetOuts() method, however the multiLogger will
// range through all of its configured loggers and execute the same SetOuts() method call
// on each of them
//
// This method has been created to comply with the Logger interface -- but it's worth underlining
// that it will overwrite all the loggers' outs. This may not be exactly the best action
// considering if there are different formats or more than one logger, it will result in
// different types of messages and / or repeated ones, respectively.
func (l *multiLogger) SetOuts(outs ...io.Writer) Logger {
	var addrMap = make([]io.Writer, 0)
	var writers = make([]io.Writer, 0)

	for _, out := range outs {
		if addr, ok := out.(*address.ConnAddr); ok {
			addrMap = append(addrMap, addr)
		} else if out == nil {
			continue
		} else {
			writers = append(writers, out)
		}
	}

	for _, log := range l.loggers {
		if l, ok := log.(*logger); ok {
			l.SetOuts(writers...)
		} else if ml, ok := log.(*multiLogger); ok {
			ml.SetOuts(writers...)
		} else {
			log.SetOuts(addrMap...)
		}

	}

	return l
}

// AddOuts method is similar to a Logger.AddOuts() method, however the multiLogger will
// range through all of its configured loggers and execute the same AddOuts() method call
// on each of them
//
// This method has been created to comply with the Logger interface -- but it's worth underlining
// that it will add the same io.Writer to all the loggers' outs. This may not be exactly
// the best action considering if there are different formats or more than one logger, it will
// result in different types of messages and / or repeated ones, respectively.
func (l *multiLogger) AddOuts(outs ...io.Writer) Logger {
	var addrMap = make([]io.Writer, 0)
	var writers = make([]io.Writer, 0)

	for _, out := range outs {
		if addr, ok := out.(*address.ConnAddr); ok {
			addrMap = append(addrMap, addr)
		} else if out == nil {
			continue
		} else {
			writers = append(writers, out)
		}
	}

	for _, log := range l.loggers {
		if l, ok := log.(*logger); ok {
			l.AddOuts(writers...)
		} else if ml, ok := log.(*multiLogger); ok {
			ml.AddOuts(writers...)
		} else {
			log.AddOuts(addrMap...)
		}

	}

	return l
}

// Prefix method is similar to a Logger.Prefix() method, however the multiLogger will
// range through all of its configured loggers and execute the same Prefix() method call
// on each of them -- applying the input prefix string as each Logger's prefix.
func (l *multiLogger) Prefix(prefix string) Logger {
	for _, logger := range l.loggers {
		logger.Prefix(prefix)
	}

	return l
}

// Fields method is similar to a Logger.Fields() method, however the multiLogger will
// range through all of its configured loggers and execute the same Fields() method call
// on each of them -- applying the input Metadata map as the Logger's metadata.
func (l *multiLogger) Fields(fields map[string]interface{}) Logger {
	for _, logger := range l.loggers {
		logger.Fields(fields)
	}
	return l
}

// Sub method is similar to a Logger.Sub() method, however the multiLogger will
// range through all of its configured loggers and execute the same Sub() method call
// on each of them -- applying the input sub-prefix string as each Logger's sub-prefix.
func (l *multiLogger) Sub(sub string) Logger {
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
// If errors are found, they are safely wrapped together and returned as a single error, since
// the io.Writer implementation involves returning an int and an error, only.
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
				err = fmt.Errorf("%w ; %s", err, e.Error())
			}
		}

		return -1, fmt.Errorf("multiple errors when writing message: %w", err)
	}

	return n, nil
}
