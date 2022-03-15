package main

import (
	"context"
	"fmt"

	"github.com/zalgonoise/zlog/log"
	"google.golang.org/grpc"
)

func main() {
	port := ":9000"

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(port, grpc.WithInsecure())

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	c := log.NewLogServiceClient(conn)

	msg := log.NewMessage().
		Level(log.LLTrace).
		Prefix("service").
		Sub("module").
		Message("grpc logging").
		Metadata(log.Field{
			"content":  true,
			"inner":    "yes",
			"quantity": 3,
		}).
		CallStack(true).
		Build().GRPC()

	response, err := c.Log(context.Background(), msg)

	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}
