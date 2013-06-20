package jsondb

import (
	"database/sql/driver"
	"os"
	"path/filepath"

	uio "github.com/metaleap/go-util/io"
)

type conn struct {
	drv    *drv
	tx     *tx
	dir    string
	tables tables
}

func newConn(drv *drv, dir string) (me *conn, err error) {
	me = &conn{drv: drv, dir: dir}
	err = me.tables.init(me, false)
	return
}

func (me *conn) doCreateTable(name string) (err error) {
	if _, ok := me.tables.all[name]; !ok {
		if fp := filepath.Join(me.dir, name+FileExt); uio.FileExists(fp) {
			err = errf("Cannot create table '%s': already exists", name)
		} else {
			err = uio.WriteTextFile(fp, "{}")
		}
	}
	if err == nil {
		_, err = me.tables.get(name)
	}
	return
}

func (me *conn) doDropTable(name string) (err error) {
	t := me.tables.all[name]
	delete(me.tables.all, name)
	if t != nil {
		err = os.Remove(t.filePath)
	} else {
		err = os.Remove(filepath.Join(me.dir, name+FileExt))
	}
	return
}

func (me *conn) doInsertInto(name string, rec interface{}) (res driver.Result, err error) {
	var t *table
	if t, err = me.tables.get(name); err == nil {
		res, err = t.insert(m(rec))
	}
	return
}

func (me *conn) doSelectFrom(name string, where interface{}) (res driver.Rows, err error) {
	var t *table
	if t, err = me.tables.get(name); err == nil {
		var recs map[string]M
		if recs, err = t.fetch(m(where)); err == nil {
			res = newRows(recs)
		}
	}
	return
}

// Begin starts and returns a new transaction.
func (me *conn) Begin() (tx driver.Tx, err error) {
	me.tx = newTx(me)
	tx = me.tx
	return
}

func (me *conn) Close() (err error) {
	me.tx = nil
	me.tables.init(me, true)
	return
}

func (me *conn) Prepare(query string) (stmt driver.Stmt, err error) {
	stmt, err = newStmt(me, query)
	return
}
