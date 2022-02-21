package log

// NewLogCh function is a helper to spawn a channeled logger function and
// channel, for an existing Logger interface.
//
// Instead of implementing the logic below everytime, this function can be
// used to spawn a go routine and use its channel to send messages:
//
//
//     logger := log.New("logger", TextFormat)
//     logCh := NewLogCh(logger)
//
//	   // then, either the "classic" channeled message approach:
//	   ch, done := logCh.Channels()
//
//     ch <- log.NewMessage().Level(log.LLTrace).Message("test message").Build()
//
//     // or using the embeded method
//     logCh.Send(log.NewMessage().Message("this works too").Build())
//
//     // and finally stop the goroutine (if needed)
//     logCh.Close()
//     // or
//     done <- struct{}{}
func NewLogCh(logger LoggerI) (logCh ChanneledLogger) {

	msgCh := make(chan *LogMessage)
	done := make(chan struct{})

	logCh = &LogChannel{
		logCh: msgCh,
		done:  done,
	}

	go func(done chan struct{}) {
		for {
			select {
			case msg, ok := <-msgCh:
				if ok {
					logger.Log(msg)
				} else {
					logger.Log(
						NewMessage().Prefix("logger").Level(LLInfo).Message("channel closed").Build(),
					)
					return
				}
			case <-done:
				// logger.Log(
				// 	NewMessage().Prefix("logger").Level(LLInfo).Message("received done signal").Build(),
				// )
				return
			}

		}
	}(done)

	return
}

type ChanneledLogger interface {
	Log(msg ...*LogMessage)
	Close()
	Channels() (logCh chan *LogMessage, done chan struct{})
}

type LogChannel struct {
	logCh chan *LogMessage
	done  chan struct{}
}

func (c LogChannel) Log(msg ...*LogMessage) {
	for _, m := range msg {
		c.logCh <- m
	}
}

func (c LogChannel) Close() {
	c.done <- struct{}{}
}

func (c LogChannel) Channels() (logCh chan *LogMessage, done chan struct{}) {
	return c.logCh, c.done
}
