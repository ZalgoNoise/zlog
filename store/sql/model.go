package sql

import (
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/zalgonoise/zlog/log"
)

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

	meta, err := json.Marshal(msg.Msg)

	if err != nil {
		return err
	}

	m.Metadata = string(meta)

	return nil
}
