package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zalgonoise/zlog/log/event"
	"gorm.io/gorm"
)

var (
	ErrInvalidEvent error = errors.New("invalid event -- missing message body")
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
	now := time.Now()

	// get values
	timestamp := msg.GetTime().AsTime()
	prefix := msg.GetPrefix()
	sub := msg.GetSub()
	level := msg.GetLevel().String()
	body := msg.GetMsg()
	meta, _ := json.Marshal(msg.GetMeta().AsMap())

	// check defaults
	if body == "" {
		return ErrInvalidEvent
	}

	if timestamp.Unix() == 0 {
		timestamp = now
	}

	m.Time = timestamp
	m.Prefix = prefix
	m.Sub = sub
	m.Level = level
	m.Msg = body

	if metafmt := string(meta); metafmt != "{}" {
		m.Metadata = metafmt
	}

	fmt.Println(m)
	return nil
}
