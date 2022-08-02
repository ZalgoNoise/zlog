package main

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/store/db/sqlite"

	"os"
)

const (
	dbPathEnv string = "SQLITE_PATH"
)

func getEnv(env string) (val string, ok bool) {
	v := os.Getenv(env)

	if v == "" {
		return v, false
	}

	return v, true
}

func setupMethodOne(dbPath string) log.Logger {
	// create a new DB writer
	db, err := sqlite.New(dbPath)

	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}

	// create a new logger with the general DB config
	logger := log.New(
		log.WithDatabase(db),
	)

	return logger
}

func setupMethodTwo(dbPath string) log.Logger {
	// create one with the package function
	return log.New(
		sqlite.WithSQLite(dbPath),
	)
}

func main() {
	// load sqlite db path from environment variable
	sqlitePath, ok := getEnv(dbPathEnv)
	if !ok {
		log.Fatalf("SQLite database path not provided, from env variable %s", dbPathEnv)
	}

	// setup a logger by preparing the DB writer
	loggerOne := setupMethodOne(sqlitePath)

	// write a message to the DB
	loggerOne.Log(
		event.New().Message("log entry #1 written to sqlite db writer #1").Build(),
		event.New().Message("log entry #2 written to sqlite db writer #1").Build(),
		event.New().Message("log entry #3 written to sqlite db writer #1").Build(),
	)

	// setup a logger directly as a DB writer
	loggerTwo := setupMethodTwo(sqlitePath)

	// write a message to the DB
	loggerTwo.Log(
		event.New().Message("log entry #1 written to sqlite db writer #2").Build(),
		event.New().Message("log entry #2 written to sqlite db writer #2").Build(),
		event.New().Message("log entry #3 written to sqlite db writer #2").Build(),
	)
}
