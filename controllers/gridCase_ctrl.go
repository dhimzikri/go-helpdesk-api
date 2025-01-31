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
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

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
	cacheKey := fmt.Sprintf("cases_page_%d_limit_%d_conditions_%s", pageNum, limitNum, whereClause)

	// Check if the data is already in the cache
	if cachedData, found := caseCache.Get(cacheKey); found {
		// Return cached data
		c.JSON(http.StatusOK, cachedData)
		return
	}
	// email_ := "a.email_"
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
			a.typeid, b.description AS typedescription, a.subtypeid, c.SubDescription AS typesubdescription,
			a.priorityid, d.Description AS prioritydescription, a.statusid, e.statusname,
			e.description AS statusdescription, a.customername, a.branchid, a.description, a.phoneno,
			a.email, a.usrupd, a.dtmupd, a.date_cr, f.contactid, f.Description AS contactdescription,
			a.relationid, g.description AS relationdescription, a.relationname, a.callerid, a.email_ , a.foragingdays
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
		request.TicketNo,
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
		request.RelationName,
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

	if request.IsSendEmail == "true" {
		go func() {
			emailErr := sendCaseEmail(config.DB, trancodeid, request)
			if emailErr != nil {
				c.JSON(http.StatusOK, gin.H{
					"success":    true,
					"msg":        "Case saved but email sending failed",
					"trancodeid": trancodeid,
					"emailError": emailErr.Error(),
				})
				return
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"msg":        "Case saved successfully",
		"trancodeid": trancodeid,
	})
}

func CloseCase(c *gin.Context) {
	var request models.CaseRequest

	// Handle potential binding errors
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid request format", "error": err.Error()})
		return
	}

	// Set default StatusID if not provided
	if request.StatusID == 0 {
		request.StatusID = 1
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
        END CATCH;
    `

	params := []interface{}{
		request.TicketNo,
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
		request.RelationName,
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

	// Send email asynchronously (if required)
	if request.IsSendEmail == "true" {
		go func() {
			emailErr := sendCaseEmail(config.DB, trancodeid, request)
			if emailErr != nil {
				// Log the email sending error (optional)
				fmt.Println("Error sending email:", emailErr.Error())
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"msg":        "Case Close successfully",
		"trancodeid": trancodeid,
	})
}

func sendCaseEmail(db *gorm.DB, trancodeid string, request models.CaseRequest) error {
	var sendEmailFlag string
	switch request.StatusID {
	case 1, 2, 3, 4:
		sendEmailFlag = "SendTicket"
	case 5:
		sendEmailFlag = "SendTicketForExtend"
	default:
		return nil
	}

	// Set ticketno same as trancodeid if empty
	ticketno := request.TicketNo
	if strings.TrimSpace(ticketno) == "" {
		ticketno = trancodeid
	}

	data := fmt.Sprintf("%s|%s", request.CustomerName, ticketno)

	// Use both trancodeid and ticketno in parameters
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
	rows, err := db.Raw(emailSQL,
		trancodeid,     // @trancodeid
		ticketno,       // @ticketno
		request.Email_, // @email
		data,           // @data
		sendEmailFlag,  // @flag
		request.UserID, // @userid
	).Rows()

	if err != nil {
		return fmt.Errorf("failed to execute email query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var subject, body, recipient, trancodeid string
		if err := rows.Scan(&subject, &body, &recipient, &trancodeid); err != nil {
			return fmt.Errorf("failed to scan email details: %v", err)
		}

		if err := SendEmail(subject, body, recipient, trancodeid); err != nil {
			return fmt.Errorf("failed to send email: %v", err)
		}
	}

	return nil
}

func SendEmail(subject, bodyEmail, recipient, trancodeid string) error {
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
