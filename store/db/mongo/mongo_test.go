package mongo

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"

	"testing"

	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

var errNoEnv error = errors.New("no environment variable found")

type testDB struct {
	host       string
	port       string
	database   string
	collection string
}

func initEnv(host, port, database, collection string, omitDB, omitCol bool) (*testDB, error) {
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
		e := fmt.Errorf(errStr, database)
		return nil, fmt.Errorf("%w: %s", errNoEnv, e.Error())
	}
	out.database = d

	c := os.Getenv(collection)
	if c == "" && !omitCol {
		e := fmt.Errorf(errStr, collection)
		return nil, fmt.Errorf("%w: %s", errNoEnv, e.Error())
	}
	out.collection = c

	return out, nil
}

var omitEnv = func(e string) (func(), error) {
	module := "Mongo"
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
	module := "Mongo"
	funcname := "New()"

	_ = module
	_ = funcname

	type test struct {
		name           string
		envHost        string
		envPort        string
		envDatabase    string
		envCollection  string
		omitUser       bool
		omitPassword   bool
		omitDatabase   bool
		omitCollection bool
	}

	// paths are fetched with os.Getenv(); you may need to export these variables
	// before executing the tests:
	//
	//     export \
	//       MONGO_HOST={address of the DB host} \
	//       MONGO_PORT={DB port} \
	//       MONGO_DATABASE={Postgres database (within the instance)} \
	//       MONGO_USER={user who will write to the database} \
	//       MONGO_PASSWORD={password for MONGO_USER}
	//
	var tests = []test{
		{
			name:          "working database",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
		},
		{
			name:          "working database",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			omitUser:      true,
		},
		{
			name:          "working database",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			omitPassword:  true,
		},
		{
			name:          "working database",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			omitDatabase:  true,
		},
		{
			name:           "working database",
			envHost:        "MONGO_HOST",
			envPort:        "MONGO_PORT",
			envDatabase:    "MONGO_DATABASE",
			envCollection:  "MONGO_COLLECTION",
			omitCollection: true,
		},
	}

	var verify = func(idx int, test test) {
		if test.omitUser {
			reset, err := omitEnv("MONGO_USER")
			if err != nil {
				t.Errorf("#%v -- %s -- action: %s", idx, err, test.name)
				return
			}
			defer reset()
		}

		if test.omitPassword {
			reset, err := omitEnv("MONGO_PASSWORD")
			if err != nil {
				t.Errorf("#%v -- %s -- action: %s", idx, err, test.name)
				return
			}
			defer reset()
		}

		if test.omitDatabase {
			reset, err := omitEnv("MONGO_DATABASE")
			if err != nil {
				t.Errorf("#%v -- %s -- action: %s", idx, err, test.name)
				return
			}
			defer reset()
		}

		if test.omitCollection {
			reset, err := omitEnv("MONGO_COLLECTION")
			if err != nil {
				t.Errorf("#%v -- %s -- action: %s", idx, err, test.name)
				return
			}
			defer reset()
		}

		cfg, err := initEnv(test.envHost, test.envPort, test.envDatabase, test.envCollection, test.omitDatabase, test.omitCollection)

		if err != nil {
			if errors.Is(err, errNoEnv) {
				e := errors.Unwrap(err)

				t.Logf(
					"#%v -- SKIPPED -- [%s] [%s] unexpected error when collecting env: %v",
					idx,
					module,
					funcname,
					e,
				)
				return
			}

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

		addr := cfg.host + ":" + cfg.port
		_, err = New(addr, cfg.database, cfg.collection)

		if err != nil && !test.omitPassword && !test.omitUser && !test.omitDatabase && !test.omitCollection {
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
	module := "Mongo"
	funcname := "Create()"

	_ = module
	_ = funcname

	type test struct {
		name          string
		envHost       string
		envPort       string
		envDatabase   string
		envCollection string
		e             []*event.Event
		ok            bool
	}

	// paths are fetched with os.Getenv(); you may need to export these variables
	// before executing the tests:
	//
	//     export \
	//       MONGO_HOST={address of the DB host} \
	//       MONGO_PORT={DB port} \
	//       MONGO_DATABASE={Postgres database (within the instance)} \
	//       MONGO_USER={user who will write to the database} \
	//       MONGO_PASSWORD={password for MONGO_USER}
	//
	var tests = []test{
		{
			name:          "create a single event",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			e: []*event.Event{
				event.New().Message("null").Build(),
			},
			ok: true,
		},
		{
			name:          "create multiple events",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			e: []*event.Event{
				event.New().Message("null_0").Build(),
				event.New().Message("null_1").Build(),
				event.New().Message("null_2").Build(),
			},
			ok: true,
		},
		{
			name:          "create with empty list of events",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			e:             []*event.Event{},
			ok:            true,
		},
		{
			name:          "create with nil events",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			e:             nil,
			ok:            true,
		},
		{
			name:          "create multiple events with some nil messages",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			e: []*event.Event{
				event.New().Message("null_0").Build(),
				nil,
				nil,
			},
			ok: true,
		},
		{
			name:          "create multiple events with all nil messages",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			e:             []*event.Event{nil, nil, nil},
			ok:            true,
		},
		{
			name:          "create invalid / empty events",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			e:             []*event.Event{{}, {}, {}},
			ok:            false,
		},
	}

	var verify = func(idx int, test test) {
		cfg, err := initEnv(test.envHost, test.envPort, test.envDatabase, test.envCollection, false, false)

		if err != nil {
			if errors.Is(err, errNoEnv) {
				e := errors.Unwrap(err)

				t.Logf(
					"#%v -- SKIPPED -- [%s] [%s] unexpected error when collecting env: %v",
					idx,
					module,
					funcname,
					e,
				)
				return
			}

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

		addr := cfg.host + ":" + cfg.port
		db, err := New(addr, cfg.database, cfg.collection)

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

		mongo, ok := db.(*Mongo)

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
			err = mongo.Create(nil)
		} else {
			err = mongo.Create(test.e...)
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
	module := "Mongo"
	funcname := "Write()"

	_ = module
	_ = funcname

	type test struct {
		name           string
		envHost        string
		envPort        string
		envDatabase    string
		envCollection  string
		e              []byte
		omitDatabase   bool
		omitCollection bool
		ok             bool
	}

	// paths are fetched with os.Getenv(); you may need to export these variables
	// before executing the tests:
	//
	//     export \
	//       MONGO_HOST={address of the DB host} \
	//       MONGO_PORT={DB port} \
	//       MONGO_DATABASE={Postgres database (within the instance)} \
	//       MONGO_USER={user who will write to the database} \
	//       MONGO_PASSWORD={password for MONGO_USER}
	//
	var tests = []test{
		{
			name:          "write a single event",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			e:             event.New().Message("null").Build().Encode(),
			ok:            true,
		},
		{
			name:          "write invalid event",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			e:             []byte("null"),
			ok:            false,
		},
		{
			name:          "write nil event",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			e:             nil,
			ok:            false,
		},
	}

	var instanceTests = []test{
		{
			name:          "write a single event",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			e:             event.New().Message("null").Build().Encode(),
			omitDatabase:  true,
			ok:            false,
		},
		{
			name:           "write a single event",
			envHost:        "MONGO_HOST",
			envPort:        "MONGO_PORT",
			envDatabase:    "MONGO_DATABASE",
			envCollection:  "MONGO_COLLECTION",
			e:              event.New().Message("null").Build().Encode(),
			omitCollection: true,
			ok:             false,
		},
		{
			name:          "write a single event",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			e:             event.New().Message("null").Build().Encode(),
			ok:            true,
		},
	}

	var initDB = func(idx int, test test) (io.WriteCloser, error) {
		cfg, err := initEnv(test.envHost, test.envPort, test.envDatabase, test.envCollection, false, false)

		if err != nil {
			if errors.Is(err, errNoEnv) {
				return nil, fmt.Errorf(
					"#%v -- SKIPPED -- [%s] [%s] unexpected error when collecting env: %w -- action: %s",
					idx,
					module,
					funcname,
					err,
					test.name,
				)
			}

			return nil, fmt.Errorf(
				"#%v -- FAILED -- [%s] [%s] unexpected error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
		}

		addr := cfg.host + ":" + cfg.port
		return New(addr, cfg.database, cfg.collection)
	}

	var verify = func(idx int, test test) {
		db, err := initDB(idx, test)

		if err != nil {
			inner := errors.Unwrap(err)

			if errors.Is(inner, errNoEnv) {
				t.Log(err)
				return
			}

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

		if db == nil && !test.ok {
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

		if err != nil {
			inner := errors.Unwrap(err)

			if errors.Is(inner, errNoEnv) {
				t.Log(err)
				return
			}

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

		if db == nil && !test.ok {
			return
		}

		mongo, ok := db.(*Mongo)

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

		mongo.db = nil

		if test.omitDatabase {
			mongo.database = ""
		}

		if test.omitCollection {
			mongo.collection = ""
		}

		if test.e == nil {
			_, err = mongo.Write(nil)
		} else {
			_, err = mongo.Write(test.e)
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
	module := "Mongo"
	funcname := "Close()"

	_ = module
	_ = funcname

	type test struct {
		name          string
		envHost       string
		envPort       string
		envDatabase   string
		envCollection string
		ok            bool
	}

	// paths are fetched with os.Getenv(); you may need to export these variables
	// before executing the tests:
	//
	//     export \
	//       MONGO_HOST={address of the DB host} \
	//       MONGO_PORT={DB port} \
	//       MONGO_DATABASE={Postgres database (within the instance)} \
	//       MONGO_USER={user who will write to the database} \
	//       MONGO_PASSWORD={password for MONGO_USER}
	//
	var tests = []test{
		{
			name:          "working closure",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			ok:            true,
		},
	}

	var initDB = func(idx int, test test) (io.WriteCloser, error) {
		if !test.ok {
			// TODO: get a decent failing state that doesn't result in a
			// nil pointer dereference panic
			return new(Mongo), nil
		}

		cfg, err := initEnv(test.envHost, test.envPort, test.envDatabase, test.envCollection, false, false)

		if err != nil {
			// skip env signal, send both values nil
			return nil, nil
		}

		addr := cfg.host + ":" + cfg.port
		return New(addr, cfg.database, cfg.collection)

	}

	var verify = func(idx int, test test) {
		db, err := initDB(idx, test)

		if err != nil {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] failed to create the database with an error: %v -- action: %s",
				idx,
				module,
				funcname,
				err,
				test.name,
			)
			return
		}

		if err == nil && db == nil {
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

		err = db.Close()

		if err != nil && test.ok {
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

func TestWithMongo(t *testing.T) {
	module := "Mongo"
	funcname := "WithMongo()"

	_ = module
	_ = funcname

	type test struct {
		name          string
		envHost       string
		envPort       string
		envDatabase   string
		envCollection string
		ok            bool
	}

	// paths are fetched with os.Getenv(); you may need to export these variables
	// before executing the tests:
	//
	//     export \
	//       MONGO_HOST={address of the DB host} \
	//       MONGO_PORT={DB port} \
	//       MONGO_DATABASE={Postgres database (within the instance)} \
	//       MONGO_USER={user who will write to the database} \
	//       MONGO_PASSWORD={password for MONGO_USER}
	//
	var tests = []test{
		{
			name:          "write a single event",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			ok:            true,
		},
		{
			name:          "write a single event",
			envHost:       "MONGO_HOST",
			envPort:       "MONGO_PORT",
			envDatabase:   "MONGO_DATABASE",
			envCollection: "MONGO_COLLECTION",
			ok:            false,
		},
	}

	var catchPanic = func(idx int, test test) {
		r := recover()

		if r == nil {
			return
		}

		regexStr := `failed to create logger config -- database creation failed:.+`
		regex := regexp.MustCompile(regexStr)

		err, ok := r.(error)

		if !ok {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] panic value is not an error: %v -- %T -- action: %s",
				idx,
				module,
				funcname,
				r,
				r,
				test.name,
			)
			return
		}

		if !regex.MatchString(err.Error()) {
			t.Errorf(
				"#%v -- FAILED -- [%s] [%s] error mismatch: wanted to match %s; got %v -- action: %s",
				idx,
				module,
				funcname,
				regexStr,
				err.Error(),
				test.name,
			)
			return
		}
	}

	var verify = func(idx int, test test) {
		defer catchPanic(idx, test)

		cfg, err := initEnv(test.envHost, test.envPort, test.envDatabase, test.envCollection, false, false)

		if err != nil {
			inner := errors.Unwrap(err)

			if errors.Is(inner, errNoEnv) {
				t.Log(err)
				return
			}

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

		if cfg == nil && !test.ok {
			return
		}

		if !test.ok {
			cfg.database = ""
		}

		addr := cfg.host + ":" + cfg.port
		c := WithMongo(addr, cfg.database, cfg.collection)

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
