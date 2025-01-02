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

// escapeSingleQuotes escapes single quotes in strings
func escapeSingleQuotes(input string) string {
	return strings.ReplaceAll(input, "'", "''")
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

		// Return the results as JSON
		c.JSON(http.StatusOK, results)
	}
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
	// Define a slice of maps to hold the query results
	var results []map[string]interface{}

	// Execute the query using GORM's raw SQL method
	if err := config.DB.Table("tblSubType").Find(&results).Error; err != nil {
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

// GetCases retrieves cases with pagination, search, and total count
var caseCache = cache.New(5*time.Minute, 10*time.Minute)

func GetCase(c *gin.Context) {
	// Get page number and limit from query parameters
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "30")

	// Convert page and limit to integers
	pageNum, limitNum := 1, 30
	pageNum, _ = strconv.Atoi(page)
	limitNum, _ = strconv.Atoi(limit)

	// Calculate offset for pagination
	offset := (pageNum - 1) * limitNum

	// Build dynamic WHERE conditions for search
	var conditions []string

	// Define a map of query parameters and their corresponding database fields
	fields := map[string]string{
		"ticketno":      "a.ticketno",
		"agreementno":   "a.agreementno",
		"applicationid": "a.applicationid",
		"customername":  "a.customername",
		"priority":      "d.Description",
		"status":        "e.statusname",
		"usrupd":        "a.usrupd",
	}

	// Iterate over the map and add conditions dynamically
	for param, field := range fields {
		if value := c.Query(param); value != "" {
			if param == "usrupd" { // Handle exact match for usrupd
				conditions = append(conditions, fmt.Sprintf("%s = '%s'", field, value))
			} else { // Handle LIKE conditions for other fields
				conditions = append(conditions, fmt.Sprintf("%s LIKE '%%%s%%'", field, value))
			}
		}
	}

	// Default condition
	conditions = append(conditions, "a.statusid <> 1")

	// Combine conditions with "AND"
	whereClause := strings.Join(conditions, " AND ")

	// Generate a unique cache key based on query parameters
	cacheKey := fmt.Sprintf("cases_page_%d_limit_%d_conditions_%s", pageNum, limitNum, whereClause)

	// Check if the data is already in the cache
	if cachedData, found := caseCache.Get(cacheKey); found {
		// Return cached data
		c.JSON(http.StatusOK, cachedData)
		return
	}

	// Build SQL query for counting total records
	sqlCount := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM [Case] a
		INNER JOIN tbltype b ON a.TypeID = b.TypeID
		INNER JOIN tblSubtype c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
		INNER JOIN Priority d ON a.PriorityID = d.PriorityID
		INNER JOIN status e ON a.statusid = e.statusid
		INNER JOIN contact f ON a.contactid = f.contactid
		INNER JOIN relation g ON a.relationid = g.relationid
		WHERE %s
	`, whereClause)

	var totalCount int64
	if err := config.DB.Raw(sqlCount).Scan(&totalCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count total records"})
		return
	}

	// Build SQL query for fetching paginated data
	sqlQuery := fmt.Sprintf(`
		SELECT 
			ROW_NUMBER() OVER (ORDER BY RIGHT(a.ticketno, 3) DESC) AS RowNumber,
			a.flagcompany, a.ticketno, a.agreementno, a.applicationid, a.customerid,
			a.typeid, b.description AS typedescriontion, a.subtypeid, c.SubDescription AS typesubdescriontion,
			a.priorityid, d.Description AS prioritydescription, a.statusid, e.statusname,
			e.description AS statusdescription, a.customername, a.branchid, a.description, a.phoneno,
			a.email, a.usrupd, a.dtmupd, a.date_cr, f.contactid, f.Description AS contactdescription,
			a.relationid, g.description AS relationdescription, a.relationname, a.callerid, a.email_, a.foragingdays
		FROM [Case] a
		INNER JOIN tbltype b ON a.TypeID = b.TypeID
		INNER JOIN tblSubtype c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
		INNER JOIN Priority d ON a.PriorityID = d.PriorityID
		INNER JOIN status e ON a.statusid = e.statusid
		INNER JOIN contact f ON a.contactid = f.contactid
		INNER JOIN relation g ON a.relationid = g.relationid
		WHERE %s
		ORDER BY RIGHT(a.ticketno, 3) DESC
		OFFSET %d ROWS FETCH NEXT %d ROWS ONLY
	`, whereClause, offset, limitNum)

	var cases []map[string]interface{}
	if err := config.DB.Raw(sqlQuery).Scan(&cases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch case data"})
		return
	}

	// Prepare the response
	response := gin.H{
		"total": totalCount,
		"cases": cases,
	}

	// Store the response in cache
	caseCache.Set(cacheKey, response, cache.DefaultExpiration)

	// Return the results in JSON format
	c.JSON(http.StatusOK, response)
}

func SaveCase(c *gin.Context) {
	var trancodeid string
	var nextID int
	var statusDescription string

	// Step 1: Check if status '2' exists in the status table
	config.DB.Raw("SELECT statusname FROM status WHERE statusid = ?", "2").Row().Scan(&statusDescription)

	// Step 2: Insert a case using sp_insertcase equivalent raw SQL
	err := config.DB.Exec(`
		EXEC sp_insertcase 
			'', 'cnaf', '410', '410101100562', '410A201107002013', '41000003166', 
			'M WAHYUDIN', '082299840427', 'info@cnaf.co.id', '2', '1', '25', '1', 
			'tesForGo', '8023', '1', '1', '', '8080', 'info@cnaf.co.id', 
			'2025-01-02', '2025-01-02', ?`, &trancodeid).Error

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Step 3: Log the insert into tbllog
	logQuery := fmt.Sprintf(`
		INSERT INTO tbllog ([tgl], [table_name], [menu], [script]) 
		SELECT GETDATE(), '[case]', 'gridcase', 
		'exec sp_insertcase , cnaf, 410, 410101100562, 410A201107002013, 41000003166, 
		M WAHYUDIN, 082299840427, info@cnaf.co.id, 2, 1, 25, 1, tesForGo, 8023, 
		1, 1, , 8080, info@cnaf.co.id, 2025-01-02, 2025-01-02'
	`)
	err = config.DB.Exec(logQuery).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Step 4: Get next ID for Case_History table based on TicketNo
	config.DB.Raw("SELECT IFNULL(MAX(id), 0) + 1 FROM Case_History WHERE TicketNo = ?", trancodeid).Row().Scan(&nextID)

	// Step 5: Insert into Case_History if not already present with statusid '2'
	var count int
	config.DB.Raw("SELECT COUNT(*) FROM Case_History WHERE TicketNo = ? AND StatusID = '2'", trancodeid).Row().Scan(&count)

	if count == 0 {
		// Insert into Case_History table
		err = config.DB.Exec(`
			INSERT INTO Case_History (id, ticketno, description, statusid, usrupd)
			VALUES (?, ?, ?, '2', '8023')`, nextID, trancodeid, statusDescription).Error

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	// Step 6: Update the status of the case to '2'
	err = config.DB.Exec("UPDATE [case] SET statusid = '2' WHERE TicketNo = ?", trancodeid).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Step 7: Log the case update to tbllog
	err = config.DB.Exec(logQuery).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Return the transaction ID as a response
	c.JSON(200, gin.H{"message": "Transaction created successfully"})
}

// Dummy function to represent email sending
// func SendEmail(subject, body, recipient string) {
// 	log.Printf("Email sent to %s with subject: %s\n", recipient, subject)
// }
