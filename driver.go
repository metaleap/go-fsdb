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
	//	Used in `selectWhere` queries, defaults to false. See `M.Match` method for explanation.
	StrCmp bool
)

//	Function that marshals an in-memory data table to a local file.
type Marshal func(v interface{}) ([]byte, error)

//	Function that unmarshals an in-memory data table from a local file.
type Unmarshal func(data []byte, v interface{}) error

//	Implements the `database/sql/driver.Driver` interface.
type drv struct {
	marshal   Marshal
	unmarshal Unmarshal
	fileExt   string
	connCache map[string]*conn
}

//	Creates a new `database/sql/driver.Driver` and returns it.
//
//	- `fileExt` -- the file name extension used for reading and writing table data files.
//
//	- `marshal`/`unmarshal` implement the actual decoding-from/encoding-to binary or textual data table files.
//
//	- `connectionCaching` -- if `true`, enables connection-caching for this `Driver`.
//	You should do so if your use-case entails many parallel go-routines concurrently
//	operating on the same database via their own `sql.DB` connections:
//	that's because, while the standard `sql` package does provide "connection pooling",
//	this is not sensible for `fsdb`, as each `fsdb.conn` does hold its own complete copies
//	of all table data files in-memory.
//	(All table writes are `sync.Mutex`-locking as necessary ONLY if connection-caching is enabled.)
func NewDriver(fileExt string, connectionCaching bool, marshal Marshal, unmarshal Unmarshal) driver.Driver {
	me := &drv{fileExt: fileExt, marshal: marshal, unmarshal: unmarshal}
	if connectionCaching {
		me.connCache = map[string]*conn{}
	}
	return me
}

//	Returns whether connection caching was enabled for `me`.
func (me *drv) ConnectionCaching() bool {
	return me.connCache != nil
}

//	Implements the `database/sql/driver.Driver.Open` interface method.
func (me *drv) Open(dirPath string) (_ driver.Conn, err error) {
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
