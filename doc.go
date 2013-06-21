// A "database driver" (compatible with Go's `database/sql` package)
// that's using a directory of JSON text files as a database of "tables".
//
// Does not implement the finer details of *real* databases (such as
// relational integrity, cascading deletes, ACID etc.) --- the *only* use-case
// is **"faster prototyping of a DB app without needing to mess with a real-world
// DB right now"** and easily inspectable, human-readable data table files.
//
// **Connection string**: any (file-system) directory path.
//
// **SQL syntax**: none. Instead, the driver uses simple JSON strings such
// as `{"createTable": "FooBars"}`. Use the documented `StmtGen` methods
// (ie. `jsondb.S.CreateTable()` and friends) to easily generate statements
// for use with sql.Exec() and sql.Query(), whether via a `sql.DB` or a `sql.Tx`.
//
// I didn't see the use in parsing real SQL syntax --- each real-world DB has its
// own syntax quirks, so when moving on from jsondb to the real DB, I'd have to adapt
// most/all SQL statements anyway. This way, it's guaranteed that I'll have to do so.
//
// **Connection pooling/caching**: works "so-so" with Go's built-in pooling: with
// many redundant in-memory copies of the same data tables, as per below.
// See documentation on the global exported `ConnectionCaching()` function for details.
//
// Each jsondb-driven `sql.DB` connection maintains a full in-memory copy of its data
// table files, auto-persisting and auto-reloading as necessary -- see documentation on
// the global exported `PersistAll()` and `ReloadAll()` functions for details.
//
// **Transactions**: they're a useful hack at best -- the idea here is for batching multiple
// writes together. Each `insertInto`/`updateWhere`/`deleteFrom` would normally persist the
// full table to disk immediately. But in the context of a transaction, they won't -- only
// the final `Tx.Commit()` will flush participating tables to disk.
package jsondb
