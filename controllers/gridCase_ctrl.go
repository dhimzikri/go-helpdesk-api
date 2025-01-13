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

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
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
	// Retrieve user_name from session
	session := sessions.Default(c)
	userName := session.Get("user_name")
	if userName == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

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

	// Add condition for statusid and username from the session
	conditions = append(conditions, "a.statusid <> 1")
	conditions = append(conditions, fmt.Sprintf("a.usrupd = '%s'", userName))

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

func getEmailFlagFromTable(ticketNo string) (int, error) {
	var emailFlag bool
	err := config.DB.Table("tblSendEmail").
		Select("issend").
		Where("ticketno = ?", ticketNo).
		Scan(&emailFlag).Error

	if err == gorm.ErrRecordNotFound {
		return 1, nil // Return 0 if no record exists
	}
	return 0, err
}

func insertEmailRecord(ticketNo, emailto, flag, userName string, data map[string]string) error {
	emailRecord := models.EmailRecord{
		TicketNo: ticketNo,
		Subject:  data["subject"],
		Body:     data["body"],
		EmailTo:  emailto,
		Flag:     flag,
		IsSend:   1,
		UsrUpd:   userName,
		DtmUpd:   time.Now(),
	}

	// First try to insert
	result := config.DB.Table("tblSendEmail").Create(&emailRecord)
	if result.Error != nil {
		// If record already exists, update it
		if strings.Contains(result.Error.Error(), "duplicate") {
			return config.DB.Table("tblSendEmail").
				Where("ticketno = ? AND flag = ?", ticketNo, flag).
				Updates(map[string]interface{}{
					"subject_": data["subject"],
					"body_":    data["body"],
					"emailto":  emailto,
					"issend":   1,
					"usrupd":   userName,
					"dtmupd":   time.Now(),
				}).Error
		}
		return result.Error
	}
	return nil
}

// Function to update or insert email flag
func updateEmailFlag(ticketNo string, isSendEmail int, emailTo string, userID string, dateUpd time.Time) error {
	// Check if record exists using count
	var count int64
	err := config.DB.Table("tblSendEmail").
		Where("ticketno = ?", ticketNo).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to check record existence: %v", err)
	}

	if count > 0 {
		// Update existing record
		log.Printf("Updating existing email record for ticket %s", ticketNo)
		updates := map[string]interface{}{
			"issend":  isSendEmail,
			"emailto": emailTo,
			"usrupd":  userID,
			"dtmupd":  dateUpd,
		}

		return config.DB.Table("tblSendEmail").
			Where("ticketno = ?", ticketNo).
			Updates(updates).Error
	} else {
		// Insert new record
		log.Printf("Creating new email record for ticket %s", ticketNo)
		newRecord := models.EmailRecord{
			TicketNo: ticketNo,
			EmailTo:  emailTo,
			Flag:     "SendTicket", // Default flag, adjust if needed
			IsSend:   isSendEmail,
			UsrUpd:   userID,
			DtmUpd:   dateUpd,
		}

		return config.DB.Table("tblSendEmail").Create(&newRecord).Error
	}
}

// Helper function to check if record exists
func checkEmailExists(ticketNo string) (bool, error) {
	var count int64
	err := config.DB.Table("tblSendEmail").
		Where("ticketno = ?", ticketNo).
		Count(&count).Error

	return count > 0, err
}

// Helper function to determine the sendEmailFlag based on the ticket status
func getSendEmailFlag(ticketNo string, statusID int) (string, error) {
	emailFlag, err := getEmailFlagFromTable(ticketNo)
	if err != nil {
		return "", fmt.Errorf("failed to get email flag: %v", err)
	}

	if emailFlag == 1 {
		switch statusID {
		case 1, 2, 3, 4:
			return "SendTicket", nil // Non-extend
		case 5:
			return "SendTicketForExtend", nil // For extend
		}
	}
	return "", nil
}

// Helper function to send email if required
func sendEmail(ticketNo, sendEmailFlag string, caseData models.Case, userName string) error {
	// Log input parameters
	log.Printf("Starting email send process for ticket %s", ticketNo)
	log.Printf("Email parameters - Flag: %s, Customer: %s, Primary Email: %s, Secondary Email: %s, UserID: %s",
		sendEmailFlag, caseData.CustomerName, caseData.Email, caseData.Email_, caseData.UserID)

	// Determine which email to use
	// Priority is given to Email_ (secondary email) as per requirements
	emailToUse := caseData.Email_
	if emailToUse == "" {
		emailToUse = caseData.Email // Fallback to primary email if secondary is empty
	}

	// Validate final email
	if emailToUse == "" {
		return fmt.Errorf("no email address available for ticket %s", ticketNo)
	}

	// Prepare data for stored procedure
	data := fmt.Sprintf("%s|%s", caseData.CustomerName, ticketNo)

	// Log stored procedure call
	log.Printf("Calling sp_getsendEmail with params - TicketNo: %s, Email: %s, Data: %s, Flag: %s, UserID: %s",
		ticketNo, emailToUse, data, sendEmailFlag, caseData.UserID)

	sqlEmail := fmt.Sprintf(
		"set nocount on; exec sp_getsendEmail '%s', '%s', '%s', '%s', '%s', NULL;",
		ticketNo, emailToUse, data, sendEmailFlag, caseData.UserID,
	)

	var emailData []models.EmailData
	if err := config.DB.Raw(sqlEmail).Scan(&emailData).Error; err != nil {
		log.Printf("Failed to execute sp_getsendEmail: %v", err)
		return fmt.Errorf("stored procedure error: %v", err)
	}

	// Log stored procedure results
	log.Printf("Retrieved %d email records from sp_getsendEmail", len(emailData))

	if len(emailData) == 0 {
		log.Printf("No email data returned from sp_getsendEmail for ticket %s", ticketNo)
		return fmt.Errorf("no email data returned from stored procedure")
	}

	for idx, row := range emailData {
		log.Printf("Processing email record %d for ticket %s", idx+1, ticketNo)

		// Use the already validated emailToUse if row.Email is empty
		if row.Email == "" {
			row.Email = emailToUse
		}

		// Validate subject and body
		if row.Subject == "" || row.Body == "" {
			log.Printf("Invalid email content for ticket %s - Subject: %v, Body length: %d",
				ticketNo, row.Subject != "", len(row.Body))
			continue
		}

		// Insert/update record in tblSendEmail
		emailData := map[string]string{
			"subject": row.Subject,
			"body":    row.Body,
		}

		if err := insertEmailRecord(ticketNo, row.Email, sendEmailFlag, userName, emailData); err != nil {
			log.Printf("Failed to insert email record: %v", err)
			return err
		}

		// Send the actual email
		log.Printf("Attempting to send email to %s for ticket %s", row.Email, ticketNo)

		if err := SettingEmail(row.Subject, row.Body, row.Email, ticketNo); err != nil {
			log.Printf("Failed to send email: %v", err)
			// Update issend to 0 if email sending fails
			updateErr := config.DB.Table("tblSendEmail").
				Where("ticketno = ? AND flag = ?", ticketNo, sendEmailFlag).
				Update("issend", 0).Error
			if updateErr != nil {
				log.Printf("Failed to update issend flag: %v", updateErr)
			}
			return err
		}

		log.Printf("Successfully sent email for ticket %s to %s", ticketNo, row.Email)
	}

	return nil
}
func SaveCaseHandler(c *gin.Context) {
	// Retrieve user_name from session
	session := sessions.Default(c)
	userName := session.Get("user_name")
	if userName == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input models.Case
	var tblSendEmail models.SendEmail
	// Bind JSON input to struct
	if err := c.BindJSON(&input); err != nil {
		log.Printf("Failed to bind input data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	// Step 1: Check if the ticket exists in the `Case` table
	var existingTicket models.Case
	var isExisting bool

	if err := config.DB.Table("Case").Where("ticketno = ?", input.TicketNo).First(&existingTicket).Error; err == nil {
		isExisting = true
	}

	ticketNo := input.TicketNo

	if isExisting {
		// Update existing case logic...
		var existingTicket models.Case
		if err := config.DB.Table("Case").Where("ticketno = ?", input.TicketNo).First(&existingTicket).Error; err == nil {
			existingTicket.Description = input.Description
			existingTicket.PriorityID = input.PriorityID
			existingTicket.ContactID = input.ContactID
			existingTicket.RelationID = input.RelationID
			existingTicket.RelationName = input.RelationName
			existingTicket.CallerID = input.CallerID
			existingTicket.Email_ = input.Email_
			existingTicket.StatusID = input.StatusID
			existingTicket.DateCr = input.DateCr

			if err := config.DB.Table("Case").Omit("statusname", "flag", "is_send_email").Save(&existingTicket).Error; err != nil {
				log.Printf("Failed to update case: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update case data"})
				return
			}
			nextID := 1
			if err := config.DB.Raw("SELECT ISNULL(MAX(id), 0) + 1 FROM Case_History WHERE ticketno = ?", input.TicketNo).Scan(&nextID).Error; err != nil {
				log.Printf("Failed to calculate next Case_History ID: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate case history ID"})
				return
			}
			insertHistoryQuery := `INSERT INTO Case_History (id, ticketno, description, statusid, usrupd, dtmupd)
				VALUES (?, ?, ?, ?, ?, GETDATE())`
			if err := config.DB.Exec(insertHistoryQuery, nextID, input.TicketNo, input.Description, input.StatusID, userName).Error; err != nil {
				log.Printf("Failed to insert into Case_History: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert case history data"})
				return
			}
		}
	} else {
		// Create new case logic...
		ticketNo = generateNewTicketNo(input.AgreementNo)
		userName, ok := userName.(string)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user name"})
			return
		}
		if err := createNewCase(input, ticketNo, userName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert case data"})
			return
		}
	}

	// (ticketNo string, isSendEmail int, emailTo string, flag string, UserID string, DateUpd time.Time )
	if err := updateEmailFlag(ticketNo, tblSendEmail.IsSendEmail, input.Email_, input.UserID, input.DateUpd); err != nil {
		log.Printf("Failed to update email flag: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update email flag"})
		return
	}

	// Insert Case History
	if err := insertCaseHistory(ticketNo, input.Description, input.StatusID, userName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert case history data"})
		return
	}

	// Check if email should be sent
	sendEmailFlag, err := getSendEmailFlag(ticketNo, input.StatusID)
	if err != nil {
		log.Printf("Error checking email flag: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check email flag"})
		return
	}

	sendEmailFlag = ""
	switch input.StatusID {
	case 1, 2, 3, 4:
		sendEmailFlag = "SendTicket"
	case 5:
		sendEmailFlag = "SendTicketForExtend"
	}

	if sendEmailFlag != "" {
		userName, ok := userName.(string)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user name"})
			return
		}

		if err := sendEmail(ticketNo, sendEmailFlag, existingTicket, userName); err != nil {
			log.Printf("Email sending failed for ticket %s: %v", ticketNo, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email", "details": err.Error()})
			return
		}
		log.Printf("Email sent successfully for ticket %s with flag %s", ticketNo, sendEmailFlag)
	} else {
		log.Printf("No email flag set for ticket %s with StatusID: %d", ticketNo, input.StatusID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Case saved successfully", "ticketNo": ticketNo})
}

// Helper function to insert into the Case_History table
func insertCaseHistory(TicketNo, description string, statusID int, userName interface{}) error {
	var nextID int
	if err := config.DB.Raw("SELECT ISNULL(MAX(id), 0) + 1 FROM Case_History WHERE ticketno = ?", TicketNo).Scan(&nextID).Error; err != nil {
		log.Printf("Failed to calculate next Case_History ID: %v", err)
		return err
	}

	insertHistoryQuery := `INSERT INTO Case_History (id, ticketno, description, statusid, usrupd, dtmupd) 
		VALUES (?, ?, ?, ?, ?, GETDATE())`
	if err := config.DB.Exec(insertHistoryQuery, nextID, TicketNo, description, statusID, userName).Error; err != nil {
		log.Printf("Failed to insert into Case_History: %v", err)
		return err
	}
	return nil
}

// Helper function to generate a new ticket number
func generateNewTicketNo(agreementNo string) string {
	agreementPrefix := agreementNo[:3]
	csPrefix := "CS"
	currentDate := time.Now().Format("060102") // Example: "250103" for Jan 3, 2025
	var lastTicketNo string
	if err := config.DB.Raw("SELECT TOP 1 ticketno FROM [case] WHERE ticketno LIKE ? ORDER BY ticketno DESC", agreementPrefix+"%").Scan(&lastTicketNo).Error; err != nil {
		log.Printf("Failed to retrieve last ticket number: %v", err)
	}

	ticketSuffix := "000001"
	if lastTicketNo != "" {
		ticketSuffix = lastTicketNo[len(agreementPrefix+csPrefix+currentDate):]
		lastIncrementedValue, _ := strconv.Atoi(ticketSuffix)
		ticketSuffix = fmt.Sprintf("%06d", lastIncrementedValue+1)
	}

	return fmt.Sprintf("%s%s%s%s", agreementPrefix, csPrefix, currentDate, ticketSuffix)
}

// Helper function to create a new case entry in the database
func createNewCase(input models.Case, ticketNo, userName string) error {
	insertCaseQuery := `INSERT INTO [case]
        (ticketno, flagcompany, branchid, agreementno, applicationid, customerid,
        customername, phoneno, email, statusid, typeid, subtypeid, priorityid,
        description, usrupd, contactid, relationid, relationname, callerid, email_, date_cr, foragingdays)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	return config.DB.Exec(insertCaseQuery,
		ticketNo, input.FlagCompany, input.BranchID, input.AgreementNo, input.ApplicationID, input.CustomerID,
		input.CustomerName, input.PhoneNo, input.Email, input.StatusID, input.TypeID, input.SubtypeID, input.PriorityID,
		input.Description, userName, input.ContactID, input.RelationID, input.RelationName, input.CallerID, input.Email_,
		input.DateCr, time.Now().Format("2006-01-02")).Error // Added IsSendEmail here
}

// SettingEmail sends an email using the gomail package
func SettingEmail(subject, bodyEmail, recipient, trancodeid string) error {
	mail := gomail.NewMessage()

	mail.SetHeader("From", "passadmin@cnaf.co.id", "INFO CS CIMBNIAGA FINANCE")
	mail.SetHeader("To", recipient)
	mail.SetHeader("Subject", subject)
	mail.SetBody("text/html", bodyEmail)

	dialer := gomail.NewDialer(
		"smtp-mail.outlook.com", // SMTP host
		587,                     // Port
		"passadmin@cnaf.co.id",  // Email address
		"123456.Aa",             // Email password
	)

	if err := dialer.DialAndSend(mail); err != nil {
		log.Printf("Failed to send email for TrancodeID %s: %v", trancodeid, err)
		return err
	}

	log.Printf("Email sent successfully for TrancodeID %s", trancodeid)
	return nil
}

// func SaveCaseHandler(c *gin.Context) {
// 	// Retrieve user_name from session
// 	session := sessions.Default(c)
// 	userName := session.Get("user_name")
// 	if userName == nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 		return
// 	}

// 	var input models.Case
// 	// Bind JSON input to struct
// 	if err := c.BindJSON(&input); err != nil {
// 		log.Printf("Failed to bind input data: %v", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
// 		return
// 	}

// 	// Step 1: Check if the ticket exists in the `Case` table
// 	var existingTicket models.Case
// 	if err := config.DB.Table("Case").Where("ticketno = ?", input.TicketNo).First(&existingTicket).Error; err == nil {
// 		// Ticket exists, perform an update
// 		existingTicket.Description = input.Description
// 		existingTicket.PriorityID = input.PriorityID
// 		existingTicket.ContactID = input.ContactID
// 		existingTicket.RelationID = input.RelationID
// 		existingTicket.RelationName = input.RelationName
// 		existingTicket.CallerID = input.CallerID
// 		existingTicket.Email_ = input.Email_
// 		existingTicket.StatusID = input.StatusID
// 		existingTicket.DateCr = input.DateCr

// 		// Update the existing case
// 		if err := config.DB.Table("Case").Omit("statusname", "is_send_email").Save(&existingTicket).Error; err != nil {
// 			log.Printf("Failed to update case: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update case data"})
// 			return
// 		}

// 		// Insert/Update Case_History
// 		nextID := 1
// 		if err := config.DB.Raw("SELECT ISNULL(MAX(id), 0) + 1 FROM Case_History WHERE ticketno = ?", input.TicketNo).Scan(&nextID).Error; err != nil {
// 			log.Printf("Failed to calculate next Case_History ID: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate case history ID"})
// 			return
// 		}

// 		insertHistoryQuery := `INSERT INTO Case_History (id, ticketno, description, statusid, usrupd, dtmupd)
// 			VALUES (?, ?, ?, ?, ?, GETDATE())`
// 		if err := config.DB.Exec(insertHistoryQuery, nextID, input.TicketNo, input.Description, input.StatusID, userName).Error; err != nil {
// 			log.Printf("Failed to insert into Case_History: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert case history data"})
// 			return
// 		}
// 		// send email
// 		sendEmailFlag := ""
// 		if existingTicket.IsSendEmail == 1 {
// 			switch existingTicket.StatusID {
// 			case 1, 2, 3, 4:
// 				sendEmailFlag = "SendTicket" // non-extend
// 			case 5:
// 				sendEmailFlag = "SendTicketForExtend" // for extend
// 			}
// 		}
// 		if sendEmailFlag != "" {
// 			data := fmt.Sprintf("%s|%s", existingTicket.CustomerName, existingTicket.TicketNo)
// 			sqlEmail := fmt.Sprintf(
// 				"set nocount on; exec sp_getsendEmail '%s', '%s', '%s', '%s', '%s', NULL;", // Pass NULL for output parameter
// 				existingTicket.TicketNo, existingTicket.Email, data, sendEmailFlag, existingTicket.UserID,
// 			)

// 			// Log the SQL query for debugging
// 			log.Printf("Executing SQL: %s", sqlEmail)

// 			var emailData []models.EmailData
// 			if err := config.DB.Raw(sqlEmail).Scan(&emailData).Error; err != nil {
// 				log.Printf("Failed to execute query: %v", err)
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed", "details": err.Error()})
// 				return
// 			}

// 			for _, row := range emailData {
// 				// Log the email address being used
// 				log.Printf("Sending email to: %s", row.Email)

// 				// Validate the email address
// 				if row.Email == "" {
// 					log.Printf("Invalid email address for TrancodeID %s: empty email", existingTicket.TicketNo)
// 					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address", "details": "Email address cannot be empty"})
// 					return
// 				}

// 				if err := SettingEmail(row.Subject, row.Body, row.Email, existingTicket.TicketNo); err != nil {
// 					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email", "details": err.Error()})
// 					return
// 				}
// 			}
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "Case updated successfully", "ticketNo": existingTicket.TicketNo})
// 		return
// 	}

// 	// Ticket doesn't exist, create a new one
// 	agreementPrefix := input.AgreementNo[:3]
// 	csPrefix := "CS"
// 	currentDate := time.Now().Format("060102")              // Example: "250103" for Jan 3, 2025
// 	currentDate_forAging := time.Now().Format("2006-01-02") // Example: "2025-01-03"

// 	// Get the last ticket number and increment
// 	var lastTicketNo string
// 	if err := config.DB.Raw("SELECT TOP 1 ticketno FROM [case] WHERE ticketno LIKE ? ORDER BY ticketno DESC", agreementPrefix+"%").Scan(&lastTicketNo).Error; err != nil {
// 		log.Printf("Failed to retrieve last ticket number: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve last ticket number"})
// 		return
// 	}

// 	if lastTicketNo == "" {
// 		lastTicketNo = agreementPrefix + csPrefix + currentDate + "000001"
// 	}

// 	ticketSuffix := lastTicketNo[len(agreementPrefix+csPrefix+currentDate):]
// 	lastIncrementedValue, err := strconv.Atoi(ticketSuffix)
// 	if err != nil {
// 		log.Printf("Failed to convert ticket suffix to integer: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert ticket suffix"})
// 		return
// 	}
// 	lastIncrementedValue++

// 	ticketNo := fmt.Sprintf("%s%s%s%06d", agreementPrefix, csPrefix, currentDate, lastIncrementedValue)

// 	// Insert into the `Case` table for new ticket
// 	insertCaseQuery := `INSERT INTO [case]
//                 (ticketno, flagcompany, branchid, agreementno, applicationid, customerid,
//                 customername, phoneno, email, statusid, typeid, subtypeid, priorityid,
//                 description, usrupd, contactid, relationid, relationname, callerid, email_, date_cr, foragingdays)
// 		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
// 	if err := config.DB.Exec(insertCaseQuery,
// 		ticketNo, input.FlagCompany, input.BranchID, input.AgreementNo, input.ApplicationID, input.CustomerID,
// 		input.CustomerName, input.PhoneNo, input.Email, input.StatusID, input.TypeID, input.SubtypeID, input.PriorityID,
// 		input.Description, userName.(string), input.ContactID, input.RelationID, input.RelationName, input.CallerID, input.Email_, input.DateCr, currentDate_forAging).Error; err != nil {
// 		log.Printf("Failed to insert into case table: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert case data"})
// 		return
// 	}
// 	// send email
// 	sendEmailFlag := ""
// 	if existingTicket.IsSendEmail == 1 {
// 		switch existingTicket.StatusID {
// 		case 1, 2, 3, 4:
// 			sendEmailFlag = "SendTicket" // non-extend
// 		case 5:
// 			sendEmailFlag = "SendTicketForExtend" // for extend
// 		}
// 	}
// 	if sendEmailFlag != "" {
// 		data := fmt.Sprintf("%s|%s", existingTicket.CustomerName, existingTicket.TicketNo)
// 		sqlEmail := fmt.Sprintf(
// 			"set nocount on; exec sp_getsendEmail '%s', '%s', '%s', '%s', '%s', NULL;", // Pass NULL for output parameter
// 			existingTicket.TicketNo, existingTicket.Email, data, sendEmailFlag, existingTicket.UserID,
// 		)

// 		// Log the SQL query for debugging
// 		log.Printf("Executing SQL: %s", sqlEmail)

// 		var emailData []models.EmailData
// 		if err := config.DB.Raw(sqlEmail).Scan(&emailData).Error; err != nil {
// 			log.Printf("Failed to execute query: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed", "details": err.Error()})
// 			return
// 		}

// 		for _, row := range emailData {
// 			// Log the email address being used
// 			log.Printf("Sending email to: %s", row.Email)

// 			// Validate the email address
// 			if row.Email == "" {
// 				log.Printf("Invalid email address for TrancodeID %s: empty email", existingTicket.TicketNo)
// 				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address", "details": "Email address cannot be empty"})
// 				return
// 			}

// 			if err := SettingEmail(row.Subject, row.Body, row.Email, existingTicket.TicketNo); err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email", "details": err.Error()})
// 				return
// 			}
// 		}
// 	}

// 	// Insert into `Case_History` table for new case
// 	nextID := 1
// 	if err := config.DB.Raw("SELECT ISNULL(MAX(id), 0) + 1 FROM Case_History").Scan(&nextID).Error; err != nil {
// 		log.Printf("Failed to calculate next Case_History ID: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate case history ID"})
// 		return
// 	}

// 	insertHistoryQuery := `INSERT INTO Case_History
// 	(id, ticketno, description, statusid, usrupd, dtmupd)
// 	VALUES (?, ?, ?, ?, ?, GETDATE())`
// 	if err := config.DB.Exec(insertHistoryQuery, nextID, ticketNo, input.Description, input.StatusID, userName).Error; err != nil {
// 		log.Printf("Failed to insert into Case_History: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert case history data"})
// 		return
// 	}

// 	// Log the operation
// 	logQuery := `INSERT INTO tbllog (tgl, table_name, menu, script)
// 	VALUES (?, '[case]', 'SaveCaseHandler', ?)`
// 	if err := config.DB.Exec(logQuery, time.Now(), fmt.Sprintf("INSERT INTO [case] VALUES (%s, ...)", ticketNo)).Error; err != nil {
// 		log.Printf("Failed to log operation: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log operation"})
// 		return
// 	}

// 	// Success Response
// 	c.JSON(http.StatusOK, gin.H{"message": "Data inserted successfully", "ticketNo": ticketNo})
// }

// // HandleSendEmail implements the email logic
// func HandleSendEmail(c *gin.Context) {
// 	// var emailReq models.EmailRequest
// 	var emailReq models.Case
// 	if err := c.ShouldBindJSON(&emailReq); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
// 		return
// 	}

// 	sendEmailFlag := ""
// 	if emailReq.IsSendEmail == 1 {
// 		switch emailReq.StatusID {
// 		case 1, 2, 3, 4:
// 			sendEmailFlag = "SendTicket" // non-extend
// 		case 5:
// 			sendEmailFlag = "SendTicketForExtend" // for extend
// 		}
// 	}

// 	if sendEmailFlag != "" {
// 		data := fmt.Sprintf("%s|%s", emailReq.CustomerName, emailReq.TicketNo)
// 		sqlEmail := fmt.Sprintf(
// 			"set nocount on; exec sp_getsendEmail '%s', '%s', '%s', '%s', '%s', NULL;", // Pass NULL for output parameter
// 			emailReq.TicketNo, emailReq.Email, data, sendEmailFlag, emailReq.UserID,
// 		)

// 		// Log the SQL query for debugging
// 		log.Printf("Executing SQL: %s", sqlEmail)

// 		var emailData []models.EmailData
// 		if err := config.DB.Raw(sqlEmail).Scan(&emailData).Error; err != nil {
// 			log.Printf("Failed to execute query: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed", "details": err.Error()})
// 			return
// 		}

// 		for _, row := range emailData {
// 			// Log the email address being used
// 			log.Printf("Sending email to: %s", row.Email)

// 			// Validate the email address
// 			if row.Email == "" {
// 				log.Printf("Invalid email address for TrancodeID %s: empty email", emailReq.TicketNo)
// 				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address", "details": "Email address cannot be empty"})
// 				return
// 			}

// 			if err := SettingEmail(row.Subject, row.Body, row.Email, emailReq.TicketNo); err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email", "details": err.Error()})
// 				return
// 			}
// 		}
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Email processed successfully"})
// }
