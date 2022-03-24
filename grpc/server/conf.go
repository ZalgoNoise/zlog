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

	defaultConfig LogServerConfig = &multiconf{
		confs: []LogServerConfig{
			WithAddr(""),
			WithLogger(),
			WithServiceLogger(),
		},
	}

	LogServerConfigs = map[int]LogServerConfig{
		0: defaultConfig,
		1: WithServiceLogger(log.New()),
		2: WithServiceLogger(log.New(log.NilConfig)),
		3: WithServiceLogger(log.New(log.ColorTextLevelFirst)),
		4: WithServiceLogger(log.New(log.JSONFormat)),
		5: WithLogger(),
		6: WithLogger(log.New(log.ColorTextLevelFirst)),
		7: WithLogger(log.New(log.JSONFormat)),
	}

	DefaultCfg        LogServerConfig = LogServerConfigs[0] // placeholder for an intialized default LogServerConfig
	ServiceLogDefault LogServerConfig = LogServerConfigs[1] // placeholder for an initialzed default logger as service logger
	ServiceLogNil     LogServerConfig = LogServerConfigs[2] // placeholder for an initialzed nil-service-logger LogServerConfig
	ServiceLogColor   LogServerConfig = LogServerConfigs[3] // placeholder for an initialized colored, level-first, service logger
	ServiceLogJSON    LogServerConfig = LogServerConfigs[4] // placeholder for an initialized JSON service logger
	LoggerDefault     LogServerConfig = LogServerConfigs[5] // placeholder for an initialized default logger
	LoggerColor       LogServerConfig = LogServerConfigs[6] // placeholder for an initialized colored, level-first logger
	LoggerJSON        LogServerConfig = LogServerConfigs[7] // placeholder for an initialized JSON logger
)

type LogServerConfig interface {
	Apply(ls *GRPCLogServer)
}

type multiconf struct {
	confs []LogServerConfig
}

// MultiConf function is a wrapper for multiple configs to be bundled (and executed) in one shot.
//
// Similar to io.MultiWriter, it will iterate through all set LogServerConfig and run the same method
// on each of them.
func MultiConf(conf ...LogServerConfig) LogServerConfig {
	if len(conf) == 0 {
		return defaultConfig
	}

	allConf := make([]LogServerConfig, 0, len(conf))
	allConf = append(allConf, conf...)

	return &multiconf{allConf}
}

// Apply method will make a multiconf-type of LogServerConfig iterate through all its objects and
// run the Apply method on the input pointer to a GRPCLogServer
func (m multiconf) Apply(lb *GRPCLogServer) {
	for _, c := range m.confs {
		c.Apply(lb)
	}
}

type LSAddr struct {
	addr string
}

type LSLogger struct {
	logger log.Logger
}

type LSServiceLogger struct {
	logger log.Logger
}

type LSOpts struct {
	opts []grpc.ServerOption
}

func (l LSAddr) Apply(ls *GRPCLogServer) {
	ls.Addr = l.addr
}

func (l LSLogger) Apply(ls *GRPCLogServer) {
	ls.Logger = l.logger
}

func (l LSServiceLogger) Apply(ls *GRPCLogServer) {
	ls.SvcLogger = l.logger
}

func (l LSOpts) Apply(ls *GRPCLogServer) {
	ls.opts = append(ls.opts, l.opts...)
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
