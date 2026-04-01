package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseConfig struct {
	Host      string
	Port      int
	User      string
	Password  string
	DBName    string
	Region    string
	RDSConfig *aws.Config
	// SSLMode is passed as sslmode query param (e.g. require, disable). Empty defaults to require in PostgresURLFromConfig.
	SSLMode string
}

func Connect(config *DatabaseConfig) *gorm.DB {
	dsn := PostgresURLFromConfig(config)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		panic(err)
	}
	log.Println("Database connected successfully")
	return db
}

func GetDB(connection *gorm.DB) *sql.DB {
	db, err := connection.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
		panic(err)
	}

	return db
}

func Auth() string {
	var dbUser string = os.Getenv("POSTGRES_USER")
	var dbHost string = os.Getenv("AWS_HOST")
	var dbPort int = 5432
	var dbEndpoint string = fmt.Sprintf("%s:%d", dbHost, dbPort)
	var region string = os.Getenv("AWS_REGION")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error: " + err.Error())
	}

	password, err := auth.BuildAuthToken(
		context.TODO(), dbEndpoint, region, dbUser, cfg.Credentials)
	if err != nil {
		panic("failed to create authentication token: " + err.Error())
	}

	return password
}
