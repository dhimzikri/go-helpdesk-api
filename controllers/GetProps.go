package controllers

import (
	"fmt"
	"golang-sqlserver-app/config"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
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

func UploadXLS(c *gin.Context) {
	sessionLogin := c.GetString("user_name")

	file, err := c.FormFile("upload_document")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "File is required"})
		return
	}

	inputFileName := "./uploads/" + file.Filename
	err = c.SaveUploadedFile(file, inputFileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"msg":     "Unable to save file: " + err.Error(),
		})
		return
	}

	f, err := excelize.OpenFile(inputFileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"msg":     "Unable to open file: " + err.Error(),
		})
		return
	}

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"msg":     "Unable to read rows: " + err.Error(),
		})
		return
	}

	if err := config.DB.Exec("TRUNCATE TABLE holiday_temp").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "Error truncating holiday_temp table"})
		return
	}

	var stringBuilder string
	count := 0

	// Start from the second row (index 1) to skip the header
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) >= 4 && row[0] != "" {
			if count > 0 {
				stringBuilder += ","
			}

			stringBuilder += fmt.Sprintf("('%s', '%s', '%s', '%s', '%s', '%s')",
				row[0],
				row[1],
				row[2],
				row[3],
				sessionLogin,
				time.Now().Format("2006-01-02 15:04:05"))
			count++

			if count == 200 {
				sqlStr := "INSERT INTO holiday_temp (tanggal, hari, keterangan, Detail, usrupd, dtmupd) VALUES " + stringBuilder
				if err := config.DB.Exec(sqlStr).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "Batch insert error"})
					return
				}
				count = 0
				stringBuilder = ""
			}
		}
	}

	if stringBuilder != "" {
		sqlStr := "INSERT INTO holiday_temp (tanggal, hari, keterangan, Detail, usrupd, dtmupd) VALUES " + stringBuilder
		if err := config.DB.Exec(sqlStr).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "Final insert error"})
			return
		}
	}

	// Validate data
	var totdata int
	config.DB.Raw("SELECT COUNT(*) FROM (SELECT tanggal, COUNT(*) as jml FROM holiday_temp GROUP BY tanggal HAVING COUNT(*) <> 1) AS Libur").Scan(&totdata)
	if totdata != 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Duplicate dates are not allowed"})
		return
	}

	config.DB.Raw("SELECT COUNT(DISTINCT YEAR(tanggal)) FROM holiday_temp").Scan(&totdata)
	if totdata > 1 {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Multiple years are not allowed"})
		return
	}

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		var year string
		tx.Raw("SELECT DISTINCT YEAR(tanggal) FROM holiday_temp").Scan(&year)

		if err := tx.Exec("DELETE FROM holiday WHERE YEAR(tanggal) = ?", year).Error; err != nil {
			return err
		}

		if err := tx.Exec("INSERT INTO holiday SELECT * FROM holiday_temp").Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "Error saving holiday data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "Data saved successfully"})
}
