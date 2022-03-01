package log

import "os"

type LoggerConfig interface {
	Apply(*LoggerBuilder)
}

var LoggerConfigs = map[int]LoggerConfig{
	0: LCDefault{},
}

var (
	DefaultConfig LoggerConfig = LoggerConfigs[0]
)

type LCDefault struct{}

func (c LCDefault) Apply(lb *LoggerBuilder) {
	lb.fmt = TextFormat
	lb.out = os.Stdout
	lb.prefix = "log"
}
