package model

import (
	"encoding/json"
	"testing"

	"github.com/zalgonoise/zlog/log/event"
)

func TestModelFrom(t *testing.T) {
	module := "DB Model"
	funcname := "From()"

	type test struct {
		name  string
		event *event.Event
		model *Event
		ok    bool
	}

	var testMessage = "null"

	var tests = []test{
		{
			name:  "default message",
			event: event.New().Message("testing").Build(),
			model: &Event{
				Prefix:   "log",
				Sub:      "",
				Level:    "info",
				Msg:      "testing",
				Metadata: "",
			},
			ok: true,
		},
		{
			name:  "default message w/ meta",
			event: event.New().Message("testing").Metadata(event.Field{"a": true}).Build(),
			model: &Event{
				Prefix:   "log",
				Sub:      "",
				Level:    "info",
				Msg:      "testing",
				Metadata: `{"a":true}`,
			},
			ok: true,
		},
		{
			name:  "invalid message",
			event: &event.Event{},
			model: nil,
		},
		{
			name:  "most basic message",
			event: &event.Event{Msg: &testMessage},
			model: &Event{
				Prefix: "log",
				Sub:    "",
				Level:  "info",
				Msg:    "null",
			},
			ok: true,
		},
	}

	var verify = func(idx int, test test, e *Event) {

		err := e.From(test.event)

		if test.ok && err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- error converting pb message: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		} else if !test.ok {
			return
		}

		if e.Prefix != test.event.GetPrefix() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- prefix mismatch: wanted %s, got %s -- action: %s",
				idx,
				module,
				funcname,
				test.event.GetPrefix(),
				e.Prefix,
				test.name,
			)
			return
		}
		if e.Sub != test.event.GetSub() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- subprefix mismatch: wanted %s, got %s -- action: %s",
				idx,
				module,
				funcname,
				test.event.GetSub(),
				e.Sub,
				test.name,
			)
			return
		}
		if e.Level != test.event.GetLevel().String() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- level mismatch: wanted %s, got %s -- action: %s",
				idx,
				module,
				funcname,
				test.event.GetLevel().String(),
				e.Level,
				test.name,
			)
			return
		}
		if e.Msg != test.event.GetMsg() {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- message mismatch: wanted %s, got %s -- action: %s",
				idx,
				module,
				funcname,
				test.event.GetMsg(),
				e.Msg,
				test.name,
			)
			return
		}

		meta, err := json.Marshal(test.event.GetMeta().AsMap())
		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- error converting metadata to JSON: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		if metafmt := string(meta); metafmt == "{}" {
			meta = []byte{}
		}

		if e.Metadata != string(meta) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] -- metadata mismatch: wanted %s, got %s -- action: %s",
				idx,
				module,
				funcname,
				string(meta),
				e.Metadata,
				test.name,
			)
			return
		}

		t.Logf(
			"#%v -- PASSED -- [%s] [%s] -- action: %s",
			idx,
			module,
			funcname,
			test.name,
		)

	}

	for idx, test := range tests {
		var e = &Event{}

		verify(idx, test, e)
	}

}
