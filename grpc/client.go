package main

import (
	"fmt"
	"time"

	// "time"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/proto/client"
)

func main() {
	logger, errCh := client.New(
		client.WithAddr(":9099"),
		client.UnaryRPC(),
	)

	msg1 := log.NewMessage().Message("all the way").Build()
	// msg2 := log.NewMessage().
	// 	Level(log.LLInfo).
	// 	Prefix("service").
	// 	Sub("module").
	// 	Message("grpc logging").
	// 	Metadata(log.Field{
	// 		"content":  true,
	// 		"inner":    "yes",
	// 		"quantity": 3,
	// 	}).
	// 	CallStack(true).
	// 	Build()

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

		logger.Prefix("").Sub("")

		t := time.Now()
		// fmt.Println(t)
		for i := 0; i < 1000; i++ {
			logger.Log(msg1)
		}
		st := time.Since(t)
		fmt.Println(st)
		// n, err := logger.Output(msg2)
		// fmt.Println(n, err)

	}()

	for {
		err := <-errCh
		fmt.Println(err)

	}

}
