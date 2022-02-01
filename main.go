package main

import "github.com/ZalgoNoise/zlog/log"

func main() {
	log := log.New()

	log.Log(2, "test log")
	log.Log(0, "another test")
	log.Log(5, "panic test")
}
