package jsondb

import (
	"path/filepath"
	"sync"

	usl "github.com/metaleap/go-util/slice"
)

type tables struct {
	sync.Mutex
	conn *conn
	all  map[string]*table
}

func (me *tables) init(conn *conn, close bool) (err error) {
	me.conn, me.all = conn, map[string]*table{}
	if !close {
		tableNames, errs := conn.enumTableFiles()
		if len(errs) > 0 {
			err = errs[0]
		} else {
			for _, tn := range tableNames {
				if _, err = me.get(tn); err != nil {
					break
				}
			}
		}
	}
	return
}

func (me *tables) get(name string) (t *table, err error) {
	if t = me.all[name]; t == nil {
		t = &table{conn: me.conn, name: name, filePath: filepath.Join(me.conn.dir, name+FileExt)}
		if err = t.reload(true); err == nil {
			if ConnectionCaching() {
				me.Lock()
				defer me.Unlock()
			}
			me.all[t.name] = t
		} else {
			t = nil
		}
	}
	return
}

func (me *tables) persistAll(tableNames ...string) (err error) {
	var e error
	for name, table := range me.all {
		if len(tableNames) == 0 || usl.StrHas(tableNames, name) {
			if e = table.persist(); e != nil && err == nil {
				err = e
			}
		}
	}
	return
}

func (me *tables) reloadAll(tableNames ...string) (err error) {
	var (
		e      error
		errs   []error
		tnames []string
	)
	if tnames, errs = me.conn.enumTableFiles(); len(errs) > 0 {
		err = errs[0]
	} else {
		for _, tn := range tnames {
			if _, err = me.get(tn); err != nil {
				break
			}
		}
	}
	if err == nil {
		if len(tableNames) == 0 {
			tableNames = tnames
		}
		if ConnectionCaching() {
			me.Lock()
			defer me.Unlock()
		}
		for name, table := range me.all {
			if !usl.StrHas(tnames, name) {
				delete(me.all, name)
			} else if usl.StrHas(tableNames, name) {
				if e = table.reload(false); e != nil && err == nil {
					err = e
				}
			}
		}
	}
	return
}