// A "database driver" (compatible with Go's `database/sql` package)
// that's using a directory of JSON text files as a database of "tables".
//
// Does not implement the finer details of *real* databases (such as
// relational integrity, cascading deletes, ACID etc.) --- the *only* use-case
// is **"faster prototyping of a DB app without needing to mess with a real-world
// DB right now"** and easily inspectable, human-readable data table files.
//
// Connection string: any (file-system) directory path.
//
// SQL syntax: none. Instead, the driver uses simple JSON strings such
// as `{"createTable": "FooBars"}`. Use the documented `StmtGen` methods
// (ie. `jsondb.S.CreateTable()` and friends) to easily generate statements
// for use with sql.Exec() and sql.Query(), whether via a `sql.DB` or a `sql.Tx`.
//
// I didn't see the use in parsing real SQL syntax --- each real-world DB has its
// own syntax quirks, so when moving on from jsondb to the real DB, I'd have to adapt
// most/all SQL statements anyway. This way, it's guaranteed that I'll have to do so.
//
// Each jsondb connection maintains a full in-memory copy of data table files,
// persisting and reloading as necessary, see documentation on the global exported
// `PersistAll()` and `ReloadAll()` functions for details.
//
// Connection pooling/caching: you can use Go's built-in pooling if you're fine
// with potentially many duplicate in-memory copies of the same data tables.
// See documentation on the global exported `ConnectionCaching()` function for details.
//
// Transactions: not quite behaving like normal database transactions
package jsondb
