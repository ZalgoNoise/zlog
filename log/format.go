package log

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	colorReset string = "\033[0m"
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

// LogTimestamp type will define the compatible timestamp formatting strings
// which can be used
type LogTimestamp string

// String method is an implementation of the Stringer interface, to idiomatically
// convert a LogTimestamp into its string representation
func (lt LogTimestamp) String() string {
	return string(lt)
}

const (
	LTRFC3339Nano LogTimestamp = time.RFC3339Nano
	LTRFC3339     LogTimestamp = time.RFC3339
	LTRFC822Z     LogTimestamp = time.RFC822Z
	LTRubyDate    LogTimestamp = time.RubyDate
	LTUnixNano    LogTimestamp = "UNIX_NANO"
	LTUnixMilli   LogTimestamp = "UNIX_MILLI"
	LTUnixMicro   LogTimestamp = "UNIX_MICRO"
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
	1:  &FmtJSON{},
	2:  &FmtCSV{},
	3:  &FmtXML{},
	4:  &FmtGob{},
	5:  NewTextFormat().Time(LTRFC3339).Build(),
	6:  NewTextFormat().Time(LTRFC822Z).Build(),
	7:  NewTextFormat().Time(LTRubyDate).Build(),
	8:  NewTextFormat().DoubleSpace().Build(),
	9:  NewTextFormat().DoubleSpace().LevelFirst().Build(),
	10: NewTextFormat().LevelFirst().Build(),
	11: NewTextFormat().DoubleSpace().Color().Build(),
	12: NewTextFormat().DoubleSpace().LevelFirst().Color().Build(),
	13: NewTextFormat().LevelFirst().Color().Build(),
	14: NewTextFormat().Color().Build(),
	15: NewTextFormat().NoHeaders().NoTimestamp().NoLevel().Build(),
	16: NewTextFormat().NoHeaders().Build(),
	17: NewTextFormat().NoTimestamp().Build(),
	18: NewTextFormat().NoTimestamp().Color().Build(),
	19: NewTextFormat().NoTimestamp().Color().Upper().Build(),
}

var (
	FormatText                LogFormatter = LogFormatters[0]  // placeholder for an initialized Text LogFormatter
	FormatJSON                LogFormatter = LogFormatters[1]  // placeholder for an initialized JSON LogFormatter
	FormatCSV                 LogFormatter = LogFormatters[2]  // placeholder for an initialized CSV LogFormatter
	FormatXML                 LogFormatter = LogFormatters[3]  // placeholder for an initialized XML LogFormatter
	FormatGob                 LogFormatter = LogFormatters[4]  // placeholder for an initialized Gob LogFormatter
	TextLongDate              LogFormatter = LogFormatters[5]  // placeholder for an initialized Text LogFormatter, with a RFC3339 date format
	TextShortDate             LogFormatter = LogFormatters[6]  // placeholder for an initialized Text LogFormatter, with a RFC822Z date format
	TextRubyDate              LogFormatter = LogFormatters[7]  // placeholder for an initialized Text LogFormatter, with a RubyDate date format
	TextDoubleSpace           LogFormatter = LogFormatters[8]  // placeholder for an initialized Text LogFormatter, with double spaces
	TextLevelFirstSpaced      LogFormatter = LogFormatters[9]  // placeholder for an initialized  LogFormatter, with level-first and double spaces
	TextLevelFirst            LogFormatter = LogFormatters[10] // placeholder for an initialized  LogFormatter, with level-first
	TextColorDoubleSpace      LogFormatter = LogFormatters[11] // placeholder for an initialized  LogFormatter, with color and double spaces
	TextColorLevelFirstSpaced LogFormatter = LogFormatters[12] // placeholder for an initialized  LogFormatter, with color, level-first and double spaces
	TextColorLevelFirst       LogFormatter = LogFormatters[13] // placeholder for an initialized  LogFormatter, with color and level-first
	TextColor                 LogFormatter = LogFormatters[14] // placeholder for an initialized  LogFormatter, with color
	TextOnly                  LogFormatter = LogFormatters[15] // placeholder for an initialized  LogFormatter, with only the text content
	TextNoHeaders             LogFormatter = LogFormatters[15] // placeholder for an initialized  LogFormatter, without headers
	TextNoTimestamp           LogFormatter = LogFormatters[15] // placeholder for an initialized  LogFormatter, without timestamp
	TextColorNoTimestamp      LogFormatter = LogFormatters[15] // placeholder for an initialized  LogFormatter, without timestamp
	TextColorUpperNoTimestamp LogFormatter = LogFormatters[15] // placeholder for an initialized  LogFormatter, without timestamp and uppercase headers
)

// FmtText struct describes the different manipulations and processing that a
// Text LogFormatter can apply to a LogMessage
type FmtText struct {
	timeFormat  string
	levelFirst  bool
	doubleSpace bool
	colored     bool
	upper       bool
	noTimestamp bool
	noHeaders   bool
	noLevel     bool
}

// FmtTextBuilder struct will define the base of a custom FmtText object,
// which will take in different options in the form of methods that will define its
// configuration.
//
// Then, the Build() method can be called to return a FmtText object
type FmtTextBuilder struct {
	timeFormat  LogTimestamp
	levelFirst  bool
	doubleSpace bool
	colored     bool
	upper       bool
	noTimestamp bool
	noHeaders   bool
	noLevel     bool
}

// NewTextFormat function will initialize a FmtTextBuilder
func NewTextFormat() *FmtTextBuilder {
	return &FmtTextBuilder{}
}

// Time method will set a FmtTextBuilder's timeFormat, and return itself
// to allow method chaining
func (b *FmtTextBuilder) Time(t LogTimestamp) *FmtTextBuilder {
	b.timeFormat = t
	return b
}

// LevelFirst method will set a FmtTextBuilder's levelFirst element to true,
// and return itself to allow method chaining
func (b *FmtTextBuilder) LevelFirst() *FmtTextBuilder {
	b.levelFirst = true
	return b
}

// DoubleSpace method will set a FmtTextBuilder's doubleSpace element to true,
// and return itself to allow method chaining
func (b *FmtTextBuilder) DoubleSpace() *FmtTextBuilder {
	b.doubleSpace = true
	return b
}

// Color method will set a FmtTextBuilder's colored element to true,
// and return itself to allow method chaining
func (b *FmtTextBuilder) Color() *FmtTextBuilder {
	b.colored = true
	return b
}

// Upper method will set a FmtTextBuilder's upper element to true,
// and return itself to allow method chaining
func (b *FmtTextBuilder) Upper() *FmtTextBuilder {
	b.upper = true
	return b
}

// NoTimestamp method will set a FmtTextBuilder's noTimestamp element to true,
// and return itself to allow method chaining
func (b *FmtTextBuilder) NoTimestamp() *FmtTextBuilder {
	b.noTimestamp = true
	return b
}

// NoHeaders method will set a FmtTextBuilder's noHeaders element to true,
// and return itself to allow method chaining
func (b *FmtTextBuilder) NoHeaders() *FmtTextBuilder {
	b.noHeaders = true
	return b
}

// NoLevel method will set a FmtTextBuilder's noLevel element to true,
// and return itself to allow method chaining
func (b *FmtTextBuilder) NoLevel() *FmtTextBuilder {
	b.noLevel = true
	return b
}

// Build method will ensure the mandatory elements of FmtText are set
// and set them as default if otherwise, returning a pointer to a
// (custom) FmtText object
func (b *FmtTextBuilder) Build() *FmtText {
	if b.timeFormat == "" {
		b.timeFormat = LTRFC3339Nano
	}

	if b.noLevel && b.levelFirst {
		b.levelFirst = false
	}

	if b.noLevel && b.colored {
		b.colored = false
	}

	if b.noLevel && b.noHeaders && b.upper {
		b.upper = false
	}

	return &FmtText{
		timeFormat:  b.timeFormat.String(),
		levelFirst:  b.levelFirst,
		doubleSpace: b.doubleSpace,
		colored:     b.colored,
		upper:       b.upper,
		noTimestamp: b.noTimestamp,
		noHeaders:   b.noHeaders,
		noLevel:     b.noLevel,
	}
}

// Format method will take in a pointer to a LogMessage; and returns a buffer and an error.
//
// This method will process the input LogMessage and marshal it according to this LogFormatter
func (f *FmtText) Format(log *LogMessage) (buf []byte, err error) {
	var sb strings.Builder

	// [info] (...)
	if f.levelFirst && !f.noHeaders {
		sb.WriteString("[")
		sb.WriteString(f.colorize(log.Level))
		sb.WriteString("]\t")
		if f.doubleSpace {
			sb.WriteString("\t")
		}
	}
	if !f.noTimestamp {
		// [time] (...)
		sb.WriteString("[")
		sb.WriteString(f.fmtTime(log.Time))
		sb.WriteString("]\t")
		if f.doubleSpace {
			sb.WriteString("\t")
		}
	}
	if !f.levelFirst && !f.noLevel {
		// (...) [info] (...)
		sb.WriteString("[")
		sb.WriteString(f.colorize(log.Level))
		sb.WriteString("]\t")
		if f.doubleSpace {
			sb.WriteString("\t")
		}
	}
	if !f.noHeaders {
		if log.Prefix != "" {
			// (...) [service] (...)
			sb.WriteString("[")
			sb.WriteString(f.capitalize(log.Prefix))
			sb.WriteString("]\t")
			if f.doubleSpace {
				sb.WriteString("\t")
			}
		}

		if log.Sub != "" {
			// (...) [module] (...)
			sb.WriteString("[")
			sb.WriteString(f.capitalize(log.Sub))
			sb.WriteString("]\t")
			if f.doubleSpace {
				sb.WriteString("\t")
			}
		}
	}

	sb.WriteString(log.Msg)

	if len(log.Metadata) > 0 {
		sb.WriteString("\t")
		if f.doubleSpace {
			sb.WriteString("\t")
		}
		sb.WriteString(f.fmtMetadata(log.Metadata))
	}
	sb.WriteString("\n")

	buf = []byte(sb.String())
	return
}

func (f *FmtText) fmtTime(t time.Time) string {
	switch f.timeFormat {
	case LTUnixNano.String():
		return strconv.Itoa(int(t.Unix()))
	case LTUnixMilli.String():
		return strconv.Itoa(int(t.UnixMilli()))
	case LTUnixMicro.String():
		return strconv.Itoa(int(t.UnixMicro()))
	default:
		return t.Format(f.timeFormat)
	}
}

func (f *FmtText) colorize(level string) string {
	if f.colored && runtime.GOOS != "windows" {
		return levelColorMap[level] + f.capitalize(level) + colorReset
	}
	return f.capitalize(level)
}

func (f *FmtText) capitalize(s string) string {
	if f.upper {
		return strings.ToUpper(s)
	}
	return s

}

func (f *FmtText) fmtMetadata(data map[string]interface{}) string {
	size := len(data)
	count := 0

	if size == 0 {
		return ""
	}

	var meta string = "[ "

	for k, v := range data {
		switch value := v.(type) {
		case []map[string]interface{}:
			meta += k + " = [ "
			for idx, m := range value {
				meta += f.fmtMetadata(m)
				if idx < len(value)-1 {
					meta += "; "
				}
			}
			meta += "] "
			count++
			if count < size {
				meta += "; "
			}

		case []Field:
			meta += k + " = [ "
			for idx, m := range value {
				meta += f.fmtMetadata(m.ToMap())
				if idx < len(value)-1 {
					meta += "; "
				}
			}
			meta += "] "
			count++
			if count < size {
				meta += "; "
			}

		case map[string]interface{}:
			meta += k + " = " + f.fmtMetadata(value)
			count++
			if count < size {
				meta += "; "
			}
		case Field:
			metadata := value.ToMap()
			meta += k + " = " + f.fmtMetadata(metadata)
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

// Apply method implements the LoggerConfig interface, allowing a FmtText object to be passed on as an
// argument, when creating a new Logger. It will define the logger's formatter as a Text LogFormatter
func (f *FmtText) Apply(lb *LoggerBuilder) {
	lb.fmt = f
}

// FmtJSON struct describes the different manipulations and processing that a JSON LogFormatter
// can apply to a LogMessage
type FmtJSON struct{}

// Format method will take in a pointer to a LogMessage; and returns a buffer and an error.
//
// This method will process the input LogMessage and marshal it according to this LogFormatter
func (f *FmtJSON) Format(log *LogMessage) (buf []byte, err error) {
	// remove trailing newline on JSON format
	if log.Msg[len(log.Msg)-1] == 10 {
		log.Msg = log.Msg[:len(log.Msg)-1]
	}

	return json.Marshal(log)
}

// Apply method implements the LoggerConfig interface, allowing a FmtJSON object to be passed on as an
// argument, when creating a new Logger. It will define the logger's formatter as a JSON LogFormatter
func (f *FmtJSON) Apply(lb *LoggerBuilder) {
	lb.fmt = f
}

// FmtCSV struct describes the different manipulations and processing that a CSV LogFormatter
// can apply to a LogMessage
type FmtCSV struct {
	unixTime bool
	jsonMeta bool
}

// FmtCSVBuilder struct allows creating custom CSV Formatters. Its default values will leave
// its supported options set as false, so it's not required to always use this struct.
//
// Its options will allow:
// - setting Unix micros as timestamp (in string format)
// - setting a JSON metadata formatter instead of text-based
type FmtCSVBuilder struct {
	unixTime bool
	jsonMeta bool
}

// NewCSVFormat function will create a new instance of a FmtCSVBuilder
func NewCSVFormat() *FmtCSVBuilder {
	return &FmtCSVBuilder{}
}

// Unix method will set the FmtCSV's timestamp as Unix micros
func (b *FmtCSVBuilder) Unix() *FmtCSVBuilder {
	b.unixTime = true
	return b
}

// JSON method will set the FmtCSV's metadata as JSON format
func (b *FmtCSVBuilder) JSON() *FmtCSVBuilder {
	b.jsonMeta = true
	return b
}

// Build method will create a (custom) FmtCSV object based on the builder's configuration,
// and return a pointer to it
func (b *FmtCSVBuilder) Build() *FmtCSV {
	return &FmtCSV{
		unixTime: b.unixTime,
		jsonMeta: b.jsonMeta,
	}
}

// Format method will take in a pointer to a LogMessage; and returns a buffer and an error.
//
// This method will process the input LogMessage and marshal it according to this LogFormatter
func (f *FmtCSV) Format(log *LogMessage) (buf []byte, err error) {
	b := bytes.NewBuffer(buf)
	w := csv.NewWriter(b)

	// prepare time value
	var t string

	if f.unixTime {
		// Unix micros in string format
		t = log.Time.Format(LTUnixMicro.String())
	} else {
		// RFC 3339 timestamp in string format
		t = log.Time.Format(LTRFC3339Nano.String())
	}

	// prepare metadata value
	var m string
	if f.jsonMeta {
		// marshal as JSON
		b, err := json.Marshal(log.Metadata)
		if err != nil {
			return nil, err
		}
		m = string(b)
	} else {
		// use FmtText to marshal the metadata
		txt := &FmtText{}
		m = txt.fmtMetadata(log.Metadata)
	}

	// default format for:
	// "timestamp","level","prefix","sub","message","metadata"
	record := []string{
		t,
		log.Level,
		log.Prefix,
		log.Sub,
		log.Msg,
		m,
	}

	if err = w.Write(record); err != nil {
		return nil, err
	}

	w.Flush()

	if err = w.Error(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil

}

// Apply method implements the LoggerConfig interface, allowing a FmtCSV object to be passed on as an
// argument, when creating a new Logger. It will define the logger's formatter as a CSV LogFormatter
func (f *FmtCSV) Apply(lb *LoggerBuilder) {
	lb.fmt = f
}

// FmtXML struct describes the different manipulations and processing that a XML LogFormatter
// can apply to a LogMessage
type FmtXML struct{}

// Format method will take in a pointer to a LogMessage; and returns a buffer and an error.
//
// This method will process the input LogMessage and marshal it according to this LogFormatter
func (f *FmtXML) Format(log *LogMessage) (buf []byte, err error) {
	// remove trailing newline on XML format
	if log.Msg[len(log.Msg)-1] == 10 {
		log.Msg = log.Msg[:len(log.Msg)-1]
	}

	type logMessage struct {
		Time     time.Time `xml:"timestamp,omitempty"`
		Prefix   string    `xml:"service,omitempty"`
		Sub      string    `xml:"module,omitempty"`
		Level    string    `xml:"level,omitempty"`
		Msg      string    `xml:"message,omitempty"`
		Metadata []field   `xml:"metadata,omitempty"`
	}

	xmlMsg := &logMessage{
		Time:     log.Time,
		Prefix:   log.Prefix,
		Sub:      log.Sub,
		Level:    log.Level,
		Msg:      log.Msg,
		Metadata: mappify(log.Metadata),
	}

	return xml.Marshal(xmlMsg)

}

// Apply method implements the LoggerConfig interface, allowing a FmtXML object to be passed on as an
// argument, when creating a new Logger. It will define the logger's formatter as a XML LogFormatter
func (f *FmtXML) Apply(lb *LoggerBuilder) {
	lb.fmt = f
}

// FmtGob struct allows marshalling a LogMessage as gob-encoded bytes
type FmtGob struct{}

// Format method will take in a pointer to a LogMessage, and return the execution of its
// encode() method; which converts it to gob-encoded bytes, returning a slice of bytes and an
// error
func (f *FmtGob) Format(log *LogMessage) ([]byte, error) {
	return log.encode()
}

// Apply method implements the LoggerConfig interface, allowing a FmtGob object to be passed on as an
// argument, when creating a new Logger. It will define the logger's formatter as a Gob LogFormatter
func (f *FmtGob) Apply(lb *LoggerBuilder) {
	lb.fmt = f
}
