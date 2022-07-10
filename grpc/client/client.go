package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/zalgonoise/zlog/grpc/address"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/log/logch"

	pb "github.com/zalgonoise/zlog/proto/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	ErrNoAddr      error = errors.New("cannot connect to gRPC server since no addresses were provided")
	ErrNoConns     error = errors.New("could not establish any successful connection with the provided address(es)")
	ErrBadResponse error = errors.New("failed to write log message in remote gRPC server")
	ErrBadWriter   error = errors.New("invalid writer -- must be of type client.ConnAddr")
)

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
	logch.ChanneledLogger
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
	msgCh chan *event.Event
	done  chan struct{}
	errCh chan error

	svcLogger log.Logger
	backoff   *Backoff

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
	backoff      *Backoff
	svcLogger    log.Logger
}

// clientInterceptors struct is a placeholder for different interceptors to be added
// to the GRPCLogServer
type clientInterceptors struct {
	streamItcp map[string]grpc.StreamClientInterceptor
	unaryItcp  map[string]grpc.UnaryClientInterceptor
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

	client := &GRPCLogClient{
		addr:      b.addr,
		opts:      append(b.opts, opts...),
		msgCh:     make(chan *event.Event),
		done:      make(chan struct{}),
		errCh:     make(chan error),
		svcLogger: b.svcLogger,
		backoff:   b.backoff,
		prefix:    "log",
		sub:       "",
		meta:      map[string]interface{}{},
	}

	client.backoff.init(b, client)

	return client

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
	var cfg []LogClientConfig

	for _, o := range opts {
		if o == nil {
			continue
		}
		cfg = append(cfg, o)
	}

	builder := newGRPCLogClient(cfg...)

	client := builder.build()

	// check input type -- create an appropriate GRPCLogger
	if builder.isUnary {
		client.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("init").Message("setting up Unary gRPC client").Build())

		return newUnaryLogger(client)
	} else {
		client.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("init").Message("setting up Stream gRPC client").Build())

		return newStreamLogger(client)
	}
}

func newUnaryLogger(c *GRPCLogClient) (GRPCLogger, chan error) {
	// launch log listener in a goroutine
	go c.listen()

	return c, c.errCh
}

func newStreamLogger(c *GRPCLogClient) (GRPCLogger, chan error) {
	// launch log listener in a goroutine
	go c.stream()

	return c, c.errCh
}

// connect method will iterate the connections map and dial each remote
//
// It stores the connection in the map or it removes the entry in case it is unhealthy
func (c *GRPCLogClient) connect() error {

	// exit if no addresses are set
	if c.addr.Len() == 0 {
		c.svcLogger.Log(event.New().Level(event.Level_fatal).Prefix("gRPC").Sub("conn").Metadata(event.Field{"error": ErrNoAddr.Error()}).Message("no addresses provided").Build())
		return ErrNoAddr
	}

	var liveConns int = 0

	for idx, remote := range c.addr.Keys() {
		c.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("conn").Metadata(event.Field{
			"index": idx,
			"addr":  remote,
		}).Message("connecting to remote").Build())

		var conn *grpc.ClientConn
		conn, err := grpc.Dial(remote, c.opts...)

		// handle dial errors
		if err != nil {
			retryErr := c.backoff.UnaryBackoffHandler(err, c.svcLogger)
			if errors.Is(retryErr, ErrBackoffLocked) {
				return retryErr
			} else if errors.Is(retryErr, ErrFailedConn) {
				return retryErr
			} else {
				c.svcLogger.Log(event.New().Level(event.Level_warn).Prefix("gRPC").Sub("retry").Metadata(event.Field{
					"error": err.Error(),
				}).Message("removing address after failed dial attempt").Build())

				// address is removed from connections map
				c.addr.Unset(remote)
				continue
			}
		}

		// once the connection is established, it's mapped to its (string) address
		c.addr.Set(remote, conn)
		liveConns++

		c.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("conn").Metadata(event.Field{
			"index": idx,
			"addr":  remote,
		}).Message("dialed the address successfully").Build())
	}

	// return ErrNoConns if the counter for live connections hasn't increased
	if liveConns == 0 {
		c.svcLogger.Log(event.New().Level(event.Level_fatal).Prefix("gRPC").Sub("conn").Metadata(event.Field{"error": ErrNoConns.Error()}).Message("all connections failed").Build())
		return ErrNoConns
	}

	return nil
}

// listen method is middleware to allow the backoff module to register an
// action call (in this case log()), which will be retried in case of failure
func (c *GRPCLogClient) listen() {
	for {
		select {
		case msg := <-c.msgCh:
			c.backoff.AddMessage(msg)
			go c.log(msg)
		case <-c.done:
			return
		}
	}
}

// log method is a go-routine function which will send all configured connections
// a Log() request.
func (c *GRPCLogClient) log(msg *event.Event) {

	// establish connections
	err := c.connect()

	// handle connection errors
	if err != nil {

		// check if errors are failed connection or backoff locked errors;
		// return so the action is canceLevel_ed
		if errors.Is(err, ErrFailedConn) || errors.Is(err, ErrBackoffLocked) {
			return
		}

		// any other errors will be sent to the error channel and logged locally;
		// then canceLevel_ing this action
		c.errCh <- err

		c.svcLogger.Log(event.New().Level(event.Level_fatal).Prefix("gRPC").Sub("log").Metadata(event.Field{
			"error": err.Error(),
		}).Message("failed to connect").Build())

		return
	}

	// there are live connections; log input message on each of the remote gRPC Log Servers
	for remote, conn := range c.addr.AsMap() {
		defer conn.Close()
		client := pb.NewLogServiceClient(conn)

		c.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("log").Metadata(event.Field{
			"remote": remote,
		}).Message("setting up log service with connection").Build())

		// generate a new context with a timeout
		bgCtx := context.Background()
		ctx, cancel := context.WithTimeout(bgCtx, defaultTimeout)

		c.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("log").Metadata(event.Field{
			"timeout": timeoutSeconds,
			"remote":  remote,
		}).Message("received a new log message to register").Build())

		// send LogMessage to remote gRPC Log Server
		// response is parsed by the interceptor; only the error is important
		_, err := client.Log(ctx, msg)

		// if the server returns any error, it's sent to the error channel, context cancelled,
		// and then return
		if err != nil {
			c.errCh <- err
			cancel()
			return
		}

		// message sent; response retrieved; context is cancelled.
		cancel()
	}
}

// stream method is a go-routine function which will establish a connection with
// all configured addresses to create a Stream().
func (c *GRPCLogClient) stream() {

	localErr := make(chan error)

	// establish connections
	err := c.connect()

	// handle connection errors
	if err != nil {

		// check if errors are failed connection or backoff locked errors;
		// return so the action is canceLevel_ed
		if errors.Is(err, ErrFailedConn) || errors.Is(err, ErrBackoffLocked) {
			return
		}

		// any other errors will be sent to the error channel and logged locally;
		// then canceLevel_ing this action
		c.errCh <- err

		c.svcLogger.Log(event.New().Level(event.Level_fatal).Prefix("gRPC").Sub("stream").Metadata(event.Field{
			"error": err.Error(),
		}).Message("failed to connect").Build())

		return
	}

	// there are live connections; setup a stream with each of the remote gRPC Log Servers
	for remote, conn := range c.addr.AsMap() {
		logClient := pb.NewLogServiceClient(conn)

		c.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("stream").Metadata(event.Field{
			"remote": remote,
		}).Message("setting up log service with connection").Build())

		// generate a new context with a timeout
		bgCtx := context.Background()
		ctx, cancel := context.WithTimeout(bgCtx, defaultStreamTimeout)

		c.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("stream").Metadata(event.Field{
			"timeout": streamTimeoutSeconds,
			"remote":  remote,
		}).Message("setting request ID for long-lived connection").Build())

		// setup a stream with the remote gRPC Log Server
		stream, err := logClient.LogStream(ctx)

		// check for errors when creating a stream
		if err != nil {

			// if it's connection refused or EOF, kick-off the backoff routine
			if errCode := status.Code(err); errCode == codes.Unavailable || errors.Is(err, io.EOF) {
				c.backoff.StreamBackoffHandler(c.errCh, cancel, c.svcLogger, c.done)
			}

			// otherwise, error is sent to the error channel, conn closed, context canceLevel_ed,
			// error logged and then return
			c.errCh <- err
			conn.Close()
			cancel()

			c.svcLogger.Log(event.New().Level(event.Level_warn).Prefix("gRPC").Sub("stream").Metadata(event.Field{
				"remote": remote,
				"error":  err.Error(),
			}).Message("failed to setup stream connection with gRPC server").Build())

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
			go c.streamHandler(
				stream,
				localErr,
				c.done,
			)

			c.handleStreamMessages(
				stream,
				localErr,
				cancel,
			)
		}()
	}
}

// streamHandler method will serve as a simple gateway for the handleStreamService method,
// to allow monitoring on the done channel, and with it closing the connection when requested
// (in a stream RPC)
func (c *GRPCLogClient) streamHandler(
	stream pb.LogService_LogStreamClient,
	localErr chan error,
	done chan struct{},
) {
	errCh := make(chan error)

	go c.handleStreamService(stream, errCh)

	for {
		select {
		case e := <-errCh:
			localErr <- e
		case <-done:
			return
		}
	}
}

// handleStreamService method is in place to break-down stream()'s functionality
// into a smaLevel_er function.
//
// The method will loop forever acting as a listener to the gRPC stream, and handling
// incoming server responses (to messages posted to it). If the response is an error
// or not OK, this is sinked to the local error channel to be processed. Otherwise,
// the response is registered.
//
// This method is ran in paraLevel_el with handleStreamMessages()
func (c *GRPCLogClient) handleStreamService(
	stream pb.LogService_LogStreamClient,
	localErr chan error,
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
// into a smaLevel_er function.
//
// It will manage the exchanged messages in a gRPC Log Client, from its channels:
// - incoming log messages which should be sent to the gRPC Log Server
// - done requests (to gracefully exit)
// - error messages (sinked into the local error channel)
//
// This method is ran in paraLevel_el with handleStreamService()
func (c *GRPCLogClient) handleStreamMessages(
	stream pb.LogService_LogStreamClient,
	localErr chan error,
	cancel context.CancelFunc,
) {
	for {
		select {

		// LogMessage is received in the message channel -- send this message to the stream
		case out := <-c.msgCh:

			// send the protofied message and check for errors (sinked to local error channel)
			err := stream.Send(out)

			if err != nil {
				localErr <- err
			}

		// done is received -- gracefully exit by canceLevel_ing context and closing the connection
		case <-c.done:
			cancel()

			c.svcLogger.Log(
				event.New().Level(event.Level_debug).Prefix("gRPC").Sub("stream").Message("received done signal").Build(),
			)

			return

		// error is received -- parse error in case it is potentially expected; handle it
		// accordingly. In case the error is unexpected, it is sent to the (outbound) error
		// channel and signal done is sent.
		case err := <-localErr:

			// Stream Deadline Exceeded -- reconnect to gRPC Log Server
			if errCode := status.Code(err); errCode == codes.DeadlineExceeded {

				c.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("stream").Metadata(event.Field{
					"error": err.Error(),
				}).Message("stream timed-out -- starting a new connection").Build())

				go c.stream()

				// Connection Refused or EOF error -- trigger backoff routine
			} else if errCode := status.Code(err); errCode == codes.Unavailable || errors.Is(err, io.EOF) {

				c.backoff.StreamBackoffHandler(c.errCh, cancel, c.svcLogger, c.done)

				// Bad Response -- send to error channel, continue
			} else if errors.Is(err, ErrBadResponse) {

				c.errCh <- err
				continue

				// default -- send to error channel, close the client
			} else {
				c.errCh <- err
				defer c.Close()

				c.svcLogger.Log(event.New().Level(event.Level_fatal).Prefix("gRPC").Sub("stream").Metadata(event.Field{
					"error": err.Error(),
				}).Message("critical error -- closing stream").Build())

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
func (c *GRPCLogClient) Close() {
	for _, conn := range c.addr.AsMap() {
		if conn == nil {
			continue
		}

		err := conn.Close()

		if err != nil {
			if errStatus, ok := status.FromError(err); ok {
				switch errStatus.Code() {
				case codes.Canceled:
					continue
				default:
					c.errCh <- err
					return
				}
			}
			c.errCh <- err
			return
		}

	}
	c.done <- struct{}{}
}

// Channels method is the implementation of the ChanneledLogger's Channels().
//
// It returns two channels (message and done channels) to allow control over the
// gRPC Log Client over a background / separate goroutine. Considering that
// creating a gRPC Log Client returns an error channel, this method will give
// the developer the three needed channels to work with the logger asynchronously
func (c *GRPCLogClient) Channels() (logCh chan *event.Event, done chan struct{}) {
	return c.msgCh, c.done
}

// Output method implements the Logger's Output().
//
// This method will simply push the incoming Log Message to the message channel,
// which is sent to a gRPC Log Server, either via a Unary or Stream RPC
func (c *GRPCLogClient) Output(m *event.Event) (n int, err error) {
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
func (c *GRPCLogClient) SetOuts(outs ...io.Writer) log.Logger {

	var o []string

	for _, remote := range outs {
		// ensure the input writer is not nil
		if remote == nil {
			continue
		}

		// ensure the input writer is of type *address.ConnAddr
		// if not, skip this writer and register this event
		if r, ok := remote.(*address.ConnAddr); !ok {

			c.svcLogger.Log(event.New().Level(event.Level_warn).Prefix("gRPC").Sub("SetOuts()").Metadata(event.Field{
				"error": ErrBadWriter.Error(),
			}).Message("invalid writer warning").Build())

			// writer is valid -- add it to connections map; register this event
		} else {
			o = append(o, r.Keys()...)
		}
	}

	if len(o) > 0 {
		// reset connections map
		c.addr.Reset()

		c.addr.Add(o...)

		c.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("SetOuts()").Metadata(event.Field{
			"addrs": o,
		}).Message("added address to connection address map").Build())

		// test connectivity to remotes -- if this returns an error, the function
		// will return nil instead of the GRPCLogClient
		err := c.connect()
		if err != nil {
			return nil
		}
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
func (c *GRPCLogClient) AddOuts(outs ...io.Writer) log.Logger {
	// don't repeat addresses
	current := c.addr.Keys()

	var o []string

	for _, remote := range outs {
		// ensure the input writer is not nil
		if remote == nil {
			continue
		}

		// ensure the input writer is of type *address.ConnAddr
		// if not, skip this writer and register this event
		if r, ok := remote.(*address.ConnAddr); !ok {

			c.svcLogger.Log(event.New().Level(event.Level_warn).Prefix("gRPC").Sub("AddOuts()").Metadata(event.Field{
				"error": ErrBadWriter.Error(),
			}).Message("invalid writer warning").Build())

			// writer is valid -- add it to connections map; register this event
		} else {
			addresses := r.Keys()
			var toAppend []string

			for a := 0; a < len(addresses); a++ {
				var isNew bool = true

				for b := 0; b < len(current); b++ {
					if current[b] == addresses[a] {
						isNew = false
						break
					}
				}

				if isNew {
					toAppend = append(toAppend, addresses[a])
				}
			}

			if len(toAppend) > 0 {
				o = append(o, toAppend...)
			}
		}
	}

	if len(o) > 0 {
		c.addr.Add(o...)

		c.svcLogger.Log(event.New().Level(event.Level_debug).Prefix("gRPC").Sub("SetOuts()").Metadata(event.Field{
			"addrs": o,
		}).Message("added address to connection address map").Build())

		// test connectivity to remotes -- if this returns an error, the function
		// will return nil instead of the GRPCLogClient
		err := c.connect()
		if err != nil {
			return nil
		}
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
func (c *GRPCLogClient) Write(p []byte) (n int, err error) {
	if p == nil || len(p) == 0 {
		return 0, nil
	}

	// decode bytes
	m, err := event.Decode(p)

	if err != nil {
		return c.Output(event.New().
			Level(event.Default_Event_Level).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(string(p)).
			Metadata(c.meta).
			Build())

	}

	// print message
	return c.Output(m)
}

// Prefix method implements the Logger interface.
//
// It will set a Logger-scoped (as opposed to message-scoped) prefix string to the logger
//
// Logger-scoped prefix strings can be set in order to avoid caLevel_ing the `MessageBuilder.Prefix()` method
// repeatedly, and instead doing so via the logger at the beginning of a service or function
//
// A logger-scoped prefix is not cleared with new Log messages, but `MessageBuilder.Prefix()` calls will
// replace it.
func (c *GRPCLogClient) Prefix(prefix string) log.Logger {
	if prefix == "" {
		c.prefix = "log"
		return c
	}
	c.prefix = prefix
	return c
}

// Sub method implements the Logger interface.
//
// Logger-scoped sub-prefix strings can be set in order to avoid caLevel_ing the `MessageBuilder.Sub()` method
// repeatedly, and instead doing so via the logger at the beginning of a service or function
//
// A logger-scoped sub-prefix is not cleared with new Log messages, but `MessageBuilder.Sub()` calls will
// replace it.
func (c *GRPCLogClient) Sub(sub string) log.Logger {
	c.sub = sub
	return c
}

// Fields method implements the Logger interface.
//
// Fields method will set Logger-scoped (as opposed to message-scoped) metadata fields to the logger
//
// Logger-scoped metadata can be set in order to avoid caLevel_ing the `MessageBuilder.Metadata()` method
// repeatedly, and instead doing so via the logger at the beginning of a service or function.
//
// Logger-scoped metadata fields are not cleared with new log messages, but only added to the existing
// metadata map. These can be cleared with a call to `Logger.Fields(nil)`
func (c *GRPCLogClient) Fields(fields map[string]interface{}) log.Logger {
	if fields == nil || len(fields) == 0 {
		c.meta = map[string]interface{}{}
		return c
	}

	c.meta = fields

	return c
}

// IsSkipExit method implements the Printer interface.
//
// IsSkipExit method returns a boolean on whether the gRPC Log Client's service logger is
// set to skip os.Exit(1) or panic() calls.
//
// It is used in functions which call these, to first evaluate if those calls should be
// executed or skipped.
//
// Considering this method is unused by the gRPC Log Client, it's merely here to
// implement the interface
func (c *GRPCLogClient) IsSkipExit() bool {
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
func (c *GRPCLogClient) Log(m ...*event.Event) {
	var queue []*event.Event

	for _, msg := range m {
		if msg == nil {
			continue
		}
		queue = append(queue, msg)
	}

	for _, msg := range queue {
		c.Output(msg)
	}
}

// Print method implements the Printer interface.
//
// It is similar to fmt.Print; and will print a message using an fmt.Sprint(v...) pattern
//
// It applies LogLevel Info
func (c *GRPCLogClient) Print(v ...interface{}) {
	c.Log(
		event.New().
			Level(event.Level_info).
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
func (c *GRPCLogClient) Println(v ...interface{}) {
	c.Log(
		event.New().
			Level(event.Level_info).
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
func (c *GRPCLogClient) Printf(format string, v ...interface{}) {
	c.Log(
		event.New().
			Level(event.Level_info).
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
func (c *GRPCLogClient) Panic(v ...interface{}) {
	body := fmt.Sprint(v...)

	c.Log(
		event.New().
			Level(event.Level_panic).
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
func (c *GRPCLogClient) Panicln(v ...interface{}) {
	body := fmt.Sprintln(v...)

	c.Log(
		event.New().
			Level(event.Level_panic).
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
func (c *GRPCLogClient) Panicf(format string, v ...interface{}) {
	body := fmt.Sprintf(format, v...)

	c.Log(
		event.New().
			Level(event.Level_panic).
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
func (c *GRPCLogClient) Fatal(v ...interface{}) {
	c.Log(
		event.New().
			Level(event.Level_fatal).
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
func (c *GRPCLogClient) Fatalln(v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_fatal).
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
func (c *GRPCLogClient) Fatalf(format string, v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_fatal).
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
func (c *GRPCLogClient) Error(v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_error).
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
func (c *GRPCLogClient) Errorln(v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_error).
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
func (c *GRPCLogClient) Errorf(format string, v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_error).
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
func (c *GRPCLogClient) Warn(v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_warn).
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
func (c *GRPCLogClient) Warnln(v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_warn).
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
func (c *GRPCLogClient) Warnf(format string, v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_warn).
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
func (c *GRPCLogClient) Info(v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_info).
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
func (c *GRPCLogClient) Infoln(v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_info).
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
func (c *GRPCLogClient) Infof(format string, v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_info).
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
func (c *GRPCLogClient) Debug(v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_debug).
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
func (c *GRPCLogClient) Debugln(v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_debug).
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
func (c *GRPCLogClient) Debugf(format string, v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_debug).
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
func (c *GRPCLogClient) Trace(v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_trace).
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
func (c *GRPCLogClient) Traceln(v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_trace).
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
func (c *GRPCLogClient) Tracef(format string, v ...interface{}) {

	c.Log(
		event.New().
			Level(event.Level_trace).
			Prefix(c.prefix).
			Sub(c.sub).
			Message(fmt.Sprintf(format, v...)).
			Metadata(c.meta).
			Build(),
	)
}
