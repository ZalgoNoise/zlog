package main

import (
	"github.com/zalgonoise/zlog/log"
)

func main() {

	// following the fmt.Print() signature, with the default standard logger
	log.Print("this is the simplest approach to entering a log message")
	log.Printf("and can include formatting: %v %v %s", 3.5, true, "string")
	log.Println("which is similar to fmt.Print() method calls")

	// with the same patterns, using methods for appropriate log levels
	log.Tracef("this is a trace message: %v", 5000)
	log.Info("this is an info message")
	log.Errorln("this is an error message")

	// also, by default, log entries can cause runtime to stop
	// such as log.Fatal() -- ends runtime with os.Exit(1)
	// and log.Panic() -- ends runtime with panic()
	// these types of exits can be overriden, but the default logger respects them
	log.Panicf("example of a logger panic event: %v", true)
}
