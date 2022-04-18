package text

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/zalgonoise/zlog/log/event"
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
	"trace": traceColor,
	"debug": debugColor,
	"info":  infoColor,
	"warn":  warnColor,
	"error": errorColor,
	"fatal": fatalColor,
	"panic": panicColor,
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

// New function will initialize a FmtTextBuilder
func New() *FmtTextBuilder {
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
func (f *FmtText) Format(log *event.Event) (buf []byte, err error) {
	var sb strings.Builder

	// [info] (...)
	if f.levelFirst && !f.noHeaders {
		sb.WriteString("[")
		sb.WriteString(f.colorize(log.GetLevel().String()))
		sb.WriteString("]\t")
		if f.doubleSpace {
			sb.WriteString("\t")
		}
	}
	if !f.noTimestamp {
		// [time] (...)
		sb.WriteString("[")
		sb.WriteString(f.fmtTime(log.GetTime().AsTime()))
		sb.WriteString("]\t")
		if f.doubleSpace {
			sb.WriteString("\t")
		}
	}
	if !f.levelFirst && !f.noLevel {
		// (...) [info] (...)
		sb.WriteString("[")
		sb.WriteString(f.colorize(log.GetLevel().String()))
		sb.WriteString("]\t")
		if f.doubleSpace {
			sb.WriteString("\t")
		}
	}
	if !f.noHeaders {
		if *log.Prefix != "" {
			// (...) [service] (...)
			sb.WriteString("[")
			sb.WriteString(f.capitalize(log.GetPrefix()))
			sb.WriteString("]\t")
			if f.doubleSpace {
				sb.WriteString("\t")
			}
		}

		if *log.Sub != "" {
			// (...) [module] (...)
			sb.WriteString("[")
			sb.WriteString(f.capitalize(log.GetSub()))
			sb.WriteString("]\t")
			if f.doubleSpace {
				sb.WriteString("\t")
			}
		}
	}

	sb.WriteString(log.GetMsg())

	if len(log.GetMeta().AsMap()) > 0 {
		sb.WriteString("\t")
		if f.doubleSpace {
			sb.WriteString("\t")
		}
		sb.WriteString(f.FmtMetadata(log.GetMeta().AsMap()))
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

func (f *FmtText) FmtMetadata(data map[string]interface{}) string {
	size := len(data)

	// exit early
	if size == 0 {
		return ""
	}

	count := 0
	var sb strings.Builder

	sb.WriteString("[ ")

	for k, v := range data {
		switch value := v.(type) {
		case []map[string]interface{}:
			sb.WriteString(k)
			sb.WriteString(" = [ ")
			for idx, m := range value {
				sb.WriteString(f.FmtMetadata(m))
				if idx < len(value)-1 {
					sb.WriteString("; ")
				}
			}
			sb.WriteString("] ")
			count++
			if count < size {
				sb.WriteString("; ")
			}

		case []event.Field:
			sb.WriteString(k)
			sb.WriteString(" = [ ")
			for idx, m := range value {
				sb.WriteString(f.FmtMetadata(m.ToMap()))
				if idx < len(value)-1 {
					sb.WriteString("; ")
				}
			}

			sb.WriteString("] ")
			count++
			if count < size {
				sb.WriteString("; ")
			}

		case map[string]interface{}:
			sb.WriteString(k)
			sb.WriteString(" = ")
			sb.WriteString(f.FmtMetadata(value))
			count++
			if count < size {
				sb.WriteString("; ")
			}
		case event.Field:
			sb.WriteString(k)
			sb.WriteString(" = ")
			sb.WriteString(f.FmtMetadata(value.ToMap()))
			count++
			if count < size {
				sb.WriteString("; ")
			}

		case string:
			sb.WriteString(k)
			sb.WriteString(" = \"")
			sb.WriteString(v.(string))
			sb.WriteString("\" ")
			count++
			if count < size {
				sb.WriteString("; ")
			}
		default:

			sb.WriteString(k)
			sb.WriteString(" = ")
			sb.WriteString(fmt.Sprint(v))
			sb.WriteString(" ")
			count++
			if count < size {
				sb.WriteString("; ")
			}
		}
	}

	sb.WriteString("] ")

	return sb.String()
}
