package mysql

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/zalgonoise/zlog/log"
	model "github.com/zalgonoise/zlog/store/db"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrNoEnv error = errors.New("no env variable provided -- ensure that the environment variables for MYSQL_USER and MYSQL_PASSWORD are set")
)

type MySQL struct {
	addr     string
	database string
	// table    string
	db *gorm.DB
}

// New function will take in a mysql DB address and database name; and create
// a new instance of a MySQL object; returning a pointer to one and an error.
func New(address, database string) (sqldb *MySQL, err error) {
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

// Create method will register any number of LogMessages in the MySQL database, returning
// an error
func (d *MySQL) Create(msg ...*log.LogMessage) error {
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

// Write method implements the io.Writer interface, for MySQL DBs to be used with Logger,
// as its writer.
//
// This implementation relies on JSON or gob-encoding the messages, so they are passed onto
// this writer. Then, it is unmarshalled into a message object which is sent in an Insert()
// call.
func (s *MySQL) Write(p []byte) (n int, err error) {
	if s.db == nil && s.addr != "" {
		if s.database == "" {
			s.database = "logs"
		}

		new, err := New(s.addr, s.database)
		if err != nil {
			return 0, err
		}
		s = new
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
	err = db.AutoMigrate(&model.LogMessage{})

	if err != nil {
		return nil, err
	}

	return db, nil
}

// LCMySQL struct defines the Logger Config object that adds a MySQL writer to a Logger
type LCMySQL struct {
	out io.Writer
	fmt log.LogFormatter
}

// WithMySQL function takes in a path to a .db file, and a table name; and returns a LoggerConfig
// so that this type of writer is defined in a Logger
func WithMySQL(address, database string) log.LoggerConfig {
	db, err := New(address, database)
	if err != nil {
		fmt.Printf("failed to open or create database with an error: %s", err)
		os.Exit(1)
	}

	//TODO(zalgonoise): benchmark this decision -- confirm if gob is more performant,
	// considering that JSON will (usually) have less bytes per (small) message
	return &LCMySQL{
		out: db,
		fmt: log.FormatGob,
	}
}

// Apply method will set the input LoggerBuilder's outputs and format to the LCMySQL object's.
func (c *LCMySQL) Apply(lb *log.LoggerBuilder) {
	lb.Out = c.out
	lb.Fmt = c.fmt
}
