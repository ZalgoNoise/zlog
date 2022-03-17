package client

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LogClientConfig interface {
	Apply(ls *GRPCLogClient)
}

const defaultTimeout = time.Second * 30

type LSContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type ContextValue string

func WithContext(ctx context.Context, cancel context.CancelFunc) LogClientConfig {
	if ctx == nil {
		return NewContext()
	}

	return &LSContext{
		ctx, cancel,
	}
}

func NewContext() LogClientConfig {
	bgCtx := context.Background()
	vctx := context.WithValue(bgCtx, ContextValue("id"), ContextValue(uuid.New().String()))

	c, cl := context.WithTimeout(vctx, defaultTimeout)
	new := &LSContext{
		c, cl,
	}

	return new
}

func (l LSContext) Apply(ls *GRPCLogClient) {
	ls.Ctx = l.ctx
}

type LSAddr struct {
	addr ConnAddr
}

func WithAddr(addr ...string) LogClientConfig {
	a := &LSAddr{}

	if len(addr) == 0 || addr == nil {
		a.addr.Add(":9099")
		return a
	}

	a.addr.Add(addr...)

	return a
}

func (l LSAddr) Apply(ls *GRPCLogClient) {
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

func (l LSOpts) Apply(ls *GRPCLogClient) {
	ls.opts = l.opts
}
