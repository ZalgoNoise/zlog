package log

type ChLogMessage struct {
	Prefix   string
	Level    LogLevel
	Msg      string
	Metadata map[string]interface{}
}
