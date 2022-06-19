package model

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

var DB *gorm.DB

func InitOrm(path string) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlDb,err := db.DB()
	if err != nil {
		panic(err)
	}
	// 连接池的空闲连接数
	sqlDb.SetMaxIdleConns(100)
	// 连接池的最大连接数
	sqlDb.SetMaxOpenConns(500)
	// 连接的最大存活时间
	sqlDb.SetConnMaxIdleTime(time.Minute * 10)
	DB = db
}
