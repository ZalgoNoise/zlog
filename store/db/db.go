package db

import (
	"io"
)

type DBWriter interface {
	io.WriteCloser
}
