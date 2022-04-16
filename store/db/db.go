package db

import (
	"io"

	"github.com/zalgonoise/zlog/log"
)

type DBWriter interface {
	io.Writer
	Create(msg ...*log.LogMessage) error
	Close() error
}
