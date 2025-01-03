package controllers

import (
	"context"
	"fmt"
	"golang-sqlserver-app/config"
	"golang-sqlserver-app/models"
	"golang-sqlserver-app/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Login User
func Login(c *gin.Context) {
	var login models.Login

	// Decode the incoming request body to populate the login model
	if err := c.ShouldBindJSON(&login); err != nil {
		utils.Response(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var user models.User

	// Context for retry logic
	ctx := context.Background()
	err := utils.DoRetry(ctx, func(ctx context.Context) error {
		// Use GORM to execute the query
		if err := config.DB2.Where("user_name = ? AND user_password = ?", login.Username, login.Password).First(&user).Error; err != nil {
			// Retry logic
			if err.Error() == "record not found" {
				utils.Response(c, http.StatusNotFound, "User not found", nil)
			} else {
				fmt.Printf("Retrying request after error: %v\n", err)
				utils.Response(c, http.StatusInternalServerError, err.Error(), nil)
			}
			return utils.RetryableError(err)
		}
		return nil
	})

	// Check if retry failed
	if err != nil {
		return
	}

	// Create token for the user
	token, err := utils.CreateToken(&user)
	if err != nil {
		utils.Response(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Prepare the response data
	responseData := map[string]interface{}{
		"data_user": user,
		"token":     token,
	}

	// Send success response
	utils.Response(c, http.StatusOK, "Successfully Logged In", responseData)
}

// Register User
func Register(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondJSON(c, http.StatusBadRequest, false, "Invalid input", nil)
		return
	}

	if err := config.DB.Create(&input).Error; err != nil {
		utils.RespondJSON(c, http.StatusInternalServerError, false, "Failed to register user", nil)
		return
	}

	utils.RespondJSON(c, http.StatusOK, true, "User registered successfully", input)
}

// Logout User
func Logout(c *gin.Context) {
	// Example logic for logout
	utils.RespondJSON(c, http.StatusOK, true, "Logged out successfully", nil)
}
