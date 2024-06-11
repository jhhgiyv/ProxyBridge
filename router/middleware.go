package router

import (
	"ProxyBridge/config"
	"github.com/gin-gonic/gin"
)

func authMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		key := context.GetHeader("X-API-KEY")
		if key == "" {
			context.JSON(401, gin.H{"error": "API key required"})
			context.Abort()
			return
		}
		if key != config.C.ApiKey {
			context.JSON(401, gin.H{"error": "Invalid API key"})
			context.Abort()
			return
		}
		context.Next()
	}
}
