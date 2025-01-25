package controllers

import (
	"fmt"
	"golang-sqlserver-app/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetTblPriority(c *gin.Context) {
	// Define a slice of maps to hold the query results
	var results []map[string]interface{}

	// Execute the query using GORM's raw SQL method
	if err := config.DB.Table("Priority").Find(&results).Error; err != nil {
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

func GetContact(c *gin.Context) {
	// Define a slice of maps to hold the query results
	var results []map[string]interface{}

	// Execute the query using GORM's raw SQL method
	if err := config.DB.Table("Contact").Find(&results).Error; err != nil {
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

func GetTblStatus(c *gin.Context) {
	// Define a slice of maps to hold the query results
	var results []map[string]interface{}

	// Execute the query using GORM's raw SQL method
	if err := config.DB.Table("Status").Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Convert all keys in the results to lowercase
	lowercaseResults := make([]map[string]interface{}, len(results))
	for i, result := range results {
		lowercaseMap := make(map[string]interface{})
		for key, value := range result {
			lowercaseMap[strings.ToLower(key)] = value
		}
		lowercaseResults[i] = lowercaseMap
	}

	// Check if any data was retrieved
	if len(lowercaseResults) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "data": []map[string]interface{}{}})
		return
	}

	// Return the lowercase results as JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   len(lowercaseResults),
		"data":    lowercaseResults,
	})
}

func GetHolidays(c *gin.Context) {
	// Define a slice of maps to hold the query results
	var results []map[string]interface{}

	// Execute the query using GORM's raw SQL method
	if err := config.DB.Table("holiday").Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Convert all keys in the results to lowercase
	lowercaseResults := make([]map[string]interface{}, len(results))
	for i, result := range results {
		lowercaseMap := make(map[string]interface{})
		for key, value := range result {
			lowercaseMap[strings.ToLower(key)] = value
		}
		lowercaseResults[i] = lowercaseMap
	}

	// Check if any data was retrieved
	if len(lowercaseResults) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "data": []map[string]interface{}{}})
		return
	}

	// Return the lowercase results as JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   len(lowercaseResults),
		"data":    lowercaseResults,
	})
}

func SaveStatus(c *gin.Context) {
	// Parse raw JSON payload
	var requestData map[string]interface{}
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid JSON payload"})
		return
	}

	// Extract fields from requestData
	statusID := "null"
	if id, ok := requestData["statusid"]; ok {
		statusID = fmt.Sprintf("'%v'", id)
	}
	statusName := requestData["statusname"].(string)
	value := requestData["value"].(string)
	description := requestData["description"].(string)

	isActive := 0
	if active, ok := requestData["isactive"]; ok && active.(bool) {
		isActive = 1
	}

	userID := "8023" // Replace with session-based user retrieval logic

	// Log the parsed data for debugging
	fmt.Printf("Parsed Data: %+v\n", requestData)

	// Build SQL query
	sqlquery := fmt.Sprintf(`
        set nocount on;
        exec sp_insertstatus 
            @statusid = %s,
            @statusname = '%s',
            @usrupd = '%s',
            @dtmupd = null,
            @value = '%s',
            @isactive = %d,
            @description = '%s';
        select '1' as data;
    `,
		statusID,
		statusName,
		userID,
		value,
		isActive,
		description)

	// Execute SQL query
	err := config.DB.Exec(sqlquery).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": err.Error()})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": ""})
}
