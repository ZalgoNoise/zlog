package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/zalgonoise/zlog/grpc/address"
	"github.com/zalgonoise/zlog/grpc/client"
	"github.com/zalgonoise/zlog/log"
)

func main() {
	var buf = &bytes.Buffer{}

	stdLogger := log.New(
		log.WithOut(buf),
		log.SkipExit,
		log.ColorText,
	)

	grpcSvcLogger := log.New(
		log.FormatJSON,
		log.ColorTextLevelFirst,
		log.SkipExit,
	)

	grpcLogger, errCh := client.New(
		client.WithAddr(":9199"),
		client.WithLogger(grpcSvcLogger),
		client.UnaryRPC(),
	)

	newAddr := address.New(":9099", ":9299", ":9399")

	multiLogger := log.MultiLogger(
		grpcLogger,
		stdLogger,
	)
	multiLogger.Log(
		log.NewMessage().Message("it works!").Build(),
	)
	multiLogger.Log(
		log.NewMessage().Message("it works!").Build(),
	)
	multiLogger.Log(
		log.NewMessage().Message("it works!").Build(),
	)

	time.Sleep(time.Second * 1)

	multiLogger.AddOuts(os.Stdout, newAddr)
	multiLogger.Log(
		log.NewMessage().Message("it works!").Build(),
	)

	fmt.Println(buf.String())

	for {
		err := <-errCh
		panic(err)
	}

}
