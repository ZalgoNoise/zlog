package log

import (
	"io"
	"os"

	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/log/format/protobuf"
	"github.com/zalgonoise/zlog/log/format/text"
	"github.com/zalgonoise/zlog/store"
	"github.com/zalgonoise/zlog/store/db"
)

// LoggerConfig interface describes the behavior that a LoggerConfig object should have
//
// The single Apply(lb *LoggerBuilder) method allows for different modules to apply changes to a
// LoggerBuilder, in a non-blocking way for other features.
//
// Each feature should implement its own structs with their own methods; where they can implement
// Apply(lb *LoggerBuilder) to set their own configurations to the input LoggerBuilder
type LoggerConfig interface {
	Apply(lb *LoggerBuilder)
}

type multiconf struct {
	confs []LoggerConfig
}

// MultiConf function is a wrapper for multiple configs to be bundled (and executed) in one shot.
//
// Similar to io.MultiWriter, it will iterate through all set LoggerConfig and run the same method
// on each of them.
func MultiConf(conf ...LoggerConfig) LoggerConfig {
	if len(conf) == 0 {
		return DefaultConfig
	}

	allConf := make([]LoggerConfig, 0, len(conf))
	allConf = append(allConf, conf...)

	return &multiconf{allConf}
}

// Apply method will make a multiconf-type of LoggerConfig iterate through all its objects and
// run the Apply method on the input pointer to a LoggerBuilder
func (m multiconf) Apply(lb *LoggerBuilder) {
	for _, c := range m.confs {
		if c != nil {
			c.Apply(lb)
		}
	}
}

var (
	// default configuration for a Logger
	DefaultConfig LoggerConfig = &multiconf{
		confs: []LoggerConfig{
			WithFormat(TextColorLevelFirst),
			WithOut(),
			WithPrefix(event.Default_Event_Prefix),
		},
	}

	// LoggerConfigs is a map of LoggerConfig indexed by an int value. This is done in a map
	// and not a list for manual ordering, spacing and manipulation of preset entries
	LoggerConfigs = map[int]LoggerConfig{
		0:  DefaultConfig,
		1:  LCSkipExit{},
		7:  WithOut(os.Stderr),
		8:  WithPrefix(event.Default_Event_Prefix),
		9:  WithFilter(event.Level_info),
		10: WithFilter(event.Level_warn),
		11: WithFilter(event.Level_error),
		12: NilLogger(),
	}

	DefaultCfg    LoggerConfig = LoggerConfigs[0]  // placeholder for an initialized default LoggerConfig
	SkipExit      LoggerConfig = LoggerConfigs[1]  // placeholder for an initialized skip-exits LoggerConfig
	StdOut        LoggerConfig = LoggerConfigs[7]  // placeholder for an initialized os.Stderr LoggerConfig
	PrefixDefault LoggerConfig = LoggerConfigs[8]  // placeholder for an initialized default-prefix LoggerConfig
	FilterInfo    LoggerConfig = LoggerConfigs[9]  // placeholder for an initialized Info-filtered LoggerConfig
	FilterWarn    LoggerConfig = LoggerConfigs[10] // placeholder for an initialized Warn-filtered LoggerConfig
	FilterError   LoggerConfig = LoggerConfigs[11] // placeholder for an initialized Error-filtered LoggerConfig
	NilConfig     LoggerConfig = LoggerConfigs[12] // placeholder for an initialized empty / nil LoggerConfig
	EmptyConfig   LoggerConfig = LoggerConfigs[12] // placeholder for an initialized empty / nil LoggerConfig
)

// LCPrefix struct is a custom LoggerConfig to define prefixes to new Loggers
type LCPrefix struct {
	p string
}

// LCSub struct is a custom LoggerConfig to define sub-prefixes to new Loggers
type LCSub struct {
	s string
}

// LCOut struct is a custom LoggerConfig to define the output io.Writer to new Loggers
type LCOut struct {
	out io.Writer
}

// LCSkipExit stuct is a custom LoggerConfig to define whether os.Exit(1) and panic() calls
// should be respected or skipped
type LCSkipExit struct{}

// LCSkipExit stuct is a custom LoggerConfig to filter Logger writes as per the message's
// log level
type LCFilter struct {
	l event.Level
}

// LCDatabase struct defines the Logger Config object that adds a DBWriter as a Logger writer
type LCDatabase struct {
	Out io.WriteCloser
	Fmt LogFormatter
}

// Apply method will set the configured prefix string to the input pointer to a LoggerBuilder
func (c *LCPrefix) Apply(lb *LoggerBuilder) {
	lb.Prefix = c.p
}

// Apply method will set the configured sub-prefix string to the input pointer to a LoggerBuilder
func (c *LCSub) Apply(lb *LoggerBuilder) {
	lb.Sub = c.s
}

// Apply method will set the configured output io.Writer to the input pointer to a LoggerBuilder
func (c *LCOut) Apply(lb *LoggerBuilder) {
	lb.Out = c.out
}

// Apply method will set the configured skipExit option to true, in the input pointer to a LoggerBuilder
//
// Not setting this option will default to "false", or to respect os.Exit(1) and panic() calls
func (c LCSkipExit) Apply(lb *LoggerBuilder) {
	lb.SkipExit = true
}

// Apply method will set the configured level filter to the input pointer to a LoggerBuilder
func (c LCFilter) Apply(lb *LoggerBuilder) {
	lb.LevelFilter = c.l.Int()
}

// Apply method will set the input LoggerBuilder's outputs and format to the LCDatabase object's.
func (c *LCDatabase) Apply(lb *LoggerBuilder) {
	lb.Out = c.Out
	lb.Fmt = c.Fmt
}

// NilLogger function will create a minimal LoggerConfig with an empty writer, and that does not
// comply with exit (os.Exit(1) and panic()) signals
func NilLogger() LoggerConfig {
	return MultiConf(
		&LCOut{
			out: store.EmptyWriter,
		},
		WithFormat(text.New().NoHeaders().NoTimestamp().NoLevel().Build()),
		&LCSkipExit{},
	)
}

// WithPrefix function will allow creating a LoggerConfig that applies a prefix string to a Logger
func WithPrefix(prefix string) LoggerConfig {
	return &LCPrefix{
		p: prefix,
	}
}

// WithSub function will allow creating a LoggerConfig that applies a sub-prefix string to a Logger
func WithSub(sub string) LoggerConfig {
	return &LCSub{
		s: sub,
	}
}

// WithOut function will allow creating a LoggerConfig that applies a (number of) io.Writer to a Logger
func WithOut(out ...io.Writer) LoggerConfig {
	if len(out) == 1 {
		return &LCOut{
			out: out[0],
		}
	}

	if len(out) > 1 {
		return &LCOut{
			out: io.MultiWriter(out...),
		}
	}

	// default
	return &LCOut{
		out: os.Stderr,
	}
}

// WithFilter function will allow filtering Logger writes, only to contain a certain log level
// and above.
//
// This method can be used to either separate different logging severities or to reduce the amount
// of bytes written to a logfile, by skipping more trivial messages
func WithFilter(level event.Level) LoggerConfig {
	return &LCFilter{
		l: level,
	}
}

// WithDatabase function creates a Logger config to use a database as a writer, while protobuf encoding
// events, to transmit to writers
func WithDatabase(dbs ...io.WriteCloser) LoggerConfig {
	if len(dbs) == 0 {
		return nil
	}

	var w io.WriteCloser

	if len(dbs) == 1 {
		w = dbs[0]
	} else if len(dbs) > 1 {
		w = db.MultiWriteCloser(dbs...)
	}

	return &LCDatabase{
		Out: w,
		Fmt: &protobuf.FmtPB{},
	}
}
