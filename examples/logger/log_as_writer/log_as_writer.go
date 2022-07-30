package main

import (
	"fmt"
	"os"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

func main() {

	logger := log.New()

	n, err := logger.Write([]byte("Hello, world!"))
	if err != nil {
		fmt.Println("errored: ", err)
		os.Exit(1)
	}

	fmt.Printf("\n---\nn: %v, err: %v\n---\n", n, err)

	n, err = logger.Write(event.New().Message("Hi, world!").Build().Encode())
	if err != nil {
		fmt.Println("errored: ", err)
		os.Exit(1)
	}

	fmt.Printf("\n---\nn: %v, err: %v\n---\n", n, err)
}
