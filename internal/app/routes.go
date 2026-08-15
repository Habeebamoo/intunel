package app

import (
	"github.com/Habeebamoo/intunel-backend/internal/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.Engine,
	notificationHandler *handlers.NotificationHandler,
) {
	v1 := router.Group("/api/v1")
	v1.POST("/notify", notificationHandler.SendNotification)
}