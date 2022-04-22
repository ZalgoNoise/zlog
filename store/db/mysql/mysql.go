package mysql

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
	model "github.com/zalgonoise/zlog/store/db/message"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrNoEnv error = errors.New("no env variable provided -- ensure that the environment variables for MYSQL_USER and MYSQL_PASSWORD are set")
)

type MySQL struct {
	addr     string
	database string
	db       *gorm.DB
}

// New function will take in a mysql DB address and database name; and create
// a new instance of a MySQL object; returning an io.WriteCloser and an error.
func New(address, database string) (sqldb io.WriteCloser, err error) {
	db, err := initialMigration(address, database)

	if err != nil {
		return nil, err
	}

	sqldb = &MySQL{
		addr:     address,
		database: database,
		db:       db,
	}

	return
}

// Create method will register any number of event.Event in the MySQL database, returning
// an error
func (d *MySQL) Create(msg ...*event.Event) error {
	if len(msg) == 0 {
		return nil
	}

	var msgs []*model.Event

	for _, m := range msg {
		var entry = &model.Event{}

		if err := entry.From(m); err != nil {
			return err
		}
		msgs = append(msgs, entry)
	}

	d.db.Create(msgs)
	return nil
}

// Write method implements the io.Writer interface, for MySQL DBs to be used with Logger,
// as its writer.
//
// The input message is expected to be a protobuf-marshalled event.Event, which is decoded
func (s *MySQL) Write(p []byte) (n int, err error) {
	if s.db == nil && s.addr != "" {
		if s.database == "" {
			s.database = "logs"
		}

		new, err := New(s.addr, s.database)
		if err != nil {
			return 0, err
		}
		s = new.(*MySQL)
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
func (d *MySQL) Close() error { return nil }

func initialMigration(address, database string) (*gorm.DB, error) {
	// "gorm:gorm@tcp(127.0.0.1:3306)/gorm?charset=utf8&parseTime=True&loc=Local"
	var uri = strings.Builder{}

	uri.WriteString(os.Getenv("MYSQL_USER"))
	uri.WriteString(":")
	uri.WriteString(os.Getenv("MYSQL_PASSWORD"))
	uri.WriteString("@tcp(")
	uri.WriteString(address)
	uri.WriteString(")/")
	uri.WriteString(database)
	uri.WriteString("?charset=utf8&parseTime=True&loc=Local")

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       uri.String(), // data source name
		DefaultStringSize:         256,          // default size for string fields
		DisableDatetimePrecision:  true,         // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,         // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,         // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false,        // auto configure based on currently MySQL version
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

// WithMySQL function takes in an address to a MySQL server, and a database name; and returns a LoggerConfig
// so that this type of writer is defined in a Logger
func WithMySQL(addr, database string) log.LoggerConfig {
	db, err := New(addr, database)
	if err != nil {
		fmt.Printf("failed to open or create database with an error: %s", err)
		os.Exit(1)
	}

	return &log.LCDatabase{
		Out: db,
		Fmt: log.FormatProtobuf,
	}
}
