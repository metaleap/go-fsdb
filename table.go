package jsondb

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	uio "github.com/metaleap/go-util/io"
)

type tables struct {
	conn *conn
	all  map[string]*table
}

func (me *tables) init(conn *conn) (err error) {
	me.conn = conn
	if errs := uio.WalkFilesIn(conn.dir, func(filePath string) bool {
		if strings.HasSuffix(strings.ToLower(filePath), FileExt) {
			fn := filepath.Base(filePath)
			_, err = me.get(fn[:len(fn)-len(FileExt)])
		}
		return err == nil
	}); len(errs) > 0 && err == nil {
		err = errs[0]
	}
	return
}

func (me *tables) get(name string) (t *table, err error) {
	if t = me.all[name]; t == nil {
		t = &table{name: name, filePath: filepath.Join(me.conn.dir, name+FileExt)}
		if err = t.lazyReload(); err == nil {
			me.all[t.name] = t
		} else {
			t = nil
		}
	}
	return
}

type table struct {
	lastLoad       time.Time
	name, filePath string
	recs           M
}

func (me *table) lazyReload() (err error) {
	var fi os.FileInfo
	if fi, err = os.Stat(me.filePath); err == nil && fi.ModTime().UnixNano() > me.lastLoad.UnixNano() {
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

func (me *table) insert(rec M) (res *result, err error) {
	id := time.Now().UnixNano()
	sid := strf("%v", id)
	if rec == nil {
		rec = M{}
	}
	if err = me.lazyReload(); err == nil {
		me.recs[sid] = rec
		if err = me.persist(); err == nil {
			res = &result{AffectedRows: 1, InsertedLast: id}
		} else {
			delete(me.recs, sid)
		}
	}
	return
}

func (me *table) persist() (err error) {
	var raw []byte
	if raw, err = json.Marshal(me.recs); err == nil {
		if err = uio.WriteBinaryFile(me.filePath, raw); err == nil {
			me.lastLoad = time.Now()
		}
	}
	return
}
