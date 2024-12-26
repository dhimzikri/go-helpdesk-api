package main

import (
	"golang-sqlserver-app/config"
	"golang-sqlserver-app/controllers"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize the database
	config.InitDB()

	// Set up the Gin router
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow this origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// CRUD endpoints
	r.GET("/gettblType", controllers.GetTblType)
	r.POST("/gettblType", controllers.GetTblType)
	r.GET("/getCase", controllers.GetCase)
	r.POST("/getCase", controllers.GetCase)

	// Start the server
	if err := r.RunTLS("0.0.0.0:8686", "combined.crt", "csr_cnaf_2024_2025.key"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
