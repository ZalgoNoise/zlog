package client

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/zalgonoise/zlog/log"
	pb "github.com/zalgonoise/zlog/proto/message"
	"google.golang.org/grpc"
)

var DeadlineError = regexp.MustCompile(`rpc error: code = DeadlineExceeded desc = context deadline exceeded`)

type GRPCLogger interface {
	log.Logger
	log.ChanneledLogger
}

type GRPCLogClient struct {
	addr  *ConnAddr
	opts  []grpc.DialOption
	msgCh chan *log.LogMessage
	done  chan struct{}

	prefix      string
	sub         string
	meta        map[string]interface{}
	skipExit    bool
	levelFilter int
}

type GRPCLogClientBuilder struct {
	addr    *ConnAddr
	opts    []grpc.DialOption
	isUnary bool
}

func newGRPCLogClient(opts ...LogClientConfig) *GRPCLogClientBuilder {
	builder := &GRPCLogClientBuilder{}

	for _, opt := range opts {
		opt.Apply(builder)
	}

	if builder.addr.Len() == 0 {
		WithAddr("").Apply(builder)
	}

	if builder.opts == nil {
		WithGRPCOpts().Apply(builder)
	}

	return builder
}

// factory
func New(opts ...LogClientConfig) (GRPCLogger, chan error) {
	builder := newGRPCLogClient(opts...)

	client := &GRPCLogClient{
		addr:  builder.addr,
		opts:  builder.opts,
		msgCh: make(chan *log.LogMessage),
		done:  make(chan struct{}),
	}

	if !builder.isUnary {
		return newStreamLogger(client)
	} else {
		return newUnaryLogger(client)
	}

}

func newUnaryLogger(c *GRPCLogClient) (GRPCLogger, chan error) {
	errCh := make(chan error)

	go c.log(errCh)

	return c, errCh
}

func newStreamLogger(c *GRPCLogClient) (GRPCLogger, chan error) {

	errCh := make(chan error)

	go c.stream(errCh)

	return c, errCh
}

func (c GRPCLogClient) connect() error {
	if c.addr.Len() == 0 {
		return errors.New("cannot connect to gRPC server since no addresses were provided")
	}

	for _, remote := range c.addr.Keys() {
		// connect
		var conn *grpc.ClientConn
		conn, err := grpc.Dial(remote, c.opts...)

		if err != nil {
			c.addr.Unset(remote)
			return err
		}
		c.addr.Set(remote, conn)

	}
	return nil
}

func (c GRPCLogClient) log(errCh chan error) {
	err := c.connect()
	if err != nil {
		errCh <- err
		return
	}

	for remote, conn := range c.addr.Map() {
		defer conn.Close()

		client := pb.NewLogServiceClient(conn)

		for msg := range c.msgCh {

			ctx, cancel := pb.NewContextTimeout()

			response, err := client.Log(ctx, msg.Proto())
			if err != nil {
				errCh <- err
				cancel()
				return
			}
			if !response.Ok {
				errCh <- fmt.Errorf("failed to write message to gRPC Log Server %s: %v", remote, response)
				cancel()
				return
			}
			cancel()
		}
	}
}

func (c GRPCLogClient) stream(errCh chan error) {

	localErr := make(chan error)

	err := c.connect()
	if err != nil {
		errCh <- err
		return
	}
	for _, conn := range c.addr.Map() {
		logClient := pb.NewLogServiceClient(conn)

		ctx, cancel := pb.NewContextTimeout()

		stream, err := logClient.LogStream(ctx)

		if err != nil {
			errCh <- err
		}

		go func() {
			respCh := make(chan bool)
			go func() {
				for {
					in, err := stream.Recv()
					if err != nil {
						localErr <- err
					}
					respCh <- in.GetOk()
				}
			}()

			for {
				select {
				case out := <-c.msgCh:
					err := stream.Send(out.Proto())
					if err != nil {
						localErr <- err
					}
				case in := <-respCh:
					if !in {
						errCh <- errors.New("failed to write log message to gRPC server")
						return
					}
				case <-c.done:
					cancel()
					return
				case err := <-localErr:
					if DeadlineError.MatchString(err.Error()) {
						go c.stream(errCh)
					} else {
						errCh <- err
						cancel()
					}
					return
				}

			}
		}()
	}
}

// implement ChanneledLogger
func (c GRPCLogClient) Close() {
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
	addr := &ConnAddr{}

	for _, remote := range outs {
		r, ok := remote.(ConnAddr)
		if !ok {
			return c
		}
		addr.Add(r.Keys()...)
	}

	c.addr = addr

	return c
}
func (c GRPCLogClient) AddOuts(outs ...io.Writer) log.Logger {
	addr := &ConnAddr{}

	for _, remote := range outs {
		r, ok := remote.(ConnAddr)
		if !ok {
			return c
		}
		addr.Add(r.Keys()...)
	}

	c.addr.Add(addr.Keys()...)

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
