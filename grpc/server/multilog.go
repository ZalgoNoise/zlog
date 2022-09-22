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

// Multilogger function takes in any number of LogServer interfaces, merging them together
// and returning a single LogServer.
//
// This is a LogServer multiplexer.
func MultiLogger(loggers ...LogServer) LogServer {

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

// Serve is the implementation of the `Serve()` method, from the LogServer interface
//
// It will cycle through all configured loggers and launching their `Serve()` method
// as a goroutine, except for the last one. This is a blocking operation.
func (l *multiLogger) Serve() {
	var idxLimit = len(l.loggers) - 2

	for i := 0; i < len(l.loggers); i++ {
		if i == idxLimit {
			l.loggers[i].Serve()
		}
		go l.loggers[i].Serve()
	}

}

// Stop is the implementation of the `Stop()` method, from the LogServer interface
//
// It will cycle through all configured loggers and launching their `Stop()` method
// as a goroutine, except for the last one. This is a blocking operation.
func (l *multiLogger) Stop() {
	var idxLimit = len(l.loggers) - 2

	for i := 0; i < len(l.loggers); i++ {
		if i == idxLimit {
			l.loggers[i].Stop()
		}
		go l.loggers[i].Stop()
	}
}

// Channels is the implementation of the `Channels()` method, from the LogServer interface
//
// It creates new event and error channels similar to a `Channels()` call, but launches three
// listeners (as goroutines) to monitor for messages and fan-out to all configured loggers
// respectively.
//
// Both `*event.Event` channels (for the _actual logger_ and the service logger) will listen
// for messages as normal, but a received message is fanned-out to all configured loggers, to
// their respective output event channel.
//
// The error channel (a read one) works the opposite way: all error channels are iterated through
// and a goroutine is launched for each of them. On _any_ error received, a copy is sent to the
//
//	output error channel
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
		for msg := range ch {
			for _, c := range chSet {
				c <- msg
			}
		}
	}(logCh, logChSet)

	// kick off goroutine for service logger channel
	// any message received on the returned channel will be fanned out to all loggers
	go func(ch chan *event.Event, chSet []chan *event.Event) {
		for msg := range ch {
			for _, c := range chSet {
				c <- msg
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
