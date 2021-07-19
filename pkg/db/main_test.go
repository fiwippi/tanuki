package db

import (
	"log"
	"os"
	"testing"
)

var testDb *DB

func TestMain(m *testing.M) {
	var err error
	testDb, err = CreateDB("test.db")
	if err != nil {
		log.Fatalf("could not create db on setup, err: %s\n", err)
	}

	code := m.Run()
	os.Remove("test.db")
	os.Exit(code)
}
