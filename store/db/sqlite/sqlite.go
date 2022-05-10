package sqlite

import (
	"fmt"
	"io"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
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
// instance of a SQLite3 object; returning an io.WriteCloser and an error.
func New(path string) (sqldb io.WriteCloser, err error) {
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

// Create method will register any number of event.Event in the SQLite database, returning
// an error
func (o *SQLite) Create(msg ...*event.Event) error {
	if len(msg) == 0 {
		return nil
	}

	var msgs []*model.Event

	for _, m := range msg {
		if m == nil {
			continue
		}

		var entry = &model.Event{}

		if err := entry.From(m); err != nil {
			return err
		}
		msgs = append(msgs, entry)
	}

	if len(msgs) == 0 {
		return nil
	}

	o.db.Create(msgs)
	return nil
}

// Write method implements the io.Writer interface, for SQLite DBs to be used with Logger,
// as its writer.
//
// The input message is expected to be a protobuf-marshalled event.Event, which is decoded
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
	err = db.AutoMigrate(&model.Event{})

	return db, err
}

// WithSQLite function takes in a path to a .db file, and a table name; and returns a LoggerConfig
// so that this type of writer is defined in a Logger
func WithSQLite(path string) log.LoggerConfig {
	db, err := New(path)
	if err != nil {
		panic(fmt.Errorf("failed to create logger config -- database creation failed: %w", err))
	}

	return &log.LCDatabase{
		Out: db,
		Fmt: log.FormatProtobuf,
	}
}
