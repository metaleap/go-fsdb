package main

import (
	"database/sql"
	"flag"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/metaleap/go-jsondb"

	ugo "github.com/metaleap/go-util"
	uio "github.com/metaleap/go-util/io"
)

var (
	custFirsts = []string{"Bob", "Alice", "Phil", "Ben", "Matt", "Rob", "Andrew", "Dave", "Kyle", "Mark"}
	custLasts  = []string{"Dylan", "Cooper", "Schumann", "Trux", "Pike", "Gerrand", "Cheney", "Isom", "Smalley"}
	custCities = []string{"Berlin", "London", "Sydney", "Phnom Penh", "Kuala Lumpur", "Jakarta", "Taipei", "Hong Kong", "San Francisco", "San Diego", "Los Santos", "San Fierro", "Liberty City", "Vice City", "Las Venturas"}

	prodAtts  = []string{"Vintage", "Luxury", "Budget", "Dick-Tracey", "Swiss", "Traditional", "Stylish", "Modern"}
	prodKinds = []string{"Dumbphone", "Console", "Toaster", "Kettle", "Tablet", "Watch"}

	numProds, numCusts int
)

func addProds(tx *sql.Tx) (err error) {
	var (
		pa  string
		rec jsondb.M
	)
	log.Println("Adding product records...")
	for _, pa1 := range prodAtts {
		for _, pk := range prodKinds {
			for _, pa2 := range prodAtts {
				if pa = pa1; pa1 != pa2 {
					pa = pa + " " + pa2
				}
				rec = jsondb.M{"Name": pa + " " + pk, "Kind": pk, "Atts": strings.Split(pa, " ")}
				numProds++
				if _, err = tx.Exec(jsondb.S.InsertInto("Products", rec)); err != nil {
					return
				}
			}
		}
	}
	log.Printf("Added %v product records", numProds)
	return
}

func addCusts(tx *sql.Tx) (err error) {
	var rec jsondb.M
	log.Println("Adding customer records...")
	for _, fn := range custFirsts {
		for _, ln := range custLasts {
			for _, c := range custCities {
				rec = jsondb.M{"Name": fn + " " + ln, "FirstName": fn, "LastName": ln, "City": c}
				numCusts++
				if _, err = tx.Exec(jsondb.S.InsertInto("Customers", rec)); err != nil {
					return
				}
			}
		}
	}
	log.Printf("Added %v customer records", numCusts)
	return
}

func addOrders(tx *sql.Tx) (err error) {
	log.Println("Adding order records...")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var (
		rec                   jsondb.M
		numOrders, t, c, o, p int
		prods                 []int
	)
	for c = 0; c < numCusts; c++ {
		numOrders = r.Intn(32) + 1
		for o = 0; o < numOrders; o++ {
			prods = make([]int, 0, r.Intn(16)+1)
			for p = 0; p < cap(prods); p++ {
				prods = append(prods, r.Intn(numProds))
			}
			rec = jsondb.M{"Customer": c, "Products": prods}
			if _, err = tx.Exec(jsondb.S.InsertInto("Orders", rec)); err != nil {
				return
			}
			t++
		}
	}
	log.Printf("Added %v order records", t)
	return
}

func main() {
	defaultDir := ugo.GopathSrcGithub("metaleap", "go-jsondb", "go-jsondb-test", "testdbs", time.Now().Format("2006-01-02_15-04-05"))
	dbDirPath := flag.String("dbdir", defaultDir, "Specify the path to a DB directory. I will open or create a JSON-DB in there.")
	flag.Parse()
	uio.EnsureDirExists(*dbDirPath)

	sql.Register(jsondb.DriverName, jsondb.NewDriver())
	db, err := sql.Open(jsondb.DriverName, *dbDirPath)
	if err == nil { // panic once at the end instead of everywhere
		log.Printf("JSON-DB location: %s", *dbDirPath)
		defer db.Close()
		var tx *sql.Tx
		if tx, err = db.Begin(); err == nil {
			if _, err = tx.Exec(jsondb.S.CreateTable("Products")); err == nil {
				if err = addProds(tx); err == nil {
					if _, err = tx.Exec(jsondb.S.CreateTable("Customers")); err == nil {
						if err = addCusts(tx); err == nil {
							if _, err = tx.Exec(jsondb.S.CreateTable("Orders")); err == nil {
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
	}
	if err != nil {
		panic(err)
	}
}
