package main

import (
	"os"

	"github.com/ZalgoNoise/zlog/log"
)

func main() {
	log := log.New(os.Stdout, "test-logs", 1)

	data := []int{
		2, 3, 5,
	}

	log.Infoln("test log")
	log.Debugf("%v\n", len(data))
	log.Warnln("big warning")
	log.Panicln("i am out")
}
