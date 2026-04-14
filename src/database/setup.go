package database

import (
	"database/sql"
	"log"

	"gorm.io/gorm"
)

func SetupDatabase(config *DatabaseConfig) (*gorm.DB, *sql.DB) {
	log.Println("Database setup starting...")
	connection := Connect(config)

	log.Println("Database migrations applied")

	db := GetDB(connection)

	if err := db.Ping(); err != nil {
		panic("Failed to ping database: " + err.Error())
	}

	log.Println("Database connected and healthy")

	return connection, db
}
