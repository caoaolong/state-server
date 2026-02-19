package orm

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// DB 返回 GORM 实例，供路由层使用
func DB() *gorm.DB {
	return db
}

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("smdb.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	db.Exec(`PRAGMA journal_mode = WAL;`)
	db.Exec(`PRAGMA synchronous = NORMAL;`)

	// 同步数据库结构
	db.AutoMigrate(&SMFlow{})

	Migrate()
}
