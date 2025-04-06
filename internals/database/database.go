package database

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/lib/pq"

	projectconfig "kano/internals/project_config"
)

var (
	dbInstance *sql.DB
	once       sync.Once
)

func ConnectDB() *sql.DB {
	once.Do(func() {
		config := projectconfig.GetConfig()

		var err error
		dbInstance, err = sql.Open("postgres", config.DatabaseURL)
		if err != nil {
			panic(fmt.Errorf("failed to connect to database: %v", err))
		}

		dbInstance.SetMaxOpenConns(15)
		dbInstance.SetMaxIdleConns(5)

	})

	return dbInstance
}

func GetDB() *sql.DB {
	if dbInstance == nil {
		// panic(errors.New("database is not initialized! Call ConnectDB() first"))
		ConnectDB()
	}

	return dbInstance
}
