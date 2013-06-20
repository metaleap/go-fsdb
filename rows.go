package jsondb

import (
	"database/sql/driver"

	usl "github.com/metaleap/go-util/slice"
)

type rows struct {
	recs []M
	cols []string
	cur  int
}

func newRows(recs []M) (me *rows) {
	me = &rows{recs: recs}
	for _, rec := range me.recs {
		for cn, _ := range rec {
			usl.StrAppendUnique(&me.cols, cn)
		}
	}
	return
}

func (me *rows) Columns() []string {
	return me.cols
}

func (me *rows) Close() (err error) {
	return
}

func (me *rows) Next(dest []driver.Value) (err error) {
	if rec := me.recs[me.cur]; rec != nil {
		var str string
		var ok bool
		for ci, cn := range me.cols {
			if str, ok = rec[cn].(string); ok {
				dest[ci] = []byte(str)
			} else {
				dest[ci] = rec[cn]
			}
		}
	}
	return
}
