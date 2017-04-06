package fsdb

type tx struct {
	conn   *conn
	tables map[*table]bool
}

func newTx(conn *conn) (me *tx) {
	me = &tx{conn: conn, tables: map[*table]bool{}}
	return
}

func (me *tx) Commit() (err error) {
	me.conn.tx = nil
	err = me.conn.tables.persistAll(me.tableNames()...)
	me.conn, me.tables = nil, nil
	return
}

func (me *tx) Rollback() (err error) {
	me.conn.tx = nil
	err = me.conn.tables.reloadAll(me.tableNames()...)
	me.conn, me.tables = nil, nil
	return
}

func (me *tx) tableNames() (tnames []string) {
	for t, _ := range me.tables {
		tnames = append(tnames, t.name)
	}
	return
}
