package controllers

import (
	"golang-sqlserver-app/config"
	"golang-sqlserver-app/models"
	"golang-sqlserver-app/utils"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Login User
func Login(c *gin.Context) {
	var login models.Login
	if err := c.ShouldBindJSON(&login); err != nil {
		utils.Response(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var user models.User
	if err := config.DB2.Where("user_name = ?", login.Username).First(&user).Error; err != nil {
		utils.Response(c, http.StatusNotFound, "User not found", nil)
		return
	}

	if err := utils.ComparePasswords(user.Password, login.Password); err != nil {
		utils.Response(c, http.StatusUnauthorized, "Invalid credentials", nil)
		return
	}

	// Save user_name in session
	session := sessions.Default(c)
	session.Set("user_name", user.Username)
	session.Save()

	token, err := utils.CreateToken(&user)
	if err != nil {
		utils.Response(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	responseData := map[string]interface{}{
		"data_user": user,
		"token":     token,
	}

	utils.Response(c, http.StatusOK, "Successfully Logged In", responseData)
}

func SaveUser(user *models.User) error {
	encryptedPassword, err := utils.EncryptPassword(user.Password)
	if err != nil {
		return err
	}

	user.Password = encryptedPassword
	return config.DB2.Create(user).Error
}

// Register User
func Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.Response(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Encrypt the user's password
	hashedPassword, err := utils.EncryptPassword(user.Password)
	if err != nil {
		utils.Response(c, http.StatusInternalServerError, "Error encrypting password", nil)
		return
	}
	user.Password = hashedPassword

	// Start a transaction
	tx := config.DB2.Begin()

	// Disable triggers for the `users` table
	if err := tx.Exec("DISABLE TRIGGER ALL ON users").Error; err != nil {
		tx.Rollback()
		utils.Response(c, http.StatusInternalServerError, "Failed to disable triggers", nil)
		return
	}

	// Save the user to the database
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		utils.Response(c, http.StatusInternalServerError, "Error saving user", nil)
		return
	}

	// Re-enable triggers for the `users` table
	if err := tx.Exec("ENABLE TRIGGER ALL ON users").Error; err != nil {
		tx.Rollback()
		utils.Response(c, http.StatusInternalServerError, "Failed to enable triggers", nil)
		return
	}

	// Commit the transaction
	tx.Commit()

	utils.Response(c, http.StatusOK, "User registered successfully", nil)
}

// Logout User
func Logout(c *gin.Context) {
	// Example logic for logout
	utils.RespondJSON(c, http.StatusOK, true, "Logged out successfully", nil)
}
