//go:build tools

package main

import (
	dbConfig "dfunani/homehub-profiles/src/config"
	"dfunani/homehub-profiles/src/database"
	"dfunani/homehub-profiles/src/database/serialisers"
	"fmt"

	_ "ariga.io/atlas-provider-gorm/gormschema"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Application starting...")
	godotenv.Load(".env")

	environment := dbConfig.SetupEnvironment()
	connection := database.SetupDatabase(&environment.DatabaseConfig)
	database.RunMigrations(&environment.DatabaseConfig)

	userID := uuid.New()
	profile := serialisers.ProfileSerialiser{
		UserID:      userID,
		DisplayName: "Test profile",
		Headline:    "Hello",
		Status:      "active",
		Links:       []byte(`[]`),
		Preferences: []byte(`{}`),
	}

	created := serialisers.CreateProfile(connection, &profile)
	fmt.Println(created)

	media := serialisers.ProfileMediaSerialiser{
		ProfileID:  created.ID,
		StorageKey: "s3://bucket/avatars/" + created.ID.String() + ".jpg",
		Kind:       "avatar",
		SortOrder:  0,
	}
	createdMedia := serialisers.CreateProfileMedia(connection, &media)
	fmt.Println(createdMedia)

	retrieved := serialisers.GetProfile(connection, created.ID, true)
	fmt.Println(retrieved)

	println("profile id match:", retrieved.ID == created.ID)
	println("media count:", len(retrieved.Media))
}
