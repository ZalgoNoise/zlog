package server

import "github.com/zalgonoise/zlog/log/event"

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

func (l *multiLogger) Channels() (logCh, logSvCh chan *event.Event, errCh chan error) {

	// make output channels
	logCh = make(chan *event.Event)
	logSvCh = make(chan *event.Event)
	errCh = make(chan error)

	// make channel slices according to configured loggers' length
	logChSet := make([]chan *event.Event, len(l.loggers))
	logSvChSet := make([]chan *event.Event, len(l.loggers))
	errChSet := make([]chan error, len(l.loggers))

	// get the channels from each configured logger
	for _, l := range l.loggers {
		lCh, lSvCh, eCh := l.Channels()

		logChSet = append(logChSet, lCh)
		logSvChSet = append(logSvChSet, lSvCh)
		errChSet = append(errChSet, eCh)
	}

	// kick off goroutine for log channel
	// any message received on the returned channel will be fanned out to all loggers
	go func(ch chan *event.Event, chSet []chan *event.Event) {
		for {
			select {
			case msg := <-ch:
				for _, c := range chSet {
					c <- msg
				}
			}
		}
	}(logCh, logChSet)

	// kick off goroutine for service logger channel
	// any message received on the returned channel will be fanned out to all loggers
	go func(ch chan *event.Event, chSet []chan *event.Event) {
		for {
			select {
			case msg := <-ch:
				for _, c := range chSet {
					c <- msg
				}
			}
		}
	}(logSvCh, logSvChSet)

	// kick off goroutine for error channel
	// any message received on any channel will be pushed to the output one
	for _, ch := range errChSet {
		go func(in, out chan error) {
			for err := range in {
				errCh <- err
			}
		}(ch, errCh)
	}

	return logCh, logSvCh, errCh
}
