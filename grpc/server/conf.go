package server

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"

	"github.com/zalgonoise/zlog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	ErrCACertAddFailed error = errors.New("failed to add server CA's certificate")
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

func WithTLS(certPath, keyPath string, caPath ...string) LogServerConfig {
	var cred credentials.TransportCredentials
	var err error

	if len(caPath) == 0 {
		cred, err = loadCreds(certPath, keyPath)

		// despite the variatic parameter, only the first element is read
		// this is so it can be fully omitted if it's for server-TLS only
	} else if len(caPath) > 0 {
		cred, err = loadCredsMutual(caPath[0], certPath, keyPath)
	} else {
		return nil
	}

	if err != nil {
		// panic since the gRPC server shouldn't start
		// if TLS is requested but invalid / errored
		panic(err)
	}

	return &LSOpts{
		opts: []grpc.ServerOption{
			grpc.Creds(cred),
		},
	}
}

func loadCreds(cert, key string) (credentials.TransportCredentials, error) {
	c, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{c},
		ClientAuth:   tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
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
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    crtPool,
	}

	return credentials.NewTLS(config), nil
}
