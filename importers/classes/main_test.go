package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
)

func TestImportClasses(t *testing.T) {
	err := os.Remove("../../sql_database/test.db")
	if err != nil {
		if os.IsNotExist(err) {
			// don't worry about it!
		} else {
			log.Fatal(err)
		}
	}
	db, err := sqlx.Open("sqlite3", "../../sql_database/test.db")
	if err != nil {
		log.Fatalf("Failed to open sqlite db: %v", err)
	}
	if err != nil {
		if os.IsNotExist(err) {
			// don't worry about it!
		} else {
			t.Fatal(err)
		}
	}
	data, err := ioutil.ReadFile("./test_data/testdata.json")
	if err != nil {
		t.Fatal(err)
	}
	classes, _ := convertJsonToClassImports(data)
	writeClassesToDB(db, classes)
}
