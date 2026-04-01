package database

import (
	"log"

	"gorm.io/gorm"
)

func SetupDatabase(config *DatabaseConfig) *gorm.DB {
	log.Println("Database setup starting...")
	connection := Connect(config)

	log.Println("Database migrations applied")

	_db := GetDB(connection)

	if err := _db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
		panic(err)
	}

	log.Println("Database connected and healthy")

	return connection
}
