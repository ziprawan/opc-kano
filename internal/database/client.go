package database

import (
	"kano/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var dbInstance *gorm.DB

func Init() {
	db, err := gorm.Open(postgres.Open(config.GetConfig().DatabaseURL))
	if err != nil {
		panic(err)
	}

	dbInstance = db
}

func GetInstance() *gorm.DB {
	if dbInstance == nil {
		Init()
	}
	return dbInstance
}
