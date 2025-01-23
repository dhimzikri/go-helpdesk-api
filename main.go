package main

import (
	"golang-sqlserver-app/config"
	"golang-sqlserver-app/routes"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize the database
	config.ConnectDB()

	// Set up the Gin router
	r := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	if err := r.SetTrustedProxies([]string{"172.20.18.103", "172.16.6.34"}); err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	// Middleware to get the client IP from X-Forwarded-For
	r.Use(func(c *gin.Context) {
		clientIP := c.Request.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = c.ClientIP() // Fallback to Gin's default method if X-Forwarded-For is not set
		}

		log.Printf("Client IP: %s", clientIP) // Log the client IP
		c.Next()
	})

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://172.16.6.34:81"}, // Allow all origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Setup routes
	routes.SetupRoutes(r)

	// Start the server using HTTP
	if err := r.Run("0.0.0.0:8686"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
