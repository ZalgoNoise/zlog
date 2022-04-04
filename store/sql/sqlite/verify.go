package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const dbTableSchema string = `PRAGMA main.table_xinfo(%s);`

var (
	tableSchemas = map[int]tableSchema{
		0: {name: "id", ftype: "INTEGER"},
		1: {name: "time", ftype: "TEXT KEY"},
		2: {name: "level", ftype: "TEXT"},
		3: {name: "prefix", ftype: "TEXT"},
		4: {name: "sub", ftype: "TEXT"},
		5: {name: "message", ftype: "TEXT"},
		6: {name: "metadata", ftype: "TEXT"},
	}
)

type tableSchema struct {
	name  string
	ftype string
}

type logTableInfo struct {
	idx      *int
	name     *string
	ftype    *string
	ncol     *int
	wr       *int
	strict   *int
	reserved *interface{}
}

func initTable() *logTableInfo {
	var (
		idx      int
		name     string
		ftype    string
		ncol     int
		wr       int
		strict   int
		reserved interface{}
	)

	return &logTableInfo{
		idx:      &idx,
		name:     &name,
		ftype:    &ftype,
		ncol:     &ncol,
		wr:       &wr,
		strict:   &strict,
		reserved: &reserved,
	}
}

func verify(db *sql.DB, table string) error {
	row, err := db.Query(fmt.Sprintf(dbTableSchema, table))
	if err != nil {
		return err
	}

	schema := initTable()
	var counter int

	for row.Next() {
		err = row.Scan(schema.idx, schema.name, schema.ftype, schema.ncol, schema.reserved, schema.wr, schema.strict)
		if err != nil {
			return err
		}

		wants := tableSchemas[counter]

		if schema.name == nil || *schema.name != wants.name {
			return fmt.Errorf(
				"column %v: name mismatch -- wanted %s ; got %s",
				counter,
				wants.name,
				*schema.name,
			)
		}

		if schema.ftype == nil || *schema.ftype != wants.ftype {
			return fmt.Errorf(
				"column %v: name mismatch -- wanted %s ; got %s",
				counter,
				wants.ftype,
				*schema.ftype,
			)
		}
		counter++
	}
	row.Close()

	return nil
}
