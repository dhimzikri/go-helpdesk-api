package main

import (
	"bufio"
	"fmt"
	"golang-sqlserver-app/config"
	"golang-sqlserver-app/controllers"
	"golang-sqlserver-app/models"
	"golang-sqlserver-app/utils"
	"log"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize the database
	config.ConnectDB()

	// Perform login before starting the server
	fmt.Println("Welcome! Please log in to continue:")
	username, password := getLoginCredentials()
	if !validateCredentials(username, password) {
		log.Fatalln("Invalid credentials. Exiting the application.")
		return
	}
	fmt.Println("Login successful! Starting the server...")

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

	// Auth endpoints
	r.POST("/login", controllers.Login)
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

	// Start the server
	if err := r.Run("0.0.0.0:8686"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

// getLoginCredentials prompts the user for username and password
func getLoginCredentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	return username, password
}

// validateCredentials checks the user's input against the database
func validateCredentials(username, password string) bool {
	var user models.User
	if err := config.DB2.Where("user_name = ?", username).First(&user).Error; err != nil {
		fmt.Println("User not found")
		return false
	}

	if err := utils.ComparePasswords(user.Password, password); err != nil {
		fmt.Println("Invalid password")
		return false
	}

	return true
}
