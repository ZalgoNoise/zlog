package main

import (
	"fmt"
	"time"

	"github.com/zalgonoise/zlog/grpc/client"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

func main() {
	logger := log.New(
		log.WithFormat(log.TextColorLevelFirst),
		// log.WithFilter(log.LLWarn),
	)

	grpcLogger, errCh := client.New(
		client.WithAddr("127.0.0.1:9099"),
		client.UnaryRPC(),
		client.WithLogger(
			logger,
		),
		client.WithGRPCOpts(),
		client.WithTLS(
			"cert/ca/ca-cert.pem",
			// "cert/client/client-cert.pem",
			// "cert/client/client-key.pem",
		),
	)
	_, done := grpcLogger.Channels()

	msg1 := event.New().Message("all the way").Build()

	grpcLogger.Log(msg1)
	grpcLogger.Log(msg1)
	grpcLogger.Log(msg1)
	grpcLogger.Log(msg1)

	msg2 := event.New().Level(event.LLWarn)
	for i := 0; i < 10000; i++ {
		grpcLogger.Log(msg2.Message(fmt.Sprintf("#%v", i)).Build())
		time.Sleep(time.Millisecond * 2000)
	}

	// done <- struct{}{}

	for {
		select {
		case err := <-errCh:
			panic(err)
		case <-done:
			return
		}
	}
}
