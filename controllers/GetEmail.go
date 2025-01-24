package controllers

import (
	"encoding/json"
	"fmt"
	"golang-sqlserver-app/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetEmailSetting(c *gin.Context) {
	// Define a slice of maps to hold the query results
	var results []map[string]interface{}

	// Execute the query using GORM's raw SQL method
	if err := config.DB.Table("tblSettingEmail").Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Check if any data was retrieved
	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "data": []map[string]interface{}{}})
		return
	}

	// Return the results as JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   len(results),
		"data":    results,
	})
}

func AddSetEmail(c *gin.Context) {
	// Read raw JSON from the request body
	rawData, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"msg":     "Failed to read request body",
		})
		return
	}

	// Parse the raw JSON into a map
	var requestBody map[string]string
	if err := json.Unmarshal(rawData, &requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"msg":     "Invalid JSON format",
		})
		return
	}

	// Extract values from the map
	id := requestBody["id"]
	name := requestBody["name"]
	typ := requestBody["type"]
	value := requestBody["value"]
	flag := requestBody["flag"]

	// Convert id to NULL if it's empty
	idValue := "NULL"
	if id != "" {
		idValue = id
	}

	// Build the SQL query
	sql := fmt.Sprintf("exec sp_insertSetEmail %s, '%s', '%s', '%s', '%s'", idValue, name, typ, value, flag)

	// Execute the query
	if err := config.DB.Exec(sql).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"msg":     err.Error(),
		})
		return
	}

	// Return success
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "Data inserted/updated successfully",
	})
}
