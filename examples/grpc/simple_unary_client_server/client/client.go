package main

import (
	"fmt"
	"os"
	"time"

	"github.com/zalgonoise/zlog/grpc/client"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

func getEnv(env string) (val string, ok bool) {
	v := os.Getenv(env)

	if v == "" {
		return v, false
	}

	return v, true
}

func getTLSConf() client.LogClientConfig {
	var tlsConf client.LogClientConfig

	var withCert bool
	var withKey bool
	var withCA bool

	certPath, ok := getEnv("TLS_CLIENT_CERT")
	if ok {
		withCert = true
	}

	keyPath, ok := getEnv("TLS_CLIENT_KEY")
	if ok {
		withKey = true
	}

	caPath, ok := getEnv("TLS_CA_CERT")
	if ok {
		withCA = true
	}

	if withCA {
		if withCert && withKey {
			tlsConf = client.WithTLS(caPath, certPath, keyPath)
		} else {
			tlsConf = client.WithTLS(caPath)
		}
	}

	return tlsConf
}

func main() {
	logger := log.New(
		log.WithFormat(log.TextColorLevelFirst),
	)

	grpcLogger, errCh := client.New(
		client.WithAddr("127.0.0.1:9099"),
		client.UnaryRPC(),
		client.WithLogger(
			logger,
		),
		client.WithGRPCOpts(),
		getTLSConf(),
	)
	_, done := grpcLogger.Channels()

	grpcLogger.Log(event.New().Message("hello from client").Build())

	for i := 0; i < 3; i++ {
		grpcLogger.Log(event.New().Level(event.Level_warn).Message(fmt.Sprintf("warning #%v", i)).Build())
		time.Sleep(time.Millisecond * 50)
	}

	for {
		select {
		case err := <-errCh:
			panic(err)
		case <-done:
			return
		}
	}
}
