package model

import (
	"encoding/json"
	"time"

	"github.com/zalgonoise/zlog/log/event"
	"gorm.io/gorm"
)

// Event struct defines the general structure of a log message, to be used
// with gorm.
//
// This model is used by different databases which are accessed with gorm.
type Event struct {
	gorm.Model
	Time     time.Time
	Prefix   string
	Sub      string
	Level    string
	Msg      string
	Metadata string
}

// From method will take in an event.Event and convert it into a (DB model) Event,
// returning any errors if existing.
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
