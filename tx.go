package jsondb

type tx struct {
	conn   *conn
	tables map[*table]bool
}

func newTx(conn *conn) (me *tx) {
	me = &tx{conn: conn, tables: map[*table]bool{}}
	return
}

func (me *tx) Commit() (err error) {
	var e error
	me.conn.tx = nil
	for t, _ := range me.tables {
		if e = t.persist(); e != nil && err == nil {
			err = e
		}
	}
	return
}

func (me *tx) Rollback() (err error) {
	var e error
	me.conn.tx = nil
	println("ROLLBACK")
	for t, _ := range me.tables {
		if e = t.reload(false); e != nil && err == nil {
			err = e
		}
	}
	return
}
