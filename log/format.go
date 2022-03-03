package log

import (
	"encoding/json"
	"fmt"
)

// LogFormatter interface describes the behavior a Formatter should have.
//
// The Format method is present to process the input LogMessage into content to be written
// (and consumed)
//
// The LoggerConfig implementation is to extend all LogFormatters to be used as LoggerConfig.
// This way, each formatter can be used directly when configuring a Logger, just by
// implementing an Apply(lb *LoggerBuilder) method
type LogFormatter interface {
	Format(log *LogMessage) (buf []byte, err error)
	LoggerConfig
}

// LogFormatters is a map of LogFormatters indexed by an int value. This is done in a map
// and not a list for manual ordering, spacing and manipulation of preset entries
var LogFormatters = map[int]LogFormatter{
	0: &TextFmt{},
	1: &JSONFmt{},
}

var (
	TextFormat LogFormatter = LogFormatters[0] // placeholder for an initialized Text LogFormatter
	JSONFormat LogFormatter = LogFormatters[1] // placeholder for an initialized JSON LogFormatter
)

// TextFmt struct describes the different manipulations and processing that a Text LogFormatter
// can apply to a LogMessage
type TextFmt struct{}

// JSONFmt struct describes the different manipulations and processing that a JSON LogFormatter
// can apply to a LogMessage
type JSONFmt struct{}

// Format method will take in a pointer to a LogMessage; and returns a buffer and an error.
//
// This method will process the input LogMessage and marshal it according to this LogFormatter
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

// Apply method implements the LoggerConfig interface, allowing a TextFmt object to be passed on as an
// argument, when creating a new Logger. It will define the logger's formatter as a Text LogFormatter
func (f *TextFmt) Apply(lb *LoggerBuilder) {
	lb.fmt = &TextFmt{}
}

// Format method will take in a pointer to a LogMessage; and returns a buffer and an error.
//
// This method will process the input LogMessage and marshal it according to this LogFormatter
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

// Apply method implements the LoggerConfig interface, allowing a JSONFmt object to be passed on as an
// argument, when creating a new Logger. It will define the logger's formatter as a JSON LogFormatter
func (f *JSONFmt) Apply(lb *LoggerBuilder) {
	lb.fmt = &JSONFmt{}
}
