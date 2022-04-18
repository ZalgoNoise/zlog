package postgres

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	dbw "github.com/zalgonoise/zlog/store/db"
	model "github.com/zalgonoise/zlog/store/db/message"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	ErrNoEnv error = errors.New("no env variable provided -- ensure that the environment variables for POSTGRES_USER and POSTGRES_PASSWORD are set")
)

type Postgres struct {
	addr     string
	port     string
	database string
	db       *gorm.DB
}

// New function will take in a postgres DB address, port and database name; and create
// a new instance of a Postgres object; returning a pointer to one and an error.
func New(address, port, database string) (sqldb dbw.DBWriter, err error) {
	db, err := initialMigration(address, port, database)

	if err != nil {
		return nil, err
	}

	sqldb = &Postgres{
		addr:     address,
		port:     port,
		database: database,
		db:       db,
	}

	return
}

// Create method will register any number of LogMessages in the Postgres database, returning
// an error
func (d *Postgres) Create(msg ...*event.Event) error {
	if len(msg) == 0 {
		return nil
	}

	var msgs []*model.LogMessage

	for _, m := range msg {
		var entry = &model.LogMessage{}

		if err := entry.From(m); err != nil {
			return err
		}
		msgs = append(msgs, entry)
	}

	d.db.Create(msgs)
	return nil
}

// Write method implements the io.Writer interface, for Postgres DBs to be used with Logger,
// as its writer.
//
// This implementation relies on JSON or gob-encoding the messages, so they are passed onto
// this writer. Then, it is unmarshalled into a message object which is sent in an Insert()
// call.
func (s *Postgres) Write(p []byte) (n int, err error) {
	if s.db == nil && s.addr != "" {
		if s.port == "" {
			s.port = "5432"
		}

		if s.database == "" {
			s.database = "logs"
		}

		new, err := New(s.addr, s.port, s.database)
		if err != nil {
			return 0, err
		}
		s = new.(*Postgres)
	}

	msg, err := event.Decode(p)
	if err != nil {
		return 0, err
	}

	err = s.Create(msg)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// Close method is implemented for compatibility with the Database interface.
//
// While this ORM doesn't force users to close the connection, MongoDB does, and the
// method should be available for use
func (d *Postgres) Close() error { return nil }

func initialMigration(address, port, database string) (*gorm.DB, error) {
	// "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=UTC"
	var uri = strings.Builder{}

	uri.WriteString("host=")
	uri.WriteString(address)
	uri.WriteString(" user=")
	uri.WriteString(os.Getenv("POSTGRES_USER"))
	uri.WriteString(" password=")
	uri.WriteString(os.Getenv("POSTGRES_PASSWORD"))
	uri.WriteString(" dbname=")
	uri.WriteString(database)
	uri.WriteString(" port=")
	uri.WriteString(port)
	uri.WriteString(" sslmode=disable TimeZone=UTC")

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  uri.String(),
		PreferSimpleProtocol: true,
	}), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	// Migrate the schema
	err = db.AutoMigrate(&model.LogMessage{})

	if err != nil {
		return nil, err
	}

	return db, nil
}

// WithPostgres function takes in an address and port to a Postgres server, and a database name;
// and returns a LoggerConfig so that this type of writer is defined in a Logger
func WithPostgres(addr, port, database string) log.LoggerConfig {
	db, err := New(addr, port, database)
	if err != nil {
		fmt.Printf("failed to open or create database with an error: %s", err)
		os.Exit(1)
	}

	//TODO(zalgonoise): benchmark this decision -- confirm if gob is more performant,
	// considering that JSON will (usually) have less bytes per (small) message
	return &log.LCDatabase{
		Out: db,
		Fmt: log.FormatJSON,
	}
}
