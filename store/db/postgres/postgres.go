package postgres

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	model "github.com/zalgonoise/zlog/store/db/message"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	ErrNoEnv           error = errors.New("no env variable provided -- ensure that the environment variables for POSTGRES_USER and POSTGRES_PASSWORD are set")
	ErrMissingUser     error = errors.New("unset Postgres user variable: please export the PSOTGRES_USER variable")
	ErrMissingPassword error = errors.New("unset Postgres password variable: please export the PSOTGRES_PASSWORD variable")
	ErrMissingDatabase error = errors.New("unset Postgres database variable: please export the PSOTGRES_DATABASE variable")
)

type Postgres struct {
	addr     string
	port     string
	database string
	db       *gorm.DB
}

// New function will take in a postgres DB address, port and database name; and create
// a new instance of a Postgres object; returning an io.WriteCloser and an error.
func New(address, port, database string) (sqldb io.WriteCloser, err error) {
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

// Create method will register any number of event.Event in the Postgres database, returning
// an error
func (d *Postgres) Create(msg ...*event.Event) error {
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

	d.db.Create(msgs)
	return nil
}

// Write method implements the io.Writer interface, for Postgres DBs to be used with Logger,
// as its writer.
//
// The input message is expected to be a protobuf-marshalled event.Event, which is decoded
func (s *Postgres) Write(p []byte) (n int, err error) {
	if s.db == nil && s.addr != "" {
		if s.port == "" {
			s.port = "5432"
		}

		if s.database == "" {
			return 0, ErrMissingDatabase
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

	if database == "" {
		return nil, ErrMissingDatabase
	}

	u := os.Getenv("POSTGRES_USER")
	if u == "" {
		return nil, ErrMissingUser
	}

	p := os.Getenv("POSTGRES_PASSWORD")
	if p == "" {
		return nil, ErrMissingPassword
	}

	uri.WriteString("host=")
	uri.WriteString(address)
	uri.WriteString(" user=")
	uri.WriteString(u)
	uri.WriteString(" password=")
	uri.WriteString(p)
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
	err = db.AutoMigrate(&model.Event{})

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
		panic(fmt.Errorf("failed to create logger config -- database creation failed: %s", err))
	}

	return &log.LCDatabase{
		Out: db,
		Fmt: log.FormatProtobuf,
	}
}
