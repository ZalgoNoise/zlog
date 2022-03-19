package main

import (
	"fmt"
	"time"

	// "time"

	"github.com/zalgonoise/zlog/grpc/client"
	"github.com/zalgonoise/zlog/log"
)

func main() {
	logger, errCh := client.New(
		client.WithAddr(":9099"),
		client.UnaryRPC(),
	)

	msgCh, done := logger.Channels()

	// msg1 := log.NewMessage().Message("all the way").Build()
	msg2 := log.NewMessage().
		Level(log.LLInfo).
		Prefix("service").
		Sub("module").
		Message("grpc logging").
		Metadata(log.Field{
			"content":  true,
			"inner":    "yes",
			"quantity": 3,
		}).
		CallStack(true).
		Build()

	msg3 := log.NewMessage().Message("loop").Level(log.LLWarn)

	// logger <- msg1

	// time.Sleep(1 * time.Second)

	// logger <- msg2
	// time.Sleep(1 * time.Second)
	go func() {

		// logger <- msg1
		// time.Sleep(1 * time.Second)
		// logger <- msg2
		// time.Sleep(1 * time.Second)

		// logger.Info("test")
		// logger.Prefix("service").Sub("module")
		// logger.Warn("serious stuff")
		msgCh <- msg2

		logger.Prefix("").Sub("")

		t := time.Now()
		// fmt.Println(t)
		for i := 0; i < 1000; i++ {
			logger.Log(msg3.Message(fmt.Sprint(i)).Build())
			time.Sleep(time.Second * 1)
		}
		st := time.Since(t)
		fmt.Println(st)
		msgCh <- msg2
		// n, err := logger.Output(msg2)
		// fmt.Println(n, err)

		done <- struct{}{}

	}()

	for {
		err := <-errCh
		if client.DeadlineError.MatchString(err.Error()) {
			fmt.Println("caught deadline exceeded error")
		} else {
			panic(err)
		}

	}

}
