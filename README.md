# fsdb
--
    import "github.com/metaleap/go-fsdb"

A "database driver" (compatible with Go's `database/sql` package) that's using a
local directory of files as a database of "tables".

Does not implement the finer details of *real* databases (such as relational
integrity, cascading deletes, ACID etc.) --- the *only* use-case is **"faster
prototyping of a DB app without needing to mess with a real-world DB right
now"**, based on easily inspectable, human-readable data table files.

**Connection string**: any (file-system) directory path.

**SQL syntax**: none. Instead, the driver uses simple JSON strings such as
`{"createTable": "FooBars"}`. Use the documented `StmtGen` methods (ie.
`fsdb.S.CreateTable` and friends) to easily generate statements for use with
sql.Exec() and sql.Query(), whether via a `sql.DB` or a `sql.Tx`.

I didn't see the use in parsing real SQL syntax --- each real-world DB has its
own syntax quirks, so when moving on from `fsdb` to the real DB, I'd have to
adapt most/all SQL statements anyway. This way, it's guaranteed that I'll have
to do so.

**Connection pooling/caching**: works "so-so" with Go's built-in pooling: with
many redundant in-memory copies of the same data tables, as per below. See
documentation on the `Driver.ConnectionCaching` method for details.

Each `fsdb`-driven `sql.DB` connection maintains a full in-memory copy of its
data table files, auto-persisting and auto-reloading as necessary -- see
documentation on the `Driver.PersistAll` and `Driver.ReloadAll` methods for
details.

**Transactions**: they're a useful hack at best -- the idea here is for batching
multiple writes together. Each `insertInto`/`updateWhere`/`deleteFrom` would
normally persist the full table to disk immediately. But in the context of a
transaction, they won't -- only the final `Tx.Commit` will flush participating
tables to disk.

## Usage

```go
const (
	//	This is not used as a object/hash property/entry in final storage
	//	but may be used in selectFrom/deleteFrom/updateWhere queries:
	IdField = "__id"
)
```

```go
var (
	//	Defaults to false. See `M.Match` method for explanation.
	StrCmp bool
)
```

#### type Driver

```go
type Driver struct {
}
```

Implements the `database/sql/driver.Driver` interface.

#### func  NewDriver

```go
func NewDriver(fileExt string, connectionCaching bool, marshal Marshal, unmarshal Unmarshal) (me *Driver)
```
Creates a new `*fsdb.Driver` and returns it.

`fileExt` -- the file name extension used by `me` for reading and writing table
data files.

`connectionCaching` -- see the `Driver.ConnectionCaching` method for details.

`marshal` and `unmarshal` implement the actual encoding from and to binary or
textual data table files.

#### func (*Driver) ConnectionCaching

```go
func (me *Driver) ConnectionCaching() bool
```
Returns whether connection caching was enabled for `me` via `fsdb.NewDriver`.

You should do so if your use-case entails many parallel go-routines concurrently
operating on the same database via their own `sql.DB` connections:

that's because, while the standard `sql` package does provide "connection
pooling", this is not sensible for `fsdb`, as each `fsdb.conn` does hold its own
complete copy of data files in-memory.

All table writes are `sync.Mutex`-locking as necessary ONLY if connection
caching is enabled.

#### func (*Driver) Open

```go
func (me *Driver) Open(dirPath string) (_ driver.Conn, err error)
```
Implements the `database/sql/driver.Driver.Open` interface method.

#### func (*Driver) PersistAll

```go
func (me *Driver) PersistAll(dbConn driver.Conn, tableNames ...string) (err error)
```
Not necessary for normal use: `me` persists tables that are being written to via
`insertInto`/`updateWhere`/`deleteFrom` immediately, or in a transaction
context, at the next `Tx.Commit`.

If no `tableNames` are specifed, persists ALL tables belonging to `dbConn`,
otherwise only the specified tables are persisted.

#### func (*Driver) ReloadAll

```go
func (me *Driver) ReloadAll(dbConn driver.Conn, tableNames ...string) (err error)
```
Not necessary for normal use: `me` lazily auto-reloads tables that have been
modified on disk if such a data-refresh is necessary for the current operation.

If no `tableNames` are specifed, reloads all tables belonging to `dbConn`.

In any event, the reload always includes new-on-disk table data files not
previously loaded, and removes in-memory data tables no longer on disk.

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
If `me` is a record, returns whether it matches the specified criteria.

- recID: the `__id` of `me`, if any (since this isn't stored in the record
itself)

- filters: one or more criteria, `AND`-ed together. Each criteria is a slice of
possible values, `OR`-ed together

- strCmp: if `false`, just compares `interface{}==interface{}`. If `true`, also
compares `fmt.Sprintf("%v", interface{}) == fmt.Sprintf("%v", interface{})`

#### type Marshal

```go
type Marshal func(v interface{}) ([]byte, error)
```

Function that marshals an in-memory data table to a local file.

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
Generates a `{"createTable":name}` statement.

#### func (*StmtGen) DeleteFrom

```go
func (me *StmtGen) DeleteFrom(name string, where M) string
```
Generates a `{"deleteFrom":name, "where": where}` statement.

#### func (*StmtGen) DropTable

```go
func (_ *StmtGen) DropTable(name string) string
```
Generates a `{"dropTable":name}` statement.

#### func (*StmtGen) InsertInto

```go
func (me *StmtGen) InsertInto(name string, rec M) string
```
Generates a `{"insertInto":name, "set": rec}` statement.

#### func (*StmtGen) SelectFrom

```go
func (me *StmtGen) SelectFrom(name string, where M) string
```
Generates a `{"selectFrom":name, "where": where}` statement.

#### func (*StmtGen) UpdateWhere

```go
func (me *StmtGen) UpdateWhere(name string, set, where M) string
```
Generates a `{"updateWhere":name, "set": set, "where": where}` statement.

#### type Unmarshal

```go
type Unmarshal func(data []byte, v interface{}) error
```

Function that unmarshals an in-memory data table from a local file.

--
**godocdown** http://github.com/robertkrimen/godocdown
