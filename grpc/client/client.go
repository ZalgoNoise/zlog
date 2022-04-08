package client

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"regexp"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/zalgonoise/zlog/grpc/address"
	"github.com/zalgonoise/zlog/log"

	pb "github.com/zalgonoise/zlog/proto/message"
	"google.golang.org/grpc"
)

const (
	// default Unary RPC timeout in seconds
	timeoutSeconds = 30

	// default Stream RPC timeout in seconds
	streamTimeoutSeconds = 3600

	// default Unary RPC timeout
	defaultTimeout = time.Second * timeoutSeconds

	// default Stream RPC timeout
	defaultStreamTimeout = time.Second * streamTimeoutSeconds
)

var (
	ErrDeadlineRegexp    = regexp.MustCompile(`rpc error: code = DeadlineExceeded desc = context deadline exceeded`)
	ErrEOFRegexp         = regexp.MustCompile(`rpc error: code = Unavailable desc = error reading from server: EOF`)
	ErrConnRefusedRegexp = regexp.MustCompile(`rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp .*: connect: connection refused"`)

	ErrNoAddr      error = errors.New("cannot connect to gRPC server since no addresses were provided")
	ErrNoConns     error = errors.New("could not establish any successful connection with the provided address(es)")
	ErrBadResponse error = errors.New("failed to write log message in remote gRPC server")
	ErrBadWriter   error = errors.New("invalid writer -- must be of type client.ConnAddr")
)

var backoff *ExpBackoff

// GRPCLogger interface will be of types Logger and ChanneledLogger, to allow
// it being used interchangeably as either, with the same methods and API
//
// In nature, it is a gRPC client for a gRPC Log Server. As such, certain
// configurations are inaccessible as they are set on the server (e.g., IsSkipExit())
//
// Also worth mentioning that MultiLogger() will support Loggers of this type, considering
// that its SetOuts() / AddOuts() methods are expecting an io.Writer of type ConnAddr
type GRPCLogger interface {
	log.Logger
	log.ChanneledLogger
}

// GRPCLogClient struct will define the elements required to build and work with
// a gRPC Log Client.
//
// It does not have exactly the same elements as a joined Logger+ChannelLogger,
// considering that it is actually sending log messages to a (remote) server that is
// configuring its own Logger.
type GRPCLogClient struct {
	addr  *address.ConnAddr
	opts  []grpc.DialOption
	msgCh chan *log.LogMessage
	done  chan struct{}

	svcLogger log.Logger

	prefix string
	sub    string
	meta   map[string]interface{}
}

// GRPCLogClientBuilder struct is an entrypoint object to create a GRPCLogClient
//
// This struct will take in multiple configurations, creating a GRPCLogClient
// during the process
type gRPCLogClientBuilder struct {
	addr         *address.ConnAddr
	opts         []grpc.DialOption
	interceptors clientInterceptors
	isUnary      bool
	expBackoff   *ExpBackoff
	svcLogger    log.Logger
}

func (b *gRPCLogClientBuilder) build() *GRPCLogClient {
	// auto merge stream / unary interceptors as []grpc.DialOption
	var opts []grpc.DialOption

	if len(b.interceptors.unaryItcp) > 0 {
		var interceptors []grpc.UnaryClientInterceptor
		for _, i := range b.interceptors.unaryItcp {
			interceptors = append(interceptors, i)
		}

		uItcp := grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(interceptors...))
		opts = append(b.opts, uItcp)
	}

	if len(b.interceptors.streamItcp) > 0 {
		var interceptors []grpc.StreamClientInterceptor
		for _, i := range b.interceptors.streamItcp {

			interceptors = append(interceptors, i)
		}

		sItcp := grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(interceptors...))
		opts = append(opts, sItcp)
	}

	// define backoff
	// TODO: make an interceptor for this
	backoff = b.expBackoff

	return &GRPCLogClient{
		addr:      b.addr,
		opts:      append(b.opts, opts...),
		msgCh:     make(chan *log.LogMessage),
		done:      make(chan struct{}),
		svcLogger: b.svcLogger,
	}
}

// clientInterceptors struct is a placeholder for different interceptors to be added
// to the GRPCLogServer
type clientInterceptors struct {
	streamItcp map[string]grpc.StreamClientInterceptor
	unaryItcp  map[string]grpc.UnaryClientInterceptor
}

func newGRPCLogClient(confs ...LogClientConfig) *gRPCLogClientBuilder {
	builder := &gRPCLogClientBuilder{
		interceptors: clientInterceptors{
			streamItcp: make(map[string]grpc.StreamClientInterceptor),
			unaryItcp:  make(map[string]grpc.UnaryClientInterceptor),
		},
	}

	// enforce defaults
	defaultConfig.Apply(builder)

	// apply input configs
	for _, config := range confs {
		config.Apply(builder)
	}

	return builder
}

// New function will serve as a GRPCLogger factory -- taking in different LogClientConfig
// options and creating either a Unary RPC or a Stream RPC GRPCLogger, along with other
// user-defined (or default) parameters
//
// This function returns not only the logger but an error channel, which should be monitored
// (preferrably in a goroutine) to ensure that the gRPC client is running without issues
func New(opts ...LogClientConfig) (GRPCLogger, chan error) {
	builder := newGRPCLogClient(opts...)

	client := builder.build()

	// check input type -- create an appropriate GRPCLogger
	if builder.isUnary {
		client.svcLogger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("init").Message("setting up Unary gRPC client").Build())
		return newUnaryLogger(client)
	} else {
		client.svcLogger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("init").Message("setting up Stream gRPC client").Build())
		return newStreamLogger(client)
	}
}

func newUnaryLogger(c *GRPCLogClient) (GRPCLogger, chan error) {
	errCh := make(chan error)

	// register log() function and error channel in the backoff module
	backoff.RegisterLog(c.log, errCh).WithDone(&c.done)

	// launch log listener in a goroutine
	go c.listen(errCh)

	return c, errCh
}

func newStreamLogger(c *GRPCLogClient) (GRPCLogger, chan error) {
	errCh := make(chan error)

	// register stream() function and error channel in the backoff module
	backoff.RegisterStream(c.stream, errCh).WithDone(&c.done)

	// launch log listener in a goroutine
	go c.stream(errCh)

	return c, errCh
}

// connect method will iterate the connections map and dial each remote
//
// It stores the connection in the map or it removes the entry in case it is unhealthy
func (c GRPCLogClient) connect() error {

	// exit if no addresses are set
	if c.addr.Len() == 0 {
		c.svcLogger.Log(log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("conn").Metadata(log.Field{"error": ErrNoAddr.Error()}).Message("no addresses provided").Build())
		return ErrNoAddr
	}

	var liveConns int = 0

	for idx, remote := range c.addr.Keys() {
		c.svcLogger.Log(
			log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("conn").
				Metadata(log.Field{
					"index": idx,
					"addr":  remote,
				}).
				Message("connecting to remote").Build(),
		)

		var conn *grpc.ClientConn
		conn, err := grpc.Dial(remote, c.opts...)

		// handle dial errors
		if err != nil {
			// retry with backoff
			c.svcLogger.Log(
				log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("conn").Metadata(log.Field{
					"error":      err,
					"iterations": backoff.Counter(),
					"maxWait":    backoff.Max(),
					"curWait":    backoff.Current(),
				}).Message("retrying connection").Build(),
			)

			// backoff locked -- skip retry until unlocked
			if backoff.IsLocked() {
				return ErrBackoffLocked
			} else {

				// backoff unlocked -- increment timer and wait
				// the Wait() method returns a registered func() to execute
				// and an error in case the backoff reaches its deadline
				call, err := backoff.Increment().Wait()

				// handle backoff deadline errors
				if err != nil {
					c.svcLogger.Log(
						log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("conn").
							Metadata(log.Field{
								"error": err.Error(),
							}).
							Message("removing address after failed dial attempt").Build(),
					)

					// address is removed from connections map
					c.addr.Unset(remote)
					continue
				} else {

					// execute registered call
					go call()
					return ErrFailedConn
				}
			}

		}

		// once the connection is established, it's mapped to its (string) address
		c.addr.Set(remote, conn)
		liveConns++

		c.svcLogger.Log(
			log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("conn").
				Metadata(log.Field{
					"index": idx,
					"addr":  remote,
				}).
				Message("dialed the address successfully").Build(),
		)
	}

	// return ErrNoConns if the counter for live connections hasn't increased
	if liveConns == 0 {
		c.svcLogger.Log(log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("conn").Metadata(log.Field{"error": ErrNoConns.Error()}).Message("all connections failed").Build())
		return ErrNoConns
	}

	return nil
}

// listen method is middleware to allow the backoff module to register an
// action call (in this case log()), which will be retried in case of failure
func (c GRPCLogClient) listen(errCh chan error) {
	for {
		msg := <-c.msgCh
		backoff.AddMessage(msg)
		go c.log(msg, errCh)
	}
}

// log method is a go-routine function which will send all configured connections
// a Log() request.
func (c GRPCLogClient) log(msg *log.LogMessage, errCh chan error) {

	// establish connections
	err := c.connect()

	// handle connection errors
	if err != nil {

		// check if errors are failed connection or backoff locked errors;
		// return so the action is cancelled
		if errors.Is(err, ErrFailedConn) || errors.Is(err, ErrBackoffLocked) {
			return
		}

		// any other errors will be sent to the error channel and logged locally;
		// then cancelling this action
		errCh <- err

		c.svcLogger.Log(
			log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("log").Metadata(log.Field{
				"error": err.Error(),
			}).Message("failed to connect").Build(),
		)

		return
	}

	// there are live connections; log input message on each of the remote gRPC Log Servers
	for remote, conn := range c.addr.Map() {
		defer conn.Close()
		client := pb.NewLogServiceClient(conn)

		c.svcLogger.Log(
			log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("log").Metadata(log.Field{
				"remote": remote,
			}).Message("setting up log service with connection").Build(),
		)

		// generate a new context with a timeout
		bgCtx := context.Background()
		ctx, cancel := context.WithTimeout(bgCtx, defaultTimeout)

		c.svcLogger.Log(
			log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("log").Metadata(log.Field{
				"timeout": timeoutSeconds,
				"remote":  remote,
			}).Message("received a new log message to register").Build(),
		)

		// send LogMessage to remote gRPC Log Server
		// response is parsed by the interceptor; only the error is important
		_, err := client.Log(ctx, msg.Proto())

		// if the server returns any error, it's sent to the error channel, context cancelled,
		// and then return
		if err != nil {
			errCh <- err
			cancel()
			return
		}

		// message sent; response retrieved; context is cancelled.
		cancel()

	}
}

// stream method is a go-routine function which will establish a connection with
// all configured addresses to create a Stream().
func (c GRPCLogClient) stream(errCh chan error) {

	localErr := make(chan error)

	// establish connections
	err := c.connect()

	// handle connection errors
	if err != nil {

		// check if errors are failed connection or backoff locked errors;
		// return so the action is cancelled
		if errors.Is(err, ErrFailedConn) || errors.Is(err, ErrBackoffLocked) {
			return
		}

		// any other errors will be sent to the error channel and logged locally;
		// then cancelling this action
		errCh <- err

		c.svcLogger.Log(
			log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("stream").Metadata(log.Field{
				"error": err.Error(),
			}).Message("failed to connect").Build(),
		)

		return
	}

	// there are live connections; setup a stream with each of the remote gRPC Log Servers
	for remote, conn := range c.addr.Map() {
		logClient := pb.NewLogServiceClient(conn)

		c.svcLogger.Log(
			log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("stream").Metadata(log.Field{
				"remote": remote,
			}).Message("setting up log service with connection").Build(),
		)

		// generate a new context with a timeout
		bgCtx := context.Background()
		ctx, cancel := context.WithTimeout(bgCtx, defaultStreamTimeout)

		c.svcLogger.Log(
			log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("stream").Metadata(log.Field{
				"timeout": streamTimeoutSeconds,
				"remote":  remote,
			}).Message("setting request ID for long-lived connection").Build(),
		)

		// setup a stream with the remote gRPC Log Server
		stream, err := logClient.LogStream(ctx)

		// check for errors when creating a stream
		if err != nil {

			// if it's connection refused or EOF, kick-off the backoff routine
			if ErrConnRefusedRegexp.MatchString(err.Error()) || ErrEOFRegexp.MatchString(err.Error()) {
				c.streamBackoff(errCh, cancel)
			}

			// otherwise, error is sent to the error channel, conn closed, context cancelled,
			// error logged and then return
			errCh <- err
			conn.Close()
			cancel()

			c.svcLogger.Log(
				log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("stream").Metadata(log.Field{
					"remote": remote,
					"error":  err.Error(),
				}).Message("failed to setup stream connection with gRPC server").Build(),
			)

			return
		}

		// stream is live; kick off a goroutine with a comm channel and two separate,
		// concurrent functions, the latter being a blocking one.
		//
		// - the first function (handleStreamService method) will listen to server responses
		// - the second one (handleStreamMessages) will listen to incoming LogMessages and
		// send them to the remote gRPC Log Server; this last function will also listen
		// to the internal comms channel and to errors (e.g., to kick-off backoff)
		go func() {
			go c.handleStreamService(
				stream,
				localErr,
				c.done,
			)

			c.handleStreamMessages(
				stream,
				localErr,
				errCh,
				cancel,
			)
		}()
	}
}

// streamBackoff method is the Stream gRPC Log Client's standard backoff flow, which
// is used when setting up a stream and when receiving an error from the gRPC Log Server
func (c GRPCLogClient) streamBackoff(
	errCh chan error,
	cancel context.CancelFunc,
) {
	// increment timer and wait
	// the Wait() method returns a registered func() to execute
	// and an error in case the backoff reaches its deadline
	call, err := backoff.Increment().Wait()

	// handle backoff deadline errors by closing the stream
	if err != nil {
		c.svcLogger.Log(
			log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("stream").Metadata(log.Field{
				"error":      err.Error(),
				"numRetries": backoff.Counter(),
			}).Message("closing stream after too many failed attempts to reconnect").Build(),
		)

		c.done <- struct{}{}
		errCh <- ErrFailedRetry
		cancel()
		return
	}

	// otherwise the stream will be recreated
	c.svcLogger.Log(
		log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("stream").Metadata(log.Field{
			"error":      err,
			"iterations": backoff.Counter(),
			"maxWait":    backoff.Max(),
			"curWait":    backoff.Current(),
		}).Message("retrying connection").Build(),
	)

	go call()
	return
}

// handleStreamService method is in place to break-down stream()'s functionality
// into a smaller function.
//
// The method will loop forever acting as a listener to the gRPC stream, and handling
// incoming server responses (to messages posted to it). If the response is an error
// or not OK, this is sinked to the local error channel to be processed. Otherwise,
// the response is registered.
//
// This method is ran in parallel with handleStreamMessages()
func (c GRPCLogClient) handleStreamService(
	stream pb.LogService_LogStreamClient,
	localErr chan error,
	done chan struct{},
) {
	for {
		// capture each incoming message (server response to Log entries)
		in, err := stream.Recv()

		if err != nil {

			// send received error to local error and register the event
			// don't break off the loop; keep listening for messages
			localErr <- err
			continue
		}

		// there are no errors in the response; check the response's OK value
		// if not OK, register this as a local bad response error and continue
		if !in.GetOk() {
			var err error
			if in.GetErr() != "" {
				err = errors.New(in.GetErr())
			} else {
				err = ErrBadResponse
			}

			localErr <- err
			continue
		}
	}
}

// handleStreamMessages method is in place to break-down stream()'s functionality
// into a smaller function.
//
// It will manage the exchanged messages in a gRPC Log Client, from its channels:
// - incoming log messages which should be sent to the gRPC Log Server
// - done requests (to gracefully exit)
// - error messages (sinked into the local error channel)
//
// This method is ran in parallel with handleStreamService()
func (c GRPCLogClient) handleStreamMessages(
	stream pb.LogService_LogStreamClient,
	localErr chan error,
	errCh chan error,
	cancel context.CancelFunc,
) {
	for {
		select {

		// LogMessage is received in the message channel -- send this message to the stream
		case out := <-c.msgCh:

			// send the protofied message and check for errors (sinked to local error channel)
			err := stream.Send(out.Proto())

			if err != nil {
				localErr <- err
			}

		// done is received -- gracefully exit by cancelling context and closing the connection
		case <-c.done:
			cancel()

			c.svcLogger.Log(
				log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("stream").Message("received done signal").Build(),
			)

			return

		// error is received -- parse error in case it is potentially expected; handle it
		// accordingly. In case the error is unexpected, it is sent to the (outbound) error
		// channel and signal done is sent.
		case err := <-localErr:

			// Stream Deadline Exceeded -- reconnect to gRPC Log Server
			if ErrDeadlineRegexp.MatchString(err.Error()) {

				c.svcLogger.Log(
					log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("stream").Metadata(log.Field{
						"error": err.Error(),
					}).Message("stream timed-out -- starting a new connection").Build(),
				)

				go c.stream(errCh)

				// Connection Refused or EOF error -- trigger backoff routine
			} else if ErrEOFRegexp.MatchString(err.Error()) || ErrConnRefusedRegexp.MatchString(err.Error()) {

				c.streamBackoff(errCh, cancel)

				// Bad Response -- send to error channel, continue
			} else if errors.Is(err, ErrBadResponse) {

				errCh <- err
				continue

				// default -- send to error channel, close the client
			} else {
				errCh <- err
				defer c.Close()

				c.svcLogger.Log(
					log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("stream").Metadata(log.Field{
						"error": err.Error(),
					}).Message("critical error -- closing stream").Build(),
				)

			}
			return
		}

	}
}

// Close method is the implementation of the ChanneledLogger's Close().
//
// For a gRPC Log Client, it is important to iterate through all (alive)
// connections in the ConnAddr map, and close them. After doing so, it
// sends the done signal to its channel, which causes all open streams to
// cancel their context and exit gracefully
func (c GRPCLogClient) Close() {
	for _, conn := range c.addr.Map() {
		conn.Close()
	}
	c.done <- struct{}{}
}

// Channels method is the implementation of the ChanneledLogger's Channels().
//
// It returns two channels (message and done channels) to allow control over the
// gRPC Log Client over a background / separate goroutine. Considering that
// creating a gRPC Log Client returns an error channel, this method will give
// the developer the three needed channels to work with the logger asynchronously
func (c GRPCLogClient) Channels() (logCh chan *log.LogMessage, done chan struct{}) {
	return c.msgCh, c.done
}

// Output method implements the Logger's Output().
//
// This method will simply push the incoming Log Message to the message channel,
// which is sent to a gRPC Log Server, either via a Unary or Stream RPC
func (c GRPCLogClient) Output(m *log.LogMessage) (n int, err error) {
	c.msgCh <- m
	return 1, nil
}

// SetOuts method implements the Logger's SetOuts().
//
// For compatibility with the Logger interface, this method must take in io.Writers.
// However, this is not how the gRPC Log Client will work to register messages.
//
// As such, the ConnAddr type will implement the Write() method to be compatible with
// this implementation -- however the type assertion (and check of the same) is required
// to ensure that only "clean" io.Writers are passed on. If so -- these will be added
// to the connection/address map. Otherwise a Bad Writer error is returned.
//
// After doing so, the method will run the client's connect() method to ensure these
// are healthy. If there are errors in this check, the retuning value will be nil instead
// of the same logger.
//
// SetOuts() will replace all the existing connections and addresses by resetting the
// map beforehand.
func (c GRPCLogClient) SetOuts(outs ...io.Writer) log.Logger {
	// reset connections map
	c.addr.Reset()

	for _, remote := range outs {
		// ensure the input writer is not nil
		if remote == nil {
			continue
		}

		// ensure the input writer is of type *address.ConnAddr
		// if not, skip this writer and register this event
		if r, ok := remote.(*address.ConnAddr); !ok {

			c.svcLogger.Log(
				log.NewMessage().Level(log.LLWarn).
					Prefix("gRPC").Sub("SetOuts()").
					Metadata(log.Field{"error": ErrBadWriter.Error()}).
					Message("invalid writer warning").Build(),
			)

			// writer is valid -- add it to connections map; register this event
		} else {
			c.addr.Add(r.Keys()...)

			c.svcLogger.Log(
				log.NewMessage().Level(log.LLDebug).
					Prefix("gRPC").Sub("SetOuts()").
					Metadata(log.Field{"addrs": r.Keys()}).
					Message("added address to connection address map").Build(),
			)
		}
	}

	// test connectivity to remotes -- if this returns an error, the function
	// will return nil instead of the GRPCLogClient
	err := c.connect()
	if err != nil {
		return nil
	}

	return c
}

// AddOuts method implements the Logger's AddOuts().
//
// For compatibility with the Logger interface, this method must take in io.Writers.
// However, this is not how the gRPC Log Client will work to register messages.
//
// As such, the ConnAddr type will implement the Write() method to be compatible with
// this implementation -- however the type assertion (and check of the same) is required
// to ensure that only "clean" io.Writers are passed on. If so -- these will be added
// to the connection/address map. Otherwise a Bad Writer error is returned.
//
// After doing so, the method will run the client's connect() method to ensure these
// are healthy. If there are errors in this check, the retuning value will be nil instead
// of the same logger.
//
// AddOuts() will add the new input io.Writers to the existing connections and addresses
func (c GRPCLogClient) AddOuts(outs ...io.Writer) log.Logger {
	for _, remote := range outs {
		// ensure the input writer is not nil
		if remote == nil {
			continue
		}

		// ensure the input writer is of type *address.ConnAddr
		// if not, skip this writer and register this event
		if r, ok := remote.(*address.ConnAddr); !ok {

			c.svcLogger.Log(
				log.NewMessage().Level(log.LLWarn).
					Prefix("gRPC").Sub("AddOuts()").
					Metadata(log.Field{"error": ErrBadWriter.Error()}).
					Message("invalid writer warning").Build(),
			)

			// writer is valid -- add it to connections map; register this event
		} else {
			c.addr.Add(r.Keys()...)

			c.svcLogger.Log(
				log.NewMessage().Level(log.LLDebug).
					Prefix("gRPC").Sub("AddOuts()").
					Metadata(log.Field{"addrs": r.Keys()}).
					Message("added address to connection address map").Build(),
			)
		}
	}

	// test connectivity to remotes -- if this returns an error, the function
	// will return nil instead of the GRPCLogClient
	err := c.connect()
	if err != nil {
		return nil
	}

	return c
}

// Write method complies with the Logger's io.Writer implementation.
//
// The Write method will allow adding this Logger as a io.Writer to other objects.
// Although this is slightly meta (the client is used as a writer to write to the server
// who will write the log message -- plus the internal logging events for these loggers)
// it is still possible, if you really want to use it that way
//
// Added bonus is support for gob-encoded messages, which is also natively supported in
// the logger's Write implementation
func (c GRPCLogClient) Write(p []byte) (n int, err error) {
	// check if it's gob-encoded
	m := &log.LogMessage{}

	buf := bytes.NewBuffer(p)
	dec := gob.NewDecoder(buf)

	err = dec.Decode(m)

	if err != nil {
		return c.Output(log.NewMessage().
			Level(log.LLInfo).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(string(p)).
			Metadata(c.meta).
			Build(),
		)

	}
	return c.Output(m)
}

// Prefix method implements the Logger interface.
//
// It will set a Logger-scoped (as opposed to message-scoped) prefix string to the logger
//
// Logger-scoped prefix strings can be set in order to avoid calling the `MessageBuilder.Prefix()` method
// repeatedly, and instead doing so via the logger at the beginning of a service or function
//
// A logger-scoped prefix is not cleared with new Log messages, but `MessageBuilder.Prefix()` calls will
// replace it.
func (c GRPCLogClient) Prefix(prefix string) log.Logger {
	if prefix == "" {
		c.prefix = "log"
		return c
	}
	c.prefix = prefix
	return c
}

// Sub method implements the Logger interface.
//
// Logger-scoped sub-prefix strings can be set in order to avoid calling the `MessageBuilder.Sub()` method
// repeatedly, and instead doing so via the logger at the beginning of a service or function
//
// A logger-scoped sub-prefix is not cleared with new Log messages, but `MessageBuilder.Sub()` calls will
// replace it.
func (c GRPCLogClient) Sub(sub string) log.Logger {
	c.sub = sub
	return c
}

// Fields method implements the Logger interface.
//
// Fields method will set Logger-scoped (as opposed to message-scoped) metadata fields to the logger
//
// Logger-scoped metadata can be set in order to avoid calling the `MessageBuilder.Metadata()` method
// repeatedly, and instead doing so via the logger at the beginning of a service or function.
//
// Logger-scoped metadata fields are not cleared with new log messages, but only added to the existing
// metadata map. These can be cleared with a call to `Logger.Fields(nil)`
func (c GRPCLogClient) Fields(fields map[string]interface{}) log.Logger {
	c.meta = fields
	return c
}

// IsSkipExit method implements the Printer interface.
//
// IsSkipExit method returns a boolean on whether the gRPC Log Client's serive logger is
// set to skip os.Exit(1) or panic() calls.
//
// It is used in functions which call these, to first evaluate if those calls should be
// executed or skipped.
//
// Considering this method is unused by the gRPC Log Client, it's merely here to
// implement the interface
func (c GRPCLogClient) IsSkipExit() bool {
	return c.svcLogger.IsSkipExit()
}

// Log method implements the Printer interface.
//
// It will take in a pointer to one or more LogMessages, and write it to the Logger's
// io.Writer without returning an error message.
//
// While the resulting error message of running `GRPCLogClient.Output()` is simply ignored, this is done
// as a blind-write for this Logger. Since these methods are simply sinking LogMessages to a channel,
// this operation is considered safe (the errors will be handled at a gRPC Log Client level, not as a Logger)
func (c GRPCLogClient) Log(m ...*log.LogMessage) {
	for _, msg := range m {
		c.Output(msg)
	}
}

// Print method implements the Printer interface.
//
// It is similar to fmt.Print; and will print a message using an fmt.Sprint(v...) pattern
//
// It applies LogLevel Info
func (c GRPCLogClient) Print(v ...interface{}) {
	c.Log(
		log.NewMessage().
			Level(log.LLInfo).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprint(v...)).
			Metadata(c.meta).
			Build(),
	)

}

// Println method implements the Printer interface.
//
// It is similar to fmt.Println; and will print a message using an fmt.Sprintln(v...) pattern
//
// It applies LogLevel Info
func (c GRPCLogClient) Println(v ...interface{}) {
	c.Log(
		log.NewMessage().
			Level(log.LLInfo).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprintln(v...)).
			Metadata(c.meta).
			Build(),
	)

}

// Printf method implements the Printer interface.
//
// It is similar to fmt.Printf; and will print a message using an fmt.Sprintf(format, v...) pattern
//
// It applies LogLevel Info
func (c GRPCLogClient) Printf(format string, v ...interface{}) {
	c.Log(
		log.NewMessage().
			Level(log.LLInfo).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprintf(format, v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Panic method implements the Printer interface.
//
// It is similar to fmt.Print; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Panic.
func (c GRPCLogClient) Panic(v ...interface{}) {
	body := fmt.Sprint(v...)

	c.Log(
		log.NewMessage().
			Level(log.LLPanic).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(body).
			Metadata(c.meta).
			Build(),
	)
}

// Panicln method implements the Printer interface.
//
// It is similar to fmt.Println; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Panic.
func (c GRPCLogClient) Panicln(v ...interface{}) {
	body := fmt.Sprintln(v...)

	c.Log(
		log.NewMessage().
			Level(log.LLPanic).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(body).
			Metadata(c.meta).
			Build(),
	)
}

// Panicf method implements the Printer interface.
//
// It is similar to fmt.Printf; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Panic.
func (c GRPCLogClient) Panicf(format string, v ...interface{}) {
	body := fmt.Sprintf(format, v...)

	c.Log(
		log.NewMessage().
			Level(log.LLPanic).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(body).
			Metadata(c.meta).
			Build(),
	)
}

// Fatal method implements the Printer interface.
//
// It is similar to fmt.Print; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Fatal.
func (c GRPCLogClient) Fatal(v ...interface{}) {
	c.Log(
		log.NewMessage().
			Level(log.LLFatal).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprint(v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Fatalln method implements the Printer interface.
//
// It is similar to fmt.Println; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Fatal.
func (c GRPCLogClient) Fatalln(v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLFatal).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprintln(v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Fatalf method implements the Printer interface.
//
// It is similar to fmt.Printf; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Fatal.
func (c GRPCLogClient) Fatalf(format string, v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLFatal).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprintf(format, v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Error method implements the Printer interface.
//
// It is similar to fmt.Print; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Error.
func (c GRPCLogClient) Error(v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLError).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprint(v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Errorln method implements the Printer interface.
//
// It is similar to fmt.Println; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Error.
func (c GRPCLogClient) Errorln(v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLError).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprintln(v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Errorf method implements the Printer interface.
//
// It is similar to fmt.Printf; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Error.
func (c GRPCLogClient) Errorf(format string, v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLError).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprintf(format, v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Warn method implements the Printer interface.
//
// It is similar to fmt.Print; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Warn.
func (c GRPCLogClient) Warn(v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLWarn).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprint(v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Warnln method implements the Printer interface.
//
// It is similar to fmt.Println; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Warn.
func (c GRPCLogClient) Warnln(v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLWarn).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprintln(v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Warnf method implements the Printer interface.
//
// It is similar to fmt.Printf; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Warn.
func (c GRPCLogClient) Warnf(format string, v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLWarn).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprintf(format, v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Info method implements the Printer interface.
//
// It is similar to fmt.Print; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Info.
func (c GRPCLogClient) Info(v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLInfo).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprint(v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Infoln method implements the Printer interface.
//
// It is similar to fmt.Println; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Info.
func (c GRPCLogClient) Infoln(v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLInfo).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprintln(v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Infof method implements the Printer interface.
//
// It is similar to fmt.Printf; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Info.
func (c GRPCLogClient) Infof(format string, v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLInfo).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprintf(format, v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Debug method implements the Printer interface.
//
// It is similar to fmt.Print; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Debug.
func (c GRPCLogClient) Debug(v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLDebug).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprint(v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Debugln method implements the Printer interface.
//
// It is similar to fmt.Println; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Debug.
func (c GRPCLogClient) Debugln(v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLDebug).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprintln(v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Debugf method implements the Printer interface.
//
// It is similar to fmt.Printf; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Debug.
func (c GRPCLogClient) Debugf(format string, v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLDebug).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprintf(format, v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Trace method implements the Printer interface.
//
// It is similar to fmt.Print; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Trace.
func (c GRPCLogClient) Trace(v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLTrace).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprint(v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Traceln method implements the Printer interface.
//
// It is similar to fmt.Println; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Trace.
func (c GRPCLogClient) Traceln(v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLTrace).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprintln(v...)).
			Metadata(c.meta).
			Build(),
	)
}

// Tracef method implements the Printer interface.
//
// It is similar to fmt.Printf; and will print a message using an fmt.Sprint(v...) pattern, while
// automatically applying LogLevel Trace.
func (c GRPCLogClient) Tracef(format string, v ...interface{}) {

	c.Log(
		log.NewMessage().
			Level(log.LLTrace).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprintf(format, v...)).
			Metadata(c.meta).
			Build(),
	)
}
