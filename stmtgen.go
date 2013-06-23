package fsdb

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

func errf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func strf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func genStmt(cmd, name string, set, where M) string {
	M := M{cmd: name, "set": set, "where": where}
	raw, _ := json.Marshal(M) // marshaling a map won't err except for brute-force malfeasance
	return string(raw)
}

//	Generates a `{"createTable":name}` statement.
func StmtCreateTable(name string) string {
	return genStmt(cmdCreateTable, name, nil, nil)
}

//	Generates a `{"dropTable":name}` statement.
func StmtDropTable(name string) string {
	return genStmt(cmdDropTable, name, nil, nil)
}

//	Generates a `{"insertInto":name, "set": rec}` statement.
func StmtInsertInto(name string, rec M) string {
	return genStmt(cmdInsertInto, name, rec, nil)
}

//	Generates a `{"selectFrom":name, "where": where}` statement.
func StmtSelectFrom(name string, where M) string {
	return genStmt(cmdSelectFrom, name, nil, where)
}

//	Generates a `{"deleteFrom":name, "where": where}` statement.
func StmtDeleteFrom(name string, where M) string {
	return genStmt(cmdDeleteFrom, name, nil, where)
}

//	Generates a `{"updateWhere":name, "set": set, "where": where}` statement.
func StmtUpdateWhere(name string, set, where M) string {
	return genStmt(cmdUpdateWhere, name, set, where)
}
