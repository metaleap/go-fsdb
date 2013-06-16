package jsondb

import (
	"database/sql/driver"
	"os"
	"path/filepath"
	"time"

	uio "github.com/metaleap/go-util/io"
)

type conn struct {
	drv      *drv
	tx       *tx
	initTime time.Time
	dir      string
	tables   tables
}

func newConn(drv *drv, dir string) (me *conn, err error) {
	me = &conn{drv: drv, dir: dir, initTime: time.Now()}
	err = me.tables.init(me)
	return
}

func (me *conn) CreateTable(name string) (err error) {
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

func (me *conn) DropTable(name string) (err error) {
	t := me.tables.all[name]
	delete(me.tables.all, name)
	if t != nil {
		err = os.Remove(t.filePath)
	} else {
		err = os.Remove(filepath.Join(me.dir, name+FileExt))
	}
	return
}

func (me *conn) InsertInto(name string, rec interface{}) (result driver.Result, err error) {
	var t *table
	m, _ := rec.(map[string]interface{})
	if t, err = me.tables.get(name); err == nil {
		result, err = t.insert(m)
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
	return
}

func (me *conn) Prepare(query string) (stmt driver.Stmt, err error) {
	stmt, err = newStmt(me, query)
	return
}
