package main

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/store/db/postgres"

	"os"
)

const (
	_         string = "POSTGRES_USER"     // account for these env variables
	_         string = "POSTGRES_PASSWORD" // account for these env variables
	dbAddrEnv string = "POSTGRES_HOST"
	dbPortEnv string = "POSTGRES_PORT"
	dbNameEnv string = "POSTGRES_DATABASE"
)

func getEnv(env string) (val string, ok bool) {
	v := os.Getenv(env)

	if v == "" {
		return v, false
	}

	return v, true
}

func setupMethodOne(dbAddr, dbPort, dbName string) log.Logger {
	// create a new DB writer
	db, err := postgres.New(dbAddr, dbPort, dbName)

	if err != nil {
		// zlog's standard logger exit
		log.Fatalf("unexpected error: %v", err)
	}

	// create a new logger with the general DB config
	logger := log.New(
		log.WithDatabase(db),
	)

	return logger
}

func setupMethodTwo(dbAddr, dbPort, dbName string) log.Logger {
	// create one with the package function, that just
	return log.New(
		postgres.WithPostgres(dbAddr, dbPort, dbName),
	)
}

func main() {
	// load postgres db details from environment variables
	postgresAddr, ok := getEnv(dbAddrEnv)

	if !ok {
		// zlog's standard logger exit
		log.Fatalf("Postgres database address not provided, from env variable %s", dbAddrEnv)
	}

	postgresPort, ok := getEnv(dbPortEnv)

	if !ok {
		// zlog's standard logger exit
		log.Fatalf("Postgres database port not provided, from env variable %s", dbPortEnv)
	}

	postgresDB, ok := getEnv(dbNameEnv)

	if !ok {
		// zlog's standard logger exit
		log.Fatalf("Postgres database name not provided, from env variable %s", dbNameEnv)
	}

	// setup a logger by preparing the DB writer
	loggerOne := setupMethodOne(postgresAddr, postgresPort, postgresDB)

	// write a message to the DB
	loggerOne.Log(
		event.New().Message("log entry #1 written to sqlite db writer #1").Build(),
		event.New().Message("log entry #2 written to sqlite db writer #1").Build(),
		event.New().Message("log entry #3 written to sqlite db writer #1").Build(),
	)

	// setup a logger directly as a DB writer
	loggerTwo := setupMethodTwo(postgresAddr, postgresPort, postgresDB)

	// write a message to the DB
	loggerTwo.Log(
		event.New().Message("log entry #1 written to sqlite db writer #2").Build(),
		event.New().Message("log entry #2 written to sqlite db writer #2").Build(),
		event.New().Message("log entry #3 written to sqlite db writer #2").Build(),
	)
}
