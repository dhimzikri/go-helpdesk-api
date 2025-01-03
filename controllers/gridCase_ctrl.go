package controllers

import (
	"fmt"
	"golang-sqlserver-app/config"
	"golang-sqlserver-app/models"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

func GetData(c *gin.Context) {
	// Initialize the result structure
	result := gin.H{"success": false}

	// Define slices to hold the query results for each table
	var tblTypeData []map[string]interface{}
	var priorityData []map[string]interface{}
	var branchData []map[string]interface{}

	// Query tblType
	if err := config.DB.Table("tblType").Find(&tblTypeData).Error; err != nil {
		result["error"] = "Failed to fetch tblType: " + err.Error()
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	// Query Priority
	if err := config.DB.Table("Priority").Find(&priorityData).Error; err != nil {
		result["error"] = "Failed to fetch Priority: " + err.Error()
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	// Query Branch (using UNION query)
	sqlStr := `
		SELECT branchid, branchfullname FROM (
			SELECT id_cost_center AS branchid, name AS branchfullname FROM portal_ext.dbo.cost_centers
			UNION
			SELECT branchid, branchfullname FROM [172.16.4.31].SGFDB.dbo.branch
		) AS a
	`
	if err := config.DB.Raw(sqlStr).Scan(&branchData).Error; err != nil {
		result["error"] = "Failed to fetch Branch: " + err.Error()
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	// Consolidate the results
	result["success"] = true
	result["tblType"] = tblTypeData
	result["Priority"] = priorityData
	result["Branch"] = branchData
	result["total"] = map[string]int{
		"tblType":  len(tblTypeData),
		"Priority": len(priorityData),
		"Branch":   len(branchData),
	}

	// Return the consolidated result as JSON
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

func SaveCaseHandler(c *gin.Context) {
	var input models.CaseInput

	// Bind JSON input to struct
	if err := c.BindJSON(&input); err != nil {
		log.Printf("Failed to bind input data: %v", err)
		c.JSON(400, gin.H{"error": "Invalid input data"})
		return
	}

	// Step 1: Extract the first 3 digits of AgreementNo
	agreementPrefix := input.AgreementNo[:3]

	// Step 2: Append "CS"
	csPrefix := "CS"

	// Step 3: Get current date in YYMMDD format
	currentDate := time.Now().Format("060102")              // Example: "250103" for Jan 3, 2025
	currentDate_forAging := time.Now().Format("2006-01-02") // Example: "250103" for Jan 3, 2025

	// Step 4: Get the last used incremented value, starting from 107274
	var lastTicketNo string
	if err := config.DB.Raw("SELECT TOP 1 ticketno FROM [case] WHERE ticketno LIKE ? ORDER BY ticketno DESC", agreementPrefix+"%").Scan(&lastTicketNo).Error; err != nil {
		log.Printf("Failed to retrieve last ticket number: %v", err)
		c.JSON(500, gin.H{"error": "Failed to retrieve last ticket number"})
		return
	}

	// Step 5: Check if we retrieved a ticket number
	if lastTicketNo == "" {
		// If no ticket number exists, we start from the first ticket number
		lastTicketNo = agreementPrefix + csPrefix + currentDate + "000001" // Starting value
	}

	// Step 6: Extract the numeric part from the last ticket number
	ticketSuffix := lastTicketNo[len(agreementPrefix+csPrefix+currentDate):] // Extract numeric part after the fixed prefix
	lastIncrementedValue, err := strconv.Atoi(ticketSuffix)                  // Convert the numeric part to an integer
	if err != nil {
		log.Printf("Failed to convert ticket suffix to integer: %v", err)
		c.JSON(500, gin.H{"error": "Failed to convert ticket suffix"})
		return
	}

	// Step 7: Increment the last ticket number
	lastIncrementedValue++ // Increment the value
	log.Printf("Incremented value: %d", lastIncrementedValue)

	// Step 8: Generate the final ticket number by concatenating all parts
	ticketNo := fmt.Sprintf("%s%s%s%06d", agreementPrefix, csPrefix, currentDate, lastIncrementedValue)

	// Log the generated ticket number
	log.Printf("Generated ticket number: %s", ticketNo)

	// Insert into the `case` table
	insertCaseQuery := `INSERT INTO [case]
                (ticketno, flagcompany, branchid, agreementno, applicationid, customerid,
                customername, phoneno, email, statusid, typeid, subtypeid, priorityid,
                description, usrupd, contactid, relationid, relationname, callerid, email_, date_cr, foragingdays)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	log.Printf("Executing query: %s with params: %v", insertCaseQuery, input)

	if err := config.DB.Exec(insertCaseQuery,
		ticketNo, input.FlagCompany, input.BranchID, input.AgreementNo, input.ApplicationID, input.CustomerID,
		input.CustomerName, input.PhoneNo, input.Email, input.StatusID, input.TypeID, input.SubtypeID, input.PriorityID,
		input.Description, input.UserID, input.ContactID, input.RelationID, input.RelationName, input.CallerID, input.Email_, input.DateCr, currentDate_forAging).Error; err != nil {
		log.Printf("Failed to insert into case table: %v", err)
		c.JSON(500, gin.H{"error": "Failed to insert case data"})
		return
	}
	nextID := 1
	selectNextIDQuery := "SELECT ISNULL(MAX(id), 0) + 1 FROM Case_History WHERE TicketNo = ?"
	log.Printf("Executing query: %s with params: %v", selectNextIDQuery, ticketNo)

	if err := config.DB.Raw(selectNextIDQuery, ticketNo).Scan(&nextID).Error; err != nil {
		log.Printf("Failed to calculate next Case_History ID: %v", err)
		c.JSON(500, gin.H{"error": "Failed to calculate case history ID"})
		return
	}

	insertHistoryQuery := `INSERT INTO Case_History 
		(id, ticketno, description, statusid, usrupd) 
		VALUES (?, ?, ?, ?, ?)`
	log.Printf("Executing query: %s with params: %v", insertHistoryQuery, nextID, ticketNo, input.StatusDesc, input.StatusID, input.UserID)

	if err := config.DB.Exec(insertHistoryQuery, nextID, ticketNo, input.StatusDesc, input.StatusID, input.UserID).Error; err != nil {
		log.Printf("Failed to insert into case history table: %v", err)
		c.JSON(500, gin.H{"error": "Failed to insert case history data"})
		return
	}

	// Log the operation
	logQuery := `INSERT INTO tbllog (tgl, table_name, menu, script) 
		VALUES (?, '[case]', 'SaveCaseHandler', ?)`
	log.Printf("Executing query: %s with params: %v", logQuery, time.Now(), fmt.Sprintf("INSERT INTO [case] VALUES (%s, ...)", ticketNo))

	if err := config.DB.Exec(logQuery, time.Now(), fmt.Sprintf("INSERT INTO [case] VALUES (%s, ...)", ticketNo)).Error; err != nil {
		log.Printf("Failed to log operation: %v", err)
		c.JSON(500, gin.H{"error": "Failed to log operation"})
		return
	}

	// Success Response
	c.JSON(200, gin.H{"message": "Data inserted successfully", "ticketNo": ticketNo})
}

// Dummy function to represent email sending
// func SendEmail(subject, body, recipient string) {
// 	log.Printf("Email sent to %s with subject: %s\n", recipient, subject)
// }
