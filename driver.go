package jsondb

import (
	"database/sql/driver"
)

const DriverName = "github.com/metaleap/go-jsondb"

type M map[string]interface{}

var (
	FileExt = ".jsondbt"

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
