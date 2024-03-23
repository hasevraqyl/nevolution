package dab

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
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

type Grade struct {
	numberTotal int
	biomeTotal  int
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
func (d Database) CheckIfBiomeExists(biome string) (e myenum) {
	rows, err := d.dt.Query("select name from biomes where name = ?", biome)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	if !rows.Next() {
		return noElem
	}
	return allClear
}
func (d Database) GradeID(grade string) (id int, e myenum) {
	tx, err := d.dt.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	rows, err := tx.Query("select _id from grades where name = ?", grade)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var gid int
	if rows.Next() {
		rows.Scan(&gid)
	} else {
		return gid, noElem
	}
	tx.Commit()
	return gid, allClear
}
func (d Database) BiomeID(biome string) (id int, e myenum) {
	tx, err := d.dt.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	rows, err := tx.Query("select _id from biomes where name = ?", biome)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var gid int
	if rows.Next() {
		rows.Scan(&gid)
	} else {
		return gid, noElem
	}
	tx.Commit()
	return gid, allClear
}
func (d Database) GradeAmount(gid int, bid int) (a int, e myenum) {
	tx, err := d.dt.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	rows, err := tx.Query("select amount from biome_grades where grade_id = ? and biome_id = ?", gid, bid)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var amount int
	if rows.Next() {
		rows.Scan(&amount)
	} else {
		return amount, noElem
	}
	tx.Commit()
	return amount, allClear
}
func (d Database) AddGradeToBiome(biome string, grade string, amount int) (e myenum) {
	tx, err := d.dt.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Fatal(err)
	}
	rows, err := tx.Query("select _id from grades where name = ?", grade)
	if err != nil {
		log.Fatal(err)
	}
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
	rows2, err := tx.Query("select _id from biomes where name = ?", biome)
	if err != nil {
		log.Fatal(err)
	}
	var biome_id int
	if rows2.Next() {
		err = rows.Scan(&biome_id)
		if err != nil {
			log.Fatal(err)
		}
	}
	rows3, err := tx.Query("select amount from biome_grades where biome_id = ? and grade_id = ?", biome_id, grade_id)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	if rows3.Next() {
		return redundantElem
	}
	rows.Close()
	rows2.Close()
	rows3.Close()
	if err != nil {
		log.Fatal(err)
	}
	s, err := tx.Prepare("insert into biome_grades(biome_id, grade_id, amount, success) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	_, err = s.Exec(biome_id, grade_id, amount, 0)
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	return allClear
}
func (d Database) DebugRemoveLater() {
	tx, err := d.dt.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	rows, err := tx.Query("select _id from grades where name = 'g'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var gid int
	if rows.Next() {
		rows.Scan(&gid)
	}
	rows2, err := tx.Query("select _id from biomes where name = 'b'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()
	var bid int
	if rows2.Next() {
		rows.Scan(&bid)
	}
	s, err := tx.Prepare("insert into biome_grades(biome_id, grade_id, amount, success) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	s.Exec(bid, gid, 100, 1)
	tx.Commit()

}
func GetSuccess(gid int, biome string) (success int) {
	return 10
}
func (d Database) Meteor() (e myenum) {
	rows, err := d.dt.Query("select amount, _id from biome_grades")
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
		s, err := tx.Prepare("update biome_grades set amount = ? where _id = ?")
		if err != nil {
			log.Fatal(err)
		}
		defer s.Close()
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
	m := make(map[int]Grade)
	tx, err := d.dt.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	rows, err := tx.Query("select amount, _id from biome_grades order by biome_id")
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	s, err := tx.Prepare("update biome_grades set amount = ? where _id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	for rows.Next() {
		var number int
		var id int
		rows.Scan(&number, &id)
		if v, ok := m[id]; ok {
			nv := Grade{
				v.biomeTotal + 1,
				v.numberTotal + number,
			}
			m[id] = nv
		} else {
			nv := Grade{
				1,
				number,
			}
			m[id] = nv
		}
		if err != nil {
			log.Fatal(err)
		}
		_, err = s.Exec(number*2, id)
		if err != nil {
			log.Fatal(err)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
	rows.Close()
	sn, err := tx.Prepare("update mutations set points_left = ? where grade_id = ? and points_left > 0")
	if err != nil {
		log.Fatal(err)
	}
	defer sn.Close()
	for k, v := range m {
		var points int
		rows2, err := d.dt.Query("select points_left from mutations where grade_id = ? and points_left > 0", k)
		if err != nil {
			log.Fatal(err)
		}
		defer rows2.Close()
		if err != nil {
			log.Fatal(err)
		}
		if rows2.Next() {
			rows2.Scan(&points)
		}
		newpoints := points - (v.numberTotal/v.biomeTotal + 1)
		if newpoints < 0 {
			newpoints = 0
		}
		sn.Exec(newpoints, k)
	}
	tx.Commit()
	return allClear
}
func (d Database) GetGradeInto(grade string) (ginfo string, e myenum) {
	var b string
	tx, err := d.dt.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Fatal(err)
	}
	rows, err := tx.Query("select _id from grades where name = ?", grade)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var id int
	if rows.Next() {
		rows.Scan(&id)
	} else {
		return b, noElem
	}
	rows2, err := tx.Query("select biome_id from biome_grades where grade_id = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()
	var info strings.Builder
	info.WriteString("Града представлена в следующих биомах:")
	for rows2.Next() {
		var biomeid int
		rows2.Scan(&biomeid)
		rows3, err := tx.Query("select name, type from biomes where _id = ?", biomeid)
		if err != nil {
			log.Fatal(err)
		}
		defer rows3.Close()
		var name string
		var ty string
		if rows3.Next() {
			rows3.Scan(&name, &ty)
		}
		info.WriteString(fmt.Sprintf("\n %v, тип %v", name, ty))
	}
	rows4, err := tx.Query("select name, points_left from mutations where grade_id = ? and points_left > 0", id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows4.Close()
	if rows4.Next() {
		var name string
		var points int
		rows4.Scan(&name, &points)
		info.WriteString(fmt.Sprintf("\nСейчас исследуется следующая мутация: %v. Осталось %v очков.", name, points))
	}
	info.WriteString("\nИмеются следующие мутации:")
	rows5, err := tx.Query("select name, points_left from mutations where grade_id = ? and points_left = 0", id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows5.Close()
	for rows5.Next() {
		var mut string
		rows5.Scan(&mut)
		info.WriteString(fmt.Sprintf("\n %v", mut))
	}
	b = info.String()
	tx.Commit()
	return b, allClear
}
func (d Database) StartMutation(grade string, mutation string) {
	rows, err := d.dt.Query("select _id from grades where name = ?", grade)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var id int
	for rows.Next() {
		rows.Scan(&id)
	}
	tx, err := d.dt.Begin()
	if err != nil {
		log.Fatal(err)
	}
	s, err := tx.Prepare("insert into mutations (grade_id, name, points_left) values (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
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
	rows, err := d.dt.Query("select _id from grades where name = ?", grade)
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
