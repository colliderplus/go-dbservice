package dbService

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"log"
)

type DB struct {
	db *gorm.DB
	err error
	user,password, host, port, database, sslmode string
}

func NewDB(user string, password string, host string, port string, database string, sslmode string) *DB {
	return &DB{user: user, password: password, host: host, port: port, database: database, sslmode: sslmode}
}

func (d *DB)Init(method ...interface {}) {
	dbinfo := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		d.user,
		d.password,
		d.host,
		d.port,
		d.database,
		d.sslmode,
	)

	d.db, d.err = gorm.Open("postgres", dbinfo)
	if d.err != nil {
		log.Println(`Failed to connect to database`)
		panic(d.err)
	}
	log.Println("Database connected")
}

func (d *DB)GetDB() *gorm.DB {
	return d.db
}

func (d *DB)CloseDB() {
	d.db.Close()
}

type Migration struct {
	Number int `gorm:"not null;PRIMARY_KEY;" json:"number"`
}
type MigrationSchema struct {
	Number  int
	Methods MigrationJobs
}

type MigrationJob = func(db *gorm.DB) error
type MigrationJobs = []MigrationJob


func Migrate(d *DB, models []interface{}, schema []MigrationSchema) {
	createModel(d,
		&Migration{},
	)
	d.db.AutoMigrate(
		models...
	)
	go migrationsMethods(d.db.Unscoped(), schema...)
}

func createModel(d *DB, models ...interface {}) {
	for _, value := range models {
		if !d.db.HasTable(value) {
			err := d.db.CreateTable(value)
			if err != nil {
				log.Println("Table already exists")
			}
		}
	}

}

func migrationsMethods(db *gorm.DB, schema ...MigrationSchema) {
	for _,sc := range schema {
		if db.Where("number = ?",sc.Number).First(&Migration{}).RecordNotFound() {
			var err error
			for _, value := range sc.Methods {
				err = value(db)
				if err != nil {
					break
				}
			}
			if err == nil {
				db.Save(&Migration{Number: sc.Number})
			}
		}
	}
}