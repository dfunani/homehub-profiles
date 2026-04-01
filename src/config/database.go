package config

import (
	"context"
	"fmt"
	"os"
	"strconv"

	dbpkg "dfunani/homehub-profiles/src/database"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
)

// DatabaseConfig is the shared DB settings type (alias to database.DatabaseConfig).
type DatabaseConfig = dbpkg.DatabaseConfig

func GetDatabaseConfig() DatabaseConfig {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	opts := []func(*awscfg.LoadOptions) error{}
	if region != "" {
		opts = append(opts, awscfg.WithRegion(region))
	}
	if p := os.Getenv("AWS_PROFILE"); p != "" {
		opts = append(opts, awscfg.WithSharedConfigProfile(p))
	}

	cfg, err := awscfg.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		panic("failed to load AWS config: " + err.Error())
	}
	port := os.Getenv("POSTGRES_PORT")
	fmt.Println("POSTGRES_PORT: ", port)
	portInt, err := strconv.Atoi(port)
	if err != nil {
		panic("failed to convert POSTGRES_PORT to int: " + err.Error())
	}
	password := dbpkg.Auth()
	return DatabaseConfig{
		Host:      os.Getenv("POSTGRES_HOST"),
		User:      os.Getenv("POSTGRES_USER"),
		Password:  password,
		DBName:    os.Getenv("POSTGRES_DB"),
		Port:      portInt,
		Region:    region,
		RDSConfig: &cfg,
		SSLMode:   os.Getenv("POSTGRES_SSLMODE"),
	}
}
