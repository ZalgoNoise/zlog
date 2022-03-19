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
	)
	grpcLogger.Serve()

}
