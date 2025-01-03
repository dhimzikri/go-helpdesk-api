// middleware.go
package middleware

import (
	"golang-sqlserver-app/utils"
	"net/http"
	"time" // import the ratelimiter package

	"github.com/gin-gonic/gin"
)

// var rateLimiter = utils.RateLimiter.NewRateLimiter(5, 1*time.Minute) // 5 requests per minute
var rateLimiter = utils.NewRateLimiter(5, 1*time.Minute) // 5 requests per minute

// RateLimitMiddleware is the Gin middleware for enforcing rate limits
func RateLimitMiddleware(c *gin.Context) {
	// Get the client's IP address
	ip := c.ClientIP()

	// Check if the IP exceeds the rate limit
	if rateLimiter.CheckRateLimit(ip) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "Too many requests, please try again later.",
		})
		c.Abort() // Stop further processing of the request
		return
	}

	// Continue with the request
	c.Next()
}
