package main

import (
	"github.com/zalgonoise/zlog/log"
)

func main() {

	// following the fmt.Print() signature, with the default standard logger
	// with the same patterns, using methods for appropriate log levels
	log.Print("this is the simplest approach to entering a log message")
	log.Tracef("and can include formatting: %v %v %s", 3.5, true, "string")
	log.Errorln("which is similar to fmt.Print() method calls")

	// also, by default, log entries can cause runtime to stop
	// such as log.Fatal() -- ends runtime with os.Exit(1)
	// and log.Panic() -- ends runtime with panic()
	// these types of exits can be overriden, but the default logger respects them
	log.Panicf("example of a logger panic event: %v", true)
}
