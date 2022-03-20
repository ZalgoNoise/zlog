package client

import (
	"github.com/zalgonoise/zlog/grpc/address"
	"github.com/zalgonoise/zlog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LogClientConfig interface {
	Apply(ls *GRPCLogClientBuilder)
}

type LSAddr struct {
	addr address.ConnAddr
}

func WithAddr(addr ...string) LogClientConfig {
	a := &LSAddr{
		addr: map[string]*grpc.ClientConn{},
	}

	if len(addr) == 0 || addr == nil {
		a.addr.Add(":9099")
		return a
	}

	a.addr.Add(addr...)

	return a
}

func (l LSAddr) Apply(ls *GRPCLogClientBuilder) {
	ls.addr = &l.addr
}

type LSOpts struct {
	opts []grpc.DialOption
}

var defaultDialOptions = []grpc.DialOption{
	grpc.WithTransportCredentials(insecure.NewCredentials()),
	grpc.FailOnNonTempDialError(true),
	grpc.WithReturnConnectionError(),
	grpc.WithDisableRetry(),
}

func WithGRPCOpts(opts ...grpc.DialOption) LogClientConfig {
	if opts != nil {
		// enforce defaults
		if len(opts) == 0 {
			return &LSOpts{opts: defaultDialOptions}
		}
		return &LSOpts{
			opts: opts,
		}
	}
	return &LSOpts{opts: defaultDialOptions}

}

func (l LSOpts) Apply(ls *GRPCLogClientBuilder) {
	ls.opts = l.opts
}

type LSType struct {
	isUnary bool
}

func StreamRPC() LogClientConfig {
	return &LSType{
		isUnary: false,
	}

}

func UnaryRPC() LogClientConfig {
	return &LSType{
		isUnary: true,
	}
}

func (l LSType) Apply(ls *GRPCLogClientBuilder) {
	ls.isUnary = l.isUnary
}

type LSLogger struct {
	logger log.Logger
}

func WithLogger(loggers ...log.Logger) LogClientConfig {
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
		logger: log.New(log.NilConfig),
	}
}

func (l LSLogger) Apply(ls *GRPCLogClientBuilder) {
	ls.svcLogger = l.logger
}
