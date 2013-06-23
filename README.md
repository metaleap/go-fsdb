# fsdb
--
    import "github.com/metaleap/go-fsdb"

A "database driver" (compatible with Go's `database/sql` package) that's using a
local directory of files as a database of "tables".

Does not implement the finer details of *real* databases (such as relational
integrity, cascading deletes, ACID etc.) --- the *only* use-case is **"faster
prototyping of a DB app without needing to mess with a real-world DB right
now"**, based on easily inspectable, human-readable data table files.

## Connection string:

any (file-system) directory path.

## Backing file format:

Use a marshal/unmarshal provider such as `metaleap/go-fsdb/jsondb` or
`metaleap/go-fsdb/tomldb`, or write your own (start by cloning `tomldb`).

## SQL syntax:

none. Instead, the `Driver` uses simple JSON strings such as `{"createTable":
"FooBars"}`. Use the documented `StmtFooBar` methods (ie. `fsdb.StmtCreateTable`
and friends) to easily generate statements for use with sql.Exec() and
sql.Query(), whether via a `sql.DB` or a `sql.Tx`.

I didn't see the use in parsing real SQL syntax --- each real-world DB has its
own syntax quirks, so when moving on from `fsdb` to the real DB, I'd have to
adapt most/all SQL statements anyway. This way, it's guaranteed that I'll have
to do so.

## Connection pooling/caching:

works "so-so" with Go's built-in pooling: with many redundant in-memory copies
of the same data tables, as per below. See documentation on the `fsdb.NewDriver`
method for details.

Each `fsdb`-driven `sql.DB` connection maintains a full in-memory copy of its
data table files, auto-persisting and auto-reloading as necessary.

## Transactions:

they're a useful hack at best -- the idea here is for batching multiple writes
together. Each `insertInto`/`updateWhere`/`deleteFrom` would normally persist
the full table to disk immediately. But in the context of a transaction, they
won't -- only the final `Tx.Commit` will flush participating tables to disk.

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
	//	Used in `selectWhere` queries, defaults to false. See `M.Match` method for explanation.
	StrCmp bool
)
```

#### func  NewDriver

```go
func NewDriver(fileExt string, connectionCaching bool, marshal Marshal, unmarshal Unmarshal) driver.Driver
```
Creates a new `database/sql/driver.Driver` and returns it.

- `fileExt` -- the file name extension used for reading and writing table data
files.

- `marshal`/`unmarshal` implement the actual decoding-from/encoding-to binary or
textual data table files.

- `connectionCaching` -- if `true`, enables connection-caching for this
`Driver`. You should do so if your use-case entails many parallel go-routines
concurrently operating on the same database via their own `sql.DB` connections:
that's because, while the standard `sql` package does provide "connection
pooling", this is not sensible for `fsdb`, as each `fsdb.conn` does hold its own
complete copies of all table data files in-memory. (All table writes are
`sync.Mutex`-locking as necessary ONLY if connection-caching is enabled.)

#### func  StmtCreateTable

```go
func StmtCreateTable(name string) string
```
Generates a `{"createTable":name}` statement.

#### func  StmtDeleteFrom

```go
func StmtDeleteFrom(name string, where M) string
```
Generates a `{"deleteFrom":name, "where": where}` statement.

#### func  StmtDropTable

```go
func StmtDropTable(name string) string
```
Generates a `{"dropTable":name}` statement.

#### func  StmtInsertInto

```go
func StmtInsertInto(name string, rec M) string
```
Generates a `{"insertInto":name, "set": rec}` statement.

#### func  StmtSelectFrom

```go
func StmtSelectFrom(name string, where M) string
```
Generates a `{"selectFrom":name, "where": where}` statement.

#### func  StmtUpdateWhere

```go
func StmtUpdateWhere(name string, set, where M) string
```
Generates a `{"updateWhere":name, "set": set, "where": where}` statement.

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

#### type Unmarshal

```go
type Unmarshal func(data []byte, v interface{}) error
```

Function that unmarshals an in-memory data table from a local file.

--
**godocdown** http://github.com/robertkrimen/godocdown
