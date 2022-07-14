package main

import (
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"github.com/zalgonoise/zlog/store/db/mongo"

	"os"
)

const (
	_         string = "MONGO_USER"     // account for these env variables
	_         string = "MONGO_PASSWORD" // account for these env variables
	dbAddrEnv string = "MONGO_HOST"
	dbPortEnv string = "MONGO_PORT"
	dbNameEnv string = "MONGO_DATABASE"
	dbCollEnv string = "MONGO_COLLECTION"
)

func getEnv(env string) (val string, ok bool) {
	v := os.Getenv(env)

	if v == "" {
		return v, false
	}

	return v, true
}

func setupMethodOne(dbAddr, dbPort, dbName, dbColl string) (log.Logger, func()) {
	// create a new DB writer
	db, err := mongo.New(dbAddr+":"+dbPort, dbName, dbColl)

	if err != nil {
		// zlog's standard logger exit
		log.Fatalf("unexpected error: %v", err)
	}

	// create a new logger with the general DB config
	logger := log.New(
		log.WithDatabase(db),
	)

	return logger, func() {
		db.Close()
	}
}

func setupMethodTwo(dbAddr, dbPort, dbName, dbColl string) log.Logger {
	// create one with the package function
	return log.New(
		mongo.WithMongo(dbAddr+":"+dbPort, dbName, dbColl),
	)
}

func main() {
	// load mongo db details from environment variables
	mongoAddr, ok := getEnv(dbAddrEnv)

	if !ok {
		// zlog's standard logger exit
		log.Fatalf("Mongo database address not provided, from env variable %s", dbAddrEnv)
	}

	mongoPort, ok := getEnv(dbPortEnv)

	if !ok {
		// zlog's standard logger exit
		log.Fatalf("Mongo database port not provided, from env variable %s", dbPortEnv)
	}

	mongoDB, ok := getEnv(dbNameEnv)

	if !ok {
		// zlog's standard logger exit
		log.Fatalf("Mongo database name not provided, from env variable %s", dbNameEnv)
	}

	mongoColl, ok := getEnv(dbCollEnv)

	if !ok {
		// zlog's standard logger exit
		log.Fatalf("Mongo collection name not provided, from env variable %s", dbCollEnv)
	}

	// setup a logger by preparing the DB writer
	loggerOne, close := setupMethodOne(mongoAddr, mongoPort, mongoDB, mongoColl)
	defer close()

	// write a message to the DB
	n, err := loggerOne.Output(
		event.New().Message("log entry written to sqlite db writer #1").Build(),
	)

	if err != nil {
		// zlog's standard logger exit
		log.Fatalf("unexpected error: %v", err)
	}

	if n == 0 {
		// zlog's standard logger exit
		log.Fatalf("zero bytes written")
	}

	log.Info("event written to DB writer #1 successfully")

	// setup a logger directly as a DB writer
	loggerTwo := setupMethodTwo(mongoAddr, mongoPort, mongoDB, mongoColl)

	// write a message to the DB
	n, err = loggerTwo.Output(
		event.New().Message("log entry written to sqlite db writer #2").Build(),
	)

	if err != nil {
		// zlog's standard logger exit
		log.Fatalf("unexpected error: %v", err)
	}

	if n == 0 {
		// zlog's standard logger exit
		log.Fatalf("zero bytes written")
	}

	log.Info("event written to DB writer #2 successfully")

}
