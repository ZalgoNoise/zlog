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

	defaultDialOptions = []grpc.DialOption{
		grpc.FailOnNonTempDialError(true),
		grpc.WithReturnConnectionError(),
		grpc.WithDisableRetry(),
	}

	defaultConfig LogClientConfig = &multiconf{
		confs: []LogClientConfig{
			WithAddr(""),
			WithGRPCOpts(),
			Insecure(),
			WithLogger(),
			WithBackoff(0, BackoffExponential()),
		},
	}

	LogClientConfigs = map[int]LogClientConfig{
		0: defaultConfig,
		1: WithBackoff(0, BackoffExponential()),
		2: WithBackoff(time.Second*30, BackoffExponential()),
	}

	DefaultCfg     LogClientConfig = LogClientConfigs[0] // placeholder for an initialized default LogClientConfig
	BackoffFiveMin LogClientConfig = LogClientConfigs[1] // placeholder for a backoff config with 5-minute deadline
	BackoffHalfMin LogClientConfig = LogClientConfigs[2] // placeholder for a backoff config with 30-second deadline

)

// LogClientConfig interface describes the behavior that a LogClientConfig object should have
//
// The single Apply(lb *GRPCLogClientBuilder) method allows for different modules to apply changes to a
// GRPCLogClientBuilder, in a non-blocking way for other features.
//
// Each feature should implement its own structs with their own methods; where they can implement
// Apply(lb *GRPCLogClientBuilder) to set their own configurations to the input GRPCLogClientBuilder
type LogClientConfig interface {
	Apply(ls *gRPCLogClientBuilder)
}

type multiconf struct {
	confs []LogClientConfig
}

// MultiConf function is a wrapper for multiple configs to be bundled (and executed) in one shot.
//
// Similar to io.MultiWriter, it will iterate through all set LogClientConfig and run the same method
// on each of them.
func MultiConf(conf ...LogClientConfig) LogClientConfig {
	if len(conf) == 0 {
		return defaultConfig
	}

	allConf := make([]LogClientConfig, 0, len(conf))
	allConf = append(allConf, conf...)

	return &multiconf{allConf}
}

// Apply method will make a multiconf-type of LogClientConfig iterate through all its objects and
// run the Apply method on the input pointer to a GRPCLogClient
func (m multiconf) Apply(lb *gRPCLogClientBuilder) {
	for _, c := range m.confs {
		c.Apply(lb)
	}
}

// LSAddr struct is a custom LogClientConfig to define addresses to new gRPC Log Client
type LSAddr struct {
	addr address.ConnAddr
}

// LSOpts struct is a custom LogClientConfig to define gRPC Dial Options to new gRPC Log Client
type LSOpts struct {
	opts []grpc.DialOption
}

// LSType struct is a custom LogClientConfig to define the type of a new gRPC Log Client (unary or stream)
type LSType struct {
	isUnary bool
}

// LSLogger struct is a custom LogClientConfig to define the service logger for the new gRPC Log Client
type LSLogger struct {
	logger log.Logger
}

// LSExpBackoff struct is a custom LogClientConfig to define the backoff configuration for the new gRPC Log Client
type LSExpBackoff struct {
	backoff *Backoff
}

// LSTiming struct is a custom LogClientConfig to add (debug) information on time taken to execute RPCs.
type LSTiming struct{}

// Apply method will set this option's address as the input gRPCLogClientBuilder's
func (l LSAddr) Apply(ls *gRPCLogClientBuilder) {
	ls.addr = &l.addr
}

// Apply method will set this option's Dial Options as the input gRPCLogClientBuilder's
func (l LSOpts) Apply(ls *gRPCLogClientBuilder) {
	ls.opts = append(ls.opts, l.opts...)
}

// Apply method will set this option's type as the input gRPCLogClientBuilder's
func (l LSType) Apply(ls *gRPCLogClientBuilder) {
	ls.isUnary = l.isUnary
}

// Apply method will set this option's logger as the input gRPCLogClientBuilder's,
// along with defining its logging interceptors with the same logger.
func (l LSLogger) Apply(ls *gRPCLogClientBuilder) {
	ls.svcLogger = l.logger

	ls.interceptors.unaryItcp["logging"] = UnaryClientLogging(l.logger, false)
	ls.interceptors.streamItcp["logging"] = StreamClientLogging(l.logger, false)
}

// Apply method will set this option's backoff as the input gRPCLogClientBuilder's
func (l LSExpBackoff) Apply(ls *gRPCLogClientBuilder) {
	ls.backoff = l.backoff
}

// Apply method will set this option's Timing interceptors as the input gRPCLogClientBuilder's
// by defining its own service logger as target
func (l LSTiming) Apply(ls *gRPCLogClientBuilder) {
	// if there is a logging interceptor configured, reconfigure it to register time
	if _, ok := ls.interceptors.streamItcp["logging"]; ok && ls.svcLogger != nil {
		ls.interceptors.streamItcp["logging"] = StreamClientLogging(ls.svcLogger, true)
		ls.interceptors.unaryItcp["logging"] = UnaryClientLogging(ls.svcLogger, true)
		return
	}

	// otherwise, if there is no logging interceptor, add a new independent timing interceptor
	ls.interceptors.streamItcp["timing"] = StreamClientTiming(ls.svcLogger)
	ls.interceptors.unaryItcp["timing"] = UnaryClientTiming(ls.svcLogger)
}

// WithAddr function will take in any amount of addresses, and create a connections
// map with them, for the gRPC client to connect to the server
//
// If these addresses are all empty (or if none is provided) defaults are applied (localhost:9099)
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

// StreamRPC function will set this gRPC Log Client type as Stream RPC
func StreamRPC() LogClientConfig {
	return &LSType{
		isUnary: false,
	}

}

// UnaryRPC function will set this gRPC Log Client type as Unary RPC
func UnaryRPC() LogClientConfig {
	return &LSType{
		isUnary: true,
	}
}

// WithLogger function will define this gRPC Log Client's service logger.
// This logger will register the gRPC Client transactions; and not the log
// messages it is handling.
//
// This function's loggers input parameter is variadic -- it supports setting
// any number of loggers. If no input is provided, then it will default to
// setting this service logger as a nil logger (one which doesn't do anything)
//
// This function configures the gRPC client's logger interceptors
func WithLogger(loggers ...log.Logger) LogClientConfig {
	var l log.Logger

	if len(loggers) == 1 {
		l = loggers[0]
	} else if len(loggers) > 1 {
		l = log.MultiLogger(loggers...)
	} else {
		l = log.New(log.NilConfig)
	}

	return &LSLogger{
		logger: l,
	}

}

// WithBackoff function will take in a time.Duration value to set as the
// exponential backoff module's retry deadline, and a BackoffFunc to
// customize the backoff pattern
//
// If deadline is set to 0 and no BackoffFunc is provided, then no backoff
// logic is applied.
//
// Otherwise, defaults are set:
//   - if a BackoffFunc is set but deadline is zero: default retry time is set
//   - if no BackoffFunc is provided, but a deadline is set: Exponential with input deadline.
func WithBackoff(deadline time.Duration, backoffFunc BackoffFunc) LogClientConfig {
	b := NewBackoff()

	if deadline == 0 && backoffFunc == nil {
		b.BackoffFunc(NoBackoff())
		return &LSExpBackoff{
			backoff: b,
		}
	} else if backoffFunc != nil {
		b.Time(defaultRetryTime)
	} else {
		b.Time(deadline)
	}

	if backoffFunc == nil {
		b.BackoffFunc(BackoffExponential())
	}

	return &LSExpBackoff{
		backoff: b,
	}
}

// WithTiming function will set a gRPC Log Client's service logger to measure
// the time taken when executing RPCs. It is only an option, and is directly tied
// to the configured service logger.
//
// Since defaults are enforced, the service logger value is never nil.
//
// This function configures the gRPC client's timer interceptors
func WithTiming() LogClientConfig {
	return &LSTiming{}
}

// WithGRPCOpts will allow passing in any number of gRPC Dial Options, which
// are added to the gRPC Log Client.
//
// Running this function with zero parameters will generate a LogClientConfig with
// the default gRPC Dial Options.
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

// Insecure function will allow creating an insecure gRPC connection (maybe for testing
// purposes) by adding a new option for insecure transport credentials.
func Insecure() LogClientConfig {
	return &LSOpts{
		opts: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	}
}

// WithTLS function allows configuring TLS / mTLS for a gRPC Log Client.
//
// If only one parameter is passed (caPath), it will run its TLS flow. If three
// parameters are set (caPath, certPath, keyPath), it will run its mTLS flow.
//
// The function will try to open the certificates that the user points to in these
// paths, so it is required that they are accessible in terms of permissions. These
// configurations will panic if they fail to execute. This is OK since it should happen
// as soon as the client is executed.
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
		RootCAs:      crtPool,
		MinVersion:   tls.VersionTLS13,
		// MaxVersion:   tls.VersionTLS13,
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
