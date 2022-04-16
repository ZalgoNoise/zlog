package store

import "io"

// EmptyWriter is an exported, initialized nilWriter
var EmptyWriter io.Writer = nilWritter{}

type nilWritter struct{}

// Write method implements the io.Writer interface, which in this case just returns
// 0, nil -- as any writes to this writer will be discarded and are not perceived as
// errors
func (nilWritter) Write(p []byte) (n int, err error) {
	return 0, nil
}
