package sqlite

import (
	"fmt"
	"io"
	"os"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	dbw "github.com/zalgonoise/zlog/store/db"
	model "github.com/zalgonoise/zlog/store/db/message"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SQLite struct is a wrapper for a SQLite database to be used as a Log Writer
type SQLite struct {
	path string
	db   *gorm.DB
}

// New function will take in a path to a .db file; and create a new
// instance of a SQLite3 object; returning a pointer to one and an error.
func New(path string) (sqldb dbw.DBWriter, err error) {
	db, err := initialMigration(path)

	if err != nil {
		return nil, err
	}

	sqldb = &SQLite{
		path: path,
		db:   db,
	}

	return
}

// Create method will register any number of LogMessages in the SQLite database, returning
// an error
func (o *SQLite) Create(msg ...*event.Event) error {
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

	o.db.Create(msgs)
	return nil
}

// Write method implements the io.Writer interface, for SQLite DBs to be used with Logger,
// as its writer.
//
// This implementation relies on JSON or gob-encoding the messages, so they are passed onto
// this writer. Then, it is unmarshalled into a message object which is sent in an Insert()
// call.
func (s *SQLite) Write(p []byte) (n int, err error) {
	if s.db == nil && s.path != "" {
		new, err := New(s.path)
		if err != nil {
			return 0, err
		}
		s = new.(*SQLite)
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
func (d *SQLite) Close() error { return nil }

func initialMigration(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	db.AutoMigrate(&model.LogMessage{})
	return db, nil
}

// LCSQLite struct defines the Logger Config object that adds a SQLite writer to a Logger
type LCSQLite struct {
	out io.Writer
	fmt log.LogFormatter
}

// WithSQLite function takes in a path to a .db file, and a table name; and returns a LoggerConfig
// so that this type of writer is defined in a Logger
func WithSQLite(path string) log.LoggerConfig {
	db, err := New(path)
	if err != nil {
		fmt.Printf("failed to open or create database with an error: %s", err)
		os.Exit(1)
	}

	//TODO(zalgonoise): benchmark this decision -- confirm if gob is more performant,
	// considering that JSON will (usually) have less bytes per (small) message
	return &LCSQLite{
		out: db,
		fmt: log.FormatGob,
	}
}

// Apply method will set the input LoggerBuilder's outputs and format to the LCSQLite object's.
func (c *LCSQLite) Apply(lb *log.LoggerBuilder) {
	lb.Out = c.out
	lb.Fmt = c.fmt
}
