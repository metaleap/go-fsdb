package jsondb

type tx struct {
}

func (me *tx) Commit() (err error) {
	return
}

func (me *tx) Rollback() (err error) {
	return
}
