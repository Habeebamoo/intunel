package app

import (
	"github.com/Habeebamoo/intunel-backend/internal/handlers"
	"github.com/Habeebamoo/intunel-backend/internal/middlewares"
	"github.com/Habeebamoo/intunel-backend/internal/store"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.Engine,
	notificationHandler *handlers.NotificationHandler,
	idempotentStore *store.IdempotencyStore,
) {
	v1 := router.Group("/api/v1")

	//health check endpoint
	v1.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1.POST("/notify", 
		middlewares.RateLimiter(), 
		middlewares.IdempotencyMiddleware(idempotentStore),
		notificationHandler.SendNotification,
	)
}