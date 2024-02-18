package dab

import (
	"database/sql"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	dt *sql.DB
}

type Migration struct {
	migration string
}

var info Migration

func Wrap(db *sql.DB) (d Database) {
	return Database{db}
}

func (d Database) Rollback() {
	f := "migrations.toml"
	if _, err := os.Stat(f); err != nil {
		log.Fatal(err)
	}
	if _, err := toml.DecodeFile(f, &info); err != nil {
		log.Fatal(err)
	}
	os.Remove("./nev.db")
	sqlStmt := info.migration
	_, err := d.dt.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
}

func (d Database) AddGrade(grade string) {
	tx, err := d.dt.Begin()
	if err != nil {
		log.Fatal(err)
	}
	s, err := tx.Prepare("insert into grades(name) values(?)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = s.Exec(grade)
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}
