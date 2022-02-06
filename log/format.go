package log

import (
	"encoding/json"
	"fmt"
)

type LogFormatter interface {
	Format(log *LogMessage) (buf []byte, err error)
}

var LogFormatters = map[int]LogFormatter{
	0: &TextFmt{},
	1: &JSONFmt{},
}

var (
	TextFormat LogFormatter = LogFormatters[0]
	JSONFormat LogFormatter = LogFormatters[1]
)

type TextFmt struct{}

type JSONFmt struct{}

func (f *TextFmt) Format(log *LogMessage) (buf []byte, err error) {
	message := fmt.Sprintf(
		"[%s]\t[%s]\t[%s]\t%s\t%s\n",
		log.Time,
		log.Level,
		log.Prefix,
		log.Msg,
		f.fmtMetadata(log.Metadata),
	)

	buf = []byte(message)
	return
}

func (f *TextFmt) fmtMetadata(data map[string]interface{}) string {
	if len(data) == 0 {
		return ""
	}
	var meta string = "[ "

	for k, v := range data {
		switch value := v.(type) {
		case map[string]interface{}:
			meta += f.fmtMetadata(value)
		case string:
			meta += fmt.Sprintf("%s = \"%s\" ; ", k, v)
		default:
			meta += fmt.Sprintf("%s = %v ; ", k, v)
		}
	}

	meta += "] "

	return meta
}

func (f *JSONFmt) Format(log *LogMessage) (buf []byte, err error) {
	// remove trailing newline on JSON format
	if log.Msg[len(log.Msg)-1] == 10 {
		log.Msg = log.Msg[:len(log.Msg)-1]
	}

	data, err := json.Marshal(log)
	if err != nil {
		return
	}
	buf = data
	return
}
