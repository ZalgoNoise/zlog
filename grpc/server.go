package main

import (
	"github.com/zalgonoise/zlog/proto/server"
)

func main() {

	grpcLogger := server.New()
	grpcLogger.Serve()

}
