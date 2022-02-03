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

type TextFmt struct{}

type JSONFmt struct{}

func (f *TextFmt) Format(log *LogMessage) (buf []byte, err error) {
	message := fmt.Sprintf(
		"[%s]\t[%s] [%s]\t%s",
		log.Time,
		log.Prefix,
		log.Level,
		log.Msg,
	)

	if log.Metadata != nil {
		message = message + "\t-- " + f.fmtMetadata(log.Metadata) + "\n"
	}

	buf = []byte(message)
	return
}

func (f *TextFmt) fmtMetadata(data map[string]interface{}) string {
	var meta string
	for k, v := range data {
		switch value := v.(type) {
		case map[string]interface{}:
			meta += " {" + f.fmtMetadata(value) + " }"
		default:
			meta += fmt.Sprintf(" %s = %v ;", k, v)
		}
	}
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
