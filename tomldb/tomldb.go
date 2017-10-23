// A "database driver" (compatible with Go's `database/sql` package)
// that's using a local directory of TOML files as a database of "tables",
// implemented on top of `github.com/metaleap/go-fsdb`.
package tomldb

import (
	"database/sql/driver"

	"github.com/go-forks/toml"
	"github.com/metaleap/go-fsdb"
	"github.com/metaleap/go-util/str"
)

var (
	//	Can be used for `sql.Register` and `sql.Open`.
	DriverName = "github.com/metaleap/go-fsdb/tomldb"

	//	File name extension for TOML data files. This is passed
	//	in `tomldb.NewDriver` to `fsdb.NewDriver(DriverName)`.
	FileExt = ".tomldbt"
)

//	Returns a `fsdb.NewDriver` initialized with `FileExt` and TOML un/marshalers.
func NewDriver(connectionCaching bool) driver.Driver {
	tomlUnmarshal := func(data []byte, v interface{}) (err error) {
		_, err = toml.Decode(string(data), v)
		return
	}
	tomlMarshal := func(v interface{}) (data []byte, err error) {
		var (
			buf     ustr.Buffer
			m, rec  map[string]interface{}
			sl      []interface{}
			fx, rx  interface{}
			ok      bool
			fn, rid string
			i       int
		)
		if m, ok = v.(map[string]interface{}); m == nil || !ok {
			m, ok = v.(fsdb.M)
		}
		if m != nil && ok {
			for rid, rx = range m {
				buf.Writeln("[%v]", rid)
				if rec, ok = rx.(map[string]interface{}); rec == nil || !ok {
					rec, ok = rx.(fsdb.M)
				}
				if rec != nil && ok {
					for fn, fx = range rec {
						if sl, ok = fx.([]interface{}); ok {
							buf.Write("%v = [", fn)
							for i, fx = range sl {
								if i > 0 {
									buf.Write(", ")
								}
								buf.Write("%#v", fx)
							}
							buf.Writeln("]")
						} else {
							buf.Writeln("%v = %#v", fn, fx)
						}
					}
				}
				buf.Writeln("")
			}
		}
		data = buf.Bytes()
		return
	}
	return fsdb.NewDriver(FileExt, connectionCaching, tomlMarshal, tomlUnmarshal)
}
