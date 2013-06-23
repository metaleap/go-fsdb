// A "database driver" (compatible with Go's `database/sql` package)
// that's using a local directory of JSON files as a database of "tables",
// implemented on top of http://github.com/metaleap/go-fsdb.
package jsondb

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/metaleap/go-fsdb"
)

var (
	//	Can be used for `sql.Register()` and `sql.Open()`.
	DriverName = "github.com/metaleap/go-fsdb/jsondb"

	//	File name extension for JSON data files. This is passed
	//	in `jsondb.NewDriver()` to `fsdb.NewDriver(DriverName)`.
	FileExt = ".jsondbt"
)

//	Returns a `fsdb.NewDriver()` initialized with `FileExt` and JSON un/marshalers.
func NewDriver() driver.Driver {
	jsonMarshal := func(v interface{}) ([]byte, error) {
		return json.MarshalIndent(v, "", " ")
	}
	return fsdb.NewDriver(FileExt, jsonMarshal, json.Unmarshal)
}
