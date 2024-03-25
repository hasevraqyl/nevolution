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
	Upstart  string
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
	nsqlStmt := info.Upstart
	_, err = d.dt.Exec(nsqlStmt)
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

func (d Database) AddBiome(biome_name string, biome_type string) (e myenum) {
	tx, err := d.dt.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	fmt.Printf("inserting into biomes a biome call %v of type %v", biome_name, biome_type)
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
	grade_id, _ := d.GradeID(grade)
	biome_id, _ := d.BiomeID(biome)
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
	rows3.Close()
	if err != nil {
		log.Fatal(err)
	}
	s, err := tx.Prepare("insert into biome_grades(biome_id, grade_id, amount, success) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	_, err = s.Exec(biome_id, grade_id, amount, GetSuccess(grade_id, "geysers"))
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
	brows, err := tx.Query("select _id from biomes where name = 'b'")
	if err != nil {
		log.Fatal(err)
	}
	defer brows.Close()
	var bid int
	if brows.Next() {
		rows.Scan(&bid)
	}
	s, err := tx.Prepare("insert into biome_grades(biome_id, grade_id, amount, success) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	s.Exec(bid, gid, 100, 10)
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
	maxes := make(map[int]int)
	mmu := make(map[int]Grade)
	tx, err := d.dt.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	rows, err := tx.Query("select distinct biome_id from biome_grades")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var bid int
		rows.Scan(&bid)
		max, err := tx.Query("select MAX(success) from biome_grades where biome_id = ?", bid)
		if err != nil {
			log.Fatal(err)
		}
		defer max.Close()
		var msuc int
		if max.Next() {
			max.Scan(&msuc)
		}
		maxes[bid] = msuc
	}
	graderows, err := tx.Query("select biome_id, grade_id, amount, success, _id from biome_grades")
	if err != nil {
		log.Fatal(err)
	}
	defer graderows.Close()
	ss, err := tx.Prepare("update biome_grades set amount = ? where _id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer ss.Close()
	for graderows.Next() {
		var bid, gid, am, suc, id, newam int
		graderows.Scan(&bid, &gid, &am, &suc, &id)
		max := maxes[bid]
		if max == suc {
			newam = (am + suc)
		} else {
			newam = (am + suc - max)
			if newam < 0 {
				newam = 0
			}
		}
		ss.Exec(newam, id)
		v, ok := mmu[gid]
		if ok {
			mmu[gid] = Grade{
				v.biomeTotal + 1,
				v.numberTotal + am,
			}
		} else {
			mmu[gid] = Grade{
				1,
				am,
			}
		}
	}
	mutrows, err := tx.Query("select _id, grade_id, points_left from mutations")
	if err != nil {
		log.Fatal(err)
	}
	defer mutrows.Close()
	s, err := tx.Prepare("update mutations set points_left = ? where _id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	for mutrows.Next() {
		var gid, pl, id, newpl int
		mutrows.Scan(&gid, &pl, &id)
		mutg := mmu[gid]
		mut := mutg.numberTotal/(mutg.biomeTotal*10) + 1
		if pl > 0 {
			newpl = pl - mut
			if newpl < 0 {
				newpl = 0
			}
			s.Exec(newpl, id)
		}
	}
	tx.Commit()
	return allClear
}
func (d Database) GetGradeInfo(grade string) (ginfo string, e myenum) {
	var b string
	tx, err := d.dt.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Fatal(err)
	}
	id, status := d.GradeID(grade)
	if status != 1 {
		return b, noElem
	}
	brows, err := tx.Query("select biome_id from biome_grades where grade_id = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	defer brows.Close()
	var info strings.Builder
	info.WriteString("Града представлена в следующих биомах:")
	for brows.Next() {
		var biomeid int
		brows.Scan(&biomeid)
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
	mutrows, err := d.dt.Query("select name from mutations where grade_id = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	defer mutrows.Close()
	for mutrows.Next() {
		var mutation string
		rows.Scan(&mutation)
		m[mutation] = struct{}{}
	}
	return m, allClear
}
