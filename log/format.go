package log

import (
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/log/format/bson"
	"github.com/zalgonoise/zlog/log/format/csv"
	"github.com/zalgonoise/zlog/log/format/gob"
	"github.com/zalgonoise/zlog/log/format/json"
	"github.com/zalgonoise/zlog/log/format/protobuf"
	"github.com/zalgonoise/zlog/log/format/text"
	"github.com/zalgonoise/zlog/log/format/xml"
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
	Format(log *event.Event) (buf []byte, err error)
}

type formatConfig struct {
	f LogFormatter
}

func (f *formatConfig) Apply(lb *LoggerBuilder) {
	lb.Fmt = f.f
}

func WithFormat(f LogFormatter) LoggerConfig {
	return &formatConfig{f: f}
}

// LogFormatters is a map of LogFormatters indexed by an int value. This is done in a map
// and not a list for manual ordering, spacing and manipulation of preset entries
var LogFormatters = map[int]LogFormatter{
	0:  text.New().Build(),
	1:  &protobuf.FmtPB{},
	2:  &json.FmtJSON{},
	3:  &csv.FmtCSV{},
	4:  &xml.FmtXML{},
	5:  &gob.FmtGob{},
	6:  &bson.FmtBSON{},
	7:  text.New().Time(text.LTRFC3339).Build(),
	8:  text.New().Time(text.LTRFC822Z).Build(),
	9:  text.New().Time(text.LTRubyDate).Build(),
	10: text.New().DoubleSpace().Build(),
	11: text.New().DoubleSpace().LevelFirst().Build(),
	12: text.New().LevelFirst().Build(),
	13: text.New().DoubleSpace().Color().Build(),
	14: text.New().DoubleSpace().LevelFirst().Color().Build(),
	15: text.New().LevelFirst().Color().Build(),
	16: text.New().Color().Build(),
	17: text.New().NoHeaders().NoTimestamp().NoLevel().Build(),
	18: text.New().NoHeaders().Build(),
	19: text.New().NoTimestamp().Build(),
	20: text.New().NoTimestamp().Color().Build(),
	21: text.New().NoTimestamp().Color().Upper().Build(),
}

var (
	FormatText                LogFormatter = LogFormatters[0]  // placeholder for an initialized Text LogFormatter
	FormatProtobuf            LogFormatter = LogFormatters[1]  // placeholder for an initialized Protobuf LogFormatter
	FormatJSON                LogFormatter = LogFormatters[2]  // placeholder for an initialized JSON LogFormatter
	FormatCSV                 LogFormatter = LogFormatters[3]  // placeholder for an initialized CSV LogFormatter
	FormatXML                 LogFormatter = LogFormatters[4]  // placeholder for an initialized XML LogFormatter
	FormatGob                 LogFormatter = LogFormatters[5]  // placeholder for an initialized Gob LogFormatter
	FormatBSON                LogFormatter = LogFormatters[6]  // placeholder for an initialized JSON LogFormatter
	TextLongDate              LogFormatter = LogFormatters[7]  // placeholder for an initialized Text LogFormatter, with a RFC3339 date format
	TextShortDate             LogFormatter = LogFormatters[8]  // placeholder for an initialized Text LogFormatter, with a RFC822Z date format
	TextRubyDate              LogFormatter = LogFormatters[9]  // placeholder for an initialized Text LogFormatter, with a RubyDate date format
	TextDoubleSpace           LogFormatter = LogFormatters[10] // placeholder for an initialized Text LogFormatter, with double spaces
	TextLevelFirstSpaced      LogFormatter = LogFormatters[11] // placeholder for an initialized  LogFormatter, with level-first and double spaces
	TextLevelFirst            LogFormatter = LogFormatters[12] // placeholder for an initialized  LogFormatter, with level-first
	TextColorDoubleSpace      LogFormatter = LogFormatters[13] // placeholder for an initialized  LogFormatter, with color and double spaces
	TextColorLevelFirstSpaced LogFormatter = LogFormatters[14] // placeholder for an initialized  LogFormatter, with color, level-first and double spaces
	TextColorLevelFirst       LogFormatter = LogFormatters[15] // placeholder for an initialized  LogFormatter, with color and level-first
	TextColor                 LogFormatter = LogFormatters[16] // placeholder for an initialized  LogFormatter, with color
	TextOnly                  LogFormatter = LogFormatters[17] // placeholder for an initialized  LogFormatter, with only the text content
	TextNoHeaders             LogFormatter = LogFormatters[18] // placeholder for an initialized  LogFormatter, without headers
	TextNoTimestamp           LogFormatter = LogFormatters[19] // placeholder for an initialized  LogFormatter, without timestamp
	TextColorNoTimestamp      LogFormatter = LogFormatters[20] // placeholder for an initialized  LogFormatter, without timestamp
	TextColorUpperNoTimestamp LogFormatter = LogFormatters[21] // placeholder for an initialized  LogFormatter, without timestamp and uppercase headers
)
