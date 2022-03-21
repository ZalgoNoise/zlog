package server

import (
	"github.com/zalgonoise/zlog/log"
	"google.golang.org/grpc"
)

type LogServerConfig interface {
	Apply(ls *GRPCLogServer)
}

type LSAddr struct {
	addr string
}

func WithAddr(addr string) LogServerConfig {
	// enforce defaults
	if addr == "" || addr == ":" {
		addr = ":9099"
	}

	return &LSAddr{
		addr: addr,
	}
}

func (l LSAddr) Apply(ls *GRPCLogServer) {
	ls.Addr = l.addr
}

type LSLogger struct {
	logger log.Logger
}

func WithLogger(loggers ...log.Logger) LogServerConfig {
	if len(loggers) == 1 {
		return &LSLogger{
			logger: loggers[0],
		}
	}

	if len(loggers) > 1 {
		return &LSLogger{
			logger: log.MultiLogger(loggers...),
		}
	}

	return &LSLogger{
		logger: log.New(),
	}
}

func (l LSLogger) Apply(ls *GRPCLogServer) {
	ls.Logger = l.logger
}

type LSServiceLogger struct {
	logger log.Logger
}

func WithServiceLogger(loggers ...log.Logger) LogServerConfig {

	if len(loggers) == 1 {
		return &LSServiceLogger{
			logger: loggers[0],
		}
	}

	if len(loggers) > 1 {
		return &LSServiceLogger{
			logger: log.MultiLogger(loggers...),
		}
	}

	return &LSServiceLogger{
		logger: log.New(log.NilConfig),
	}
}

func (l LSServiceLogger) Apply(ls *GRPCLogServer) {
	ls.SvcLogger = l.logger
}

type LSOpts struct {
	opts []grpc.ServerOption
}

func WithGRPCOpts(opts ...grpc.ServerOption) LogServerConfig {
	if opts != nil {
		return &LSOpts{
			opts: opts,
		}
	}
	return &LSOpts{
		opts: []grpc.ServerOption{},
	}

}

func (l LSOpts) Apply(ls *GRPCLogServer) {
	if len(ls.opts) == 0 {
		ls.opts = l.opts
		return
	}
	ls.opts = append(ls.opts, l.opts...)
}
