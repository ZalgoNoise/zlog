package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zalgonoise/zlog/log"
)

const (
	dbCreateTable string = `CREATE TABLE IF NOT EXISTS log (id INTEGER PRIMARY KEY, time TEXT KEY, level TEXT, prefix TEXT, sub TEXT, message TEXT, metadata TEXT);`
	dbInsertValue string = `INSERT INTO log (time, level, prefix, sub, message, metadata) VALUES (?, ?, ?, ?, ?, ?);`
)

var (
	ErrDBNotSetUp error = errors.New("database is not setup -- nil pointer")
)

type SQLite3 struct {
	path string
	db   *sql.DB
}

func New(path string) (*SQLite3, error) {

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	op, err := db.Prepare(dbCreateTable)
	if err != nil {
		return nil, err
	}

	defer op.Close()

	_, err = op.Exec()
	if err != nil {
		return nil, err
	}

	return &SQLite3{
		path: path,
		db:   db,
	}, nil

}

func Load(path string) (*SQLite3, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	err = verify(db)
	if err != nil {
		return nil, err
	}

	return &SQLite3{
		path: path,
		db:   db,
	}, nil
}

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

		op, err := s.db.Prepare(dbInsertValue)
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
