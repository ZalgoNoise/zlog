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
	LogClientConfRequired = []LogClientConf{0, 1, 2, 3, 4, 5}
	LogClientConfKeys     = map[string]LogClientConf{
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
	addrCfg    = config.New(LSAddress.String(), gRPCLogClientBuilderType)
	typeCfg    = config.New(LSType.String(), gRPCLogClientBuilderType)
	logCfg     = config.New(LSLogging.String(), gRPCLogClientBuilderType)
	tlsCfg     = config.New(LSTLS.String(), gRPCLogClientBuilderType)
	optsCfg    = config.New(LSGRPCOpts.String(), gRPCLogClientBuilderType)
	backoffCfg = config.New(LSBackoff.String(), gRPCLogClientBuilderType)

	LogClientDefaults = map[LogClientConf]config.Config{
		0: config.WithDefault(addrCfg, func() config.Config {
			return WithAddr("")
		}),
		1: config.WithDefault(typeCfg, func() config.Config {
			return StreamRPC()
		}),
		2: config.WithDefault(logCfg, func() config.Config {
			return WithLogger()
		}),
		3: config.WithDefault(tlsCfg, func() config.Config {
			return Insecure()
		}),
		4: config.WithDefault(backoffCfg, func() config.Config {
			return WithBackoff(0)
		}),
		5: config.WithDefault(optsCfg, func() config.Config {
			return WithGRPCOpts()
		}),
	}
)

func (c LogClientConf) Int32() int32 {
	return int32(c)
}

func (c LogClientConf) String() string {
	return LogClientConfVals[c.Int32()]
}
func (c LogClientConf) Default() config.Config {
	return LogClientDefaults[c].Default()
}

func WithAddr(addr ...string) config.Config {
	var connAddr address.ConnAddr = map[string]*grpc.ClientConn{}

	if len(addr) == 0 || addr == nil {
		connAddr.Add(":9099")
		return config.WithValue(addrCfg, &connAddr)
	}

	connAddr.Add(addr...)
	return config.WithValue(addrCfg, &connAddr)
}

func WithGRPCOpts(opts ...grpc.DialOption) config.Config {
	if len(opts) == 0 {
		// enforce defaults
		return config.WithValue(optsCfg, defaultDialOptions)
	}

	return config.WithValue(optsCfg, opts)
}

func Insecure() config.Config {
	return config.WithValue(tlsCfg, insecureDialOptions)
}

func WithTLS(caPath string, certKeyPair ...string) config.Config {
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

func StreamRPC() config.Config {
	return config.WithValue(typeCfg, "stream")
}

func UnaryRPC() config.Config {
	return config.WithValue(typeCfg, "unary")
}

func WithLogger(loggers ...log.Logger) config.Config {
	if len(loggers) == 1 {
		return config.WithValue(logCfg, loggers[0])
	}

	if len(loggers) > 1 {
		return config.WithValue(logCfg, log.MultiLogger(loggers...))
	}

	return config.WithValue(logCfg, log.New(log.NilConfig))
}

func WithBackoff(t time.Duration) config.Config {
	b := NewBackoff()

	// default config
	if t == 0 || t == defaultRetryTime {
		return config.WithValue(backoffCfg, b.Time(defaultRetryTime))
	}

	return config.WithValue(backoffCfg, b.Time(t))
}
