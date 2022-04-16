package db

import (
	"encoding/json"
	"io"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/zalgonoise/zlog/log"
)

type Database interface {
	io.Writer
	Create(msg ...*log.LogMessage) error
	Close() error
}

type LogMessage struct {
	gorm.Model
	Time     time.Time
	Prefix   string
	Sub      string
	Level    string
	Msg      string
	Metadata string
}

func (m *LogMessage) From(msg *log.LogMessage) error {
	m.Time = msg.Time
	m.Prefix = msg.Prefix
	m.Sub = msg.Sub
	m.Level = msg.Level
	m.Msg = msg.Msg

	meta, err := json.Marshal(msg.Metadata)

	if err != nil {
		return err
	}

	if metafmt := string(meta); metafmt != "{}" {
		m.Metadata = metafmt
	}

	return nil
}
