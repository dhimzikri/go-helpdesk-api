// middleware.go
package middleware

import (
	"golang-sqlserver-app/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var rateLimiter = utils.NewRateLimiter(5, 10*time.Second) //set reset 10Sec for debug

func RateLimitMiddleware(c *gin.Context) {
	ip := c.ClientIP()

	// Check if the IP exceeds the rate limit
	if rateLimiter.CheckRateLimit(ip) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "Too many requests, please try again later.",
		})
		c.Abort()
		return
	}
	c.Next()
}
