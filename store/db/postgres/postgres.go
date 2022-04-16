package postgres

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/zalgonoise/zlog/log"
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
func (d *Postgres) Create(msg ...*log.LogMessage) error {
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

	err = s.Create(out)
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

// LCPostgres struct defines the Logger Config object that adds a Postgres writer to a Logger
type LCPostgres struct {
	out io.Writer
	fmt log.LogFormatter
}

// WithPostgres function takes in a path to a .db file, and a table name; and returns a LoggerConfig
// so that this type of writer is defined in a Logger
func WithPostgres(address, port, database string) log.LoggerConfig {
	db, err := New(address, port, database)
	if err != nil {
		fmt.Printf("failed to open or create database with an error: %s", err)
		os.Exit(1)
	}

	//TODO(zalgonoise): benchmark this decision -- confirm if gob is more performant,
	// considering that JSON will (usually) have less bytes per (small) message
	return &LCPostgres{
		out: db,
		fmt: log.FormatJSON,
	}
}

// Apply method will set the input LoggerBuilder's outputs and format to the LCPostgres object's.
func (c *LCPostgres) Apply(lb *log.LoggerBuilder) {
	lb.Out = c.out
	lb.Fmt = c.fmt
}
