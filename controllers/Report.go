package controllers

import (
	"database/sql"
	"fmt"
	"golang-sqlserver-app/config"
	"log"
	"net/http"
	"strconv"

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
	userid := "8023" // Assuming you have middleware setting this

	var src string
	if userid == "admin" || readHeadCS(userid) { // You'll need to implement readHeadCS function
		src = fmt.Sprintf("CONVERT(DATE, a.dtmupd) >= '%s' and CONVERT(DATE, a.dtmupd) <= '%s'", startDate, endDate)
	} else {
		src = fmt.Sprintf("0=0 and a.dtmupd >= '%s' and a.dtmupd < dateadd(day, 1, '%s')", startDate, endDate)
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

func ExportXLS(c *gin.Context) {
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

	// Log the generated SQL query
	log.Printf("Generated SQL Query: %s", sqlQuery)

	var cases []map[string]interface{}
	if err := config.DB.Raw(sqlQuery).Scan(&cases).Error; err != nil {
		// Log database query error
		log.Printf("Failed to fetch case data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch case data"})
		return
	}

	// Log the number of cases fetched
	log.Printf("Number of cases fetched: %d", len(cases))

	// Execute the query
	rows, err := config.DB.Raw(sqlQuery).Rows()
	if err != nil {
		// Log rows error
		log.Printf("Failed to execute query: %v", err)
		return
	}
	defer rows.Close()

	// Log the start of Excel file creation
	log.Println("Creating Excel file")

	// Create a new Excel file
	f := excelize.NewFile()
	sheetName := "Sheet1"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		// Log Excel sheet creation error
		log.Printf("Failed to create new sheet: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new sheet: " + err.Error()})
		return
	}
	f.SetActiveSheet(index)

	// Write headers to the Excel file
	headers := []string{"FlagCompany", "TicketNo", "AgreementNo", "ApplicationID", "CustomerID", "TypeID", "TypeDescription", "SubTypeID", "SubTypeDescription", "PriorityID", "PriorityDescription", "StatusID", "StatusName", "StatusDescription", "CustomerName", "BranchID", "Description", "PhoneNo", "Email", "UsrUpd", "Jml", "ContactID", "ContactDescription", "EmployeeID", "RelationID", "RelationDescription", "RelationName", "Cabang", "Channel", "ForagingDays", "TanggalPenyelesaian", "WaktuPenyelesaian", "TglExt", "DateCR"}
	// headers := []string{"FlagCompany", "TicketNo", "AgreementNo", "ApplicationID", "CustomerID", "TypeID", "TypeDescription", "SubTypeID", "SubTypeDescription", "PriorityID", "PriorityDescription", "StatusID", "StatusName", "StatusDescription", "CustomerName", "BranchID", "Description", "PhoneNo", "Email", "UsrUpd", "Jml", "ContactID", "ContactDescription", "EmployeeID", "RelationID", "RelationDescription", "RelationName", "Cabang", "Channel", "ForagingDays", "TanggalPenyelesaian", "WaktuPenyelesaian", "TglExt", "DateCR"}
	for col, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+col)
		f.SetCellValue(sheetName, cell, header)
	}

	// Log the headers written to Excel
	log.Printf("Headers written to Excel: %v", headers)

	// Write data to the Excel file
	rowIndex := 2
	for rows.Next() {
		var (
			RowNumber, flagCompany, ticketNo, agreementNo, applicationID, customerID, typeID, customerName, statusName, branchID, description, phoneNo,
			email, jml, contactID, relationID, relationName, channel, tglExt string
			typeDescription, subTypeDescription, priorityDescription, statusDescription, contactDescription   sql.NullString
			relationDescription, tanggalPenyelesaian, waktuPenyelesaian, cabang, dateCR, foragingDays, usrUpd sql.NullString // Use sql.NullString for nullable string columns
			priorityID, statusID, subTypeID, employeeID                                                       sql.NullInt64  // Use sql.NullInt64 for nullable integer columnssubTypeID                                                                                                                                                                                                                                                                                  sql.NullInt64  // Use sql.NullInt64 for nullable integer columns
		)

		// Scan the row data
		if err := rows.Scan(
			&RowNumber, &flagCompany, &ticketNo, &agreementNo, &applicationID, &customerID, &typeID, &typeDescription,
			&subTypeID, &subTypeDescription, &priorityID, &priorityDescription, &statusID, &statusName,
			&statusDescription, &customerName, &branchID, &description, &phoneNo, &email, &usrUpd, &jml,
			&contactID, &contactDescription, &employeeID, &relationID, &relationDescription, &relationName,
			&cabang, &channel, &foragingDays, &tanggalPenyelesaian, &waktuPenyelesaian, &tglExt, &dateCR,
		); err != nil {
			// Log row scanning error
			log.Printf("Failed to scan row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row: " + err.Error()})
			return
		}

		// Log the row data being written to Excel
		rowData := []interface{}{
			RowNumber, flagCompany, ticketNo, agreementNo, applicationID, customerID, typeID, statusName, customerName, branchID,
			description, phoneNo, email, jml, contactID, relationName, relationID, channel,
			tglExt,
			getStringFromNullString(usrUpd),
			getStringFromNullString(foragingDays),
			getStringFromNullString(dateCR),
			getStringFromNullString(cabang),
			getStringFromNullString(waktuPenyelesaian),
			getStringFromNullString(tanggalPenyelesaian),
			getStringFromNullString(typeDescription),
			getStringFromNullString(subTypeDescription),
			getStringFromNullString(priorityDescription),
			getStringFromNullString(contactDescription),
			getStringFromNullString(relationDescription),
			getStringFromNullString(statusDescription),
			getIntFromNullInt64(subTypeID),
			getIntFromNullInt64(priorityID),
			getIntFromNullInt64(statusID),
			getIntFromNullInt64(employeeID),
		}
		log.Printf("Row Data: %v", rowData)

		// Write row data to the Excel file
		for col, data := range rowData {
			cell := fmt.Sprintf("%c%d", 'A'+col, rowIndex)
			f.SetCellValue(sheetName, cell, data)
		}
		rowIndex++

		// Write row data to the Excel file
		for col, data := range rowData {
			cell := fmt.Sprintf("%c%d", 'A'+col, rowIndex)
			f.SetCellValue(sheetName, cell, data)
		}
		rowIndex++
	}

	// Log the number of rows written to Excel
	log.Printf("Number of rows written to Excel: %d", rowIndex-2)

	// Save the Excel file
	fileName := "export.xlsx"
	if err := f.SaveAs(fileName); err != nil {
		// Log Excel file save error
		log.Printf("Failed to save Excel file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save Excel file: " + err.Error()})
		return
	}

	// Log the file path
	log.Printf("Excel file saved: %s", fileName)

	// Serve the file for download
	c.FileAttachment(fileName, fileName)

	// Log the end of the function
	log.Println("ExportXLS function completed successfully")
}

func getStringFromNullString(nullString sql.NullString) string {
	if nullString.Valid {
		return nullString.String
	}
	return "" // Default value for NULL
}

// Helper function to handle sql.NullInt64
func getIntFromNullInt64(nullInt sql.NullInt64) int64 {
	if nullInt.Valid {
		return nullInt.Int64
	}
	return 0 // Default value for NULL
}
