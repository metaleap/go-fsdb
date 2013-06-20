# jsondb
--
    import "github.com/metaleap/go-jsondb"

	Package description

## Usage

```go
const (
	DriverName = "github.com/metaleap/go-jsondb"

	IdField = "__id"
)
```

```go
var (
	FileExt = ".jsondbt"
	StrCmp  bool
)
```

#### func  ConnectionCaching

```go
func ConnectionCaching() bool
```

#### func  NewDriver

```go
func NewDriver() (me driver.Driver)
```

#### func  PersistAll

```go
func PersistAll(connection driver.Conn, tableNames ...string) (err error)
```

#### func  ReloadAll

```go
func ReloadAll(connection driver.Conn, tableNames ...string) (err error)
```

#### func  SetConnectionCaching

```go
func SetConnectionCaching(enableCaching bool)
```

#### type M

```go
type M map[string]interface{}
```


#### func (M) Match

```go
func (me M) Match(recId string, filters M, strCmp bool) (isMatch bool)
```

#### type StmtGen

```go
type StmtGen struct {
}
```


```go
var S StmtGen
```

#### func (*StmtGen) CreateTable

```go
func (_ *StmtGen) CreateTable(name string) string
```

#### func (*StmtGen) DeleteFrom

```go
func (me *StmtGen) DeleteFrom(name string, where M) string
```

#### func (*StmtGen) DropTable

```go
func (_ *StmtGen) DropTable(name string) string
```

#### func (*StmtGen) InsertInto

```go
func (me *StmtGen) InsertInto(name string, rec M) string
```

#### func (*StmtGen) SelectFrom

```go
func (me *StmtGen) SelectFrom(name string, where M) string
```

#### func (*StmtGen) UpdateWhere

```go
func (me *StmtGen) UpdateWhere(name string, set, where M) string
```

--
**godocdown** http://github.com/robertkrimen/godocdown