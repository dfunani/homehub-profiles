package routes

import (
	"dfunani/homehub-profiles/src/database/serialisers"
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
	err := context.BindJSON(&media)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	media.ProfileID = profileID

	createdMedia := serialisers.CreateProfileMedia(connection, &media)
	fmt.Println(createdMedia)
	context.JSON(http.StatusOK, createdMedia)
}

func createProfile(context *gin.Context, connection *gorm.DB) {
	var profile serialisers.ProfileSerialiser
	err := context.BindJSON(&profile)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
