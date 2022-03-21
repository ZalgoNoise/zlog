package store

import "io"

var EmptyWriter io.Writer = nilWritter{}

type nilWritter struct{}

func (nilWritter) Write(p []byte) (n int, err error) {
	return 0, nil
}
