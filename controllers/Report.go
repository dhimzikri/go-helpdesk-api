package controllers

import (
	"fmt"
	"golang-sqlserver-app/config"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
						FORMAT(dbo.getTanggalPenyelesaian(a.ticketno), 'dd MMM yyyy') AS tanggalpenyelesaian, a.date_cr,
						---dbo.getWaktuPenyelesaian(a.ticketno) AS waktupenyelesaian,
						CASE
							WHEN dbo.getTanggalPenyelesaian(a.ticketno) IS NULL THEN NULL 
							WHEN dbo.GetBusinessDays(a.foragingdays, dbo.getTanggalPenyelesaian(a.ticketno)) = 0 
							THEN CONCAT('0 Day(s), ', FORMAT(dbo.getTanggalPenyelesaian(a.ticketno), 'HH:mm:ss'))
							ELSE CONCAT(
							    dbo.GetBusinessDays(a.foragingdays, dbo.getTanggalPenyelesaian(a.ticketno)), 
							    ' Day(s), ', 
							    FORMAT(dbo.getTanggalPenyelesaian(a.ticketno), 'HH:mm:ss')
							)
						END AS waktupenyelesaian,
						CASE 
							WHEN xx.statusid = 5 
							THEN FORMAT(xx.dtmupd, 'dd MMM yyyy') 
							ELSE '' 
						END AS tgl_ext
						---FORMAT(CONVERT(DATE, a.date_cr), 'dd MMM yyyy') AS date_cr
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
						FORMAT(dbo.getTanggalPenyelesaian(a.ticketno), 'dd MMM yyyy') AS tanggalpenyelesaian, a.date_cr,
						---dbo.getWaktuPenyelesaian(a.ticketno) AS waktupenyelesaian,
						CASE
							WHEN dbo.getTanggalPenyelesaian(a.ticketno) IS NULL THEN NULL 
							WHEN dbo.GetBusinessDays(a.foragingdays, dbo.getTanggalPenyelesaian(a.ticketno)) = 0 
							THEN CONCAT('0 Day(s), ', FORMAT(dbo.getTanggalPenyelesaian(a.ticketno), 'HH:mm:ss'))
							ELSE CONCAT(
								dbo.GetBusinessDays(a.foragingdays, dbo.getTanggalPenyelesaian(a.ticketno)), 
								' Day(s), ', 
								FORMAT(dbo.getTanggalPenyelesaian(a.ticketno), 'HH:mm:ss')
							)
						END AS waktupenyelesaian,
						CASE 
							WHEN xx.StatusName = 'Extend' 
							THEN FORMAT(xx.dtmupd, 'dd MMM yyyy') 
							ELSE '' 
						END AS tgl_ext
						---FORMAT(CONVERT(DATE, a.date_cr), 'dd MMM yyyy') AS date_cr
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
			where RowNumber>%d and RowNumber<=%d order by a.foragingdays asc
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
