package config

import (
	"log"
)

type Environment struct {
	DatabaseConfig DatabaseConfig
}

func SetupEnvironment() Environment {
	log.Println("Environment setup starting...")
	environment := Environment{
		DatabaseConfig: GetDatabaseConfig(),
	}
	log.Println("Environment setup complete")
	return environment
}
