package jsondb

import (
	"database/sql/driver"
)

type rows struct {
}

func (me *rows) Columns() (cols []string) {
	return
}

func (me *rows) Close() (err error) {
	return
}

func (me *rows) Next(dest []driver.Value) (err error) {
	return
}
