package client

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/zalgonoise/zlog/config"
	"github.com/zalgonoise/zlog/grpc/address"
	"github.com/zalgonoise/zlog/log"
	pb "github.com/zalgonoise/zlog/proto/message"
	"google.golang.org/grpc"
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
var gRPCLogClientBuilderType *GRPCLogClientBuilder = &GRPCLogClientBuilder{}

type GRPCLogger interface {
	log.Logger
	log.ChanneledLogger
}

type GRPCLogClient struct {
	addr  *address.ConnAddr
	opts  []grpc.DialOption
	msgCh chan *log.LogMessage
	done  chan struct{}

	svcLogger log.Logger

	prefix      string
	sub         string
	meta        map[string]interface{}
	skipExit    bool
	levelFilter int
}

type GRPCLogClientBuilder struct {
	conf *config.Configs
}

func newGRPCLogClient(confs ...config.Config) *GRPCLogClientBuilder {
	// enforce defaults on builder
	builder := &GRPCLogClientBuilder{
		conf: gRPCLogClientDefaults(),
	}

	// apply input configs
	for _, config := range confs {
		config.Apply(builder.conf)
	}

	// type check configs
	for _, config := range builder.conf.Map() {
		if !config.Is(gRPCLogClientBuilderType) {
			return nil
		}
	}

	// temp -- backoff handler
	// TODO: move conf logic to gRPC client struct
	backoff = builder.conf.Get(LSBackoff.String()).(*ExpBackoff)

	return builder
}

// factory
func New(opts ...config.Config) (GRPCLogger, chan error) {
	builder := newGRPCLogClient(opts...)

	// join []grpc.DialOption
	opt := builder.conf.Get(LSGRPCOpts.String()).([]grpc.DialOption)
	tls := builder.conf.Get(LSTLS.String()).([]grpc.DialOption)

	grpcOpts := []grpc.DialOption{}
	grpcOpts = append(grpcOpts, opt...)
	grpcOpts = append(grpcOpts, tls...)

	client := &GRPCLogClient{
		addr:      builder.conf.Get(LSAddress.String()).(*address.ConnAddr),
		opts:      grpcOpts,
		msgCh:     make(chan *log.LogMessage),
		done:      make(chan struct{}),
		svcLogger: builder.conf.Get(LSLogging.String()).(log.Logger),
	}

	if builder.conf.Get(LSType.String()).(string) == "stream" {

		client.svcLogger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("init").Message("setting up Stream gRPC client").Build())
		return newStreamLogger(client)

	} else {
		client.svcLogger.Log(log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("init").Message("setting up Unary gRPC client").Build())
		return newUnaryLogger(client)

	}

}

func newUnaryLogger(c *GRPCLogClient) (GRPCLogger, chan error) {
	errCh := make(chan error)

	backoff.RegisterLog(c.log, errCh).WithDone(&c.done)

	go c.listen(errCh)

	return c, errCh
}

func newStreamLogger(c *GRPCLogClient) (GRPCLogger, chan error) {

	errCh := make(chan error)

	backoff.RegisterStream(c.stream, errCh).WithDone(&c.done)

	go c.stream(errCh)

	return c, errCh
}

func (c GRPCLogClient) listen(errCh chan error) {
	for {
		msg := <-c.msgCh
		backoff.AddMessage(msg)
		go c.log(msg, errCh)
	}
}

func (c GRPCLogClient) connect() error {

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

		if err != nil {
			c.svcLogger.Log(
				log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("conn").Metadata(log.Field{
					"error":      err,
					"iterations": backoff.Counter(),
					"maxWait":    backoff.Max(),
					"curWait":    backoff.Current(),
				}).Message("retrying connection").Build(),
			)

			if backoff.IsLocked() {
				return ErrBackoffLocked
			} else {

				call, err := backoff.Increment().Wait()

				if err != nil {
					c.svcLogger.Log(
						log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("conn").
							Metadata(log.Field{
								"error": err.Error(),
							}).
							Message("removing address after failed dial attempt").Build(),
					)
					c.addr.Unset(remote)
					continue
				} else {
					go call()
					return ErrFailedConn
				}
			}

		}

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
	if liveConns == 0 {
		c.svcLogger.Log(log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("conn").Metadata(log.Field{"error": ErrNoConns.Error()}).Message("all connections failed").Build())
		return ErrNoConns
	}

	return nil
}

func (c GRPCLogClient) log(msg *log.LogMessage, errCh chan error) {
	err := c.connect()

	if err != nil {
		if errors.Is(err, ErrFailedConn) || errors.Is(err, ErrBackoffLocked) {
			return
		}

		errCh <- err

		c.svcLogger.Log(
			log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("log").Metadata(log.Field{
				"error": err.Error(),
			}).Message("failed to connect").Build(),
		)

		return
	}

	for remote, conn := range c.addr.Map() {
		defer conn.Close()

		client := pb.NewLogServiceClient(conn)

		c.svcLogger.Log(
			log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("log").Metadata(log.Field{
				"remote": remote,
			}).Message("setting up log service with connection").Build(),
		)

		ctx, cancel, reqID := pb.NewContextTimeout(pb.DefaultTimeout)

		c.svcLogger.Log(
			log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("log").Metadata(log.Field{
				"timeout": pb.TimeoutSeconds,
				"id":      reqID,
				"remote":  remote,
			}).Message("received a new log message to register").Build(),
		)

		response, err := client.Log(ctx, msg.Proto())

		if err != nil {
			errCh <- err
			cancel()

			c.svcLogger.Log(
				log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("log").Metadata(log.Field{
					"id":     reqID,
					"remote": remote,
					"error":  err.Error(),
				}).Message("failed to send message to gRPC server").Build(),
			)

			return
		}
		if !response.Ok {
			errCh <- ErrBadResponse
			cancel()

			c.svcLogger.Log(
				log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("log").Metadata(log.Field{
					"id":     reqID,
					"remote": remote,
					"error":  err.Error(),
					"response": log.Field{
						"ok": response.GetOk(),
					},
				}).Message("failed to send message to gRPC server").Build(),
			)

			return
		}
		cancel()

	}
}

func (c GRPCLogClient) stream(errCh chan error) {

	localErr := make(chan error)

	err := c.connect()
	if err != nil {
		if errors.Is(err, ErrFailedConn) || errors.Is(err, ErrBackoffLocked) {
			return
		}
		errCh <- err

		c.svcLogger.Log(
			log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("stream").Metadata(log.Field{
				"error": err.Error(),
			}).Message("failed to connect").Build(),
		)

		return
	}
	for remote, conn := range c.addr.Map() {
		logClient := pb.NewLogServiceClient(conn)

		c.svcLogger.Log(
			log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("stream").Metadata(log.Field{
				"remote": remote,
			}).Message("setting up log service with connection").Build(),
		)

		ctx, cancel, reqID := pb.NewContextTimeout(pb.DefaultStreamTimeout)

		c.svcLogger.Log(
			log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("stream").Metadata(log.Field{
				"timeout": pb.StreamTimeoutSeconds,
				"id":      reqID,
				"remote":  remote,
			}).Message("setting request ID for long-lived connection").Build(),
		)

		stream, err := logClient.LogStream(ctx)

		if err != nil {
			if ErrConnRefusedRegexp.MatchString(err.Error()) || ErrEOFRegexp.MatchString(err.Error()) {
				// conn.Close()
				call, err := backoff.Increment().Wait()

				if err != nil {
					fmt.Println("stream err listner: ", ErrFailedRetry)
					cancel()

					errCh <- ErrFailedRetry

					c.svcLogger.Log(
						log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("stream").Metadata(log.Field{
							"id":         reqID,
							"error":      err.Error(),
							"numRetries": backoff.Counter(),
						}).Message("closing stream after too many failed attempts to reconnect").Build(),
					)
					return

				}

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

			errCh <- err
			conn.Close()
			cancel()

			c.svcLogger.Log(
				log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("stream").Metadata(log.Field{
					"id":     reqID,
					"remote": remote,
					"error":  err.Error(),
				}).Message("failed to setup stream connection with gRPC server").Build(),
			)

			return
		}

		go func() {
			respCh := make(chan bool)
			go c.handleStreamService(
				reqID,
				stream,
				localErr,
				respCh,
				c.done,
			)

			c.handleStreamMessages(
				reqID,
				stream,
				localErr,
				errCh,
				respCh,
				cancel,
			)
		}()
	}
}

func (c GRPCLogClient) handleStreamService(
	reqID string,
	stream pb.LogService_LogStreamClient,
	localErr chan error,
	respCh chan bool,
	done chan struct{},
) {
	for {
		if in, err := stream.Recv(); err != nil {
			localErr <- err

			c.svcLogger.Log(
				log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("stream").Metadata(log.Field{
					"id":    reqID,
					"error": err.Error(),
				}).Message("issue receiving message from stream").Build(),
			)
			continue
		} else {
			c.svcLogger.Log(
				log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("stream").Metadata(log.Field{
					"id": reqID,
				}).Message("response received from gRPC server").Build(),
			)

			respCh <- in.GetOk()
		}
	}
}

func (c GRPCLogClient) handleStreamMessages(
	reqID string,
	stream pb.LogService_LogStreamClient,
	localErr chan error,
	errCh chan error,
	respCh chan bool,
	cancel context.CancelFunc,
) {
	for {
		select {
		case out := <-c.msgCh:

			c.svcLogger.Log(
				log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("stream").Metadata(log.Field{
					"id": reqID,
				}).Message("incoming log message to send").Build(),
			)

			err := stream.Send(out.Proto())
			if err != nil {
				localErr <- err

				c.svcLogger.Log(
					log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("stream").Metadata(log.Field{
						"id":    reqID,
						"error": err.Error(),
					}).Message("issue sending log message to gRPC server").Build(),
				)
			}
		case in := <-respCh:
			c.svcLogger.Log(
				log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("stream").Metadata(log.Field{
					"id": reqID,
					"response": log.Field{
						"ok": in,
					},
				}).Message("registering server response").Build(),
			)

			if !in {
				errCh <- ErrBadResponse

				c.svcLogger.Log(
					log.NewMessage().Level(log.LLWarn).Prefix("gRPC").Sub("stream").Metadata(log.Field{
						"id": reqID,
						"response": log.Field{
							"ok": in,
						},
						"error": ErrBadResponse.Error(),
					}).Message("failed to write log message").Build(),
				)

				return
			}
		case <-c.done:
			cancel()

			c.svcLogger.Log(
				log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("stream").Metadata(log.Field{
					"id": reqID,
				}).Message("received done signal").Build(),
			)

			return

		case err := <-localErr:
			if ErrDeadlineRegexp.MatchString(err.Error()) {

				c.svcLogger.Log(
					log.NewMessage().Level(log.LLDebug).Prefix("gRPC").Sub("stream").Metadata(log.Field{
						"id":    reqID,
						"error": err.Error(),
					}).Message("stream timed-out -- starting a new connection").Build(),
				)

				go c.stream(errCh)

			} else if ErrEOFRegexp.MatchString(err.Error()) || ErrConnRefusedRegexp.MatchString(err.Error()) {

				// TODO: implement exponential backoff
				// this is a temp fallback to retry the connection
				// using a 5 second timer before trying again
				call, err := backoff.Increment().Wait()

				if err != nil {

					fmt.Println("localErr listner: ", ErrFailedRetry)

					c.svcLogger.Log(
						log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("stream").Metadata(log.Field{
							"id":         reqID,
							"error":      err.Error(),
							"numRetries": backoff.Counter(),
						}).Message("closing stream after too many failed attempts to reconnect").Build(),
					)
					c.done <- struct{}{}
					errCh <- ErrFailedRetry
					cancel()
					return

				} else {
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

			} else {
				errCh <- err
				cancel()

				c.svcLogger.Log(
					log.NewMessage().Level(log.LLFatal).Prefix("gRPC").Sub("stream").Metadata(log.Field{
						"id":    reqID,
						"error": err.Error(),
					}).Message("critical error -- closing stream").Build(),
				)

			}
			return
		}

	}
}

// implement ChanneledLogger
func (c GRPCLogClient) Close() {
	for _, conn := range c.addr.Map() {
		conn.Close()
	}
	c.done <- struct{}{}
}

func (c GRPCLogClient) Channels() (logCh chan *log.LogMessage, done chan struct{}) {
	return c.msgCh, c.done
}

// implement Logger

func (c GRPCLogClient) Output(m *log.LogMessage) (n int, err error) {
	if c.levelFilter > log.LogTypeKeys[m.Level] {
		return 0, nil
	}

	c.msgCh <- m
	return 1, nil

}

func (c GRPCLogClient) SetOuts(outs ...io.Writer) log.Logger {
	c.addr.Reset()

	for _, remote := range outs {
		if r, ok := remote.(*address.ConnAddr); !ok {

			c.svcLogger.Log(
				log.NewMessage().Level(log.LLWarn).
					Prefix("gRPC").Sub("SetOuts()").
					Metadata(log.Field{"error": ErrBadWriter.Error()}).
					Message("invalid writer warning").Build(),
			)

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

	err := c.connect()
	if err != nil {
		return nil
	}

	return c
}
func (c GRPCLogClient) AddOuts(outs ...io.Writer) log.Logger {
	for _, remote := range outs {
		if r, ok := remote.(*address.ConnAddr); !ok {

			c.svcLogger.Log(
				log.NewMessage().Level(log.LLWarn).
					Prefix("gRPC").Sub("AddOuts()").
					Metadata(log.Field{"error": ErrBadWriter.Error()}).
					Message("invalid writer warning").Build(),
			)

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

	err := c.connect()
	if err != nil {
		return nil
	}

	return c
}

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

func (c GRPCLogClient) Prefix(prefix string) log.Logger {
	if prefix == "" {
		c.prefix = "log"
		return c
	}
	c.prefix = prefix
	return c
}

func (c GRPCLogClient) Sub(sub string) log.Logger {
	c.sub = sub
	return c
}

func (c GRPCLogClient) Fields(fields map[string]interface{}) log.Logger {
	c.meta = fields
	return c
}

func (c GRPCLogClient) IsSkipExit() bool {
	return c.skipExit
}

func (c GRPCLogClient) Log(m ...*log.LogMessage) {
	for _, msg := range m {
		c.Output(msg)
	}
}

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

	if !c.skipExit {
		panic(body)
	}
}

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

	if !c.skipExit {
		panic(body)
	}
}

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

	if !c.skipExit {
		panic(body)
	}
}

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

	if !c.skipExit {
		os.Exit(1)
	}
}

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

	if !c.skipExit {
		os.Exit(1)
	}
}

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

	if !c.skipExit {
		os.Exit(1)
	}
}

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
