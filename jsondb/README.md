# jsondb
--
    import "github.com/metaleap/go-fsdb/jsondb"

A "database driver" (compatible with Go's `database/sql` package) that's using a
local directory of JSON files as a database of "tables", implemented on top of
`github.com/metaleap/go-fsdb`.

## Usage

```go
var (
	//	Can be used for `sql.Register` and `sql.Open`.
	DriverName = "github.com/metaleap/go-fsdb/jsondb"

	//	File name extension for JSON data files. This is passed
	//	in `jsondb.NewDriver` to `fsdb.NewDriver(DriverName)`.
	FileExt = ".jsondbt"
)
```

#### func  NewDriver

```go
func NewDriver(connectionCaching bool) driver.Driver
```
Returns a `fsdb.NewDriver` initialized with `FileExt` and JSON un/marshalers.

--
**godocdown** http://github.com/robertkrimen/godocdown
