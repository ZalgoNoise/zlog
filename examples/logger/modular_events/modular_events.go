package main

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

func main() {

	// Events can be created and customized with a builder pattern,
	// where each element is defined with a chained method until the
	// Build() method is called.
	//
	// This last method will apply the timestamp to the event and any
	// defaults for missing (required) fields.
	log.Log(
		event.New().
			Prefix("module").
			Sub("service").
			Level(event.Level_warn).
			Metadata(event.Field{
				"data": true,
			}).
			Build(),
		event.New().
			Prefix("mod").
			Sub("svc").
			Level(event.Level_debug).
			Metadata(event.Field{
				"debug": "something something",
			}).
			Build(),
	)
}
