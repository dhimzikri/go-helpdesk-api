package controllers

import (
	"encoding/json"
	"fmt"
	"golang-sqlserver-app/config"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
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

func GetUsers(c *gin.Context) {
	// Query parameters for filtering and pagination
	query := c.Query("query")
	col := c.Query("col")
	start, _ := strconv.Atoi(c.DefaultQuery("start", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	// Base SQL query
	baseSQL := `
		SELECT user_name, real_name, b.name AS divisi
		FROM portal_ext.dbo.users a
		INNER JOIN portal_ext.dbo.cost_centers b ON a.cost_center_id = b.id_cost_center
		WHERE a.user_name NOT IN ('admin', 'opr') AND a.is_active = 1
	`

	// Add filtering
	if query != "" && col != "" {
		baseSQL += " AND " + col + " LIKE ?"
	}

	// Total count query
	var total int64
	countSQL := `
		SELECT COUNT(1) 
		FROM portal_ext.dbo.users a
		INNER JOIN portal_ext.dbo.cost_centers b ON a.cost_center_id = b.id_cost_center
		WHERE a.user_name NOT IN ('admin', 'opr') AND a.is_active = 1
	`
	if query != "" && col != "" {
		countSQL += " AND " + col + " LIKE ?"
		config.DB.Raw(countSQL, "%"+query+"%").Scan(&total)
	} else {
		config.DB.Raw(countSQL).Scan(&total)
	}

	// Paginated query
	paginatedSQL := baseSQL + " ORDER BY a.user_id OFFSET ? ROWS FETCH NEXT ? ROWS ONLY"
	var users []map[string]interface{}
	if query != "" && col != "" {
		config.DB.Raw(paginatedSQL, "%"+query+"%", start, limit).Scan(&users)
	} else {
		config.DB.Raw(paginatedSQL, start, limit).Scan(&users)
	}

	// Return JSON response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   total,
		"data":    users,
	})
}

// GetAssignments fetches assignment data without predefined structs
func GetAssignments(c *gin.Context) {

	// Query parameters for pagination
	start, _ := strconv.Atoi(c.DefaultQuery("start", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	query := c.Query("query")
	col := c.Query("col")

	// Base SQL query
	baseSQL := `SELECT * FROM v_tblAssignment WHERE 1=1`

	// Add filtering
	if query != "" && col != "" {
		baseSQL += " AND " + col + " LIKE ?"
	}

	// Total count query
	var total int64
	countSQL := `SELECT COUNT(1) FROM v_tblAssignment WHERE 1=1`
	if query != "" && col != "" {
		countSQL += " AND " + col + " LIKE ?"
		config.DB.Raw(countSQL, "%"+query+"%").Scan(&total)
	} else {
		config.DB.Raw(countSQL).Scan(&total)
	}

	// Paginated query
	paginatedSQL := baseSQL + " ORDER BY id OFFSET ? ROWS FETCH NEXT ? ROWS ONLY"
	var assignments []map[string]interface{}
	if query != "" && col != "" {
		config.DB.Raw(paginatedSQL, "%"+query+"%", start, limit).Scan(&assignments)
	} else {
		config.DB.Raw(paginatedSQL, start, limit).Scan(&assignments)
	}

	// Return JSON response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   total,
		"data":    assignments,
	})
}

func SaveAssign(c *gin.Context) {
	// Parse JSON input
	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid input"})
		return
	}

	// Extract values from the input
	id := request["id"].(string)
	costcenterid := request["costcenterid"].(string)
	employeeid := request["employeeid"].(string)
	isactive := 0
	if request["isactive"] == "on" {
		isactive = 1
	}
	employeeid_alternate := request["employeeid_alternate"].(string)
	isactive_alternate := 0
	if request["isactive_alternate"] == "on" {
		isactive_alternate = 1
	}

	// Assuming `userid` is retrieved from a session or request context
	userid := "example_user" // Replace this with your session-based or token-based user retrieval logic

	// Construct the SQL query using fmt.Sprintf
	sqlQuery := fmt.Sprintf(`
		set nocount on;
		exec sp_insertAssignment '%s', '%s', '%s', '%d', '%s', '%d', '%s';

		insert into tbllog ([tgl], [table_name], [menu], [script])
		select GETDATE(), '[tblAssignment]', 'assignment',
			   'exec sp_insertAssignment %s, %s, %s, %d, %s, %d, %s';
	`, id, costcenterid, employeeid, isactive, employeeid_alternate, isactive_alternate, userid,
		id, costcenterid, employeeid, isactive, employeeid_alternate, isactive_alternate, userid)

	// Execute the SQL query
	err := config.DB.Exec(sqlQuery).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": err.Error()})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": ""})
}
