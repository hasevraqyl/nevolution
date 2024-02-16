package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	_ "github.com/mattn/go-sqlite3"
)

type Migration struct {
	migration string
}

var info Migration

func Rollback() {
	f := "migrations.toml"
	if _, err := os.Stat(f); err != nil {
		log.Fatal(err)
	}
	if _, err := toml.DecodeFile(f, &info); err != nil {
		log.Fatal(err)
	}
	os.Remove("./nev.db")
	db, err := sql.Open("sqlite3", "./nev.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	sqlStmt := info.migration
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
}
