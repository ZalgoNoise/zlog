package main

import (
	"encoding/json"
	"fmt"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/store/db/mongo"
)

func main() {

	var msgs = []*event.Event{
		event.New().Message("basic").Build(),
		event.New().Prefix("test").Sub("tester").Level(event.Level_warn).Message("longer, more complete message").Metadata(event.Field{"a": 1}).Build(),
		event.New().Prefix("super").Sub("large").Level(event.Level_warn).Message("even longer, more complete message").Metadata(event.Field{
			"a": 1,
			"b": "two",
			"c": true,
			"d": []int{1, 2, 3},
			"e": []string{"a", "b", "c"},
			"f": []bool{true, true, false},
			"g": map[string]interface{}{
				"a": "nested",
				"b": 1,
				"c": true,
				"d": map[string]interface{}{
					"further_nesting": []string{"yep", "yes", "yea"},
				},
			},
			"h": []map[string]interface{}{
				{
					"a": 1,
				},
				{
					"b": "two",
				},
				{
					"c": true,
				},
			},
		}).CallStack(true).Build(),
	}

	var bufs [][]byte

	fmt.Println("first test -- enc/dec")

	for _, m := range msgs {
		b := m.Encode()
		fmt.Println(b, "\n", len(b))
		bufs = append(bufs, b)
	}

	for _, m := range bufs {
		e, err := event.Decode(m)
		if err != nil {
			panic(err)
		}

		b, err := json.Marshal(e)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	}

	db, err := mongo.New("172.16.0.10:27017", "logger", "zlog")
	if err != nil {
		panic(err)
	}

	logger := log.New(
		// log.WithFormat(log.TextColorLevelFirstSpaced),
		log.WithDatabase(db),
		log.SkipExit,
	)

	for _, m := range msgs {
		b, err := logger.Output(m)
		if err != nil {
			panic(err)
		}

		fmt.Printf("wrote %v bytes\n", b)
	}

	// logger := log.New(
	// 	log.WithDatabase(db),
	// )

	// for i := 0; i < 1000; i++ {
	// 	var msgs = []*event.EventBuilder{
	// 		event.New().Message("hi").Prefix("complex").Sub("test").Metadata(event.Field{
	// 			"a": 0,
	// 			"b": "a",
	// 			"c": []string{
	// 				"a", "b", "c",
	// 			},
	// 			"d": []int{
	// 				0, 1, 2,
	// 			},
	// 			"e": true,
	// 			"f": []bool{true, false, true},
	// 			"g": map[string]interface{}{
	// 				"a": 0,
	// 				"b": "a",
	// 				"c": []string{
	// 					"a", "b", "c",
	// 				},
	// 			},
	// 			"h": event.Field{
	// 				"a": 0,
	// 				"b": "a",
	// 				"c": []string{
	// 					"a", "b", "c",
	// 				},
	// 			},
	// 		}),
	// 		event.New().Message("hello"),
	// 		event.New().Message("hey"),
	// 	}

	// 	for _, m := range msgs {
	// 		out := m.Build()
	// 		_, err := logger.Output(out)
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 	}
	// }

}
