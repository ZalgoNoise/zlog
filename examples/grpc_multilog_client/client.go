package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/zalgonoise/zlog/grpc/address"
	"github.com/zalgonoise/zlog/grpc/client"
	"github.com/zalgonoise/zlog/log"
)

func main() {
	var buf = &bytes.Buffer{}

	stdLogger := log.New(
		log.WithOut(buf),
		log.SkipExitCfg,
		log.ColorText,
	)

	grpcSvcLogger := log.New(
		log.JSONCfg,
		log.ColorTextLevelFirst,
		log.SkipExitCfg,
	)

	grpcLogger, errCh := client.New(
		client.WithAddr(":9999"),
		client.WithLogger(grpcSvcLogger),
		// client.UnaryRPC(),
	)

	newAddr := address.New(":9099")

	multiLogger := log.MultiLogger(
		grpcLogger,
		stdLogger,
	)

	multiLogger.AddOuts(os.Stdout, newAddr)

	multiLogger.Log(
		log.NewMessage().Message("it works!").Build(),
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

	fmt.Println(buf.String())

	for {
		err := <-errCh
		panic(err)
	}

}
