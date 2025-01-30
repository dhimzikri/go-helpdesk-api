package controllers

import (
	"fmt"
	"golang-sqlserver-app/config"
	"golang-sqlserver-app/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func GetHistory(c *gin.Context) {
	query := c.Query("query")
	col := c.Query("col")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// Handle pagination
	start := 0
	if c.Query("start") != "" {
		start, _ = strconv.Atoi(c.Query("start"))
	}
	limit, _ := strconv.Atoi(c.Query("limit"))

	// Get user from session
	userid := c.GetString("user_name") // Assuming you have middleware setting this

	var src string
	if userid == "admin" || readHeadCS(userid) { // You'll need to implement readHeadCS function
		src = fmt.Sprintf("CONVERT(DATE, a.dtmupd) >= '%s' and CONVERT(DATE, a.dtmupd) <= '%s'", startDate, endDate)
	} else {
		src = fmt.Sprintf("0=0 and a.dtmupd >= '%s' and a.dtmupd < dateadd(day, 1, '%s') and a.usrupd = '%s'", startDate, endDate, userid)
	}

	if query != "" && col != "" {
		src = src + " AND " + col + " LIKE '%" + query + "%'"
	}
	// Generate a unique cache key based on query parameters
	cacheKey := fmt.Sprintf("cases_page_%d_limit_%d_conditions_%s", start, limit, src)

	// Check if the data is already in the cache
	if cachedData, found := caseCache.Get(cacheKey); found {
		// Return cached data
		c.JSON(http.StatusOK, cachedData)
		return
	}
	// email_ := "a.email_"

	// Build SQL query for fetching paginated data
	sqlQuery := fmt.Sprintf(`SET NOCOUNT ON;

			DECLARE @jml AS int;

			SELECT @jml = COUNT(a.ticketno)
			FROM (
				SELECT a.ticketno
				FROM [Case] a
				INNER JOIN tbltype b ON a.TypeID = b.TypeID
				INNER JOIN (
					SELECT a.*, 
						CASE 
							WHEN b.isactive = 1 AND b.isactive_alternate = 0 THEN b.employeeid
							WHEN b.isactive = 0 AND b.isactive_alternate = 1 THEN b.employeeid_alternate
							ELSE b.employeeid
						END AS employeeid
					FROM tblSubtype a
					LEFT JOIN tblassignment b ON a.cost_center = b.costcenterid
				) c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
				INNER JOIN [Priority] d ON a.PriorityID = d.PriorityID
				INNER JOIN [status] e ON a.statusid = e.statusid
				INNER JOIN [contact] f ON a.contactid = f.contactid
				INNER JOIN [relation] g ON a.relationid = g.relationid
				WHERE %s
				
				UNION ALL
				
				SELECT a.ticketno
				FROM [Case_NewCustomer] a
				INNER JOIN tbltype b ON a.TypeID = b.TypeID
				INNER JOIN (
					SELECT a.*, 
						CASE 
							WHEN b.isactive = 1 AND b.isactive_alternate = 0 THEN b.employeeid
							WHEN b.isactive = 0 AND b.isactive_alternate = 1 THEN b.employeeid_alternate
							ELSE b.employeeid
						END AS employeeid
					FROM tblSubtype a
					LEFT JOIN tblassignment b ON a.cost_center = b.costcenterid
				) c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
				INNER JOIN [Priority] d ON a.PriorityID = d.PriorityID
				INNER JOIN [status] e ON a.statusid = e.statusid
				INNER JOIN [contact] f ON a.contactid = f.contactid
				INNER JOIN [relation] g ON a.relationid = g.relationid
				WHERE %s
			) AS a;

			SELECT *
			FROM (
				SELECT ROW_NUMBER() OVER (ORDER BY RIGHT(a.ticketno, 3) desc) AS 'RowNumber', *
				FROM (
					SELECT 
						a.flagcompany, a.ticketno, a.agreementno, a.applicationid, a.customerid, 
						a.typeid, b.description AS typedescriontion,
						a.subtypeid, c.SubDescription AS typesubdescriontion, a.priorityid, 
						d.Description AS prioritydescription, 
						a.statusid, e.statusname, e.description AS statusdescription, 
						a.customername, a.branchid, a.description, a.phoneno, a.email, 
						u.real_name AS usrupd,
						@jml AS jml, f.contactid, f.Description AS contactdescription, 
						c.employeeid, a.relationid, g.description AS relationdescription, 
						a.relationname, dbo.FnGetFullBranchName(a.branchid) AS cabang,
						'CS HO' AS channel, CONVERT(varchar, a.foragingdays, 106) AS foragingdays,
						FORMAT(dbo.getTanggalPenyelesaian(a.ticketno), 'dd MMM yyyy') AS tanggalpenyelesaian,
						dbo.getWaktuPenyelesaian(a.ticketno) AS waktupenyelesaian,
						CASE 
							WHEN xx.statusid = 5 
							THEN FORMAT(xx.dtmupd, 'dd MMM yyyy') 
							ELSE '' 
						END AS tgl_ext,
						FORMAT(CONVERT(DATE, a.date_cr), 'dd MMM yyyy') AS date_cr
					FROM [Case] a
					INNER JOIN tbltype b ON a.TypeID = b.TypeID
					INNER JOIN (
						SELECT a.*, 
							CASE 
								WHEN b.isactive = 1 AND b.isactive_alternate = 0 THEN b.employeeid
								WHEN b.isactive = 0 AND b.isactive_alternate = 1 THEN b.employeeid_alternate
								ELSE b.employeeid
							END AS employeeid
						FROM tblSubtype a
						LEFT JOIN tblassignment b ON a.cost_center = b.costcenterid
					) c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
					INNER JOIN [Priority] d ON a.PriorityID = d.PriorityID
					INNER JOIN [status] e ON a.statusid = e.statusid
					INNER JOIN [contact] f ON a.contactid = f.contactid
					INNER JOIN [relation] g ON a.relationid = g.relationid
					LEFT JOIN [Portal_EXT].[dbo].[users] u ON a.usrupd = u.user_name
					LEFT JOIN (
						SELECT a.statusid, a.ticketno, a.dtmupd, b.StatusName
						FROM Case_History a
						LEFT JOIN Status b ON a.statusid = b.StatusID
						WHERE b.StatusID = 5
					) xx ON a.ticketno = xx.ticketno 
					WHERE %s
					
					UNION
					SELECT 
						a.flagcompany, a.ticketno, a.agreementno, a.applicationid, a.customerid, 
						a.typeid, b.description AS typedescriontion,
						a.subtypeid, c.SubDescription AS typesubdescriontion, a.priorityid, 
						d.Description AS prioritydescription, 
						a.statusid, e.statusname, e.description AS statusdescription, 
						a.customername, a.branchid, a.description, a.phoneno, a.email, 
						u.real_name AS usrupd,
						@jml AS jml, f.contactid, f.Description AS contactdescription, 
						c.employeeid, a.relationid, g.description AS relationdescription, 
						a.relationname, dbo.FnGetFullBranchName(a.branchid) AS cabang,
						'CS HO' AS channel, CONVERT(varchar, a.foragingdays, 106) AS foragingdays,
						FORMAT(dbo.getTanggalPenyelesaian(a.ticketno), 'dd MMM yyyy') AS tanggalpenyelesaian,
						dbo.getWaktuPenyelesaian(a.ticketno) AS waktupenyelesaian,
						CASE 
							WHEN xx.StatusName = 'Extend' 
							THEN FORMAT(xx.dtmupd, 'dd MMM yyyy') 
							ELSE '' 
						END AS tgl_ext,
						FORMAT(CONVERT(DATE, a.date_cr), 'dd MMM yyyy') AS date_cr
					FROM [Case_NewCustomer] a
					INNER JOIN tbltype b ON a.TypeID = b.TypeID
					INNER JOIN (
						SELECT a.*, 
							CASE 
								WHEN b.isactive = 1 AND b.isactive_alternate = 0 THEN b.employeeid
								WHEN b.isactive = 0 AND b.isactive_alternate = 1 THEN b.employeeid_alternate
								ELSE b.employeeid
							END AS employeeid
						FROM tblSubtype a
						LEFT JOIN tblassignment b ON a.cost_center = b.costcenterid
					) c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
					INNER JOIN [Priority] d ON a.PriorityID = d.PriorityID
					INNER JOIN [status] e ON a.statusid = e.statusid
					INNER JOIN [contact] f ON a.contactid = f.contactid
					INNER JOIN [relation] g ON a.relationid = g.relationid
					LEFT JOIN [Portal_EXT].[dbo].[users] u ON a.usrupd = u.user_name
					LEFT JOIN (
						SELECT a.statusid, a.ticketno, b.StatusName, a.dtmupd
						FROM Case_History a
						LEFT JOIN Status b ON a.statusid = b.StatusID
						WHERE b.StatusName = 'Extend'
					) xx ON a.ticketno = xx.ticketno
					WHERE %s 
				) AS a
			) AS a
			where RowNumber>%d and RowNumber<=%d order by a.foragingdays desc
	`, src, src, src, src, start, limit)

	var cases []map[string]interface{}
	if err := config.DB.Raw(sqlQuery).Scan(&cases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch case data"})
		return
	}

	// Prepare the response
	response := gin.H{
		// "total": totalCount,
		"cases": cases,
	}

	// Store the response in cache
	// caseCache.Set(cacheKey, response, cache.DefaultExpiration)

	// Return the results in JSON format
	c.JSON(http.StatusOK, response)
}

func readHeadCS(username string) bool {
	row := config.DB.Raw("EXEC sp_getHeadCS @username = ?", username).Row()
	if row.Err() != nil {
		// You might want to log the error here
		return false
	}

	var count int
	if err := row.Scan(&count); err != nil {
		// You might want to log the error here
		return false
	}

	return count > 0
}

// Fetch case data
func ForExport(src string, start, limit int) ([]models.CaseData, error) {
	var cases []models.CaseData
	sqlQuery := fmt.Sprintf(`SET NOCOUNT ON;

	DECLARE @jml AS int;

	SELECT @jml = COUNT(a.TicketNo)
	FROM (
		SELECT a.TicketNo
		FROM [Case] a
		INNER JOIN tbltype b ON a.TypeID = b.TypeID
		INNER JOIN (
			SELECT a.*, 
				CASE 
					WHEN b.isactive = 1 AND b.isactive_alternate = 0 THEN b.EmployeeID
					WHEN b.isactive = 0 AND b.isactive_alternate = 1 THEN b.EmployeeID_alternate
					ELSE b.EmployeeID
				END AS EmployeeID
			FROM tblSubtype a
			LEFT JOIN tblassignment b ON a.cost_center = b.costcenterid
		) c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
		INNER JOIN [Priority] d ON a.PriorityID = d.PriorityID
		INNER JOIN [status] e ON a.StatusID = e.StatusID
		INNER JOIN [contact] f ON a.ContactID = f.ContactID
		INNER JOIN [relation] g ON a.RelationID = g.RelationID
		WHERE %s
		
		UNION ALL
		
		SELECT a.TicketNo
		FROM [Case_NewCustomer] a
		INNER JOIN tbltype b ON a.TypeID = b.TypeID
		INNER JOIN (
			SELECT a.*, 
				CASE 
					WHEN b.isactive = 1 AND b.isactive_alternate = 0 THEN b.EmployeeID
					WHEN b.isactive = 0 AND b.isactive_alternate = 1 THEN b.EmployeeID_alternate
					ELSE b.EmployeeID
				END AS EmployeeID
			FROM tblSubtype a
			LEFT JOIN tblassignment b ON a.cost_center = b.costcenterid
		) c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
		INNER JOIN [Priority] d ON a.PriorityID = d.PriorityID
		INNER JOIN [status] e ON a.StatusID = e.StatusID
		INNER JOIN [contact] f ON a.ContactID = f.ContactID
		INNER JOIN [relation] g ON a.RelationID = g.RelationID
		WHERE %s
	) AS a;

	SELECT *
	FROM (
		SELECT ROW_NUMBER() OVER (ORDER BY RIGHT(a.TicketNo, 3) desc) AS 'RowNumber', *
		FROM (
			SELECT 
				a.FlagCompany, a.TicketNo, a.AgreementNo, a.ApplicationID, a.CustomerID, 
				a.TypeID, b.description AS TypeDescription,
				a.SubTypeID, c.SubDescription AS SubTypeDescription, a.PriorityID, 
				d.Description AS PriorityDescription, 
				a.StatusID, e.StatusName, e.description AS StatusDescription, 
				a.CustomerName, a.BranchID, a.description, a.PhoneNo, a.Email, 
				u.real_name AS UsrUpd,
				@jml AS jml, f.ContactID, f.Description AS ContactDescription, 
				c.EmployeeID, a.RelationID, g.description AS RelationDescription, 
				a.RelationName, dbo.FnGetFullBranchName(a.BranchID) AS cabang,
				'CS HO' AS channel, CONVERT(varchar, a.ForAgingDays, 106) AS ForAgingDays,
				FORMAT(dbo.getTanggalPenyelesaian(a.TicketNo), 'dd MMM yyyy') AS TanggalPenyelesaian,
				dbo.getWaktuPenyelesaian(a.TicketNo) AS WaktuPenyelesaian,
				CASE 
					WHEN xx.StatusID = 5 
					THEN FORMAT(xx.dtmupd, 'dd MMM yyyy') 
					ELSE '' 
				END AS tgl_ext,
				FORMAT(CONVERT(DATE, a.date_cr), 'dd MMM yyyy') AS date_cr
			FROM [Case] a
			INNER JOIN tbltype b ON a.TypeID = b.TypeID
			INNER JOIN (
				SELECT a.*, 
					CASE 
						WHEN b.isactive = 1 AND b.isactive_alternate = 0 THEN b.EmployeeID
						WHEN b.isactive = 0 AND b.isactive_alternate = 1 THEN b.EmployeeID_alternate
						ELSE b.EmployeeID
					END AS EmployeeID
				FROM tblSubtype a
				LEFT JOIN tblassignment b ON a.cost_center = b.costcenterid
			) c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
			INNER JOIN [Priority] d ON a.PriorityID = d.PriorityID
			INNER JOIN [status] e ON a.StatusID = e.StatusID
			INNER JOIN [contact] f ON a.ContactID = f.ContactID
			INNER JOIN [relation] g ON a.RelationID = g.RelationID
			LEFT JOIN [Portal_EXT].[dbo].[users] u ON a.UsrUpd = u.user_name
			LEFT JOIN (
				SELECT a.StatusID, a.TicketNo, a.dtmupd, b.StatusName
				FROM Case_History a
				LEFT JOIN Status b ON a.StatusID = b.StatusID
				WHERE b.StatusID = 5
			) xx ON a.TicketNo = xx.ticketno 
			WHERE %s
			
			UNION
			SELECT 
				a.FlagCompany, a.TicketNo, a.AgreementNo, a.ApplicationID, a.CustomerID, 
				a.TypeID, b.description AS TypeDescription,
				a.SubTypeID, c.SubDescription AS SubTypeDescription, a.PriorityID, 
				d.Description AS PriorityDescription, 
				a.StatusID, e.StatusName, e.description AS StatusDescription, 
				a.CustomerName, a.BranchID, a.description, a.PhoneNo, a.Email, 
				u.real_name AS UsrUpd,
				@jml AS jml, f.ContactID, f.Description AS ContactDescription, 
				c.EmployeeID, a.RelationID, g.description AS RelationDescription, 
				a.RelationName, dbo.FnGetFullBranchName(a.BranchID) AS cabang,
				'CS HO' AS channel, CONVERT(varchar, a.ForAgingDays, 106) AS ForAgingDays,
				FORMAT(dbo.getTanggalPenyelesaian(a.TicketNo), 'dd MMM yyyy') AS TanggalPenyelesaian,
				dbo.getWaktuPenyelesaian(a.TicketNo) AS WaktuPenyelesaian,
				CASE 
					WHEN xx.StatusName = 'Extend' 
					THEN FORMAT(xx.dtmupd, 'dd MMM yyyy') 
					ELSE '' 
				END AS tgl_ext,
				FORMAT(CONVERT(DATE, a.date_cr), 'dd MMM yyyy') AS date_cr
			FROM [Case_NewCustomer] a
			INNER JOIN tbltype b ON a.TypeID = b.TypeID
			INNER JOIN (
				SELECT a.*, 
					CASE 
						WHEN b.isactive = 1 AND b.isactive_alternate = 0 THEN b.EmployeeID
						WHEN b.isactive = 0 AND b.isactive_alternate = 1 THEN b.EmployeeID_alternate
						ELSE b.EmployeeID
					END AS EmployeeID
				FROM tblSubtype a
				LEFT JOIN tblassignment b ON a.cost_center = b.costcenterid
			) c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
			INNER JOIN [Priority] d ON a.PriorityID = d.PriorityID
			INNER JOIN [status] e ON a.StatusID = e.StatusID
			INNER JOIN [contact] f ON a.ContactID = f.ContactID
			INNER JOIN [relation] g ON a.RelationID = g.RelationID
			LEFT JOIN [Portal_EXT].[dbo].[users] u ON a.UsrUpd = u.user_name
			LEFT JOIN (
				SELECT a.StatusID, a.TicketNo, b.StatusName, a.dtmupd
				FROM Case_History a
				LEFT JOIN Status b ON a.StatusID = b.StatusID
				WHERE b.StatusName = 'Extend'
			) xx ON a.TicketNo = xx.ticketno
			WHERE %s 
		) AS a
	) AS a
	where RowNumber>%d and RowNumber<=%d order by a.ForAgingDays desc
`, src, src, src, src, start, limit)

	result := config.DB.Raw(sqlQuery).Scan(&cases)
	if result.Error != nil {
		return nil, result.Error
	}

	return cases, nil
}

// API endpoint to return JSON data
func GetJSON(c *gin.Context) {
	// Log the start of the function
	log.Println("ExportXLS function started")

	// Define the SQL query (replace placeholders with actual values)
	query := c.Query("query")
	col := c.Query("col")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// Log query parameters
	log.Printf("Query Parameters - query: %s, col: %s, startDate: %s, endDate: %s", query, col, startDate, endDate)

	// Handle pagination
	start := 0
	if c.Query("start") != "" {
		start, _ = strconv.Atoi(c.Query("start"))
	}
	limit, _ := strconv.Atoi(c.Query("limit"))

	// Log pagination parameters
	log.Printf("Pagination Parameters - start: %d, limit: %d", start, limit)

	// Get user from session
	userid := "8023" // Assuming you have middleware setting this
	log.Printf("User ID: %s", userid)

	var src string
	if userid == "admin" || readHeadCS(userid) { // You'll need to implement readHeadCS function
		src = fmt.Sprintf("CONVERT(DATE, a.dtmupd) >= '%s' and CONVERT(DATE, a.dtmupd) <= '%s'", startDate, endDate)
	} else {
		src = fmt.Sprintf("0=0 and a.dtmupd >= '%s' and a.dtmupd < dateadd(day, 1, '%s')", startDate, endDate)
	}

	if query != "" && col != "" {
		src = src + " AND " + col + " LIKE '%" + query + "%'"
	}

	// Log the generated SQL condition
	log.Printf("Generated SQL Condition: %s", src)

	// Generate a unique cache key based on query parameters
	cacheKey := fmt.Sprintf("cases_page_%d_limit_%d_conditions_%s", start, limit, src)

	// Log the cache key
	log.Printf("Cache Key: %s", cacheKey)

	// Check if the data is already in the cache
	if cachedData, found := caseCache.Get(cacheKey); found {
		// Log cache hit
		log.Println("Cache hit - returning cached data")
		c.JSON(http.StatusOK, cachedData)
		return
	} else {
		// Log cache miss
		log.Println("Cache miss - querying database")
	}
	cases, err := ForExport(src, start, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cases)
}

func GetXLSX(c *gin.Context) {
	// Log the start of the function
	log.Println("ExportXLS function started")

	// Define the SQL query (replace placeholders with actual values)
	query := c.Query("query")
	col := c.Query("col")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// Log query parameters
	log.Printf("Query Parameters - query: %s, col: %s, startDate: %s, endDate: %s", query, col, startDate, endDate)

	// Handle pagination
	start := 0
	if c.Query("start") != "" {
		start, _ = strconv.Atoi(c.Query("start"))
	}
	limit, _ := strconv.Atoi(c.Query("limit"))

	// Log pagination parameters
	log.Printf("Pagination Parameters - start: %d, limit: %d", start, limit)

	// Get user from session
	userid := "8023" // Assuming you have middleware setting this
	log.Printf("User ID: %s", userid)

	var src string
	if userid == "admin" || readHeadCS(userid) { // You'll need to implement readHeadCS function
		src = fmt.Sprintf("CONVERT(DATE, a.dtmupd) >= '%s' and CONVERT(DATE, a.dtmupd) <= '%s'", startDate, endDate)
	} else {
		src = fmt.Sprintf("0=0 and a.dtmupd >= '%s' and a.dtmupd < dateadd(day, 1, '%s')", startDate, endDate)
	}

	if query != "" && col != "" {
		src = src + " AND " + col + " LIKE '%" + query + "%'"
	}

	// Log the generated SQL condition
	log.Printf("Generated SQL Condition: %s", src)

	// Generate a unique cache key based on query parameters
	cacheKey := fmt.Sprintf("cases_page_%d_limit_%d_conditions_%s", start, limit, src)

	// Log the cache key
	log.Printf("Cache Key: %s", cacheKey)

	// Check if the data is already in the cache
	if cachedData, found := caseCache.Get(cacheKey); found {
		// Log cache hit
		log.Println("Cache hit - returning cached data")
		c.JSON(http.StatusOK, cachedData)
		return
	} else {
		// Log cache miss
		log.Println("Cache miss - querying database")
	}
	cases, err := ForExport(src, start, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f := excelize.NewFile()
	sheetName := "Cases"
	f.SetSheetName("Sheet1", sheetName)

	headers := []string{"FlagCompany", "TicketNo", "AgreementNo", "ApplicationID", "CustomerID", "TypeID", "TypeDescription",
		"SubTypeID", "SubTypeDescription", "PriorityID", "PriorityDescription", "StatusID", "StatusName", "StatusDescription",
		"CustomerName", "BranchID", "Description", "PhoneNo", "Email", "UsrUpd", "ContactID", "ContactDescription",
		"EmployeeID", "RelationID", "RelationDescription", "RelationName", "Cabang", "Channel", "ForAgingDays",
		"TanggalPenyelesaian", "WaktuPenyelesaian", "TglExt", "DateCr"}

	// Set headers in the first row
	for colIndex, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIndex+1, 1) // A1, B1, C1...
		f.SetCellValue(sheetName, cell, header)
	}

	// Write case data to the sheet
	for rowIndex, caseData := range cases {
		row := rowIndex + 2 // Data starts from the second row

		values := []interface{}{
			caseData.FlagCompany, caseData.TicketNo, caseData.AgreementNo, caseData.ApplicationID, caseData.CustomerID,
			caseData.TypeID, caseData.TypeDescription, caseData.SubTypeID, caseData.SubTypeDescription,
			caseData.PriorityID, caseData.PriorityDescription, caseData.StatusID, caseData.StatusName, caseData.StatusDescription,
			caseData.CustomerName, caseData.BranchID, caseData.Description, caseData.PhoneNo, caseData.Email, caseData.UsrUpd,
			caseData.ContactID, caseData.ContactDescription, caseData.EmployeeID, caseData.RelationID,
			caseData.RelationDescription, caseData.RelationName, caseData.Cabang, caseData.Channel, caseData.ForAgingDays,
			caseData.TanggalPenyelesaian, caseData.WaktuPenyelesaian, caseData.TglExt, caseData.DateCr,
		}

		for colIndex, value := range values {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, row)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	// Prepare file for download
	filename := fmt.Sprintf("cases_%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	_ = f.Write(c.Writer)
}
