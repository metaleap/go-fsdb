package fsdb

import (
	"database/sql/driver"

	"github.com/go-utils/ugo"
)

const (
	//	This is not used as a object/hash property/entry in final storage
	//	but may be used in selectFrom/deleteFrom/updateWhere queries:
	IdField = "__id"
)

var (
	//	Defaults to false. See `M.Match` method for explanation.
	StrCmp bool
)

//	Function that marshals an in-memory data table to a local file.
type Marshal func(v interface{}) ([]byte, error)

//	Function that unmarshals an in-memory data table from a local file.
type Unmarshal func(data []byte, v interface{}) error

//	Implements the `database/sql/driver.Driver` interface.
type Driver struct {
	marshal   Marshal
	unmarshal Unmarshal
	fileExt   string
	connCache map[string]*conn
}

//	Creates a new `*fsdb.Driver` and returns it.
//
//	`fileExt` -- the file name extension used by `me` for reading and writing table data files.
//
//	`connectionCaching` -- see the `Driver.ConnectionCaching` method for details.
//
//	`marshal` and `unmarshal` implement the actual encoding from and to binary or textual data table files.
func NewDriver(fileExt string, connectionCaching bool, marshal Marshal, unmarshal Unmarshal) (me *Driver) {
	if me = (&Driver{fileExt: fileExt, marshal: marshal, unmarshal: unmarshal}); connectionCaching {
		me.connCache = map[string]*conn{}
	}
	return
}

//	Implements the `database/sql/driver.Driver.Open` interface method.
func (me *Driver) Open(dirPath string) (_ driver.Conn, err error) {
	var conn *conn
	if me.connCache != nil {
		conn, _ = me.connCache[dirPath]
	}
	if conn == nil {
		conn, err = newConn(me, dirPath)
		if me.connCache != nil {
			me.connCache[dirPath] = conn
		}
	}
	return conn, err
}

//	Returns whether connection caching was enabled for `me` via `fsdb.NewDriver`.
//
//	You should do so if your use-case entails many parallel go-routines concurrently
//	operating on the same database via their own `sql.DB` connections:
//
//	that's because, while the standard `sql` package does provide "connection pooling",
//	this is not sensible for `fsdb`, as each `fsdb.conn` does hold its own complete copy
//	of data files in-memory.
//
//	All table writes are `sync.Mutex`-locking as necessary ONLY if connection caching is enabled.
func (me *Driver) ConnectionCaching() bool {
	return me.connCache != nil
}

//	Not necessary for normal use: `me` persists tables that are being
//	written to via `insertInto`/`updateWhere`/`deleteFrom` immediately, or in
//	a transaction context, at the next `Tx.Commit`.
//
//	If no `tableNames` are specifed, persists ALL tables belonging to `dbConn`,
//	otherwise only the specified tables are persisted.
func (me *Driver) PersistAll(dbConn driver.Conn, tableNames ...string) (err error) {
	if c, _ := dbConn.(*conn); c != nil {
		err = c.tables.persistAll(tableNames...)
	} else {
		err = errf("fsdb.PersistAll() needs a *fsdb.conn, not a %#v", dbConn)
	}
	return
}

//	Not necessary for normal use: `me` lazily auto-reloads tables that have been
//	modified on disk if such a data-refresh is necessary for the current operation.
//
//	If no `tableNames` are specifed, reloads all tables belonging to `dbConn`.
//
//	In any event, the reload always includes new-on-disk table data files not
//	previously loaded, and removes in-memory data tables no longer on disk.
func (me *Driver) ReloadAll(dbConn driver.Conn, tableNames ...string) (err error) {
	if c, _ := dbConn.(*conn); c != nil {
		err = c.tables.reloadAll(tableNames...)
	} else {
		err = errf("fsdb.ReloadAll() needs a *fsdb.conn, not a %#v", dbConn)
	}
	return
}

//	A convenience short-hand. Used for actual records, as well as `where` criteria (in
//	selectFrom, deleteFrom, updateWhere) and `set` data (in insertInto and updateWhere).
type M map[string]interface{}

//	If `me` is a record, returns whether it matches the specified criteria.
//
//	- recID: the `__id` of `me`, if any (since this isn't stored in the record itself)
//
//	- filters: one or more criteria, `AND`-ed together. Each criteria is a slice of possible values, `OR`-ed together
//
//	- strCmp: if `false`, just compares `interface{}==interface{}`. If `true`, also compares `fmt.Sprintf("%v", interface{}) == fmt.Sprintf("%v", interface{})`
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
		if fn != IdField || len(recId) > 0 {
			if !matchAny(fn, ugo.Ifx(fn == IdField, recId, me[fn]), interfaces(fx)) {
				return
			}
		}
	}
	isMatch = true
	return
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
