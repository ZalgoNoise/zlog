package postgres

import (
	"fmt"
	"io"
	"os"
	"regexp"

	"testing"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

func getEnv(env string) (val string, ok bool) {
	v := os.Getenv(env)

	if v == "" {
		return v, false
	}

	return v, true
}

type testDB struct {
	host     string
	port     string
	database string
}

func initEnv(host, port, database string, omitDB bool) (*testDB, error) {
	errStr := "invalid input, missing env %s"
	out := &testDB{}

	h := os.Getenv(host)
	if h == "" {
		out.host = "127.0.0.1"
	}
	out.host = h

	p := os.Getenv(port)
	if p == "" {
		out.port = "3306"
	}
	out.port = p

	d := os.Getenv(database)
	if d == "" && !omitDB {
		return nil, fmt.Errorf(errStr, database)
	}
	out.database = d

	return out, nil
}

var omitEnv = func(e string) (func(), error) {
	module := "Postgres"
	funcname := "omitEnv(%s)"

	env := os.Getenv(e)
	err := os.Setenv(e, "")

	if err != nil {
		return nil, fmt.Errorf(
			"FAILED -- [%s] [%s] unexpected error: %v",
			module,
			fmt.Sprintf(funcname, e),
			err,
		)
	}

	return func() {
		os.Setenv(e, env)
	}, nil
}

func TestNew(t *testing.T) {
	module := "Postgres"
	funcname := "New()"

	_ = module
	_ = funcname

	type test struct {
		name         string
		envHost      string
		envPort      string
		envDatabase  string
		omitUser     bool
		omitPassword bool
		omitDatabase bool
	}

	// paths are fetched with os.Getenv(); you may need to export these variables
	// before executing the tests:
	//
	//     export \
	//       POSTGRES_HOST={address of the DB host} \
	//       POSTGRES_PORT={DB port} \
	//       POSTGRES_DATABASE={Postgres database (within the instance)} \
	//       POSTGRES_USER={user who will write to the database} \
	//       POSTGRES_PASSWORD={password for POSTGRES_USER}
	//
	var tests = []test{
		{
			name:        "working database",
			envHost:     "POSTGRES_HOST",
			envPort:     "POSTGRES_PORT",
			envDatabase: "POSTGRES_DATABASE",
		},
		{
			name:        "working database",
			envHost:     "POSTGRES_HOST",
			envPort:     "POSTGRES_PORT",
			envDatabase: "POSTGRES_DATABASE",
			omitUser:    true,
		},
		{
			name:         "working database",
			envHost:      "POSTGRES_HOST",
			envPort:      "POSTGRES_PORT",
			envDatabase:  "POSTGRES_DATABASE",
			omitPassword: true,
		},
		{
			name:         "working database",
			envHost:      "POSTGRES_HOST",
			envPort:      "POSTGRES_PORT",
			envDatabase:  "POSTGRES_DATABASE",
			omitDatabase: true,
		},
	}

	var verify = func(idx int, test test) {
		if test.omitUser {
			reset, err := omitEnv("POSTGRES_USER")
			if err != nil {
				t.Errorf("#%v -- %s -- action: %s", idx, err, test.name)
				return
			}
			defer reset()
		}

		if test.omitPassword {
			reset, err := omitEnv("POSTGRES_PASSWORD")
			if err != nil {
				t.Errorf("#%v -- %s -- action: %s", idx, err, test.name)
				return
			}
			defer reset()
		}

		if test.omitDatabase {
			reset, err := omitEnv("POSTGRES_DATABASE")
			if err != nil {
				t.Errorf("#%v -- %s -- action: %s", idx, err, test.name)
				return
			}
			defer reset()
		}

		cfg, err := initEnv(test.envHost, test.envPort, test.envDatabase, test.omitDatabase)

		if err != nil {
			t.Logf(
				"#%v -- SKIPPED -- [%s] [%s] unexpected error when collecting env: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		_, err = New(cfg.host, cfg.port, cfg.database)

		if err != nil && !test.omitPassword && !test.omitUser && !test.omitDatabase {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestCreate(t *testing.T) {
	module := "Postgres"
	funcname := "Create()"

	_ = module
	_ = funcname

	type test struct {
		name        string
		envHost     string
		envPort     string
		envDatabase string
		e           []*event.Event
		ok          bool
	}

	// paths are fetched with os.Getenv(); you may need to export these variables
	// before executing the tests:
	//
	//     export \
	//       POSTGRES_HOST={address of the DB host} \
	//       POSTGRES_PORT={DB port} \
	//       POSTGRES_DATABASE={Postgres database (within the instance)} \
	//       POSTGRES_USER={user who will write to the database} \
	//       POSTGRES_PASSWORD={password for POSTGRES_USER}
	//
	var tests = []test{
		{
			name:        "create a single event",
			envHost:     "POSTGRES_HOST",
			envPort:     "POSTGRES_PORT",
			envDatabase: "POSTGRES_DATABASE",
			e: []*event.Event{
				event.New().Message("null").Build(),
			},
			ok: true,
		},
		{
			name:        "create multiple events",
			envHost:     "POSTGRES_HOST",
			envPort:     "POSTGRES_PORT",
			envDatabase: "POSTGRES_DATABASE",
			e: []*event.Event{
				event.New().Message("null_0").Build(),
				event.New().Message("null_1").Build(),
				event.New().Message("null_2").Build(),
			},
			ok: true,
		},
		{
			name:        "create with empty list of events",
			envHost:     "POSTGRES_HOST",
			envPort:     "POSTGRES_PORT",
			envDatabase: "POSTGRES_DATABASE",
			e:           []*event.Event{},
			ok:          true,
		},
		{
			name:        "create with nil events",
			envHost:     "POSTGRES_HOST",
			envPort:     "POSTGRES_PORT",
			envDatabase: "POSTGRES_DATABASE",
			e:           nil,
			ok:          true,
		},
		{
			name:        "create multiple events with some nil messages",
			envHost:     "POSTGRES_HOST",
			envPort:     "POSTGRES_PORT",
			envDatabase: "POSTGRES_DATABASE",
			e: []*event.Event{
				event.New().Message("null_0").Build(),
				nil,
				nil,
			},
			ok: true,
		},
		{
			name:        "create multiple events with all nil messages",
			envHost:     "POSTGRES_HOST",
			envPort:     "POSTGRES_PORT",
			envDatabase: "POSTGRES_DATABASE",
			e:           []*event.Event{nil, nil, nil},
			ok:          true,
		},
		{
			name:        "create invalid / empty events",
			envHost:     "POSTGRES_HOST",
			envPort:     "POSTGRES_PORT",
			envDatabase: "POSTGRES_DATABASE",
			e:           []*event.Event{{}, {}, {}},
			ok:          false,
		},
	}

	var verify = func(idx int, test test) {
		cfg, err := initEnv(test.envHost, test.envPort, test.envDatabase, false)

		if err != nil {
			t.Logf(
				"#%v -- SKIPPED -- [%s] [%s] unexpected error when collecting env: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		db, err := New(cfg.host, cfg.port, cfg.database)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		mysql, ok := db.(*Postgres)

		if !ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] DB creation did not return a Postgres pointer -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		if test.e == nil {
			err = mysql.Create(nil)
		} else {
			err = mysql.Create(test.e...)
		}

		if test.ok && err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}
}

func TestWrite(t *testing.T) {
	module := "Postgres"
	funcname := "Write()"

	_ = module
	_ = funcname

	type test struct {
		name         string
		envHost      string
		envPort      string
		envDatabase  string
		e            []byte
		omitDatabase bool
		ok           bool
	}

	// paths are fetched with os.Getenv(); you may need to export these variables
	// before executing the tests:
	//
	//     export \
	//       POSTGRES_HOST={address of the DB host} \
	//       POSTGRES_PORT={DB port} \
	//       POSTGRES_DATABASE={Postgres database (within the instance)} \
	//       POSTGRES_USER={user who will write to the database} \
	//       POSTGRES_PASSWORD={password for POSTGRES_USER}
	//
	var tests = []test{
		{
			name:        "write a single event",
			envHost:     "POSTGRES_HOST",
			envPort:     "POSTGRES_PORT",
			envDatabase: "POSTGRES_DATABASE",
			e:           event.New().Message("null").Build().Encode(),
			ok:          true,
		},
		{
			name:        "write invalid event",
			envHost:     "POSTGRES_HOST",
			envPort:     "POSTGRES_PORT",
			envDatabase: "POSTGRES_DATABASE",
			e:           []byte("null"),
			ok:          false,
		},
		{
			name:        "write nil event",
			envHost:     "POSTGRES_HOST",
			envPort:     "POSTGRES_PORT",
			envDatabase: "POSTGRES_DATABASE",
			e:           nil,
			ok:          false,
		},
	}

	var instanceTests = []test{
		{
			name:         "write a single event",
			envHost:      "POSTGRES_HOST",
			envPort:      "POSTGRES_PORT",
			envDatabase:  "POSTGRES_DATABASE",
			e:            event.New().Message("null").Build().Encode(),
			omitDatabase: true,
			ok:           false,
		},
		{
			name:        "write a single event",
			envHost:     "POSTGRES_HOST",
			envPort:     "POSTGRES_PORT",
			envDatabase: "POSTGRES_DATABASE",
			e:           event.New().Message("null").Build().Encode(),
			ok:          true,
		},
	}

	var initDB = func(idx int, test test) (io.WriteCloser, error) {
		cfg, err := initEnv(test.envHost, test.envPort, test.envDatabase, false)

		if err != nil {
			return nil, fmt.Errorf(
				"#%v -- SKIPPED -- [%s] [%s] unexpected error when collecting env: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
		}

		return New(cfg.host, cfg.port, cfg.database)
	}

	var verify = func(idx int, test test) {
		db, err := initDB(idx, test)

		if test.ok && err != nil {
			t.Logf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		if test.e == nil {
			_, err = db.Write(nil)
		} else {
			_, err = db.Write(test.e)
		}

		if test.ok && err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}
	}

	var verifyNewInstance = func(idx int, test test) {
		db, err := initDB(idx, test)

		if test.ok && err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		mysql, ok := db.(*Postgres)

		if !ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error when converting Postgres DB interface -- action: %s",
				idx,
				module,
				funcname,
				test.name,
			)
			return
		}

		mysql.db = nil

		if test.omitDatabase {
			mysql.database = ""
		}

		if test.e == nil {
			_, err = mysql.Write(nil)
		} else {
			_, err = mysql.Write(test.e)
		}

		if test.ok && err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}

	for idx, test := range instanceTests {
		verifyNewInstance(idx, test)
	}
}

func TestClose(t *testing.T) {
	module := "Postgres"
	funcname := "Close()"

	_ = module
	_ = funcname

	type test struct {
	}

	// paths are fetched with os.Getenv(); you may need to export these variables
	// before executing the tests:
	//
	//     export \
	//       POSTGRES_HOST={address of the DB host} \
	//       POSTGRES_PORT={DB port} \
	//       POSTGRES_DATABASE={Postgres database (within the instance)} \
	//       POSTGRES_USER={user who will write to the database} \
	//       POSTGRES_PASSWORD={password for POSTGRES_USER}
	//
	var tests = []test{
		{},
	}

	var verify = func(idx int, test test) {

		db := &Postgres{}

		err := db.Close()

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v",
				idx,
				module,
				funcname,
				err,
			)
			return
		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}

}

func TestWithPostgres(t *testing.T) {
	module := "Postgres"
	funcname := "WithPostgres()"

	_ = module
	_ = funcname

	type test struct {
		name        string
		envHost     string
		envPort     string
		envDatabase string
		ok          bool
	}

	// paths are fetched with os.Getenv(); you may need to export these variables
	// before executing the tests:
	//
	//     export \
	//       POSTGRES_HOST={address of the DB host} \
	//       POSTGRES_PORT={DB port} \
	//       POSTGRES_DATABASE={Postgres database (within the instance)} \
	//       POSTGRES_USER={user who will write to the database} \
	//       POSTGRES_PASSWORD={password for POSTGRES_USER}
	//
	var tests = []test{
		{
			name:        "write a single event",
			envHost:     "POSTGRES_HOST",
			envPort:     "POSTGRES_PORT",
			envDatabase: "POSTGRES_DATABASE",
			ok:          true,
		},
		{
			name:        "write a single event",
			envHost:     "POSTGRES_HOST",
			envPort:     "POSTGRES_PORT",
			envDatabase: "POSTGRES_DATABASE",
			ok:          false,
		},
	}

	var catchPanic = func(idx int, test test) {
		r := recover()

		if r == nil && test.ok {
			return
		}

		regexStr := `failed to create logger config -- database creation failed:.+`
		regex := regexp.MustCompile(regexStr)

		if !regex.MatchString(r.(error).Error()) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] error mismatch: wanted to match %s; got %v -- action: %s",
				idx,
				module,
				funcname,
				regexStr,
				r.(string),
				test.name,
			)
			return
		}
	}

	var verify = func(idx int, test test) {
		defer catchPanic(idx, test)

		cfg, err := initEnv(test.envHost, test.envPort, test.envDatabase, false)

		if err != nil {
			t.Logf(
				"#%v -- SKIPPED -- [%s] [%s] unexpected error when collecting env: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		if !test.ok {
			cfg.database = ""
		}

		c := WithPostgres(cfg.host, cfg.port, cfg.database)

		out, ok := c.(*log.LCDatabase)

		if !ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] operation didn't return expected type: %T -- action: %s",
				idx,
				module,
				funcname,
				out,
				test.name,
			)
			return

		}

	}

	for idx, test := range tests {
		verify(idx, test)
	}

}
