package model

import (
	"encoding/json"
	"time"

	"github.com/zalgonoise/zlog/log/event"
	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	Time     time.Time
	Prefix   string
	Sub      string
	Level    string
	Msg      string
	Metadata string
}

func (m *Event) From(msg *event.Event) error {
	m.Time = msg.GetTime().AsTime()
	m.Prefix = msg.GetPrefix()
	m.Sub = msg.GetSub()
	m.Level = msg.GetLevel().String()
	m.Msg = msg.GetMsg()

	meta, err := json.Marshal(msg.Meta.AsMap())

	if err != nil {
		return err
	}

	if metafmt := string(meta); metafmt != "{}" {
		m.Metadata = metafmt
	}

	return nil
}
