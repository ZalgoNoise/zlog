package main

import (
	"time"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/log/logch"
)

func main() {

	// create a new, basic logger directly as a channeled logger
	chLogger := logch.New(log.New())

	// send messages using its Log() method directly; like the simple one:
	chLogger.Log(
		event.New().Message("one").Build(),
		event.New().Message("two").Build(),
		event.New().Message("three").Build(),
	)

	// or, call its Channels() method to work with the channels directly:
	msgCh, done := chLogger.Channels()

	// send the messages in a separate goroutine, then close the logger
	go func() {
		msgCh <- event.New().Message("four").Build()
		msgCh <- event.New().Message("five").Build()
		msgCh <- event.New().Message("six").Build()

		// give it a millisecond to allow the last message to be printed
		time.Sleep(time.Millisecond)

		// send done signal to stop the process
		done <- struct{}{}
	}()

	// keep-alive until the done signal is received
	for {
		select {
		case <-done:
			return
		}
	}

}
