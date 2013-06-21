# jsondb
--
    import "github.com/metaleap/go-jsondb"

A "database driver" (compatible with Go's `database/sql` package)
that's using a directory of JSON text files as a database of "tables".

Does not implement the finer details of *real* databases (such as
relational integrity, cascading deletes, ACID etc.) --- the *only* use-case
is **"faster prototyping of a DB app without needing to mess with a real-world
DB right now"** and easily inspectable, human-readable data table files.

Connection string: any (file-system) directory path.

SQL syntax: none. Instead, the driver uses simple JSON strings such
as `{"createTable": "FooBars"}`. Use the documented `StmtGen` methods
(ie. `jsondb.S.CreateTable()` and friends) to easily generate statements
for use with sql.Exec() and sql.Query(), whether via a `sql.DB` or a `sql.Tx`.

I didn't see the use in parsing real SQL syntax --- each real-world DB has its
own syntax quirks, so when moving on from jsondb to the real DB, I'd have to adapt
most/all SQL statements anyway. This way, it's guaranteed that I'll have to do so.

Each jsondb connection maintains a full in-memory copy of data table files,
persisting and reloading as necessary, see documentation on the global exported
`PersistAll()` and `ReloadAll()` functions for details.

Connection pooling/caching: you can use Go's built-in pooling if you're fine
with potentially many duplicate in-memory copies of the same data tables.
See documentation on the global exported `ConnectionCaching()` function for details.

Transactions: not quite behaving like normal database transactions

## Usage

```go
const (
	//	Always use this for:
	//	- sql.Register(jsondb.DriverName, jsondb.NewDriver())
	//	- sql.Open(jsondb.DriverName, yourDbDirPath)
	DriverName = "github.com/metaleap/go-jsondb"

	//	This is not used as a JSON hash/object property in final storage
	//	but may be used in selectFrom/deleteFrom/updateWhere queries:
	IdField = "__id"
)
```

```go
var (
	//	File name extension for data files. If this is to be customized,
	//	set this in your init() or at least before starting to use jsondb.
	FileExt = ".jsondbt"

	//	Defaults to false. See `M.Match()` method for explanation.
	StrCmp bool
)
```

#### func  ConnectionCaching

```go
func ConnectionCaching() bool
```
Returns whether connection caching is currently enabled, defaulting to false.

Connection caching can be enabled via `SetConnectionCaching(bool)`.

You should do so if your use-case entails many parallel go-routines concurrently
operating on the same database via their own sql.DB connections:

that's because, while the standard `sql` package does provide "connection
pooling", this is not sensible for jsondb, as each jsondb.conn does hold its own
complete copy of data files in-memory.

Table writes are Mutex-locking as necessary if (and only if) connection caching
is enabled.

#### func  NewDriver

```go
func NewDriver() (me driver.Driver)
```
Usage: `sql.Register(jsondb.DriverName, jsondb.NewDriver())`

#### func  PersistAll

```go
func PersistAll(dbConn driver.Conn, tableNames ...string) (err error)
```
Not necessary for normal use: jsondb persists tables that are being written to
via insertInto/updateWhere/deleteFrom immediately, or in a transaction context,
at the next Tx.Commit().

If no tableNames are specifed, persists all tables belonging to dbConn, else
only the specified tables are persisted.

#### func  ReloadAll

```go
func ReloadAll(dbConn driver.Conn, tableNames ...string) (err error)
```
Not necessary for normal use: jsondb lazily auto-reloads tables that have been
modified on disk if such a data-refresh is necessary for the current operation.

If no tableNames are specifed, reloads all tables belonging to dbConn.

The reload includes new-on-disk table data files not previously loaded, and
removes in-memory data tables no longer on disk.

#### func  SetConnectionCaching

```go
func SetConnectionCaching(enableCaching bool)
```
Enables or disables connection caching depending on the specified bool. For
details on connection caching, see `ConnectionCaching()`.

#### type M

```go
type M map[string]interface{}
```

A convenience short-hand. Used for actual records, as well as `where` criteria
(in selectFrom, deleteFrom, updateWhere) and `set` data (in insertInto and
updateWhere).

#### func (M) Match

```go
func (me M) Match(recId string, filters M, strCmp bool) (isMatch bool)
```
If me is a record, returns whether it matches the specified criteria.

- recID: the __id of me, if any (since this isn't stored in the record itself)

- filters: one or more criteria, AND-ed together. Each criteria is a slice of
possible values, OR-ed together

- strCmp: if false, just compares `interface{}==interface{}`. If true, also
compares `strf("%v", interface{}) == strf("%v", interface{})`

#### type StmtGen

```go
type StmtGen struct {
}
```

Stateless struct, use via the exported `S` global singleton.

```go
var (
	//	Generates statements for sql.Exec() and sql.Query().
	S StmtGen
)
```

#### func (*StmtGen) CreateTable

```go
func (_ *StmtGen) CreateTable(name string) string
```
Generates a `{"createTable":name}` statement

#### func (*StmtGen) DeleteFrom

```go
func (me *StmtGen) DeleteFrom(name string, where M) string
```
Generates a `{"deleteFrom":name, "where": where}` statement

#### func (*StmtGen) DropTable

```go
func (_ *StmtGen) DropTable(name string) string
```
Generates a `{"dropTable":name}` statement

#### func (*StmtGen) InsertInto

```go
func (me *StmtGen) InsertInto(name string, rec M) string
```
Generates a `{"insertInto":name, "set": rec}` statement

#### func (*StmtGen) SelectFrom

```go
func (me *StmtGen) SelectFrom(name string, where M) string
```
Generates a `{"selectFrom":name, "where": where}` statement

#### func (*StmtGen) UpdateWhere

```go
func (me *StmtGen) UpdateWhere(name string, set, where M) string
```
Generates a `{"updateWhere":name, "set": set, "where": where}` statement

--
**godocdown** http://github.com/robertkrimen/godocdown