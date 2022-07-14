package main

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/store/db/mysql"

	"os"
)

const (
	_         string = "MYSQL_USER"     // account for these env variables
	_         string = "MYSQL_PASSWORD" // account for these env variables
	dbAddrEnv string = "MYSQL_HOST"
	dbPortEnv string = "MYSQL_PORT"
	dbNameEnv string = "MYSQL_DATABASE"
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
	db, err := mysql.New(dbAddr+":"+dbPort, dbName)

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
	// create one with the package function
	return log.New(
		mysql.WithMySQL(dbAddr+":"+dbPort, dbName),
	)
}

func main() {
	// load mysql db details from environment variables
	mysqlAddr, ok := getEnv(dbAddrEnv)

	if !ok {
		// zlog's standard logger exit
		log.Fatalf("MySQL database address not provided, from env variable %s", dbAddrEnv)
	}

	mysqlPort, ok := getEnv(dbPortEnv)

	if !ok {
		// zlog's standard logger exit
		log.Fatalf("MySQL database port not provided, from env variable %s", dbPortEnv)
	}

	mysqlDB, ok := getEnv(dbNameEnv)

	if !ok {
		// zlog's standard logger exit
		log.Fatalf("MySQL database name not provided, from env variable %s", dbNameEnv)
	}

	// setup a logger by preparing the DB writer
	loggerOne := setupMethodOne(mysqlAddr, mysqlPort, mysqlDB)

	// write a message to the DB
	loggerOne.Log(
		event.New().Message("log entry #1 written to sqlite db writer #1").Build(),
		event.New().Message("log entry #2 written to sqlite db writer #1").Build(),
		event.New().Message("log entry #3 written to sqlite db writer #1").Build(),
	)

	// setup a logger directly as a DB writer
	loggerTwo := setupMethodTwo(mysqlAddr, mysqlPort, mysqlDB)

	// write a message to the DB
	loggerTwo.Log(
		event.New().Message("log entry #1 written to sqlite db writer #2").Build(),
		event.New().Message("log entry #2 written to sqlite db writer #2").Build(),
		event.New().Message("log entry #3 written to sqlite db writer #2").Build(),
	)
}
