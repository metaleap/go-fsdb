# tomldb
--
    import "github.com/metaleap/go-fsdb/tomldb"

A "database driver" (compatible with Go's `database/sql` package) that's using a
local directory of TOML files as a database of "tables", implemented on top of
`github.com/metaleap/go-fsdb`.

## Usage

```go
var (
	//	Can be used for `sql.Register` and `sql.Open`.
	DriverName = "github.com/metaleap/go-fsdb/tomldb"

	//	File name extension for TOML data files. This is passed
	//	in `tomldb.NewDriver` to `fsdb.NewDriver(DriverName)`.
	FileExt = ".tomldbt"
)
```

#### func  NewDriver

```go
func NewDriver(connectionCaching bool) *fsdb.Driver
```
Returns a `fsdb.NewDriver` initialized with `FileExt` and TOML un/marshalers.

--
**godocdown** http://github.com/robertkrimen/godocdown
