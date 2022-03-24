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
		server.WithAddr("127.0.0.1:9099"),
		server.WithGRPCOpts(),
		server.WithTLS(
			"cert/server/server-cert.pem",
			"cert/server/server-key.pem",
			// "cert/ca/ca-cert.pem",
		),
	)
	grpcLogger.Serve()

}
