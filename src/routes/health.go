package routes

import (
	"dfunani/homehub-profiles/src/mapping"
	"dfunani/homehub-profiles/src/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func healthHandler(context *gin.Context, connection *gorm.DB) {
	log.Printf("[Health] handler received request")
	health := mapping.HealthResponse{
		Version:     utils.GetVersion(),
		Environment: utils.GetEnvironment(),
		Database:    utils.GetDatabaseConnectionStatus(connection),
		Status:      utils.GetServiceStatus(),
	}
	context.JSON(http.StatusOK, health)
}

func infoHandler(context *gin.Context, _ *gorm.DB) {
	log.Printf("[Info] handler received request")
	info := mapping.InfoResponse{
		ServiceName: utils.GetServiceName(),
		Version:     utils.GetVersion(),
		Services:    utils.GetServices(),
		Environment: utils.GetEnvironment(),
	}
	context.JSON(http.StatusOK, info)
}
