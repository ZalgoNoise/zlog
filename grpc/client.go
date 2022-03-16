package main

import (
	"fmt"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/proto/client"
)

func main() {
	logger := client.New(
		client.WithAddr(":9099"),
	)

	msg1 := log.NewMessage().Message("all the way").Build()
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

	go func() {

		logger.Info("test")
		logger.Prefix("service").Sub("module")
		logger.Warn("serious stuff")

		logger.Prefix("").Sub("")

		logger.Log(msg1)
		n, err := logger.Output(msg2)
		fmt.Println(n, err)

	}()

	for {
		err := <-logger.ErrCh
		fmt.Println(err)

	}

}
