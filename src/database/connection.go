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
		panic("Failed to connect to database: " + err.Error())
	}
	log.Println("Database connected successfully")
	return db
}

func GetDB(connection *gorm.DB) *sql.DB {
	db, err := connection.DB()
	if err != nil {
		panic("Failed to get database connection: " + err.Error())
	}

	return db
}

func RDSAuth(host string, user string) string {
	log.Println("RDSAuth starting...")
	var dbHost string = os.Getenv("AWS_ENDPOINT_URL")
	var dbEndpoint string = fmt.Sprintf("%s:%d", dbHost, 5432)
	var region string = os.Getenv("AWS_REGION")

	if dbHost == "" {
		dbHost = host
	}
	if region == "" {
		region = "eu-north-1"
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("Failed to load AWS config: " + err.Error())
	}

	password, err := auth.BuildAuthToken(
		context.TODO(), dbEndpoint, region, user, cfg.Credentials)
	if err != nil {
		panic("Failed to create authentication token: " + err.Error())
	}

	log.Println("RDSAuth complete")
	return password
}
