package controllers

import (
	"fmt"
	"golang-sqlserver-app/config"
	"golang-sqlserver-app/models"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetCases retrieves cases with pagination, search, and total count
// var caseCache = cache.New(5*time.Minute, 10*time.Minute)

func GetCase(c *gin.Context) {
	// Skip session logic - No need to retrieve user_name from session

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

	// Add condition for statusid (you may choose to remove or modify this based on your requirements)
	conditions = append(conditions, "a.statusid <> 1")

	// Combine conditions with "AND"
	whereClause := strings.Join(conditions, " AND ")

	// Generate a unique cache key based on query parameters
	// cacheKey := fmt.Sprintf("cases_page_%d_limit_%d_conditions_%s", pageNum, limitNum, whereClause)

	// Check if the data is already in the cache
	// if cachedData, found := caseCache.Get(cacheKey); found {
	// Return cached data
	// 	c.JSON(http.StatusOK, cachedData)
	// 	return
	// }

	// Build SQL query for counting total records
	sqlCount := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM [Case] a
		INNER JOIN tbltype b ON a.TypeID = b.TypeID
		LEFT JOIN tblSubtype c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
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
			a.typeid, b.description AS typedescription, a.subtypeid, c.SubDescription AS typesubdescription,
			a.priorityid, d.Description AS prioritydescription, a.statusid, e.statusname,
			e.description AS statusdescription, a.customername, a.branchid, a.description, a.phoneno,
			a.email, a.usrupd, a.dtmupd, a.date_cr, f.contactid, f.Description AS contactdescription,
			a.relationid, g.description AS relationdescription, a.relationname, a.callerid, a.email_, a.foragingdays
		FROM [Case] a
		INNER JOIN tbltype b ON a.TypeID = b.TypeID
		LEFT JOIN tblSubtype c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
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
	// caseCache.Set(cacheKey, response, cache.DefaultExpiration)

	// Return the results in JSON format
	c.JSON(http.StatusOK, response)
}

func SaveCase(c *gin.Context) {
	var request models.CaseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid request format", "error": err.Error()})
		return
	}

	// Convert status to string for comparison
	statusStr := fmt.Sprintf("%d", request.StatusID)

	sql := `
    BEGIN TRY  
        SET NOCOUNT ON;
        DECLARE @trancodeid VARCHAR(20);
        
        EXEC sp_insertcase ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, @trancodeid OUTPUT;
        
        -- Log the operation
        INSERT INTO tbllog ([tgl], [table_name], [menu], [script])
        SELECT GETDATE(), '[case]', 'gridcase', 'Executed insert case with status: ' + ?;
        
        SELECT @trancodeid AS trancodeid;
    END TRY
    BEGIN CATCH
        SELECT ERROR_NUMBER() AS ErrorNumber, 
               ERROR_MESSAGE() AS ErrorMessage;
    END CATCH;`

	params := []interface{}{
		nullIfEmpty(request.TicketNo),
		request.FlagCompany,
		request.BranchID,
		strings.TrimSpace(request.AgreementNo),
		strings.TrimSpace(request.ApplicationID),
		strings.TrimSpace(request.CustomerID),
		request.CustomerName,
		request.PhoneNo,
		request.Email,
		statusStr,
		request.TypeID,
		request.SubTypeID,
		request.PriorityID,
		request.Description,
		request.UserID,
		request.ContactID,
		request.RelationID,
		nullIfEmpty(request.RelationName),
		request.CallerID,
		request.Email_,
		request.DateCr,
		"0",       // forAgingDays
		statusStr, // for the log message
	}

	// Execute query and scan result
	rows, err := config.DB.Raw(sql, params...).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"msg":     "Failed to execute query",
			"error":   err.Error(),
		})
		return
	}
	defer rows.Close()

	var trancodeid string
	if rows.Next() {
		if err := rows.Scan(&trancodeid); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"msg":     "Failed to scan result",
				"error":   err.Error(),
			})
			return
		}
	}

	// Handle email sending if required
	if request.IsSendEmail == "true" {
		emailErr := SendCaseEmail(config.DB, trancodeid, request)
		if emailErr != nil {
			c.JSON(http.StatusOK, gin.H{
				"success":    true,
				"msg":        "Case saved but email sending failed",
				"trancodeid": trancodeid,
				"emailError": emailErr.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"msg":        "Case saved successfully",
		"trancodeid": trancodeid,
	})
}

func SendCaseEmail(db *gorm.DB, trancodeid string, request models.CaseRequest) error {
	var sendEmailFlag string
	switch request.StatusID {
	case 1, 2, 3, 4:
		sendEmailFlag = "SendTicket"
	case 5:
		sendEmailFlag = "SendTicketForExtend"
	default:
		return nil
	}

	ticketno := request.TicketNo
	if strings.TrimSpace(ticketno) == "" {
		ticketno = trancodeid
	}

	data := fmt.Sprintf("%s|%s", request.CustomerName, ticketno)

	emailSQL := `
    SET NOCOUNT ON;
    EXEC sp_getsendEmail 
        @trancodeid = ?,
        @ticketno = ?,
        @emailto = ?,
        @data = ?,
        @flag = ?,
        @usrupd = ?
    `

	// Log the email query parameters for debugging
	log.Printf("Executing email query with params: trancodeid=%s, ticketno=%s, emailto=%s, data=%s, flag=%s, usrupd=%s",
		trancodeid, ticketno, request.Email_, data, sendEmailFlag, request.UserID)

	rows, err := db.Raw(emailSQL,
		trancodeid,
		ticketno,
		request.Email_,
		data,
		sendEmailFlag,
		request.UserID,
	).Rows()

	if err != nil {
		return fmt.Errorf("failed to execute email query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %v", err)
		}

		// Create holders for the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the holders
		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}

		// Create a map of column name to value
		result := make(map[string]string)
		for i, col := range columns {
			val := values[i]
			if val != nil {
				result[strings.ToLower(col)] = fmt.Sprintf("%v", val)
			}
		}

		// Log the retrieved values for debugging
		log.Printf("Retrieved email data: %+v", result)

		// Extract required fields
		subject := result["subject"]
		body := result["body"]
		recipient := result["emailto"] // Adjusted to match SP column name

		// Validate email data
		if subject == "" || body == "" || recipient == "" {
			log.Printf("Warning: Missing email data - subject: %v, body length: %v, recipient: %v",
				subject, len(body), recipient)
			continue
		}

		// Send the email
		if err := SettingEmail(subject, body, recipient, trancodeid); err != nil {
			log.Printf("Failed to send email: %v", err)
			return fmt.Errorf("failed to send email: %v", err)
		}
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %v", err)
	}

	return nil
}
