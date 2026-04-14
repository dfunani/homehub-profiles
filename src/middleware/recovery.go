package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery returns middleware that catches panics, logs the stack, and responds with JSON.
// In debug mode the response includes the panic value; in release mode the message is generic.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(os.Stderr, func(c *gin.Context, recovered interface{}) {
		log.Printf("[panic] %s %s: %v\n%s",
			c.Request.Method, c.Request.URL.Path, recovered, debug.Stack())

		if c.Writer.Written() {
			return
		}

		message := "An unexpected error occurred."
		if gin.Mode() == gin.DebugMode {
			message = fmt.Sprintf("%v", recovered)
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_server_error",
			"message": message,
		})
	})
}
