package xml

import (
	"encoding/xml"
	"time"

	e "github.com/zalgonoise/zlog/log/event"
)

// FmtXML struct describes the different manipulations and processing that a XML LogFormatter
// can apply to an event.Event
type FmtXML struct{}

type entry struct {
	Time     time.Time `xml:"timestamp,omitempty"`
	Prefix   string    `xml:"service,omitempty"`
	Sub      string    `xml:"module,omitempty"`
	Level    string    `xml:"level,omitempty"`
	Msg      string    `xml:"message,omitempty"`
	Metadata []Field   `xml:"metadata,omitempty"`
}

// Format method will take in a pointer to an event.Event; and returns a buffer and an error.
//
// This method will process the input event.Event and marshal it according to this LogFormatter
func (f *FmtXML) Format(log *e.Event) (buf []byte, err error) {
	// remove trailing newline on XML format
	if log.GetMsg()[len(log.GetMsg())-1] == 10 {
		*log.Msg = log.GetMsg()[:len(log.GetMsg())-1]
	}

	xmlMsg := &entry{
		Time:     log.Time.AsTime(),
		Prefix:   log.GetPrefix(),
		Sub:      log.GetSub(),
		Level:    log.GetLevel().String(),
		Msg:      log.GetMsg(),
		Metadata: Mappify(log.GetMeta().AsMap()),
	}

	return xml.Marshal(xmlMsg)

}

// Field type designates how a metadata element should be displayed, in XML
//
// As such, each mapped item in an event.Event's metadata will be converted to
// an object containing key / value elements.
type Field struct {
	Key string      `xml:"key,omitempty"`
	Val interface{} `xml:"value,omitempty"`
}

// Mappify function will take in a metadata map[string]interface{}, and convert it
// into a slice of (XML) Fields.
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
		case []e.Field:
			f := []Field{}

			for _, im := range value {
				ifield := Field{}
				for ik, iv := range im.AsMap() {
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
		case e.Field:
			fields = append(fields, Field{
				Key: k,
				Val: Mappify(value.AsMap()),
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
