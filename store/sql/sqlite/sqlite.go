package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zalgonoise/zlog/log"
)

const (
	dbCreateTable string = `CREATE TABLE IF NOT EXISTS %s (id INTEGER PRIMARY KEY, time TEXT KEY, level TEXT, prefix TEXT, sub TEXT, message TEXT, metadata TEXT);`
	dbInsertValue string = `INSERT INTO %s (time, level, prefix, sub, message, metadata) VALUES (?, ?, ?, ?, ?, ?);`
)

var (
	ErrDBNotSetUp error = errors.New("database is not setup -- nil pointer")
)

// SQLite3 struct is a wrapper for a SQLite database to be used as a Log Writer
type SQLite3 struct {
	path  string
	table string
	db    *sql.DB
}

// New function will take in a path to a .db file, and a table name; and create a new
// instance of a SQLite3 object; returning a pointer to one and an error.
func New(path, table string) (*SQLite3, error) {

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	op, err := db.Prepare(fmt.Sprintf(dbCreateTable, table))
	if err != nil {
		return nil, err
	}

	defer op.Close()

	_, err = op.Exec()
	if err != nil {
		return nil, err
	}

	return &SQLite3{
		path:  path,
		table: table,
		db:    db,
	}, nil

}

// Load function will take in a path to a .db file, and a table name; and create a new
// instance of a SQLite3 object based on an existing database; returning a pointer to
// one and an error.
func Load(path, table string) (*SQLite3, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	err = verify(db, table)
	if err != nil {
		return nil, err
	}

	return &SQLite3{
		path:  path,
		table: table,
		db:    db,
	}, nil
}

// Insert method will register any number of LogMessages in the SQLite database, returning
// an error
func (s *SQLite3) Insert(msg ...*log.LogMessage) error {

	if len(msg) == 0 {
		return nil
	}

	for _, m := range msg {
		// unix
		// t := strconv.FormatInt(m.Time.Unix(), 10)
		// RFC3339
		t := m.Time.Format(time.RFC3339)
		meta, err := json.Marshal(m.Metadata)
		if err != nil {
			return err
		}

		op, err := s.db.Prepare(fmt.Sprintf(dbInsertValue, s.table))
		if err != nil {
			return err
		}

		_, err = op.Exec(
			t,
			m.Level,
			m.Prefix,
			m.Sub,
			m.Msg,
			string(meta),
		)

		if err != nil {
			return err
		}
	}

	return nil
}

// Write method implements the io.Writer interface, for SQLite DBs to be used with Logger,
// as its writer.
//
// This implementation relies on JSON or gob-encoding the messages, so they are passed onto
// this writer. Then, it is unmarshalled into a message object which is sent in an Insert()
// call.
func (s *SQLite3) Write(p []byte) (n int, err error) {
	if s.db == nil {
		return 0, ErrDBNotSetUp
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

	err = s.Insert(out)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// LCSQLite struct defines the Logger Config object that adds a SQLite writer to a Logger
type LCSQLite struct {
	out io.Writer
	fmt log.LogFormatter
}

// WithSQLite function takes in a path to a .db file, and a table name; and returns a LoggerConfig
// so that this type of writer is defined in a Logger
func WithSQLite(path, table string) log.LoggerConfig {
	var db = &SQLite3{}
	var lerr, nerr error

	db, lerr = Load(path, table)
	if lerr != nil {
		db, nerr = New(path, table)
		if nerr != nil {
			fmt.Printf("failed to open or create database with errors; open: %s ; create: %s", lerr, nerr)
			os.Exit(1)
		}
	}

	return &LCSQLite{
		out: db,
		fmt: log.FormatJSON,
	}
}

// Apply method will set the input LoggerBuilder's outputs and format to the LCSQLite object's.
func (c *LCSQLite) Apply(lb *log.LoggerBuilder) {
	lb.Out = c.out
	lb.Fmt = c.fmt
}
