package log

import (
	"io"
	"os"
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
		return defaultConfig
	}

	allConf := make([]LoggerConfig, 0, len(conf))
	allConf = append(allConf, conf...)

	return &multiconf{allConf}
}

// Apply method will make a multiconf-type of LoggerConfig iterate through all its objects and
// run the Apply method on the input pointer to a LoggerBuilder
func (m multiconf) Apply(lb *LoggerBuilder) {
	for _, c := range m.confs {
		c.Apply(lb)
	}
}

var defaultConfig LoggerConfig = &multiconf{
	confs: []LoggerConfig{
		TextFormat,
		WithOut(),
		WithPrefix("log"),
	},
}

// LoggerConfigs is a map of LoggerConfig indexed by an int value. This is done in a map
// and not a list for manual ordering, spacing and manipulation of preset entries
var LoggerConfigs = map[int]LoggerConfig{
	0:  defaultConfig,
	1:  LCSkipExit{},
	5:  TextFormat,
	6:  JSONFormat,
	7:  WithOut(os.Stdout),
	8:  WithPrefix("log"),
	9:  WithFilter(LLInfo),
	10: WithFilter(LLWarn),
	11: WithFilter(LLError),
}

var (
	DefaultCfg   LoggerConfig = LoggerConfigs[0]  // placeholder for an initialized default LoggerConfig
	SkipExitCfg  LoggerConfig = LoggerConfigs[1]  // placeholder for an initialized skip-exits LoggerConfig
	TextCfg      LoggerConfig = LoggerConfigs[5]  // placeholder for an initialized Text-LogFormatter LoggerConfig
	JSONCfg      LoggerConfig = LoggerConfigs[6]  // placeholder for an initialized JSON-LogFormatter LoggerConfig
	StdOutCfg    LoggerConfig = LoggerConfigs[7]  // placeholder for an initialized os.Stdout LoggerConfig
	DefPrefixCfg LoggerConfig = LoggerConfigs[8]  // placeholder for an initialized default-prefix LoggerConfig
	InfoFilter   LoggerConfig = LoggerConfigs[9]  // placeholder for an initialized Info-filtered LoggerConfig
	WarnFilter   LoggerConfig = LoggerConfigs[10] // placeholder for an initialized Warn-filtered LoggerConfig
	ErrorFilter  LoggerConfig = LoggerConfigs[11] // placeholder for an initialized Error-filtered LoggerConfig
)

// LCPrefix struct is a custom LoggerConfig to define prefixes to new Loggers
type LCPrefix struct {
	p string
}

// WithPrefix function will allow creating a LoggerConfig that applies a prefix string to a Logger
func WithPrefix(prefix string) LoggerConfig {
	return &LCPrefix{
		p: prefix,
	}
}

// Apply method will set the configured prefix string to the input pointer to a LoggerBuilder
func (c *LCPrefix) Apply(lb *LoggerBuilder) {
	lb.prefix = c.p
}

// LCOut struct is a custom LoggerConfig to define the output io.Writer to new Loggers
type LCOut struct {
	out io.Writer
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
		out: os.Stdout,
	}
}

// Apply method will set the configured output io.Writer to the input pointer to a LoggerBuilder
func (c *LCOut) Apply(lb *LoggerBuilder) {
	lb.out = c.out
}

// LCSkipExit stuct is a custom LoggerConfig to define whether os.Exit(1) and panic() calls
// should be respected or skipped
type LCSkipExit struct{}

// Apply method will set the configured skipExit option to true, in the input pointer to a LoggerBuilder
//
// Not setting this option will default to "false", or to respect os.Exit(1) and panic() calls
func (c LCSkipExit) Apply(lb *LoggerBuilder) {
	lb.skipExit = true
}

// LCSkipExit stuct is a custom LoggerConfig to filter Logger writes as per the message's
// log level
type LCFilter struct {
	l LogLevel
}

// WithFilter function will allow filtering Logger writes, only to contain a certain log level
// and above.
//
// This method can be used to either separate different logging severities or to reduce the amount
// of bytes written to a logfile, by skipping more trivial messages
func WithFilter(level LogLevel) LoggerConfig {
	return &LCFilter{
		l: level,
	}
}

// Apply method will set the configured level filter to the input pointer to a LoggerBuilder
func (c LCFilter) Apply(lb *LoggerBuilder) {
	lb.levelFilter = c.l.Int()
}
