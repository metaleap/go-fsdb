package jsondb

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
	return
}

func (me *stmt) NumInput() (num int) {
	return
}

func (me *stmt) Exec(args []driver.Value) (res driver.Result, err error) {
	switch me.cmd {
	case cmdCreateTable:
		err = me.conn.CreateTable(me.table)
	case cmdDropTable:
		err = me.conn.DropTable(me.table)
	case cmdInsertInto:
		res, err = me.conn.InsertInto(me.table, me.query["set"])
	default:
		err = errf("Nothing to Exec()")
	}
	if err == nil && res == nil {
		res = &result{}
	}
	return
}

func (me *stmt) Query(args []driver.Value) (res driver.Rows, err error) {
	return
}
