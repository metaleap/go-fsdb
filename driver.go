package jsondb

import (
	"database/sql/driver"
)

const DriverName = "github.com/metaleap/go-jsondb"

type M map[string]interface{}

var (
	FileExt = ".jsondbt"
)

type drv struct {
}

func NewDriver() (me driver.Driver) {
	me = &drv{}
	return
}

func (me *drv) Open(dirPath string) (conn driver.Conn, err error) {
	conn, err = newConn(me, dirPath)
	return
}
