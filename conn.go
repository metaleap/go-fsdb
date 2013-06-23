package fsdb

import (
	"database/sql/driver"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-utils/ufs"
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

func (me *conn) enumTableFiles() (tableNames []string, errs []error) {
	errs = ufs.WalkFilesIn(me.dir, func(filePath string) bool {
		if strings.HasSuffix(filePath, me.drv.fileExt) {
			fn := filepath.Base(filePath)
			tableNames = append(tableNames, fn[:len(fn)-len(me.drv.fileExt)])
		}
		return true
	})
	return
}

func (me *conn) doCreateTable(name string) (err error) {
	if _, ok := me.tables.all[name]; !ok {
		if fp := filepath.Join(me.dir, name+me.drv.fileExt); ufs.FileExists(fp) {
			err = errf("Cannot create table '%s': already exists", name)
		} else {
			var data []byte
			if data, err = me.drv.marshal(M{}); err == nil {
				err = ufs.WriteBinaryFile(fp, data)
			}
		}
	}
	if err == nil {
		_, err = me.tables.get(name)
	}
	return
}

func (me *conn) doDeleteFrom(name string, where interface{}) (res driver.Result, err error) {
	var t *table
	if t, err = me.tables.get(name); err == nil {
		var recs map[string]M
		if recs, err = t.fetch(m(where)); err == nil {
			rids := make([]string, 0, len(recs))
			for rid, _ := range recs {
				rids = append(rids, rid)
			}
			res, err = t.delete(rids)
		}
	}
	return
}

func (me *conn) doDropTable(name string) (err error) {
	t := me.tables.all[name]
	delete(me.tables.all, name)
	if t != nil {
		err = os.Remove(t.filePath)
	} else {
		err = os.Remove(filepath.Join(me.dir, name+me.drv.fileExt))
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

func (me *conn) doUpdateWhere(name string, set, where interface{}) (res driver.Result, err error) {
	var (
		t   *table
		num int64
	)
	upd := m(set)
	if t, err = me.tables.get(name); err == nil && len(upd) > 0 {
		if err = t.reload(true); err == nil {
			var recs map[string]M
			if recs, err = t.fetch(m(where)); err == nil {
				for _, rec := range recs {
					for fn, fv := range upd {
						rec[fn] = fv
					}
					num++
				}
				if num > 0 {
					err = t.persist()
				}
			}
		}
	}
	if err == nil {
		res = &result{AffectedRows: num}
	}
	return
}

func (me *conn) Prepare(query string) (stmt driver.Stmt, err error) {
	stmt, err = newStmt(me, query)
	return
}
