package log

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"
)

const (
	baseTextFormat     string = "[%s]\t[%s]\t[%s]\t%s\t%s\n"
	extendedTextFormat string = "[%s]\t\t[%s]\t\t[%s]\t\t%s\t\t%s\n"
	colorReset         string = "\033[0m"
)

const (
	traceColor string = "\033[97m"
	debugColor string = "\033[37m"
	infoColor  string = "\033[36m"
	warnColor  string = "\033[33m"
	errorColor string = "\033[31m"
	fatalColor string = "\033[31m"
	panicColor string = "\033[35m"
)

var levelColorMap = map[string]string{
	LLTrace.String(): traceColor,
	LLDebug.String(): debugColor,
	LLInfo.String():  infoColor,
	LLWarn.String():  warnColor,
	LLError.String(): errorColor,
	LLFatal.String(): fatalColor,
	LLPanic.String(): panicColor,
}

type LogTimestamp string

func (lt LogTimestamp) String() string {
	return string(lt)
}

const (
	LTRFC3339Nano LogTimestamp = time.RFC3339Nano
	LTRFC3339     LogTimestamp = time.RFC3339
	LTRFC822Z     LogTimestamp = time.RFC822Z
	LTRubyDate    LogTimestamp = time.RubyDate
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
	0:  NewTextFormat().Build(),
	1:  &JSONFmt{},
	5:  NewTextFormat().Time(LTRFC3339).Build(),
	6:  NewTextFormat().Time(LTRFC822Z).Build(),
	7:  NewTextFormat().Time(LTRubyDate).Build(),
	8:  NewTextFormat().DoubleSpace().Build(),
	9:  NewTextFormat().DoubleSpace().LevelFirst().Build(),
	10: NewTextFormat().LevelFirst().Build(),
	11: NewTextFormat().DoubleSpace().Color().Build(),
	12: NewTextFormat().DoubleSpace().LevelFirst().Color().Build(),
	14: NewTextFormat().Color().Build(),
}

var (
	TextFormat                LogFormatter = LogFormatters[0] // placeholder for an initialized Text LogFormatter
	JSONFormat                LogFormatter = LogFormatters[1] // placeholder for an initialized JSON LogFormatter
	TextLongDate              LogFormatter = LogFormatters[5] // placeholder for an initialized Text LogFormatter, with a RFC3339 date format
	TextShortDate             LogFormatter = LogFormatters[6] // placeholder for an initialized Text LogFormatter, with a RFC822Z date format
	TextRubyDate              LogFormatter = LogFormatters[7] // placeholder for an initialized Text LogFormatter, with a RubyDate date format
	TextDoubleSpace           LogFormatter = LogFormatters[8]
	TextLevelFirstSpaced      LogFormatter = LogFormatters[9]
	TextLevelFirst            LogFormatter = LogFormatters[10]
	ColorTextDoubleSpace      LogFormatter = LogFormatters[11]
	ColorTextLevelFirstSpaced LogFormatter = LogFormatters[12]
	ColorTextLevelFirst       LogFormatter = LogFormatters[13]
	ColorText                 LogFormatter = LogFormatters[14]
)

// TextFmt struct describes the different manipulations and processing that a Text LogFormatter
// can apply to a LogMessage
type TextFmt struct {
	timeFormat  string
	levelFirst  bool
	doubleSpace bool
	colored     bool
	upper       bool
}

type TextFmtBuilder struct {
	timeFormat  LogTimestamp
	levelFirst  bool
	doubleSpace bool
	colored     bool
	upper       bool
}

func NewTextFormat() *TextFmtBuilder {
	return &TextFmtBuilder{}
}

func (b *TextFmtBuilder) Time(t LogTimestamp) *TextFmtBuilder {
	b.timeFormat = t
	return b
}

func (b *TextFmtBuilder) LevelFirst() *TextFmtBuilder {
	b.levelFirst = true
	return b
}

func (b *TextFmtBuilder) DoubleSpace() *TextFmtBuilder {
	b.doubleSpace = true
	return b
}

func (b *TextFmtBuilder) Color() *TextFmtBuilder {
	b.colored = true
	return b
}

func (b *TextFmtBuilder) Upper() *TextFmtBuilder {
	b.upper = true
	return b
}

func (b *TextFmtBuilder) Build() *TextFmt {
	if b.timeFormat == "" {
		b.timeFormat = LTRFC3339Nano
	}

	return &TextFmt{
		timeFormat:  b.timeFormat.String(),
		levelFirst:  b.levelFirst,
		doubleSpace: b.doubleSpace,
		colored:     b.colored,
		upper:       b.upper,
	}
}

// JSONFmt struct describes the different manipulations and processing that a JSON LogFormatter
// can apply to a LogMessage
type JSONFmt struct{}

// Format method will take in a pointer to a LogMessage; and returns a buffer and an error.
//
// This method will process the input LogMessage and marshal it according to this LogFormatter
func (f *TextFmt) Format(log *LogMessage) (buf []byte, err error) {
	var message string
	var format string

	if f.doubleSpace {
		format = extendedTextFormat
	} else {
		format = baseTextFormat
	}

	if f.levelFirst {
		message = fmt.Sprintf(
			format,
			f.colorize(log.Level),
			log.Time.Format(f.timeFormat),
			f.capitalize(log.Prefix),
			log.Msg,
			f.fmtMetadata(log.Metadata),
		)
	} else {
		message = fmt.Sprintf(
			format,
			log.Time.Format(f.timeFormat),
			f.colorize(log.Level),
			f.capitalize(log.Prefix),
			log.Msg,
			f.fmtMetadata(log.Metadata),
		)
	}

	buf = []byte(message)
	return
}

func (f *TextFmt) colorize(level string) string {
	if f.colored && runtime.GOOS != "windows" {
		return levelColorMap[level] + f.capitalize(level) + colorReset
	}
	return level
}

func (f *TextFmt) capitalize(s string) string {
	if f.upper {
		return strings.ToUpper(s)
	}
	return s

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
	lb.fmt = f
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
	lb.fmt = f
}
