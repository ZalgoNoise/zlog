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
	0: defaultConfig,
	5: TextFormat,
	6: JSONFormat,
	7: WithOut(os.Stdout),
	8: WithPrefix("log"),
}

var (
	DefaultCfg   LoggerConfig = LoggerConfigs[0] // placeholder for an initialized default LoggerConfig
	TextCfg      LoggerConfig = LoggerConfigs[5] // placeholder for an initialized Text-LogFormatter LoggerConfig
	JSONCfg      LoggerConfig = LoggerConfigs[6] // placeholder for an initialized JSON-LogFormatter LoggerConfig
	StdOutCfg    LoggerConfig = LoggerConfigs[7] // placeholder for an initialized os.Stdout LoggerConfig
	DefPrefixCfg LoggerConfig = LoggerConfigs[8] // placeholder for an initialized default-prefix LoggerConfig
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
	if len(out) == 0 {
		return &LCOut{
			out: os.Stdout,
		}
	}

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
