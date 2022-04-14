package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/zalgonoise/zlog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrNoURI error = errors.New("no URI provided -- supply an environment variable containing the mongodb URI, or set it in the MONGODB_URI environment variable")
)

type Mongo struct {
	env        string
	uri        string
	database   string
	collection string
	db         *mongo.Client
	ctx        context.Context
	cancel     context.CancelFunc
}

func New(envURI, database, collection string) (*Mongo, error) {
	uri, err := checkEnv(envURI)

	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		defer cancel()
		return nil, err
	}

	return &Mongo{
		env:        envURI,
		uri:        uri,
		database:   database,
		collection: collection,
		db:         client,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

func checkEnv(env string) (string, error) {
	if env == "" {
		uri := os.Getenv("MONGODB_URI")

		if uri == "" {
			return "", ErrNoURI
		}

		return uri, nil
	}
	uri := os.Getenv(env)

	if uri == "" {
		return "", ErrNoURI
	}

	return uri, nil
}

func (d *Mongo) Close() error {
	defer d.cancel()
	if err := d.db.Disconnect(d.ctx); err != nil {
		return err
	}
	return nil
}

func (d *Mongo) Create(msg ...*log.LogMessage) error {
	if len(msg) == 0 {
		return nil
	}

	fmt.Println("database: ", d.database, "; collection: ", d.collection)

	var coll = d.db.Database(d.database).Collection(d.collection)
	var msgs []interface{}

	for _, m := range msg {
		var entry = bson.D{
			{Key: "timestamp", Value: m.Time},
			{Key: "service", Value: m.Prefix},
			{Key: "module", Value: m.Sub},
			{Key: "level", Value: m.Level},
			{Key: "message", Value: m.Msg},
			{Key: "metadata", Value: m.Metadata},
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

func (d *Mongo) Write(p []byte) (n int, err error) {
	if d.db == nil && d.env != "" {
		if d.database == "" {
			d.database = "logging"
		}
		if d.collection == "" {
			d.collection = "logs"
		}

		new, err := New(d.env, d.database, d.collection)
		if err != nil {
			return 0, err
		}
		d = new
	}

	var out *log.LogMessage

	// check if it's gob-encoded
	msg, err := log.NewMessage().FromGob(p)
	out = msg

	if err != nil {
		// fall back to JSON
		var msg = &log.LogMessage{}
		jerr := json.Unmarshal(p, msg)
		if jerr != nil {
			return 0, fmt.Errorf("unable to decode input message; gob: %s -- json: %s", err, jerr)
		}
		out = msg
	}

	err = d.Create(out)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// LCMongo struct defines the Logger Config object that adds a Mongo writer to a Logger
type LCMongo struct {
	out io.Writer
	fmt log.LogFormatter
}

// WithMongo function takes in a path to a .db file, and a table name; and returns a LoggerConfig
// so that this type of writer is defined in a Logger
func WithMongo(envURI, database, collection string) log.LoggerConfig {
	db, err := New(envURI, database, collection)
	if err != nil {
		fmt.Printf("failed to open or create database with an error: %s", err)
		os.Exit(1)
	}

	return &LCMongo{
		out: db,
		fmt: log.FormatGob,
	}
}

// Apply method will set the input LoggerBuilder's outputs and format to the LCMongo object's.
func (c *LCMongo) Apply(lb *log.LoggerBuilder) {
	lb.Out = c.out
	lb.Fmt = c.fmt
}
