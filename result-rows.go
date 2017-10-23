package fsdb

import (
	"database/sql/driver"
	"io"

	"github.com/metaleap/go-util/slice"
)

type result struct {
	InsertedLast, AffectedRows int64
}

func (me *result) LastInsertId() (id int64, err error) {
	id = me.InsertedLast
	return
}

func (me *result) RowsAffected() (num int64, err error) {
	num = me.AffectedRows
	return
}

type rows struct {
	cols []string
	rids []string
	recs []M
	cur  int
}

func newRows(recs map[string]M) (me *rows) {
	me = &rows{recs: make([]M, 0, len(recs)), rids: make([]string, 0, len(recs))}
	me.cols = append(me.cols, IdField)
	for rid, rec := range recs {
		for cn, _ := range rec {
			uslice.StrAppendUnique(&me.cols, cn)
		}
		me.recs = append(me.recs, rec)
		me.rids = append(me.rids, rid)
	}
	return
}

func (me *rows) Columns() []string {
	return me.cols
}

func (me *rows) Close() (err error) {
	me.cur = 0
	return
}

func (me *rows) Next(dest []driver.Value) (err error) {
	if me.cur < len(me.recs) {
		if rec := me.recs[me.cur]; rec != nil {
			var str string
			var ok bool
			for ci, cn := range me.cols {
				if cn == IdField {
					dest[ci] = me.rids[me.cur]
				} else if str, ok = rec[cn].(string); ok {
					dest[ci] = []byte(str)
				} else {
					dest[ci] = rec[cn]
				}
			}
		}
		me.cur++
	} else {
		err = io.EOF
	}
	return
}
