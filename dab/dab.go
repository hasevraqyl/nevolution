package dab

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	_ "github.com/mattn/go-sqlite3"
)

type myenum int

const (
	allClear      myenum = 1
	noElem        myenum = 2
	invalidElem   myenum = 3
	redundantElem myenum = 4
)

type Database struct {
	dt *sql.DB
}

type Migration struct {
	Rollback string
}

var info Migration

func Wrap(db *sql.DB) (d Database) {
	return Database{db}
}

func (e myenum) Text(text string) string {
	if e == allClear {
		return text
	} else if e == noElem {
		return fmt.Sprintf("Элемента %v нет", text)
	} else if e == invalidElem {
		return fmt.Sprintf("Элемента %v не бывает", text)
	} else {
		return "Такой элемент уже есть"
	}
}
func (d Database) Rollback() {
	f := "./migrations.toml"
	if _, err := os.Stat(f); err != nil {
		log.Fatal(err)
	}
	if _, err := toml.DecodeFile(f, &info); err != nil {
		log.Fatal(err)
	}
	os.Remove("./nev.db")
	sqlStmt := info.Rollback
	_, err := d.dt.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
}
func (d Database) CloseDB() {
	d.dt.Close()
}

func (d Database) AddGrade(grade string) (e myenum) {
	rows, err := d.dt.Query("select name from grades where name = ?", grade)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		return redundantElem
	}
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
	return allClear
}

func (d Database) AddBiomePreliminary(biome_type string) (e myenum) {
	b_t_l := []string{"гейзеры", "курильщики", "пелагиаль", "пресные воды", "эндолиты", "атмосфера", "литораль"}
	for _, v := range b_t_l {
		if biome_type == v {
			return allClear
		}
	}
	return invalidElem
}

func (d Database) AddBiome(biome_name string, biome_type string) (e myenum) {
	tx, err := d.dt.Begin()
	if err != nil {
		log.Fatal(err)
	}
	s, err := tx.Prepare("insert into biomes(name, type) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = s.Exec(biome_name, biome_type)
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	return allClear
}
func (d Database) AddGradeToBiomePreliminary(biome string) (e myenum) {
	rows, err := d.dt.Query("select name from biomes where name = ?", biome)
	if err != nil {
		log.Fatal(err)
	}
	if !rows.Next() {
		return noElem
	}
	return allClear
}

func (d Database) AddGradeToBiome(biome string, grade string) (e myenum) {
	rows, err := d.dt.Query("select id from grades where name = ?", grade)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	var grade_id int
	if rows.Next() {
		err = rows.Scan(&grade_id)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		return noElem
	}
	rows2, err := d.dt.Query("select id from biomes where name = ?", biome)
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()
	var biome_id int
	if rows2.Next() {
		err = rows.Scan(&biome_id)
		if err != nil {
			log.Fatal(err)
		}
	}
	rows3, err := d.dt.Query("select amount from biome_grades where biome_id = ?, grade_id = ?", biome_id, grade_id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows3.Close()
	if err != nil {
		log.Fatal(err)
	}
	if rows3.Next() {
		return redundantElem
	}
	tx, err := d.dt.Begin()
	if err != nil {
		log.Fatal(err)
	}
	s, err := tx.Prepare("insert into biome_grades(biome_id, grade_id, amount, success, type) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = s.Exec(biome_id, grade_id, 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	return allClear
}

func (d Database) Meteor() (e myenum) {
	rows, err := d.dt.Query("select amount, id from biome_grades")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	for rows.Next() {
		var number int
		var id int
		rows.Scan(&number, &id)
		i := int64(number) - (r.Int63n(300) + 300)
		if i < 0 {
			i = 0
		}
		tx, err := d.dt.Begin()
		if err != nil {
			log.Fatal(err)
		}
		s, err := tx.Prepare("update biome_grades set amount = ? where id = ?")
		if err != nil {
			log.Fatal(err)
		}
		_, err = s.Exec(i, id)
		if err != nil {
			log.Fatal(err)
		}
		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}
	}
	return allClear
}

func (d Database) Turn() (e myenum) {
	rows, err := d.dt.Query("select amount, id from biome_grades")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var number int
		var id int
		rows.Scan(&number, &id)
		tx, err := d.dt.Begin()
		if err != nil {
			log.Fatal(err)
		}
		s, err := tx.Prepare("update biome_grades set amount = ? where id = ?")
		if err != nil {
			log.Fatal(err)
		}
		_, err = s.Exec(number*2, id)
		if err != nil {
			log.Fatal(err)
		}
		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}
	}
	return allClear
}
func (d Database) StartMutation(grade string, mutation string) {
	rows, err := d.dt.Query("select id from grades where name = ?", grade)
	if err != nil {
		log.Fatal(err)
	}
	var id int
	for rows.Next() {
		rows.Scan(&id)
	}
	tx, err := d.dt.Begin()
	if err != nil {
		log.Fatal(err)
	}
	s, err := tx.Prepare("insert into mutations values (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = s.Exec(id, mutation, 300)
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}
func (d Database) GetGradeMutations(grade string) (mutations map[string]struct{}, e myenum) {
	m := make(map[string]struct{})
	rows, err := d.dt.Query("select id from grades where name = ?", grade)
	if err != nil {
		log.Fatal(err)
	}
	var id int
	if rows.Next() {
		rows.Scan(&id)
	} else {
		return m, noElem
	}
	defer rows.Close()
	rows2, err := d.dt.Query("select name from mutations where grade_id = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var mutation string
		rows.Scan(&mutation)
		m[mutation] = struct{}{}
	}
	return m, allClear
}
