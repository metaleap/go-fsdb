package fsdb

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
)

type stmt struct {
	conn       *conn
	query      M
	cmd, table string
}

func newStmt(conn *conn, query string) (me *stmt, err error) {
	me = &stmt{conn: conn}
	if !strings.HasPrefix(query, "{") {
		query = "{" + query
	}
	if !strings.HasSuffix(query, "}") {
		query = query + "}"
	}
	if err = json.Unmarshal([]byte(query), &me.query); err == nil {
		for k, v := range me.query {
			switch k {
			case cmdCreateTable, cmdDropTable, cmdInsertInto, cmdSelectFrom, cmdUpdateWhere, cmdDeleteFrom:
				me.cmd, me.table = k, strf("%v", v)
				break // for
			}
		}
	}
	if err != nil {
		me = nil
	}
	return
}

func (me *stmt) Close() (err error) {
	me.query = nil
	return
}

func (me *stmt) NumInput() (num int) {
	return
}

func (me *stmt) Exec(args []driver.Value) (res driver.Result, err error) {
	switch me.cmd {
	case cmdCreateTable:
		err = me.conn.doCreateTable(me.table)
	case cmdDropTable:
		err = me.conn.doDropTable(me.table)
	case cmdInsertInto:
		res, err = me.conn.doInsertInto(me.table, me.query["set"])
	case cmdDeleteFrom:
		res, err = me.conn.doDeleteFrom(me.table, me.query["where"])
	case cmdUpdateWhere:
		res, err = me.conn.doUpdateWhere(me.table, me.query["set"], me.query["where"])
	default:
		err = errf("Cannot Exec() via '%s', try Query()", me.cmd)
	}
	if err == nil && res == nil {
		res = &result{}
	}
	return
}

func (me *stmt) Query(args []driver.Value) (res driver.Rows, err error) {
	switch me.cmd {
	case cmdSelectFrom:
		res, err = me.conn.doSelectFrom(me.table, me.query["where"])
	default:
		err = errf("Cannot Query() via '%s', try Exec()", me.cmd)
	}
	return
}
