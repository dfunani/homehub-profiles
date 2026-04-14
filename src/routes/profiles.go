package routes

import (
	"dfunani/homehub-profiles/src/database/serialisers"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func createProfileMedia(context *gin.Context, connection *gorm.DB, id string) {
	profileID := uuid.MustParse(id)

	profile := serialisers.GetProfile(connection, profileID, true)
	fmt.Println(profile)

	var media serialisers.ProfileMediaSerialiser
	json.NewDecoder(context.Request.Body).Decode(&media)
	media.ProfileID = profileID

	createdMedia := serialisers.CreateProfileMedia(connection, &media)
	fmt.Println(createdMedia)
	context.JSON(http.StatusOK, createdMedia)
}

func createProfile(context *gin.Context, connection *gorm.DB) {
	var profile serialisers.ProfileSerialiser
	json.NewDecoder(context.Request.Body).Decode(&profile)
	links, err := json.Marshal(profile.Links)
	if err != nil {
		log.Printf("[Profile] Error marshalling links: %s", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	profile.Links = links
	preferences, err := json.Marshal(profile.Preferences)
	if err != nil {
		log.Printf("[Profile] Error marshalling preferences: %s", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	profile.Preferences = preferences
	created := serialisers.CreateProfile(connection, &profile)
	log.Printf("[Profile] Profile created: %s", created.ID)
	context.JSON(http.StatusOK, created)
}

func getProfile(context *gin.Context, connection *gorm.DB, id string) {
	profileID := uuid.MustParse(id)
	profile := serialisers.GetProfile(connection, profileID, true)

	log.Printf("[Profile] Profile retrieved: %s", profile.ID)
	context.JSON(http.StatusOK, profile)
}

func getProfileMedia(context *gin.Context, connection *gorm.DB, id string) {
	profileID := uuid.MustParse(id)
	media := serialisers.GetProfileMedia(connection, profileID)
	log.Printf("[Profile] Media retrieved: %s", media.ID)
	context.JSON(http.StatusOK, media)
}
