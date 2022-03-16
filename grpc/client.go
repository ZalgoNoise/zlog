package main

import (
	"context"
	"fmt"

	"github.com/zalgonoise/zlog/log"
	pb "github.com/zalgonoise/zlog/proto/message"
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

	c := pb.NewLogServiceClient(conn)

	msg := log.NewMessage().
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
		Build().Proto()

	for i := 0; i < 10; i++ {
		r, err := sendMsg(c, msg)

		if err != nil {
			panic(err)
		}

		fmt.Println(r)
	}

}

func sendMsg(client pb.LogServiceClient, msg *pb.MessageRequest) (*pb.MessageResponse, error) {
	response, err := client.Log(context.Background(), msg)

	if err != nil {
		return nil, err
	}

	return response, nil
}
