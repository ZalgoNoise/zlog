package mongo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrNoURI error = errors.New("no URI provided -- supply an environment variable containing the mongodb URI, or set it in the MONGODB_URI environment variable")
)

// Mongo struct is a wrapper for a MongoDB database to be used as a Log Writer
type Mongo struct {
	uri        string
	addr       string
	database   string
	collection string
	db         *mongo.Client
}

// New function will take in a MongoDB address, database and collection names; and
// create a new instance of a Mongo object; returning an io.WriteCloser and an error.
func New(address, database, collection string) (io.WriteCloser, error) {
	// getting the target URI
	//   mongodb://user:password@127.0.0.1:27017/?maxPoolSize=20&w=majority
	var uri = strings.Builder{}

	uri.WriteString("mongodb://")
	uri.WriteString(os.Getenv("MONGO_USER"))
	uri.WriteString(":")
	uri.WriteString(os.Getenv("MONGO_PASSWORD"))
	uri.WriteString("@")
	uri.WriteString(address)
	uri.WriteString("/?maxPoolSize=20&w=majority")

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri.String()))
	if err != nil {
		return nil, err
	}

	return &Mongo{
		uri:        uri.String(),
		addr:       address,
		database:   database,
		collection: collection,
		db:         client,
	}, nil
}

// Close method is used to terminate the live connection to the MongoDB instance.
func (d *Mongo) Close() error {
	if err := d.db.Disconnect(context.Background()); err != nil {
		return err
	}
	return nil
}

// Create method will register any number of event.Event in the Postgres database, returning
// an error
func (d *Mongo) Create(msg ...*event.Event) error {
	if len(msg) == 0 {
		return nil
	}

	var coll = d.db.Database(d.database).Collection(d.collection)
	var msgs []interface{}

	for _, m := range msg {
		var entry = bson.D{
			{Key: "timestamp", Value: m.GetTime().AsTime()},
			{Key: "service", Value: m.GetPrefix()},
			{Key: "module", Value: m.GetSub()},
			{Key: "level", Value: m.GetLevel()},
			{Key: "message", Value: m.GetMsg()},
			{Key: "metadata", Value: m.Meta.AsMap()},
		}
		msgs = append(msgs, entry)
	}

	if len(msgs) == 1 {
		_, err := coll.InsertOne(context.Background(), msgs[0])
		if err != nil {
			return err
		}
		return nil
	}

	_, err := coll.InsertMany(context.Background(), msgs)
	if err != nil {
		return err
	}
	return nil
}

// Write method implements the io.Writer interface, for Postgres DBs to be used with Logger,
// as its writer.
//
// The input message is expected to be a protobuf-marshalled event.Event, which is decoded
func (d *Mongo) Write(p []byte) (n int, err error) {
	if d.db == nil && d.addr != "" {
		if d.database == "" {
			d.database = "logging"
		}
		if d.collection == "" {
			d.collection = "logs"
		}

		new, err := New(d.addr, d.database, d.collection)
		if err != nil {
			return 0, err
		}
		d = new.(*Mongo)
	}

	msg, err := event.Decode(p)
	if err != nil {
		return 0, err
	}

	err = d.Create(msg)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// WithMongo function takes in the address to the mongo server, and a database and collection name;
// and returns a LoggerConfig so that this type of writer is defined in a Logger
func WithMongo(addr, database, collection string) log.LoggerConfig {
	db, err := New(addr, database, collection)
	if err != nil {
		fmt.Printf("failed to open or create database with an error: %s", err)
		os.Exit(1)
	}

	return &log.LCDatabase{
		Out: db,
		Fmt: log.FormatProtobuf,
	}
}
