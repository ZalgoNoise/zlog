package main

import (
	"fmt"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

func main() {

	logger := log.New()

	n, err := logger.Write([]byte("Hello, world!"))

	fmt.Printf("\n---\nn: %v, err: %v\n---\n", n, err)

	n, err = logger.Write(event.New().Message("Hi, world!").Build().Encode())

	fmt.Printf("\n---\nn: %v, err: %v\n---\n", n, err)
}
