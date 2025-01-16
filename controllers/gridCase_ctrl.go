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
	"gopkg.in/gomail.v2"
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
	// caseCache.Set(cacheKey, response, cache.DefaultExpiration)

	// Return the results in JSON format
	c.JSON(http.StatusOK, response)
}

func SaveCaseHandler(c *gin.Context) {
	// Retrieve user_name from session
	// session := sessions.Default(c)
	userName := "8023"
	// if userName == nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	// 	return
	// }

	var input models.Case
	// Bind JSON input to struct
	if err := c.BindJSON(&input); err != nil {
		log.Printf("Failed to bind input data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	// Helper function to send email
	sendCaseEmail := func(ticketNo, customerName, email, status string) error {
		sendEmailFlag := ""
		switch status {
		case "1", "2", "3", "4":
			sendEmailFlag = "SendTicket"
		case "5":
			sendEmailFlag = "SendTicketForExtend"
		}

		if sendEmailFlag != "" {
			data := fmt.Sprintf("%s|%s", customerName, ticketNo)
			sqlEmail := fmt.Sprintf(
				"set nocount on; exec sp_getsendEmail '%s', '%s', '%s', '%s', '%s', NULL;",
				ticketNo, email, data, sendEmailFlag, userName,
			)

			var emailData []models.EmailData
			if err := config.DB.Raw(sqlEmail).Scan(&emailData).Error; err != nil {
				return fmt.Errorf("failed to execute email query: %v", err)
			}

			for _, row := range emailData {
				if row.Email == "" {
					return fmt.Errorf("invalid email address: empty email")
				}

				if err := SettingEmail(row.Subject, row.Body, row.Email, ticketNo); err != nil {
					return fmt.Errorf("failed to send email: %v", err)
				}
			}
		}
		return nil
	}

	// Step 1: Check if the ticket exists in the `Case` table
	var existingTicket models.Case
	dtmupd := time.Now()
	if err := config.DB.Table("Case").Where("ticketno = ?", input.TicketNo).First(&existingTicket).Error; err == nil {
		// ============enable for debug only=====================================
		// Ticket exists, perform an update
		// existingTicket.TicketNo = input.TicketNo //it must disable (enable it for debug only)
		// existingTicket.FlagCompany = input.FlagCompany //it must disable (enable it for debug only)
		// existingTicket.BranchID = input.BranchID //it must disable (enable it for debug only)
		// existingTicket.AgreementNo = input.AgreementNo //it must disable (enable it for debug only)
		// existingTicket.ApplicationID = input.ApplicationID //it must disable (enable it for debug only)
		// existingTicket.CustomerID = input.CustomerID //it must disable (enable it for debug only)
		// existingTicket.IsSendEmail = input.IsSendEmail
		// existingTicket.Email = input.Email //it must disable (enable it for debug only)
		// existingTicket.CallerID = input.CallerID //it must disable (enable it for debug only)
		// existingTicket.StatusDescription = input.StatusDescription
		// existingTicket.ForAgingDays = input.ForAgingDays
		// ============enable for debug only=====================================
		existingTicket.CustomerName = input.CustomerName //currently enabled, use in debug
		existingTicket.PhoneNo = input.PhoneNo
		existingTicket.Email_ = input.Email_
		existingTicket.StatusID = input.StatusID
		existingTicket.TypeID = input.TypeID
		existingTicket.SubTypeID = input.SubTypeID
		existingTicket.PriorityID = input.PriorityID
		existingTicket.Description = input.Description
		existingTicket.UserID = input.UserID
		existingTicket.ContactID = input.ContactID
		existingTicket.RelationID = input.RelationID
		existingTicket.RelationName = input.RelationName
		existingTicket.DateUpd = dtmupd
		existingTicket.DateCr = input.DateCr

		// Update the existing case
		if err := config.DB.Table("Case").Omit("statusname", "IsSendEmail", "is_send_email", "flag").Save(&existingTicket).Error; err != nil {
			log.Printf("Failed to update case: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update case data"})
			return
		}

		// Insert/Update Case_History
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

		// Send email for updated case
		if err := sendCaseEmail(existingTicket.TicketNo, existingTicket.CustomerName, existingTicket.Email_, fmt.Sprintf("%d", existingTicket.StatusID)); err != nil {
			log.Printf("Email sending failed: %v", err)
			// Continue execution even if email fails
		}

		c.JSON(http.StatusOK, gin.H{"message": "Case updated successfully", "ticketNo": existingTicket.TicketNo})
		return
	}

	// Ticket doesn't exist, create a new one
	if len(input.AgreementNo) < 3 {
		log.Printf("Invalid AgreementNo: '%s'. Length is less than 3.", input.AgreementNo)
		c.JSON(http.StatusBadRequest, gin.H{"error": "AgreementNo must be at least 3 characters long"})
		return
	}

	// Extract the prefix safely
	agreementPrefix := input.AgreementNo[:3]
	csPrefix := "CS"
	currentDate := time.Now().Format("060102")
	currentDate_forAging := time.Now().Format("2006-01-02")

	// Get the last ticket number and increment
	var lastTicketNo string
	if err := config.DB.Raw("SELECT TOP 1 ticketno FROM [case] WHERE ticketno LIKE ? ORDER BY ticketno DESC", agreementPrefix+"%").Scan(&lastTicketNo).Error; err != nil {
		log.Printf("Failed to retrieve last ticket number: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve last ticket number"})
		return
	}

	if lastTicketNo == "" {
		lastTicketNo = agreementPrefix + csPrefix + currentDate + "000001"
	}

	ticketSuffix := lastTicketNo[len(agreementPrefix+csPrefix+currentDate):]
	lastIncrementedValue, err := strconv.Atoi(ticketSuffix)
	if err != nil {
		log.Printf("Failed to convert ticket suffix to integer: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert ticket suffix"})
		return
	}
	lastIncrementedValue++

	ticketNo := fmt.Sprintf("%s%s%s%06d", agreementPrefix, csPrefix, currentDate, lastIncrementedValue)

	// Insert into the `Case` table for new ticket
	insertCaseQuery := `INSERT INTO [case]
                (ticketno, flagcompany, branchid, agreementno, applicationid, customerid,
                customername, phoneno, email, statusid, typeid, subtypeid, priorityid,
                description, usrupd, contactid, relationid, relationname, callerid, email_, date_cr, foragingdays)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if err := config.DB.Exec(insertCaseQuery,
		ticketNo, input.FlagCompany, input.BranchID, input.AgreementNo, input.ApplicationID, input.CustomerID,
		input.CustomerName, input.PhoneNo, input.Email, input.StatusID, input.TypeID, input.SubTypeID, input.PriorityID,
		input.Description, userName, input.ContactID, input.RelationID, input.RelationName, input.CallerID, input.Email_, input.DateCr, currentDate_forAging).Error; err != nil {
		log.Printf("Failed to insert into case table: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert case data"})
		return
	}

	// Insert into `Case_History` table for new case
	nextID := 1
	if err := config.DB.Raw("SELECT ISNULL(MAX(id), 0) + 1 FROM Case_History").Scan(&nextID).Error; err != nil {
		log.Printf("Failed to calculate next Case_History ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate case history ID"})
		return
	}

	insertHistoryQuery := `INSERT INTO Case_History 
	(id, ticketno, description, statusid, usrupd, dtmupd) 
	VALUES (?, ?, ?, ?, ?, GETDATE())`
	if err := config.DB.Exec(insertHistoryQuery, nextID, ticketNo, input.Description, input.StatusID, userName).Error; err != nil {
		log.Printf("Failed to insert into Case_History: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert case history data"})
		return
	}

	// Log the operation
	logQuery := `INSERT INTO tbllog (tgl, table_name, menu, script) 
	VALUES (?, '[case]', 'SaveCaseHandler', ?)`
	if err := config.DB.Exec(logQuery, time.Now(), fmt.Sprintf("INSERT INTO [case] VALUES (%s, ...)", ticketNo)).Error; err != nil {
		log.Printf("Failed to log operation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log operation"})
		return
	}

	// Send email for new case
	if err := sendCaseEmail(ticketNo, input.CustomerName, input.Email_, fmt.Sprintf("%d", input.StatusID)); err != nil {
		log.Printf("Email sending failed: %v", err)
		// Continue execution even if email fails
	}

	// Success Response
	c.JSON(http.StatusOK, gin.H{"message": "Data inserted successfully", "ticketNo": ticketNo})
}

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
