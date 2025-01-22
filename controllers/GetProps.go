package controllers

import (
	"fmt"
	"golang-sqlserver-app/config"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

func GetTblType(c *gin.Context) {
	// Define a slice of maps to hold the query results
	var results []map[string]interface{}

	// Execute the query using GORM's raw SQL method
	if err := config.DB.Table("tblType").Find(&results).Error; err != nil {
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

func GetBranch(c *gin.Context) {
	// Parse query parameters
	query := c.Query("query") // e.g., ?query=some_value
	col := c.Query("col")     // e.g., ?col=name

	// Initialize the result structure
	result := gin.H{"success": false}

	// Base query with UNION
	sqlStr := `
		SELECT branchid, branchfullname FROM (
			SELECT id_cost_center AS branchid, name AS branchfullname FROM portal_ext.dbo.cost_centers
			UNION
			SELECT branchid, branchfullname FROM [172.16.4.31].SGFDB.dbo.branch
		) AS a
	`

	// Add filtering if query and col are provided
	if query != "" && col != "" {
		sqlStr += " WHERE " + col + " LIKE ?"
	}

	// Execute the query
	var branches []map[string]interface{}
	var err error

	if query != "" && col != "" {
		err = config.DB.Raw(sqlStr, "%"+query+"%").Scan(&branches).Error
	} else {
		err = config.DB.Raw(sqlStr).Scan(&branches).Error
	}

	if err != nil {
		result["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	// Check if data is retrieved
	if len(branches) == 0 {
		result["data"] = []map[string]interface{}{} // Return an empty array
		result["total"] = 0
	} else {
		result["success"] = true
		result["data"] = branches
		result["total"] = len(branches)
	}

	// Return the result
	c.JSON(http.StatusOK, result)
}

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

func GetTblStatus(c *gin.Context) {
	// Define a slice of maps to hold the query results
	var results []map[string]interface{}

	// Execute the query using GORM's raw SQL method
	if err := config.DB.Table("Status").Find(&results).Error; err != nil {
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

// Initialize the cache with a default expiration of 5 minutes and cleanup interval of 10 minutes
var c = cache.New(5*time.Minute, 10*time.Minute)

// ExecuteStoredProcedure executes the stored procedure with caching
func ReadAgreement(flagCompany string, searchParams map[string]string) ([]map[string]interface{}, error) {
	db := config.GetDB() // Get the database connection

	// Create a unique cache key based on flagCompany and searchParams
	cacheKey := fmt.Sprintf("%s-%v", flagCompany, searchParams)

	// Check if the result is already in the cache
	if cachedResult, found := c.Get(cacheKey); found {
		fmt.Println("Cache hit")
		return cachedResult.([]map[string]interface{}), nil
	}

	// Build the where clause dynamically based on search parameters
	whereClause := "0=0"
	if agreementNo, ok := searchParams["agreementno"]; ok && agreementNo != "" {
		whereClause += fmt.Sprintf(" and a.agreementno like '%%%s%%'", escapeSingleQuotes(agreementNo))
	}
	if customerName, ok := searchParams["customername"]; ok && customerName != "" {
		whereClause += fmt.Sprintf(" and b.name like '%%%s%%'", escapeSingleQuotes(customerName))
	}

	// Correctly escape the query
	query := fmt.Sprintf("EXEC sp_getInformation @flagcompany = '%s', @where = '%s'",
		escapeSingleQuotes(flagCompany), escapeSingleQuotes(whereClause))

	// Execute the query
	rows, err := db.Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Parse the results from the query
	var results []map[string]interface{}
	for rows.Next() {
		columns, _ := rows.Columns()
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)
		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	// Store the result in the cache for future use
	c.Set(cacheKey, results, cache.DefaultExpiration)

	return results, nil
}

// GetInfoHandler handles the API request for executing the stored procedure with caching
func AgreementNoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get query parameters
		flagCompany := c.Query("flagcompany")
		searchParams := map[string]string{
			"agreementno":   c.Query("agreementno"),
			"applicationid": c.Query("applicationid"),
			"customername":  c.Query("customername"),
		}

		// Call the stored procedure
		results, err := ReadAgreement(flagCompany, searchParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Wrap results in a "data" field
		response := gin.H{
			"data": results,
		}

		// Return the results as JSON
		c.JSON(http.StatusOK, response)
	}
}

// escapeSingleQuotes escapes single quotes in strings
func escapeSingleQuotes(input string) string {
	return strings.ReplaceAll(input, "'", "''")
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

func GetSubType(c *gin.Context) {
	typeID, _ := strconv.Atoi(c.Query("typeid"))
	query := c.Query("query")
	col := c.Query("col")

	// Base query for tblSubType
	src := config.DB.Table("tblSubType").Select("tblSubType.*, cost_centers.name AS cost_center_name")

	// Build raw join query manually using db2 for cost_centers
	joinQuery := "(SELECT id, name FROM portal_ext.dbo.cost_centers) AS cost_centers"
	src = src.Joins("LEFT JOIN " + joinQuery + " ON tblSubType.cost_center = cost_centers.id")

	// Add filters
	src = src.Where("typeid = ? AND isactive = ?", typeID, 1)
	if query != "" && col != "" {
		src = src.Where(col+" LIKE ?", "%"+query+"%")
	}

	// Fetch results into a map
	var resultData []map[string]interface{}
	result := src.Find(&resultData)

	// Handle errors or empty results
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   result.Error.Error(),
		})
		return
	}

	if len(resultData) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"total":   0,
		})
		return
	}
	// Convert keys to lowercase
	for i := range resultData {
		lowercaseMap := make(map[string]interface{})
		for key, value := range resultData[i] {
			lowercaseMap[strings.ToLower(key)] = value
		}
		resultData[i] = lowercaseMap
	}

	// Send successful response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   result.RowsAffected,
		"data":    resultData,
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

type CostCenter struct {
	BranchID       int    `json:"branchid"`
	BranchFullName string `json:"branchfullname"`
}

// Define the response structure
type Response struct {
	Total   int          `json:"total"`
	Success bool         `json:"success"`
	Data    []CostCenter `json:"data"`
}

func GetBranchID(c *gin.Context) {
	// Define a slice of maps to hold the query results
	var results []map[string]interface{}

	// Execute the query using GORM's raw SQL method
	if err := config.DB2.Table("cost_centers").
		Select("id_cost_center as branchid, name as branchfullname").
		Find(&results).Error; err != nil {
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

func CreateTblType(c *gin.Context) {
	// Parse the request parameters
	tblTypeID := c.PostForm("tblTypeid")
	name := c.PostForm("name")
	void := c.PostForm("void")

	// Convert the "void" value to a boolean-compatible integer
	voidInt := 0
	if void == "on" {
		voidInt = 1
	}

	// Simulate fetching the current user from session or token
	userID := "current_user" // Replace with actual logic to fetch user from session or context

	// Build the SQL query to execute the stored procedure
	sql := fmt.Sprintf("EXEC sp_insert_tblType '%s', '%s', '%d', '%s'", tblTypeID, name, voidInt, userID)

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
