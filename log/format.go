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
// The Format method is present to process the input event.Event into content to be written
// (and consumed)
type LogFormatter interface {
	Format(log *event.Event) (buf []byte, err error)
}

type formatConfig struct {
	f LogFormatter
}

// Apply method is the default implementation of LoggerConfig, where the builder's formatter
// is set as the one in this formatConfig
func (f *formatConfig) Apply(lb *LoggerBuilder) {
	lb.Fmt = f.f
}

// WithFormat is a wrapper to use LogFormatters as LoggerConfigs.
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

// LogFormatConfigs is a map of LoggerConfigs indexed by an int value. This is done in a map
// and not a list for manual ordering, spacing and manipulation of preset entries
var LogFormatConfigs = map[int]LoggerConfig{
	0:  WithFormat(LogFormatters[0]),
	1:  WithFormat(LogFormatters[1]),
	2:  WithFormat(LogFormatters[2]),
	3:  WithFormat(LogFormatters[3]),
	4:  WithFormat(LogFormatters[4]),
	5:  WithFormat(LogFormatters[5]),
	6:  WithFormat(LogFormatters[6]),
	7:  WithFormat(LogFormatters[7]),
	8:  WithFormat(LogFormatters[8]),
	9:  WithFormat(LogFormatters[9]),
	10: WithFormat(LogFormatters[10]),
	11: WithFormat(LogFormatters[11]),
	12: WithFormat(LogFormatters[12]),
	13: WithFormat(LogFormatters[13]),
	14: WithFormat(LogFormatters[14]),
	15: WithFormat(LogFormatters[15]),
	16: WithFormat(LogFormatters[16]),
	17: WithFormat(LogFormatters[17]),
	18: WithFormat(LogFormatters[18]),
	19: WithFormat(LogFormatters[19]),
	20: WithFormat(LogFormatters[20]),
	21: WithFormat(LogFormatters[21]),
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

var (
	CfgFormatText                LoggerConfig = LogFormatConfigs[0]  // placeholder for an initialized Text LoggerConfig
	CfgFormatProtobuf            LoggerConfig = LogFormatConfigs[1]  // placeholder for an initialized Protobuf LoggerConfig
	CfgFormatJSON                LoggerConfig = LogFormatConfigs[2]  // placeholder for an initialized JSON LoggerConfig
	CfgFormatCSV                 LoggerConfig = LogFormatConfigs[3]  // placeholder for an initialized CSV LoggerConfig
	CfgFormatXML                 LoggerConfig = LogFormatConfigs[4]  // placeholder for an initialized XML LoggerConfig
	CfgFormatGob                 LoggerConfig = LogFormatConfigs[5]  // placeholder for an initialized Gob LoggerConfig
	CfgFormatBSON                LoggerConfig = LogFormatConfigs[6]  // placeholder for an initialized JSON LoggerConfig
	CfgTextLongDate              LoggerConfig = LogFormatConfigs[7]  // placeholder for an initialized Text LoggerConfig, with a RFC3339 date format
	CfgTextShortDate             LoggerConfig = LogFormatConfigs[8]  // placeholder for an initialized Text LoggerConfig, with a RFC822Z date format
	CfgTextRubyDate              LoggerConfig = LogFormatConfigs[9]  // placeholder for an initialized Text LoggerConfig, with a RubyDate date format
	CfgTextDoubleSpace           LoggerConfig = LogFormatConfigs[10] // placeholder for an initialized Text LoggerConfig, with double spaces
	CfgTextLevelFirstSpaced      LoggerConfig = LogFormatConfigs[11] // placeholder for an initialized  LoggerConfig, with level-first and double spaces
	CfgTextLevelFirst            LoggerConfig = LogFormatConfigs[12] // placeholder for an initialized  LoggerConfig, with level-first
	CfgTextColorDoubleSpace      LoggerConfig = LogFormatConfigs[13] // placeholder for an initialized  LoggerConfig, with color and double spaces
	CfgTextColorLevelFirstSpaced LoggerConfig = LogFormatConfigs[14] // placeholder for an initialized  LoggerConfig, with color, level-first and double spaces
	CfgTextColorLevelFirst       LoggerConfig = LogFormatConfigs[15] // placeholder for an initialized  LoggerConfig, with color and level-first
	CfgTextColor                 LoggerConfig = LogFormatConfigs[16] // placeholder for an initialized  LoggerConfig, with color
	CfgTextOnly                  LoggerConfig = LogFormatConfigs[17] // placeholder for an initialized  LoggerConfig, with only the text content
	CfgTextNoHeaders             LoggerConfig = LogFormatConfigs[18] // placeholder for an initialized  LoggerConfig, without headers
	CfgTextNoTimestamp           LoggerConfig = LogFormatConfigs[19] // placeholder for an initialized  LoggerConfig, without timestamp
	CfgTextColorNoTimestamp      LoggerConfig = LogFormatConfigs[20] // placeholder for an initialized  LoggerConfig, without timestamp
	CfgTextColorUpperNoTimestamp LoggerConfig = LogFormatConfigs[21] // placeholder for an initialized  LoggerConfig, without timestamp and uppercase headers
)
