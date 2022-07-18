package server

type multiLogger struct {
	loggers []LogServer
}

func (ml *multiLogger) addLoggers(l ...LogServer) {
	ml.loggers = make([]LogServer, 0, len(l))
	for _, logger := range l {
		ml.addLogger(logger)
	}
}

func (ml *multiLogger) addLogger(l LogServer) {
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
	return

}

func (ml *multiLogger) build() LogServer {
	if len(ml.loggers) == 0 {
		return nil
	}

	if len(ml.loggers) == 1 {
		return ml.loggers[0]
	}

	return ml
}

func MultiLogger(loggers ...LogServer) LogServer {

	if loggers == nil || len(loggers) == 0 {
		return nil
	}

	if len(loggers) == 1 {
		return loggers[0]
	}

	ml := new(multiLogger)

	ml.addLoggers(loggers...)
	return ml.build()
}

func (l *multiLogger) Serve() {
	var idxLimit = len(l.loggers) - 2

	for i := 0; i < len(l.loggers); i++ {
		if i == idxLimit {
			l.loggers[i].Serve()
		}
		go l.loggers[i].Serve()
	}

}
func (l *multiLogger) Stop() {
	var idxLimit = len(l.loggers) - 2

	for i := 0; i < len(l.loggers); i++ {
		if i == idxLimit {
			l.loggers[i].Stop()
		}
		go l.loggers[i].Stop()
	}
}
