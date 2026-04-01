package main

import (
	"fmt"

	"dfunani/homehub-profiles/src/database"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	fmt.Println(database.Auth())
}
