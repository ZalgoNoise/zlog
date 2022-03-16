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
	logger log.LoggerI
}

func WithLogger(loggers ...log.LoggerI) LogServerConfig {
	l := &LSLogger{}

	// enforce defaults
	if len(loggers) == 0 {
		l.logger = log.New()
	}

	if len(loggers) == 1 {
		l.logger = loggers[0]
	}

	if len(loggers) > 1 {
		l.logger = log.MultiLogger(loggers...)
	}

	return l
}

func (l LSLogger) Apply(ls *GRPCLogServer) {
	ls.Logger = l.logger
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
	ls.opts = l.opts
}
