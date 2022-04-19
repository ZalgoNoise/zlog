package main

import (
	"github.com/zalgonoise/zlog/log"

	"github.com/zalgonoise/zlog/grpc/server"
)

func main() {

	grpcLogger := server.New(
		server.WithLogger(
			log.New(
				log.WithFormat(log.TextColorLevelFirstSpaced),
			),
		),
		server.WithServiceLogger(
			log.New(
				log.WithFormat(log.TextColorLevelFirstSpaced),
			),
		),
		server.WithAddr("127.0.0.1:9099"),
		server.WithGRPCOpts(),
		server.WithTLS(
			"cert/server/server.pem",
			"cert/server/server.key",
			// "cert/ca/cacert.pem",
		),
	)
	grpcLogger.Serve()

}
