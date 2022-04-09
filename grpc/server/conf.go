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
		3: WithServiceLogger(log.New(log.TextColorLevelFirst)),
		4: WithServiceLogger(log.New(log.FormatJSON)),
		5: WithLogger(),
		6: WithLogger(log.New(log.TextColorLevelFirst)),
		7: WithLogger(log.New(log.FormatJSON)),
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

// LogServerConfig interface describes the behavior that a LogServerConfig object should have
//
// The single Apply(lb *GRPCLogServer) method allows for different modules to apply changes to a
// GRPCLogServer, in a non-blocking way for other features.
//
// Each feature should implement its own structs with their own methods; where they can implement
// Apply(lb *GRPCLogServer) to set their own configurations to the input GRPCLogServer
type LogServerConfig interface {
	Apply(ls *gRPCLogServerBuilder)
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
func (m multiconf) Apply(lb *gRPCLogServerBuilder) {
	for _, c := range m.confs {
		c.Apply(lb)
	}
}

// LSAddr struct is a custom LogServerConfig to define addresses to new gRPC Log Server
type LSAddr struct {
	addr string
}

// LSLogger struct is a custom LogServerConfig to define the (served) logger for the new gRPC Log Server
type LSLogger struct {
	logger log.Logger
}

// LSServiceLogger struct is a custom LogServerConfig to define the service logger for the new gRPC Log Server
type LSServiceLogger struct {
	logger     log.Logger
	streamItcp grpc.StreamServerInterceptor
	unaryItcp  grpc.UnaryServerInterceptor
}

// LSOpts struct is a custom LogServerConfig to define gRPC Dial Options to new gRPC Log Server
type LSOpts struct {
	opts []grpc.ServerOption
}

// Apply method will set this option's address as the input GRPCLogServer's
func (l LSAddr) Apply(ls *gRPCLogServerBuilder) {
	ls.addr = l.addr
}

// Apply method will set this option's logger as the input GRPCLogServer's logger
func (l LSLogger) Apply(ls *gRPCLogServerBuilder) {
	ls.logger = l.logger
}

// Apply method will set this option's logger as the input GRPCLogServer's service logger
func (l LSServiceLogger) Apply(ls *gRPCLogServerBuilder) {
	ls.svcLogger = l.logger
	ls.interceptors.streamItcp["logging"] = l.streamItcp
	ls.interceptors.unaryItcp["logging"] = l.unaryItcp
}

// Apply method will set this option's Dial Options as the input GRPCLogServer's
func (l LSOpts) Apply(ls *gRPCLogServerBuilder) {
	ls.opts = append(ls.opts, l.opts...)
}

// WithAddr function will take one address for the gRPC Log Server to listen to.
//
// If this address is empty, defaults are applied (localhost:9099)
func WithAddr(addr string) LogServerConfig {
	// enforce defaults
	if addr == "" || addr == ":" {
		addr = ":9099"
	}

	return &LSAddr{
		addr: addr,
	}
}

// WithLogger function will define this gRPC Log Server's logger.
//
// This logger will register the gRPC Client incoming log messages, from either
// Unary or Stream RPCs.
//
// This function's loggers input parameter is variadic -- it supports setting
// any number of loggers. If no input is provided, then it will default to
// setting this logger as a default logger (with its output set to os.Stderr)
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

// WithServiceLogger function will define this gRPC Log Server's service logger.
//
// This logger will register the gRPC Server's transactions, and not the client's
// incoming log messages.
//
// This function's loggers input parameter is variadic -- it supports setting
// any number of loggers. If no input is provided, then it will default to
// setting this service logger as a nil logger (one which doesn't do anything)
func WithServiceLogger(loggers ...log.Logger) LogServerConfig {
	var l log.Logger

	if len(loggers) == 1 {
		l = loggers[0]
	} else if len(loggers) > 1 {
		l = log.MultiLogger(loggers...)
	} else {
		l = log.New(log.NilConfig)
	}

	return &LSServiceLogger{
		logger:     l,
		streamItcp: StreamServerLogging(l),
		unaryItcp:  UnaryServerLogging(l),
	}

}

// WithGRPCOpts will allow passing in any number of gRPC Server Options, which
// are added to the gRPC Log Server.
//
// Running this function with zero parameters will generate a LogServerConfig with
// the default gRPC Server Options.
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

// WithTLS function allows configuring TLS / mTLS for a gRPC Log Server.
//
// If only two parameters are passed (certPath, keyPath), it will run its TLS flow. If three
// parameters are set (certPath, keyPath, caPath), it will run its mTLS flow.
//
// The function will try to open the certificates that the user points to in these
// paths, so it is required that they are accessible in terms of permissions. These
// configurations will panic if they fail to execute. This is OK since it should happen
// as soon as the server is executed.
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

	if !crtPool.AppendCertsFromPEM(ca) {
		return nil, ErrCACertAddFailed
	}

	c, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{c},
		ClientCAs:    crtPool,
		RootCAs:      crtPool,
		MinVersion:   tls.VersionTLS13,
		// MaxVersion:   tls.VersionTLS13,
	}

	return credentials.NewTLS(config), nil
}
