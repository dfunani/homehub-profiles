package main

import (
	"fmt"
	"os"

	"dfunani/homehub-profiles/src/database"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load("aws.env")
	host := os.Getenv("POSTGRES_HOST")
	user := os.Getenv("POSTGRES_USER")
	password := database.RDSAuth(host, user)
	fmt.Println(password)
}
