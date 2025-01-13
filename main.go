package main

import (
	"golang-sqlserver-app/config"
	"golang-sqlserver-app/controllers"
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
	// r.Use(middleware.RateLimitMiddleware)
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	// Auth endpoints
	r.POST("/login", controllers.Login) // /user/login route
	r.POST("/register", controllers.Register)
	r.GET("/logout", controllers.Logout)

	// CRUD endpoints
	r.GET("/getCase", controllers.GetCase)
	r.POST("/getCase", controllers.GetCase)
	r.POST("/getSaveCase", controllers.SaveCaseHandler)
	r.GET("/getRelation", controllers.GetRelation)
	r.GET("/getStatus", controllers.GetTblStatus)
	r.GET("/gettblType", controllers.GetTblType)
	r.GET("/getAgreement", controllers.AgreementNoHandler())
	r.GET("/getContact", controllers.GetContact)
	r.GET("/getSubType", controllers.GetSubType)
	r.GET("/gettblPriority", controllers.GetTblPriority)
	// r.POST("/sendEmail", controllers.HandleSendEmail)

	// Start the server using HTTP
	if err := r.Run("0.0.0.0:8686"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
