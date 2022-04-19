package main

import (
	"os"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/store/db/sqlite"
)

var msgs = []*event.Event{
	event.New().
		Prefix("testing").
		Sub("fake").
		Message("log2db").
		Metadata(event.Field{
			"a": true,
			"b": 1,
			"c": "data",
			"d": map[string]interface{}{
				"e": "inner",
				"f": []string{
					"g", "h", "i",
				},
			},
		}).
		Build(),
	event.New().
		Prefix("test2").
		Sub("faker").
		Message("log2db").
		Metadata(event.Field{
			"a": true,
			"b": 1,
			"c": "data",
			"d": map[string]interface{}{
				"e": "inner",
				"f": []string{
					"g", "h", "i",
				},
			},
		}).
		Build(),
	event.New().
		Prefix("tester").
		Sub("faked with space").
		Message("log2db").
		Metadata(event.Field{
			"a": true,
			"b": 1,
			"c": "data",
			"d": map[string]interface{}{
				"e": "inner",
				"f": []string{
					"g", "h", "i",
				},
			},
		}).
		Build(),
}

func main() {

	homedir := os.Getenv("HOME")
	db, err := sqlite.New(homedir + "/tmp/db/log.db")
	// db, err := sqlite.Load(homedir + "/tmp/db/log.db")

	if err != nil {
		panic(err)
	}

	logger := log.New(
		log.WithDatabase(db),
		log.SkipExit,
	)

	for _, m := range msgs {
		_, err := logger.Output(m)
		if err != nil {
			panic(err)
		}
	}

}
