package routes

import (
	"log"

	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func BuildRoutes(service_router *gin.RouterGroup, connection *gorm.DB) {
	log.Printf("[Routes] registering routes")
	buildHealthRoutes(service_router, connection)
	buildProfileRoutesV1(service_router, connection)
	log.Printf("[Routes] all routes registered")
}

func buildHealthRoutes(service_router *gin.RouterGroup, connection *gorm.DB) {
	health_router := service_router.Group("")
	health_router.GET("/health", logger.SetLogger(), func(c *gin.Context) {
		healthHandler(c, connection)
	})
	health_router.GET("/info", logger.SetLogger(), func(c *gin.Context) {
		infoHandler(c, connection)
	})
	log.Printf("[Health] Routes registered")
}

func buildProfileRoutesV1(service_router *gin.RouterGroup, connection *gorm.DB) {
	profile_router := service_router.Group("/api/v1")
	profile_router.POST("/profiles", logger.SetLogger(), func(c *gin.Context) {
		createProfile(c, connection)
	})
	profile_router.POST("/profiles/:id/media", logger.SetLogger(), func(c *gin.Context) {
		id := c.Param("id")
		createProfileMedia(c, connection, id)
	})
	profile_router.GET("/profiles/:id", logger.SetLogger(), func(c *gin.Context) {
		id := c.Param("id")
		getProfile(c, connection, id)
	})
	profile_router.GET("/profiles/:id/media", logger.SetLogger(), func(c *gin.Context) {
		id := c.Param("id")
		getProfileMedia(c, connection, id)
	})
	log.Printf("[Profile] v1 Routes registered")
}
