package main

import (
	"github.com/zalgonoise/zlog/log"

	"github.com/zalgonoise/zlog/grpc/server"
)

func main() {

	grpcLogger := server.New(
		server.WithLogger(
			log.New(
				log.ColorTextLevelFirstSpaced,
			),
		),
		server.WithServiceLogger(
			log.New(
				log.ColorTextLevelFirstSpaced,
			),
		),
		server.WithAddr(":9099"),
	)
	grpcLogger.Serve()

}
