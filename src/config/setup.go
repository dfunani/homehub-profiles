package config

import (
	"log"
	"os"
)

type Environment struct {
	DatabaseConfig DatabaseConfig
	Environment    string
}

func SetupEnvironment() Environment {
	environment := os.Getenv("ENVIRONMENT")
	log.Printf("%s Environment setup starting...", environment)
	environmentConfig := Environment{
		DatabaseConfig: GetDatabaseConfig(environment),
		Environment:    environment,
	}
	log.Printf("%s Environment setup complete", environmentConfig.Environment)
	return environmentConfig
}
