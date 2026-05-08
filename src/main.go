package main

import (
	dbConfig "dfunani/homehub-profiles/src/config"
	"dfunani/homehub-profiles/src/database"
	"dfunani/homehub-profiles/src/middleware"
	"dfunani/homehub-profiles/src/routes"
	"fmt"
	"time"

	_ "ariga.io/atlas-provider-gorm/gormschema"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Application starting...")
	godotenv.Load(".env")

	environment := dbConfig.SetupEnvironment()
	connection, db := database.SetupDatabase(&environment.DatabaseConfig)
	defer db.Close()
	database.RunMigrations(&environment.DatabaseConfig)

	app := gin.New()
	app.Use(middleware.Recovery())
	app.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(params gin.LogFormatterParams) string {
			return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
				params.ClientIP,
				params.TimeStamp.Format(time.RFC1123),
				params.Method,
				params.Path,
				params.Request.Proto,
				params.StatusCode,
				params.Latency,
				params.Request.UserAgent(),
				params.ErrorMessage,
			)
		},
	}))

	service_router := app.Group("/profiles")

	routes.BuildRoutes(service_router, connection)

	app.Run(":80")
}
