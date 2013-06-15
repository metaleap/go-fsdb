package jsondb

import (
	"encoding/json"
	"fmt"
)

const (
	cmdCreateTable = "createTable"
	cmdDropTable   = "dropTable"
	cmdInsertInto  = "insertInto"
	cmdSelectFrom  = "selectFrom"
	cmdUpdateWhere = "updateWhere"
	cmdDeleteFrom  = "deleteFrom"
)

var S StmtGen

func errf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func strf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

type StmtGen struct {
}

func (_ *StmtGen) genStmt(cmd, name string, set, where M) string {
	M := M{cmd: name, "set": set, "where": where}
	raw, _ := json.Marshal(M) // marshaling a map won't err except for brute-force malfeasance
	return string(raw)
}

func (_ *StmtGen) CreateTable(name string) string {
	return S.genStmt(cmdCreateTable, name, nil, nil)
}

func (_ *StmtGen) DropTable(name string) string {
	return S.genStmt(cmdDropTable, name, nil, nil)
}

func (me *StmtGen) InsertInto(name string, rec M) string {
	return S.genStmt(cmdInsertInto, name, rec, nil)
}

func (me *StmtGen) SelectFrom(name string, where M) string {
	return S.genStmt(cmdSelectFrom, name, nil, where)
}

func (me *StmtGen) DeleteFrom(name string, where M) string {
	return S.genStmt(cmdDeleteFrom, name, nil, where)
}

func (me *StmtGen) UpdateWhere(name string, set, where M) string {
	return S.genStmt(cmdUpdateWhere, name, set, where)
}
