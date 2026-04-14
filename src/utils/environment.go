package utils

import "os"

func GetEnvironment() string {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		return "local"
	}
	return environment
}
