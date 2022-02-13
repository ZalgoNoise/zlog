package log

import (
	"fmt"
	"io"
	"os"
)

type LoggerI interface {
	Output(m *LogMessage) error
	SetOuts(outs ...io.Writer) LoggerI
	AddOuts(outs ...io.Writer) LoggerI
	SetPrefix(prefix string) LoggerI
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

type multiLogger struct {
	loggers []LoggerI
}

func MultiLogger(loggers ...LoggerI) LoggerI {
	allLoggers := make([]LoggerI, 0, len(loggers))
	allLoggers = append(allLoggers, loggers...)

	return &multiLogger{allLoggers}
}

func (l *multiLogger) Output(m *LogMessage) error {
	for _, logger := range l.loggers {
		err := logger.Output(m)
		if err != nil {
			return err
		}
	}
	return nil
}

// output setter methods

func (l *multiLogger) SetOuts(outs ...io.Writer) LoggerI {
	for _, logger := range l.loggers {
		logger.SetOuts(outs...)
	}

	return l
}

func (l *multiLogger) AddOuts(outs ...io.Writer) LoggerI {
	for _, logger := range l.loggers {
		logger.AddOuts(outs...)
	}

	return l
}

// prefix setter methods

func (l *multiLogger) SetPrefix(prefix string) LoggerI {
	for _, logger := range l.loggers {
		logger.SetPrefix(prefix)
	}

	return l
}

// metadata methods

func (l *multiLogger) Fields(fields map[string]interface{}) LoggerI {
	for _, logger := range l.loggers {
		logger.Fields(fields)
	}
	return l
}

// print methods

func (l *multiLogger) Print(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Print(v...)
	}
}

func (l *multiLogger) Println(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Println(v...)
	}
}

func (l *multiLogger) Printf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Printf(format, v...)
	}
}

// log methods

func (l *multiLogger) Log(m *LogMessage) {
	for _, logger := range l.loggers {
		logger.Log(m)
	}
}

// panic methods

func (l *multiLogger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)

	for _, logger := range l.loggers {
		logger.Panic(s)
	}

	panic(s)
}

func (l *multiLogger) Panicln(v ...interface{}) {
	s := fmt.Sprint(v...)

	for _, logger := range l.loggers {
		logger.Panicln(s)
	}

	panic(s)
}

func (l *multiLogger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprint(v...)

	for _, logger := range l.loggers {
		logger.Panicf(format, s)
	}

	panic(s)
}

// fatal methods

func (l *multiLogger) Fatal(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Fatal(v...)
	}
	os.Exit(1)
}

func (l *multiLogger) Fatalln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Fatalln(v...)
	}
	os.Exit(1)
}

func (l *multiLogger) Fatalf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Fatalf(format, v...)
	}
	os.Exit(1)
}

// error methods

func (l *multiLogger) Error(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Error(v...)
	}
}

func (l *multiLogger) Errorln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Errorln(v...)
	}
}

func (l *multiLogger) Errorf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Errorf(format, v...)
	}
}

// warn methods

func (l *multiLogger) Warn(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Warn(v...)
	}
}

func (l *multiLogger) Warnln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Warnln(v...)
	}
}

func (l *multiLogger) Warnf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Warnf(format, v...)
	}
}

// info methods

func (l *multiLogger) Info(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Info(v...)
	}

}

func (l *multiLogger) Infoln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Infoln(v...)
	}
}

func (l *multiLogger) Infof(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Infof(format, v...)
	}
}

// debug methods

func (l *multiLogger) Debug(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Debug(v...)
	}
}

func (l *multiLogger) Debugln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Debugln(v...)
	}
}

func (l *multiLogger) Debugf(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Debugf(format, v...)
	}
}

// trace methods

func (l *multiLogger) Trace(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Trace(v...)
	}
}

func (l *multiLogger) Traceln(v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Traceln(v...)
	}
}

func (l *multiLogger) Tracef(format string, v ...interface{}) {
	for _, logger := range l.loggers {
		logger.Tracef(format, v...)
	}
}
