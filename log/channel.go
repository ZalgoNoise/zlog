package log

// NewLogCh function is a helper to spawn a channeled logger function and
// channel, for an existing Logger interface.
//
// Instead of implementing the logic below everytime, this function can be
// used to spawn a go routine and use its channel to send messages:
//
//
//     logger := log.New("logger", TextFormat)
//     logCh, chLogger := NewLogCh(logger)
//
//     go chLogger()
//
//     logCh <- log.NewMessage().Level(log.LLTrace).Message("test message").Build()
//
func NewLogCh(logger LoggerI) (logCh chan *LogMessage, chLogger func()) {
	logCh = make(chan *LogMessage)
	chLogger = func() {
		for {
			msg, ok := <-logCh
			if ok {
				logger.Log(msg)
			} else {
				logger.Log(
					NewMessage().Prefix("logger").Level(LLInfo).Message("channel closed").Build(),
				)
				break
			}
		}
	}
	return
}
