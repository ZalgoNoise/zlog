package xml

import (
	"encoding/xml"
	"time"

	"github.com/zalgonoise/zlog/log/event"
)

// FmtXML struct describes the different manipulations and processing that a XML LogFormatter
// can apply to a LogMessage
type FmtXML struct{}

type logMessage struct {
	Time     time.Time `xml:"timestamp,omitempty"`
	Prefix   string    `xml:"service,omitempty"`
	Sub      string    `xml:"module,omitempty"`
	Level    string    `xml:"level,omitempty"`
	Msg      string    `xml:"message,omitempty"`
	Metadata []Field   `xml:"metadata,omitempty"`
}

// Format method will take in a pointer to a LogMessage; and returns a buffer and an error.
//
// This method will process the input LogMessage and marshal it according to this LogFormatter
func (f *FmtXML) Format(log *event.Event) (buf []byte, err error) {
	// remove trailing newline on XML format
	if log.Msg[len(log.Msg)-1] == 10 {
		log.Msg = log.Msg[:len(log.Msg)-1]
	}

	xmlMsg := &logMessage{
		Time:     log.Time,
		Prefix:   log.Prefix,
		Sub:      log.Sub,
		Level:    log.Level,
		Msg:      log.Msg,
		Metadata: Mappify(log.Metadata),
	}

	return xml.Marshal(xmlMsg)

}

type Field struct {
	Key string      `xml:"key,omitempty"`
	Val interface{} `xml:"value,omitempty"`
}

func Mappify(data map[string]interface{}) []Field {
	var fields []Field

	for k, v := range data {
		switch value := v.(type) {
		case []map[string]interface{}:
			f := []Field{}

			for _, im := range value {
				ifield := Field{}
				for ik, iv := range im {
					ifield.Key = ik
					ifield.Val = iv
				}

				f = append(f, ifield)
			}

			fields = append(fields, Field{
				Key: k,
				Val: f,
			})
		case []event.Field:
			f := []Field{}

			for _, im := range value {
				ifield := Field{}
				for ik, iv := range im.ToMap() {
					ifield.Key = ik
					ifield.Val = iv
				}

				f = append(f, ifield)
			}

			fields = append(fields, Field{
				Key: k,
				Val: f,
			})
		case map[string]interface{}:
			fields = append(fields, Field{
				Key: k,
				Val: Mappify(value),
			})
		case event.Field:
			fields = append(fields, Field{
				Key: k,
				Val: Mappify(value.ToMap()),
			})
		default:
			fields = append(fields, Field{
				Key: k,
				Val: value,
			})
		}
	}

	return fields
}
