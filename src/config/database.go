package config

import (
	"context"
	"os"
	"strconv"

	dbpkg "dfunani/homehub-profiles/src/database"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
)

// DatabaseConfig is the shared DB settings type (alias to database.DatabaseConfig).
type DatabaseConfig = dbpkg.DatabaseConfig

func GetDatabaseConfig(environment string) DatabaseConfig {
	port := os.Getenv("POSTGRES_PORT")
	portInt, err := strconv.Atoi(port)
	if err != nil {
		portInt = 5432
	}
	host := os.Getenv("POSTGRES_HOST")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "eu-north-1"
	}

	cfg, err := awscfg.LoadDefaultConfig(context.TODO(), awscfg.WithRegion(region))
	if err != nil {
		panic("failed to load AWS config: " + err.Error())
	}

	if environment == "local" {
		return DatabaseConfig{
			Host:      host,
			User:      user,
			Password:  password,
			DBName:    dbName,
			Port:      portInt,
			Region:    region,
			RDSConfig: &cfg,
			SSLMode:   "disable",
		}
	}

	authToken := dbpkg.RDSAuth(host, user)
	return DatabaseConfig{
		Host:      host,
		User:      user,
		Password:  authToken,
		DBName:    dbName,
		Port:      portInt,
		Region:    region,
		RDSConfig: &cfg,
		SSLMode:   "require",
	}
}
