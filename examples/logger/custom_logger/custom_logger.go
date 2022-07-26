package main

import (
	"bytes"
	"fmt"

	"github.com/zalgonoise/zlog/log"
)

func main() {

	// just calling the New() function will spawn a standard logger
	// that is -- one that prints to os.Stdout with the default configuration
	logger := log.New(
		log.WithPrefix("svc"),
		log.WithSub("mod"),
	)

	logger.Printf("message from custom logger: %v", true)
	logger.Trace("with different methods for different log levels")
	logger.Warnln("for the three fmt.Print() methods you're used to already")

	// you can then customize the logger itself by:
	//   - adding / changing writers (outputs)
	//   - persisting prefix, sub-prefix and structured metadata
	buf := new(bytes.Buffer)
	logger.SetOuts(buf)
	logger.Prefix("service")
	logger.Sub("module")

	logger.Info("message written to a new buffer")
	logger.Trace("with a new prefix and sub-prefix")
	logger.Warnln("that will be persisted through these messages")

	fmt.Print(buf.String())
	logger.SetOuts()
	logger.Prefix("")
	logger.Sub("")

	logger.Info("still works like the default state")

	// and can be reset
}
