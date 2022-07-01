package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/zalgonoise/zlog/grpc/address"
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
	var buf = &bytes.Buffer{}

	stdLogger := log.New(
		log.WithOut(buf),
		log.SkipExit,
		log.WithFormat(log.TextColor),
	)

	grpcSvcLogger := log.New(
		log.WithFormat(log.TextColorLevelFirst),
		log.SkipExit,
	)

	grpcLogger, errCh := client.New(
		client.WithAddr("127.0.0.1:9299"),
		client.WithLogger(grpcSvcLogger),
		client.UnaryRPC(),
		getTLSConf(), // loaded from provided env variables
	)

	newAddr := address.New("127.0.0.1:9399")

	multiLogger := log.MultiLogger(
		grpcLogger,
		stdLogger,
	)
	multiLogger.Log(
		event.New().Message("it works!").Build(),
	)
	multiLogger.Log(
		event.New().Message("it works!").Build(),
	)
	multiLogger.Log(
		event.New().Message("it works!").Build(),
	)

	time.Sleep(time.Millisecond * 100)

	multiLogger.AddOuts(os.Stdout, newAddr)
	multiLogger.Log(
		event.New().Message("it works!").Build(),
	)

	fmt.Println(buf.String())

	for {
		err := <-errCh
		panic(err)
	}

}
