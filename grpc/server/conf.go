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
		3: WithServiceLogger(log.New(log.WithFormat(log.TextColorLevelFirst))),
		4: WithServiceLogger(log.New(log.WithFormat(log.FormatJSON))),
		5: WithLogger(),
		6: WithLogger(log.New(log.WithFormat(log.TextColorLevelFirst))),
		7: WithLogger(log.New(log.WithFormat(log.FormatJSON))),
	}

	DefaultCfg        LogServerConfig = LogServerConfigs[0] // placeholder for an initialized default LogServerConfig
	ServiceLogDefault LogServerConfig = LogServerConfigs[1] // placeholder for an initialized default logger as service logger
	ServiceLogNil     LogServerConfig = LogServerConfigs[2] // placeholder for an initialized nil-service-logger LogServerConfig
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
	allConf := make([]LogServerConfig, 0, len(conf))
	for _, c := range conf {
		if c == nil {
			continue
		}
		allConf = append(allConf, c)
	}

	if len(allConf) == 0 {
		return defaultConfig
	}

	if len(allConf) == 1 {
		return allConf[0]
	}

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
	logger  log.Logger
	verbose bool
}

// LSOpts struct is a custom LogServerConfig to define gRPC Dial Options to new gRPC Log Server
type LSOpts struct {
	opts []grpc.ServerOption
}

// LSTiming struct is a custom LogServerConfig to add a timing module to exchanged RPCs
type LSTiming struct{}

// Apply method will set this option's address as the input GRPCLogServer's
func (l LSAddr) Apply(ls *gRPCLogServerBuilder) {
	ls.addr = l.addr
}

// Apply method will set this option's logger as the input GRPCLogServer's logger
func (l LSLogger) Apply(ls *gRPCLogServerBuilder) {
	ls.logger = l.logger
}

// Apply method will set this option's logger as the input GRPCLogServer's service logger,
// and its logger interceptors
func (l LSServiceLogger) Apply(ls *gRPCLogServerBuilder) {
	ls.svcLogger = l.logger

	if l.verbose {
		ls.interceptors.streamItcp["logging"] = StreamServerLogging(l.logger, false)
		ls.interceptors.unaryItcp["logging"] = UnaryServerLogging(l.logger, false)
	}
}

// Apply method will set this option's Dial Options as the input GRPCLogServer's
func (l LSOpts) Apply(ls *gRPCLogServerBuilder) {
	ls.opts = append(ls.opts, l.opts...)
}

// Apply method will set the input GRPCLogServer's service logger to time the exchanged
// messages (if existing), otherwise to configure a new module for timing
func (l LSTiming) Apply(ls *gRPCLogServerBuilder) {
	// if there is a logging interceptor configured, reconfigure it to register time
	if _, ok := ls.interceptors.streamItcp["logging"]; ok && ls.svcLogger != nil {
		ls.interceptors.streamItcp["logging"] = StreamServerLogging(ls.svcLogger, true)
		ls.interceptors.unaryItcp["logging"] = UnaryServerLogging(ls.svcLogger, true)
		return
	}

	// otherwise, if there is no logging interceptor, add a new independent timing interceptor
	ls.interceptors.streamItcp["timing"] = StreamServerTiming(ls.svcLogger)
	ls.interceptors.unaryItcp["timing"] = UnaryServerTiming(ls.svcLogger)
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
	var l log.Logger

	if len(loggers) == 1 {
		l = loggers[0]
	} else if len(loggers) > 1 {
		l = log.MultiLogger(loggers...)
	} else {
		l = log.New()
	}

	return &LSLogger{
		logger: l,
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
		logger: l,
	}
}

// WithServiceLoggerV function will define this gRPC Log Server's service logger,
// in verbose mode -- capturing interactions for each RPC. This differs from a log
// level filter as it will add a logging interceptor as a module.
//
// This logger will register the gRPC Server's transactions, and not the client's
// incoming log messages.
//
// This function's loggers input parameter is variadic -- it supports setting
// any number of loggers. If no input is provided, then it will default to
// setting this service logger as a nil logger (one which doesn't do anything)
func WithServiceLoggerV(loggers ...log.Logger) LogServerConfig {
	var l log.Logger

	if len(loggers) == 1 {
		l = loggers[0]
	} else if len(loggers) > 1 {
		l = log.MultiLogger(loggers...)
	} else {
		l = log.New(log.NilConfig)
	}

	return &LSServiceLogger{
		logger:  l,
		verbose: true,
	}
}

// WithTiming function will set a gRPC Log Server's service logger to measure
// the time taken when executing RPCs. It is only an option, and is directly tied
// to the configured service logger.
//
// Since defaults are enforced, the service logger value is never nil.
//
// This function configures the gRPC server's timer interceptors
func WithTiming() LogServerConfig {
	return &LSTiming{}
}

// WithGRPCOpts will allow passing in any number of gRPC Server Options, which
// are added to the gRPC Log Server.
//
// Running this function with zero parameters will generate a LogServerConfig with
// the default gRPC Server Options.
func WithGRPCOpts(opts ...grpc.ServerOption) LogServerConfig {
	var o []grpc.ServerOption

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		o = append(o, opt)
	}

	return &LSOpts{
		opts: o,
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
	} else {
		cred, err = loadCredsMutual(caPath[0], certPath, keyPath)
	}

	// panic since the gRPC server shouldn't start
	// if TLS is requested but invalid / errored
	if err != nil {
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
