package client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LogClientConfig interface {
	Apply(ls *GRPCLogClientBuilder)
}

type LSAddr struct {
	addr ConnAddr
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

func WithGRPCOpts(opts ...grpc.DialOption) LogClientConfig {
	if opts != nil {
		// enforce defaults
		if len(opts) == 0 {
			return &LSOpts{
				opts: []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
			}
		}
		return &LSOpts{
			opts: opts,
		}
	}
	return &LSOpts{
		opts: []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	}

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
