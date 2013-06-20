package jsondb

import (
	"database/sql/driver"

	ugo "github.com/metaleap/go-util"
)

const (
	DriverName = "github.com/metaleap/go-jsondb"

	idField = "__id"
)

type M map[string]interface{}

func (me M) Match(recId string, filters M, strCmp bool) (isMatch bool) {
	matchAny := func(fn string, rvx interface{}, fvx []interface{}) bool {
		for _, fv := range fvx {
			if rvx == fv || strCmp && strf("%v", rvx) == strf("%v", fv) {
				return true
			}
		}
		return false
	}
	for fn, fx := range filters {
		if fn != idField || len(recId) > 0 {
			if !matchAny(fn, ugo.Ifx(fn == idField, recId, me[fn]), interfaces(fx)) {
				return
			}
		}
	}
	isMatch = true
	return
}

var (
	FileExt = ".jsondbt"
	StrCmp  bool

	connCache map[string]driver.Conn
)

type drv struct {
}

func NewDriver() (me driver.Driver) {
	me = &drv{}
	return
}

func (me *drv) Open(dirPath string) (conn driver.Conn, err error) {
	if connCache != nil {
		conn, _ = connCache[dirPath]
	}
	if conn == nil {
		conn, err = newConn(me, dirPath)
		if connCache != nil {
			connCache[dirPath] = conn
		}
	}
	return
}

func ConnectionCaching() bool {
	return connCache != nil
}

func interfaces(ix interface{}) (slice []interface{}) {
	var ok bool
	if slice, ok = ix.([]interface{}); (!ok) && ix != nil {
		slice = append(slice, ix)
	}
	return
}

func m(ix interface{}) (m M) {
	if m, _ = ix.(M); m == nil {
		if mm := ix.(map[string]interface{}); mm != nil {
			m = M(mm)
		}
	}
	return
}

func PersistAll(connection driver.Conn, tableNames ...string) (err error) {
	if c, _ := connection.(*conn); c != nil {
		err = c.tables.persistAll(tableNames...)
	} else {
		err = errf("jsondb.PersistAll() needs a *jsondb.conn, not a %#v", connection)
	}
	return
}

func ReloadAll(connection driver.Conn, tableNames ...string) (err error) {
	if c, _ := connection.(*conn); c != nil {
		err = c.tables.reloadAll(tableNames...)
	} else {
		err = errf("jsondb.ReloadAll() needs a *jsondb.conn, not a %#v", connection)
	}
	return
}

func SetConnectionCaching(enableCaching bool) {
	if isEnabled := ConnectionCaching(); isEnabled != enableCaching {
		if enableCaching {
			connCache = map[string]driver.Conn{}
		} else {
			connCache = nil
		}
	}
}
