# go-jsondb-test
--
This program demonstrates how to use the `go-jsondb` package:

It creates a new database inside the directory specified via the `-dbdir=""`
command-line flag, or if not present, in a new temporary directory under
$GOPATH/src/github.com/metaleap/go-jsondb/go-jsondb-test

In this newly created (or overwritten) database:
- via `createTable`, creates 3 'tables'/collections: Customers, Products, Orders
- via `insertInto`, populates those with semi-random records
- via `selectFrom`, queries the DB to find all Customers with *LastName=Collins*
- via `deleteFrom`, deletes all Orders belonging to those customers
- via `updateWhere`, for all *FirstName=Alice&City=Berlin* Customers, sets their City to Seattle

--
**godocdown** http://github.com/robertkrimen/godocdown