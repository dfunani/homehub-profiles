package utils

import (
	"dfunani/homehub-profiles/src/config"
	"dfunani/homehub-profiles/src/database"
	"dfunani/homehub-profiles/src/mapping"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func GetServiceName() string {
	return "HomeHub Profiles"
}

func GetVersion() string {
	file := os.Getenv("VERSION")
	if file == "" {
		return "unknown"
	}
	return file
}

func GetServices() []string {
	return []string{"Create Profile", "Create Profile Media", "Get Profile", "Get Profile Media"}
}

func GetDatabaseConnectionStatus(connection *gorm.DB) mapping.DatabaseConnectionStatus {
	godotenv.Load(".env")
	tables := []string{"profiles", "profile_media"}

	environment := config.SetupEnvironment()
	connection, db := database.SetupDatabase(&environment.DatabaseConfig)
	defer db.Close()

	for _, table := range tables {
		log.Println("Checking table:", table)
		query := fmt.Sprintf("SELECT 1 from %s limit 1", table)
		if err := connection.Exec(query).Error; err != nil {
			return mapping.Disconnected
		}
	}

	return mapping.Connected
}

func GetServiceStatus() mapping.ServiceStatus {
	httpClient := &http.Client{
		Timeout: 3 * time.Second,
	}
	response, err := httpClient.Get("http://localhost:8081/profiles/info")
	if err != nil {
		return mapping.Error
	}
	if response.StatusCode != 200 {
		return mapping.Error
	}
	return mapping.Ok
}
