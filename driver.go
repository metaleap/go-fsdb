package jsondb

import (
	"database/sql/driver"

	ugo "github.com/metaleap/go-util"
)

const (
	//	Always use this for:
	//	- sql.Register(jsondb.DriverName, jsondb.NewDriver())
	//	- sql.Open(jsondb.DriverName, yourDbDirPath)
	DriverName = "github.com/metaleap/go-jsondb"

	//	This is not used as a JSON hash/object property in final storage
	//	but may be used in selectFrom/deleteFrom/updateWhere queries:
	IdField = "__id"
)

var (
	//	File name extension for data files. If this is to be customized,
	//	set this in your init() or at least before starting to use jsondb.
	FileExt = ".jsondbt"

	//	Defaults to false. See `M.Match()` method for explanation.
	StrCmp bool

	connCache map[string]driver.Conn
)

//	A convenience short-hand. Used for actual records, as well as `where` criteria (in
//	selectFrom, deleteFrom, updateWhere) and `set` data (in insertInto and updateWhere).
type M map[string]interface{}

//	If me is a record, returns whether it matches the specified criteria.
//
//	- recID: the __id of me, if any (since this isn't stored in the record itself)
//
//	- filters: one or more criteria, AND-ed together. Each criteria is a slice of possible values, OR-ed together
//
//	- strCmp: if false, just compares `interface{}==interface{}`. If true, also compares `strf("%v", interface{}) == strf("%v", interface{})`
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

type drv struct {
}

//	Usage: `sql.Register(jsondb.DriverName, jsondb.NewDriver())`
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

//	Returns whether connection caching is currently enabled, defaulting to false.
//
//	Connection caching can be enabled via `SetConnectionCaching(bool)`.
//
//	You should do so if your use-case entails many parallel go-routines concurrently
//	operating on the same database via their own sql.DB connections:
//
//	that's because, while the standard `sql` package does provide "connection pooling",
//	this is not sensible for jsondb, as each jsondb.conn does hold its own complete copy
//	of data files in-memory.
//
//	Table writes are Mutex-locking as necessary if (and only if) connection caching is enabled.
func ConnectionCaching() bool {
	return connCache != nil
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

//	Not necessary for normal use: jsondb persists tables that are being
//	written to via insertInto/updateWhere/deleteFrom immediately, or in
//	a transaction context, at the next Tx.Commit().
//
//	If no tableNames are specifed, persists all tables belonging to dbConn,
//	else only the specified tables are persisted.
func PersistAll(dbConn driver.Conn, tableNames ...string) (err error) {
	if c, _ := dbConn.(*conn); c != nil {
		err = c.tables.persistAll(tableNames...)
	} else {
		err = errf("jsondb.PersistAll() needs a *jsondb.conn, not a %#v", dbConn)
	}
	return
}

//	Not necessary for normal use: jsondb lazily auto-reloads tables that have been
//	modified on disk if such a data-refresh is necessary for the current operation.
//
//	If no tableNames are specifed, reloads all tables belonging to dbConn.
//
//	The reload includes new-on-disk table data files not previously loaded, and
//	removes in-memory data tables no longer on disk.
func ReloadAll(dbConn driver.Conn, tableNames ...string) (err error) {
	if c, _ := dbConn.(*conn); c != nil {
		err = c.tables.reloadAll(tableNames...)
	} else {
		err = errf("jsondb.ReloadAll() needs a *jsondb.conn, not a %#v", dbConn)
	}
	return
}

//	Enables or disables connection caching depending on the specified bool.
//	For details on connection caching, see `ConnectionCaching()`.
func SetConnectionCaching(enableCaching bool) {
	if isEnabled := ConnectionCaching(); isEnabled != enableCaching {
		if enableCaching {
			connCache = map[string]driver.Conn{}
		} else {
			connCache = nil
		}
	}
}
