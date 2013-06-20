package jsondb

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	uio "github.com/metaleap/go-util/io"
	usl "github.com/metaleap/go-util/slice"
)

type tables struct {
	conn *conn
	all  map[string]*table
}

func (me *tables) init(conn *conn, close bool) (err error) {
	me.conn, me.all = conn, map[string]*table{}
	if !close {
		if errs := uio.WalkFilesIn(conn.dir, func(filePath string) bool {
			if strings.HasSuffix(strings.ToLower(filePath), FileExt) {
				fn := filepath.Base(filePath)
				_, err = me.get(fn[:len(fn)-len(FileExt)])
			}
			return err == nil
		}); len(errs) > 0 && err == nil {
			err = errs[0]
		}
	}
	return
}

func (me *tables) get(name string) (t *table, err error) {
	if t = me.all[name]; t == nil {
		t = &table{conn: me.conn, name: name, filePath: filepath.Join(me.conn.dir, name+FileExt)}
		if err = t.reload(true); err == nil {
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
	var e error
	for name, table := range me.all {
		if len(tableNames) == 0 || usl.StrHas(tableNames, name) {
			if e = table.reload(false); e != nil && err == nil {
				err = e
			}
		}
	}
	return
}

type table struct {
	conn           *conn
	lastLoad       time.Time
	name, filePath string
	recs           M
}

func (me *table) fetch(where M) (recs map[string]M, err error) {
	var rec M
	recs = map[string]M{}
	// fast map[id] pre-fetches if where has id query:
	if idQuery := interfaces(where[IdField]); len(idQuery) > 0 {
		var ok bool
		var str string
		for _, id := range idQuery {
			if str, ok = id.(string); ok {
				if rec = m(me.recs[str]); rec != nil && rec.Match("", where, StrCmp) {
					recs[str] = rec
				}
			}
		}
	} else {
		for rid, rix := range me.recs {
			if rec = m(rix); rec != nil {
				if rec.Match(rid, where, StrCmp) {
					recs[rid] = rec
				}
			}
		}
	}
	return
}

func (me *table) reload(lazy bool) (err error) {
	var fi os.FileInfo
	if fi, err = os.Stat(me.filePath); err == nil && ((!lazy) || me.recs == nil || me.lastLoad.UnixNano() == 0 || (me.conn.tx == nil && fi.ModTime().UnixNano() > me.lastLoad.UnixNano())) {
		var raw []byte
		if raw, err = ioutil.ReadFile(me.filePath); err == nil {
			recs := M{}
			if err = json.Unmarshal(raw, &recs); err == nil {
				me.recs, me.lastLoad = recs, time.Now()
			}
		}
	}
	return
}

func (me *table) delete(recIDs []string) (res *result, err error) {
	var (
		num int64
		ok  bool
	)
	if err = me.reload(true); err == nil && len(recIDs) > 0 {
		for _, rid := range recIDs {
			if _, ok = me.recs[rid]; ok {
				delete(me.recs, rid)
				num++
			}
		}
		if num > 0 {
			err = me.persist()
		}
	}
	if err == nil {
		res = &result{AffectedRows: num}
	}
	return
}

func (me *table) insert(rec M) (res *result, err error) {
	if rec == nil {
		err = errf("Cannot insert nil")
	} else if err = me.reload(true); err == nil {
		id := int64(len(me.recs))
		sid := strf("%v", id)
		if _, ok := me.recs[sid]; ok {
			err = errf("Cannot insert: duplicate record ID")
		} else {
			me.recs[sid] = rec
			if err = me.persist(); err == nil {
				res = &result{AffectedRows: 1, InsertedLast: id}
			} else {
				delete(me.recs, sid)
			}
		}
	}
	return
}

func (me *table) persist() (err error) {
	if me.conn.tx == nil {
		var raw []byte
		if raw, err = json.MarshalIndent(me.recs, "", " "); err == nil {
			if err = uio.WriteBinaryFile(me.filePath, raw); err == nil {
				me.lastLoad = time.Now()
			}
		}
	} else {
		me.conn.tx.tables[me] = true
	}
	return
}
