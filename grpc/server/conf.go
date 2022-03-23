package server

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"

	"github.com/zalgonoise/zlog/config"
	"github.com/zalgonoise/zlog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	ErrCACertAddFailed error = errors.New("failed to add server CA's certificate")

	gRPCLogServerBuilderType *GRPCLogServerBuilder = &GRPCLogServerBuilder{}
)

type LogServerConf int32

const (
	LSAddress LogServerConf = iota
	LSOpts
	LSLogger
	LSSvcLogger
	LSTLS
)

var (
	LogServerConfKeys = map[string]LogServerConf{
		"addr":   0,
		"opts":   1,
		"logger": 2,
		"svclog": 3,
		"tls":    4,
	}
	LogServerConfVals = map[int32]string{
		0: "addr",
		1: "opts",
		2: "logger",
		3: "svclog",
		4: "tls",
	}
)

func (c LogServerConf) Int32() int32 {
	return int32(c)
}

func (c LogServerConf) String() string {
	return LogServerConfVals[c.Int32()]
}

func gRPCLogServerDefaults() *config.Configs {
	return config.NewMap(
		WithAddr(""),
		WithLogger(),
		WithServiceLogger(),
		WithGRPCOpts(),
		noTLS(),
	)
}

// type LogServerConfig interface {
// 	Apply(ls *GRPCLogServer)
// }

// type LSAddr struct {
// 	addr string
// }

func WithAddr(addr string) config.Config {
	var cfg = config.New(LSAddress.String(), gRPCLogServerBuilderType)

	// enforce defaults
	if addr == "" || addr == ":" {
		addr = ":9099"
	}

	return config.WithValue(cfg, addr)
}

// func (l LSAddr) Apply(ls *GRPCLogServer) {
// 	ls.Addr = l.addr
// }

// type LSLogger struct {
// 	logger log.Logger
// }

func WithLogger(loggers ...log.Logger) config.Config {
	var cfg = config.New(LSLogger.String(), gRPCLogServerBuilderType)

	if len(loggers) == 1 {
		return config.WithValue(cfg, loggers[0])
	}

	if len(loggers) > 1 {
		return config.WithValue(cfg, log.MultiLogger(loggers...))
	}

	return config.WithValue(cfg, log.New())
}

// func (l LSLogger) Apply(ls *GRPCLogServer) {
// 	ls.Logger = l.logger
// }

// type LSServiceLogger struct {
// 	logger log.Logger
// }

func WithServiceLogger(loggers ...log.Logger) config.Config {
	var cfg = config.New(LSSvcLogger.String(), gRPCLogServerBuilderType)

	if len(loggers) == 1 {
		return config.WithValue(cfg, loggers[0])
	}

	if len(loggers) > 1 {
		return config.WithValue(cfg, log.MultiLogger(loggers...))
	}

	return config.WithValue(cfg, log.New(log.NilConfig))
}

// func (l LSServiceLogger) Apply(ls *GRPCLogServer) {
// 	ls.SvcLogger = l.logger
// }

// type LSOpts struct {
// 	opts []grpc.ServerOption
// }

func WithGRPCOpts(opts ...grpc.ServerOption) config.Config {
	var cfg = config.New(LSOpts.String(), gRPCLogServerBuilderType)

	if opts != nil {
		return config.WithValue(cfg, opts)
	}

	return config.WithValue(cfg, []grpc.ServerOption{})

}

// func (l LSOpts) Apply(ls *GRPCLogServer) {
// 	if len(ls.opts) == 0 {
// 		ls.opts = l.opts
// 		return
// 	}
// 	ls.opts = append(ls.opts, l.opts...)
// }

func WithTLS(certPath, keyPath string, caPath ...string) config.Config {
	var cfg = config.New(LSTLS.String(), gRPCLogServerBuilderType)

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

	return config.WithValue(cfg, []grpc.ServerOption{
		grpc.Creds(cred),
	})

}

func noTLS() config.Config {
	var cfg = config.New(LSTLS.String(), gRPCLogServerBuilderType)
	return config.WithValue(cfg, []grpc.ServerOption{})
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
