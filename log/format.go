package log

import (
	"encoding/json"
	"fmt"
)

type LogFormatter interface {
	Format(log *LogMessage) (buf []byte, err error)
	Apply(lb *LoggerBuilder)
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
	size := len(data)
	count := 0

	if size == 0 {
		return ""
	}

	var meta string = "[ "

	for k, v := range data {
		switch value := v.(type) {
		case map[string]interface{}:
			meta += k + " = " + f.fmtMetadata(value)
			count++
			if count < size {
				meta += "; "
			}
		case string:
			meta += fmt.Sprintf("%s = \"%s\" ", k, v)
			count++
			if count < size {
				meta += "; "
			}
		default:
			meta += fmt.Sprintf("%s = %v ", k, v)
			count++
			if count < size {
				meta += "; "
			}
		}
	}

	meta += "] "

	return meta
}

func (f *TextFmt) Apply(lb *LoggerBuilder) {
	lb.fmt = &TextFmt{}
}

func (f *JSONFmt) Format(log *LogMessage) (buf []byte, err error) {
	// remove trailing newline on JSON format
	if log.Msg[len(log.Msg)-1] == 10 {
		log.Msg = log.Msg[:len(log.Msg)-1]
	}

	data, err := json.Marshal(log)
	if err != nil {
		return nil, err
	}
	buf = data
	return
}

func (f *JSONFmt) Apply(lb *LoggerBuilder) {
	lb.fmt = &JSONFmt{}
}
