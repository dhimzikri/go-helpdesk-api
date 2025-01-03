package main

import (
	"golang-sqlserver-app/config"
	"golang-sqlserver-app/controllers"
	"golang-sqlserver-app/middleware"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize the database
	config.ConnectDB()

	// Set up the Gin router
	r := gin.Default()
	r.Use(middleware.RateLimitMiddleware)

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Auth endpoints
	r.POST("/login", controllers.Login) // /user/login route
	r.POST("/register", controllers.Register)
	r.GET("/logout", controllers.Logout)

	// CRUD endpoints
	r.GET("/getData", controllers.GetData)
	r.GET("/getCase", controllers.GetCase)
	r.POST("/getCase", controllers.GetCase)
	r.GET("/getRelation", controllers.GetRelation)
	r.GET("/getStatus", controllers.GetTblStatus)
	r.GET("/getAgreement", controllers.AgreementNoHandler())
	r.GET("/getContact", controllers.GetContact)
	r.GET("/getSubType", controllers.GetSubType)
	r.POST("/getSaveCase", controllers.SaveCaseHandler)

	// Start the server using HTTP
	if err := r.Run("0.0.0.0:8686"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
