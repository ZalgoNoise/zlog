package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"time"

	"github.com/zalgonoise/zlog/grpc/address"
	"github.com/zalgonoise/zlog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrCACertAddFailed error = errors.New("failed to add server CA's certificate")
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

func Insecure() LogClientConfig {
	return &LSOpts{
		opts: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	}
}

func WithTLS(caPath string, certKeyPair ...string) LogClientConfig {
	var cred credentials.TransportCredentials
	var err error

	if len(certKeyPair) == 0 {
		cred, err = loadCreds(caPath)
	} else if len(certKeyPair) > 1 {

		// despite the variatic parameter, only the first two elements are read
		// this is so it can be fully omitted if it's for server-TLS only
		cred, err = loadCredsMutual(caPath, certKeyPair[0], certKeyPair[1])

	} else {
		return nil
	}

	if err != nil {
		// panic since the gRPC client shouldn't start
		// if TLS is requested but invalid / errored
		panic(err)
	}

	return &LSOpts{
		opts: []grpc.DialOption{
			grpc.WithTransportCredentials(cred),
		},
	}
}

func loadCredsMutual(caCert, cert, key string) (credentials.TransportCredentials, error) {
	ca, err := ioutil.ReadFile(caCert)

	if err != nil {
		return nil, err
	}

	crtPool := x509.NewCertPool()

	if ok := crtPool.AppendCertsFromPEM(ca); !ok {
		return nil, ErrCACertAddFailed
	}

	c, err := tls.LoadX509KeyPair(cert, key)

	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{c},
		ClientCAs:    crtPool,
	}

	return credentials.NewTLS(config), nil
}

func loadCreds(caCert string) (credentials.TransportCredentials, error) {
	ca, err := ioutil.ReadFile(caCert)
	if err != nil {
		return nil, err
	}

	crtPool := x509.NewCertPool()

	if ok := crtPool.AppendCertsFromPEM(ca); !ok {
		return nil, ErrCACertAddFailed
	}

	config := &tls.Config{
		RootCAs: crtPool,
	}

	return credentials.NewTLS(config), nil
}

func (l LSOpts) Apply(ls *GRPCLogClientBuilder) {
	if len(ls.opts) == 0 {
		ls.opts = l.opts
		return
	}
	ls.opts = append(ls.opts, l.opts...)
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

type LSExpBackoff struct {
	backoff time.Duration
}

func WithBackoff(t time.Duration) LogClientConfig {
	// default config
	if t == 0 || t == defaultRetryTime {
		return &LSExpBackoff{
			backoff: defaultRetryTime,
		}
	}

	return &LSExpBackoff{
		backoff: t,
	}
}

func (l LSExpBackoff) Apply(ls *GRPCLogClientBuilder) {
	ls.expBackoff = l.backoff
}
