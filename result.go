package jsondb

type result struct {
	InsertedLast, AffectedRows int64
}

func (me *result) LastInsertId() (id int64, err error) {
	id = me.InsertedLast
	return
}

func (me *result) RowsAffected() (num int64, err error) {
	num = me.AffectedRows
	return
}
