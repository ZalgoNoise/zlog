package log

import (
	"io"
	"os"
)

type LoggerConfig interface {
	Apply(*LoggerBuilder)
}

type multiconf struct {
	confs []LoggerConfig
}

func MultiConf(conf ...LoggerConfig) LoggerConfig {
	allConf := make([]LoggerConfig, 0, len(conf))
	allConf = append(allConf, conf...)

	return &multiconf{allConf}
}

func (m multiconf) Apply(lb *LoggerBuilder) {
	for _, c := range m.confs {
		c.Apply(lb)
	}
}

var defaultConfig LoggerConfig = &multiconf{
	confs: []LoggerConfig{
		LCTextFormat{},
		LCStdOut{},
		LCDefaultPrefix{},
	},
}

var LoggerConfigs = map[int]LoggerConfig{
	0: defaultConfig,
	5: LCTextFormat{},
	6: LCJSONFormat{},
	7: LCStdOut{},
	8: LCDefaultPrefix{},
}

var (
	DefaultCfg   LoggerConfig = LoggerConfigs[0]
	TextCfg      LoggerConfig = LoggerConfigs[5]
	JSONCfg      LoggerConfig = LoggerConfigs[6]
	StdOutCfg    LoggerConfig = LoggerConfigs[7]
	DefPrefixCfg LoggerConfig = LoggerConfigs[8]
)

type LCTextFormat struct{}

func (c LCTextFormat) Apply(lb *LoggerBuilder) {
	lb.fmt = TextFormat
}

type LCJSONFormat struct{}

func (c LCJSONFormat) Apply(lb *LoggerBuilder) {
	lb.fmt = JSONFormat
}

type LCStdOut struct{}

func (c LCStdOut) Apply(lb *LoggerBuilder) {
	lb.out = os.Stdout
}

type LCDefaultPrefix struct{}

func (c LCDefaultPrefix) Apply(lb *LoggerBuilder) {
	lb.prefix = "log"
}

type LCPrefix struct {
	p string
}

func WithPrefix(prefix string) *LCPrefix {
	return &LCPrefix{
		p: prefix,
	}
}

func (c *LCPrefix) Apply(lb *LoggerBuilder) {
	lb.prefix = c.p
}

type LCOut struct {
	out io.Writer
}

func WithOut(out ...io.Writer) *LCOut {
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

func (c *LCOut) Apply(lb *LoggerBuilder) {
	lb.out = c.out
}
