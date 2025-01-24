package controllers

import (
	"fmt"
	"golang-sqlserver-app/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetRelation(c *gin.Context) {
	// Define a slice of maps to hold the query results
	var results []map[string]interface{}

	// Execute the query using GORM's raw SQL method
	if err := config.DB.Table("Relation").Find(&results).Error; err != nil {
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

func AddRelation(c *gin.Context) {
	// Parse the request parameters
	relationid := c.PostForm("relationid")
	description := c.PostForm("description")
	isactive := c.PostForm("isactive")

	// Simulate fetching the current user from session or token
	userID := "8023" // Replace with actual logic to fetch user from session or context

	// Build the SQL query to execute the stored procedure
	sql := fmt.Sprintf("exec sp_insertrelation '%s','%s','%s','%s'", relationid, description, isactive, userID)

	// Execute the query
	if err := config.DB.Exec(sql).Error; err != nil {
		// If there's an error, return it in the response
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"msg":     err.Error(),
		})
		return
	}

	// If the query was successful, return success
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "Data inserted successfully",
	})
}
