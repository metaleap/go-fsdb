// This program demonstrates how to use the `go-fsdb` and `go-fsdb/{foo}db` packages:
//
// It creates a new database inside the directory specified via the `-dbdir=""`
// command-line flag, or if not present, in a new temporary directory under
// $GOPATH/src/github.com/metaleap/go-fsdb/go-fsdb-test
//
// In this newly created (or overwritten) database:
// - via `createTable`, creates 3 'tables'/collections: Customers, Products, Orders
// - via `insertInto`, populates those with semi-random records
// - via `selectFrom`, queries the DB to find all Customers with *LastName=Collins*
// - via `deleteFrom`, deletes all Orders belonging to those customers
// - via `updateWhere`, for all *FirstName=Alice&City=Berlin* Customers, sets their City to Seattle
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	fsdb "github.com/metaleap/go-fsdb"
	fsdb_json "github.com/metaleap/go-fsdb/jsondb"
	fsdb_toml "github.com/metaleap/go-fsdb/tomldb"

	"github.com/go-utils/udb"
	"github.com/go-utils/ufs"
	"github.com/go-utils/ugo"
)

var (
	dbDrvModes = []string{"json", "toml"}

	custFirsts = []string{"Bob", "Alice", "Phil", "Edwyn", "Matt", "Rob", "Andrew", "Dave", "Kyle", "Mark"}
	custLasts  = []string{"Dylan", "Cooper", "Collins", "Trux", "Pike", "Gerrand", "Cheney", "Isom", "Smalley"}
	custCities = []string{"Berlin", "London", "Sydney", "Phnom Penh", "Kuala Lumpur", "Jakarta", "Taipei", "Hong Kong", "San Francisco", "San Diego", "Los Santos", "San Fierro", "Liberty City", "Vice City", "Las Venturas"}

	prodAtts  = []string{"Vintage", "Luxury", "Budget", "Dick-Tracey", "Swiss", "Traditional", "Stylish", "Modern"}
	prodKinds = []string{"Dumbphone", "Console", "Toaster", "Kettle", "Tablet", "Watch"}

	numProds, numCusts int
)

func addProds(tx *sql.Tx) (err error) {
	var (
		pa  string
		rec fsdb.M
	)
	log.Println("Adding records to 'Products'...")
	for _, pa1 := range prodAtts {
		for _, pk := range prodKinds {
			for _, pa2 := range prodAtts {
				if pa = pa1; pa1 != pa2 {
					pa = pa + " " + pa2
				}
				rec = fsdb.M{"Name": pa + " " + pk, "Kind": pk, "Atts": strings.Split(pa, " ")}
				if _, err = tx.Exec(fsdb.StmtInsertInto("Products", rec)); err != nil {
					return
				}
				numProds++
			}
		}
	}
	log.Printf("Added %v 'Products' records", numProds)
	return
}

func addCusts(tx *sql.Tx) (err error) {
	var rec fsdb.M
	log.Println("Adding records to 'Customers'...")
	for _, fn := range custFirsts {
		for _, ln := range custLasts {
			for _, c := range custCities {
				rec = fsdb.M{"FullName": fn + " " + ln, "FirstName": fn, "LastName": ln, "City": c}
				if _, err = tx.Exec(fsdb.StmtInsertInto("Customers", rec)); err != nil {
					return
				}
				numCusts++
			}
		}
	}
	log.Printf("Added %v 'Customers' records", numCusts)
	return
}

func addOrders(tx *sql.Tx) (err error) {
	log.Println("Adding records to 'Orders'...")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var (
		rec                   fsdb.M
		numOrders, t, c, o, p int
		prods                 []string
	)
	for c = 0; c < numCusts; c++ {
		numOrders = r.Intn(32) + 1
		for o = 0; o < numOrders; o++ {
			prods = make([]string, 0, r.Intn(16)+1)
			for p = 0; p < cap(prods); p++ {
				prods = append(prods, strconv.Itoa(r.Intn(numProds)))
			}
			rec = fsdb.M{"Customer": strconv.Itoa(c), "Products": prods}
			if _, err = tx.Exec(fsdb.StmtInsertInto("Orders", rec)); err != nil {
				return
			}
			t++
		}
	}
	log.Printf("Added %v 'Orders' records", t)
	return
}

func conn(dbDrvMode, dbDirPath string) (db *sql.DB, err error) {
	switch dbDrvMode {
	case "json":
		sql.Register(fsdb_json.DriverName, fsdb_json.NewDriver(false))
		db, err = sql.Open(fsdb_json.DriverName, dbDirPath)
	case "toml":
		sql.Register(fsdb_toml.DriverName, fsdb_toml.NewDriver(false))
		db, err = sql.Open(fsdb_toml.DriverName, dbDirPath)
	default:
		err = fmt.Errorf("Unknown -drv flag value %#v: must be one of: %v", dbDrvMode, dbDrvModes)
	}
	return
}

func main() {
	defaultDir := ugo.GopathSrcGithub("metaleap", "go-fsdb", "go-fsdb-test", "testdbs", time.Now().Format("2006-01-02_15-04-05"))
	dbDirPath := flag.String("dbdir", defaultDir, "Specify the path to a DB directory. I will open or create a JSON-DB in there.")
	dbDrvMode := flag.String("drv", dbDrvModes[0], fmt.Sprintf("Must be one of: %v.", dbDrvModes))
	flag.Parse()
	ufs.EnsureDirExists(*dbDirPath)

	db, err := conn(*dbDrvMode, *dbDirPath)
	if err == nil { // panic once at the end instead of everywhere
		log.Printf("JSON-DB location: %s", *dbDirPath)
		defer db.Close()
		var tx *sql.Tx
		if tx, err = db.Begin(); err == nil {
			if _, err = tx.Exec(fsdb.StmtCreateTable("Products")); err == nil {
				if err = addProds(tx); err == nil {
					if _, err = tx.Exec(fsdb.StmtCreateTable("Customers")); err == nil {
						if err = addCusts(tx); err == nil {
							if _, err = tx.Exec(fsdb.StmtCreateTable("Orders")); err == nil {
								err = addOrders(tx)
							}
						}
					}
				}
			}
			if err == nil {
				err = tx.Commit()
			} else if err2 := tx.Rollback(); err2 != nil {
				log.Printf("Rollback error: %v", err2)
			}
		}
		var rows *sql.Rows
		queryName := "Collins"
		var recIds []string
		if rows, err = db.Query(fsdb.StmtSelectFrom("Customers", fsdb.M{"LastName": queryName})); err == nil {
			defer rows.Close()
			var cursor udb.SqlCursor
			if err = cursor.PrepareColumns(rows); err == nil {
				var rec map[string]interface{}
				for rows.Next() {
					if rec, err = cursor.Scan(rows); err == nil {
						// log.Printf("Record found for LastName=%#v:\t%v\n", queryName, rec)
						recIds = append(recIds, rec[fsdb.IdField].(string))
					} else {
						break
					}
				}
			}
			if err == nil {
				err = rows.Err()
			}
			if err == nil {
				log.Printf("Found %v 'Customers' with LastName=%#v---deleting all their 'Orders':", len(recIds), queryName)
				var numRows int64
				if numRows, err = udb.Exec(db, false, fsdb.StmtDeleteFrom("Orders", fsdb.M{"Customer": recIds})); err == nil {
					log.Printf("..deletion affected %v rows", numRows)
					queryName = "Alice"
					log.Printf("Updating all FirstName=%#v 'Customers' from Berlin to Seattle:", queryName)
					if numRows, err = udb.Exec(db, false, fsdb.StmtUpdateWhere("Customers", fsdb.M{"City": "Seattle"}, fsdb.M{"City": "Berlin", "FirstName": "Alice"})); err == nil {
						log.Printf("..update affected %v records.", numRows)
					}
				}
			}
		}

	}

	if err != nil {
		panic(err)
	}
}
