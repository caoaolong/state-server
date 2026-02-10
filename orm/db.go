package orm

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

// DB 返回 GORM 实例，供路由层使用
func DB() *gorm.DB {
	return db
}

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("smdb.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	db.Exec(`PRAGMA journal_mode = WAL;`)
	db.Exec(`PRAGMA synchronous = NORMAL;`)
	Migrate()
}
