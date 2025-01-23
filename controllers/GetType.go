package controllers

import (
	"fmt"
	"golang-sqlserver-app/config"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

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

func CreateTblSubType(c *gin.Context) {
	// Parse the request parameters
	tblTypeID := c.PostForm("typeid")
	subtypeid := c.PostForm("subtypeid")
	subdescription := c.PostForm("subdescription")
	isactive := c.PostForm("isactive")
	costcenter := c.PostForm("cost_center")
	estimasi := c.PostForm("estimasi")

	// // Simulate fetching the current user from session or token
	userID := "8023" // Replace with actual logic to fetch user from session or context

	// Build the SQL query to execute the stored procedure
	// sql := fmt.Sprintf("exec sp_insert_tblSubType '%s', '%s', '%s', '%s','%s', '%s', '%s'", tblTypeID, subtypeid, subdescription, estimasi, cost_center, isactive, userID)
	sql := fmt.Sprintf("exec sp_insert_tblSubType '%s','%s','%s','%s','%s','%s','%s'", tblTypeID, subtypeid, subdescription, isactive, userID, costcenter, estimasi)

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

func GetSubType(c *gin.Context) {
	db := config.GetDB()
	// Get query parameters
	query := c.DefaultQuery("query", "") // Default to an empty string if not provided
	col := c.DefaultQuery("col", "")     // Default to an empty string if not provided
	typeID := c.DefaultQuery("typeid", "0")

	// Convert typeID to integer
	typeIDInt, err := strconv.Atoi(typeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid typeid"})
		return
	}

	// Base condition
	src := fmt.Sprintf("typeid=%d AND isactive=1", typeIDInt)

	// Add filtering if query and col are provided
	if query != "" && col != "" {
		src = fmt.Sprintf("%s AND %s LIKE '%%%s%%'", src, col, query)
	}

	// Build SQL query
	sqlStr := fmt.Sprintf(`
		SELECT a.*, b.name AS cost_center_name
		FROM tblSubType a
		LEFT JOIN portal_ext..cost_centers b ON a.cost_center = b.id
		WHERE %s`, src)

	// Execute the query
	rows, err := db.Raw(sqlStr).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute query"})
		return
	}
	defer rows.Close()

	// Prepare results
	var results []map[string]interface{}
	for rows.Next() {
		// Dynamically scan columns into a map
		columns, _ := rows.Columns()
		values := make([]interface{}, len(columns))
		valuePointers := make([]interface{}, len(columns))
		for i := range values {
			valuePointers[i] = &values[i]
		}

		// Scan the row into value pointers
		if err := rows.Scan(valuePointers...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
			return
		}

		// Build the row as a map
		row := make(map[string]interface{})
		for i, colName := range columns {
			var v interface{}
			val := values[i]

			// Handle NULL values
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			row[colName] = v
		}

		results = append(results, row)
	}

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
