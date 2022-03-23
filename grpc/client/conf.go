package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"time"

	"github.com/zalgonoise/zlog/config"
	"github.com/zalgonoise/zlog/grpc/address"
	"github.com/zalgonoise/zlog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrCACertAddFailed error = errors.New("failed to add server CA's certificate")

	gRPCLogClientBuilderType = &GRPCLogClientBuilder{}

	defaultDialOptions = []grpc.DialOption{
		grpc.FailOnNonTempDialError(true),
		grpc.WithReturnConnectionError(),
		grpc.WithDisableRetry(),
	}

	insecureDialOptions = []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
)

type LogClientConf int32

const (
	LSAddress LogClientConf = iota
	LSType
	LSLogging
	LSTLS
	LSBackoff
	LSGRPCOpts
)

var (
	LogClientConfKeys = map[string]LogClientConf{
		"addr":     0,
		"type":     1,
		"logging":  2,
		"tls":      3,
		"backoff":  4,
		"grpcopts": 5,
	}

	LogClientConfVals = map[int32]string{
		0: "addr",
		1: "type",
		2: "logging",
		3: "tls",
		4: "backoff",
		5: "grpcopts",
	}
)

func (c LogClientConf) Int32() int32 {
	return int32(c)
}

func (c LogClientConf) String() string {
	return LogClientConfVals[c.Int32()]
}

func gRPCLogClientDefaults() *config.Configs {
	return config.NewMap(
		WithAddr(""),
		StreamRPC(),
		WithLogger(),
		Insecure(),
		WithBackoff(0),
		WithGRPCOpts(),
	)
}

func WithAddr(addr ...string) config.Config {
	var connAddr address.ConnAddr = map[string]*grpc.ClientConn{}
	var addrCfg = config.New(LSAddress.String(), gRPCLogClientBuilderType)

	if len(addr) == 0 || addr == nil {
		connAddr.Add(":9099")
		return config.WithValue(addrCfg, &connAddr)
	}

	connAddr.Add(addr...)
	return config.WithValue(addrCfg, &connAddr)
}

func StreamRPC() config.Config {
	var typeCfg = config.New(LSType.String(), gRPCLogClientBuilderType)
	return config.WithValue(typeCfg, "stream")
}

func UnaryRPC() config.Config {
	var typeCfg = config.New(LSType.String(), gRPCLogClientBuilderType)
	return config.WithValue(typeCfg, "unary")
}

func WithLogger(loggers ...log.Logger) config.Config {
	var logCfg = config.New(LSLogging.String(), gRPCLogClientBuilderType)
	if len(loggers) == 1 {
		return config.WithValue(logCfg, loggers[0])
	}

	if len(loggers) > 1 {
		return config.WithValue(logCfg, log.MultiLogger(loggers...))
	}

	return config.WithValue(logCfg, log.New(log.NilConfig))
}

func Insecure() config.Config {
	var tlsCfg = config.New(LSTLS.String(), gRPCLogClientBuilderType)
	return config.WithValue(tlsCfg, insecureDialOptions)
}

func WithTLS(caPath string, certKeyPair ...string) config.Config {
	var tlsCfg = config.New(LSTLS.String(), gRPCLogClientBuilderType)
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

	return config.WithValue(tlsCfg, []grpc.DialOption{
		grpc.WithTransportCredentials(cred),
	})
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

func WithBackoff(t time.Duration) config.Config {
	var backoffCfg = config.New(LSBackoff.String(), gRPCLogClientBuilderType)
	b := NewBackoff()

	// default config
	if t == 0 || t == defaultRetryTime {
		return config.WithValue(backoffCfg, b.Time(defaultRetryTime))
	}

	return config.WithValue(backoffCfg, b.Time(t))
}

func WithGRPCOpts(opts ...grpc.DialOption) config.Config {
	var optsCfg = config.New(LSGRPCOpts.String(), gRPCLogClientBuilderType)

	if len(opts) == 0 {
		// enforce defaults
		return config.WithValue(optsCfg, defaultDialOptions)
	}

	return config.WithValue(optsCfg, opts)
}
