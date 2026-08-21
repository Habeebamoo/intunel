package middlewares

import (
	"net/http"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-gonic/gin"
)

func RateLimiter() gin.HandlerFunc {
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Minute,
		Limit: 30,
	})

	return ratelimit.RateLimiter(store, &ratelimit.Options{
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
		ErrorHandler: func(c *gin.Context, info ratelimit.Info) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "Too many request"})
		},
	})
}