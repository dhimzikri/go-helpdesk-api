package controllers

import (
	"fmt"
	"golang-sqlserver-app/config"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

// GetCases retrieves cases with pagination, search, and total count
var caseCache = cache.New(5*time.Minute, 10*time.Minute)

func GetJoinCases(c *gin.Context) {
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
	userid := c.GetString("admin") // Assuming you have middleware setting this

	var src string
	if userid == "admin" || readHeadCS(userid) { // You'll need to implement readHeadCS function
		src = fmt.Sprintf("CONVERT(DATE, a.dtmupd) >= '%s' and CONVERT(DATE, a.dtmupd) <= '%s'", startDate, endDate)
	} else {
		src = fmt.Sprintf("0=0 and a.dtmupd >= '%s' and a.dtmupd < dateadd(day, 1, '%s') and a.usrupd = '%s'", startDate, endDate, userid)
	}

	if query != "" && col != "" {
		src = src + " AND " + col + " LIKE '%" + query + "%'"
	}

	var results []map[string]interface{}
	// Main data query
	mainQuery := `SET NOCOUNT ON;

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
				WHERE ` + src + `
				
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
				WHERE ` + src + `
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
					WHERE ` + src + `
					
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
					WHERE ` + src + ` 
				) AS a
			) AS a
			where RowNumber>0 and RowNumber<=100 order by a.foragingdays asc`

	err := config.DB.Raw(mainQuery, start, start+limit).Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if len(results) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			// "total":   totalCount,
			"data": results,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
		})
	}
}

// func readHeadCS(userid string) bool {
// 	panic("unimplemented")
// }

func readHeadCS(username string) bool {
	// Fix the parameter name to match the stored procedure expectation
	var result int
	err := config.DB.Raw("EXEC sp_getHeadCS @user_name = ?", username).Scan(&result).Error

	if err != nil {
		// Log the error
		fmt.Printf("Error executing sp_getHeadCS: %v\n", err)
		return false
	}

	return result > 0
}

// // Helper function to build the query for [Case] table
// func buildCaseQuery(whereClause string) string {
// 	return fmt.Sprintf(`
// 		SELECT
//     ROW_NUMBER() OVER (ORDER BY RIGHT(a.ticketno, 3) DESC) AS RowNumber,
//     a.flagcompany, a.ticketno, a.agreementno, a.applicationid, a.customerid,
//     a.typeid, b.description AS typedescription, a.subtypeid, c.SubDescription AS typesubdescription,
//     a.priorityid, d.Description AS prioritydescription, a.statusid, e.statusname,
//     e.description AS statusdescription, a.customername, a.branchid, a.description, a.phoneno,
//     a.email, a.usrupd, a.dtmupd, a.date_cr, f.contactid, f.Description AS contactdescription,
//     a.relationid, g.description AS relationdescription, a.relationname, a.callerid, a.foragingdays,
//     dbo.FnGetFullBranchName(a.branchid) as cabang,
//    --- dbo.getWaktuPenyelesaian(a.ticketno) AS waktupenyelesaian,
//     CASE
//         WHEN dbo.getTanggalPenyelesaian(a.ticketno) IS NULL THEN NULL
//         WHEN dbo.GetBusinessDays(a.foragingdays, dbo.getTanggalPenyelesaian(a.ticketno)) = 0
//         THEN CONCAT('0 Day(s), ', FORMAT(dbo.getTanggalPenyelesaian(a.ticketno), 'HH:mm:ss'))
//         ELSE CONCAT(
//             dbo.GetBusinessDays(a.foragingdays, dbo.getTanggalPenyelesaian(a.ticketno)),
//             ' Day(s), ',
//             FORMAT(dbo.getTanggalPenyelesaian(a.ticketno), 'HH:mm:ss')
//         )
//     END AS waktupenyelesaian,
//     CASE
//         WHEN xx.statusid = 5
//         THEN FORMAT(xx.dtmupd, 'dd MMM yyyy')
//         ELSE ''
//     END AS tgl_ext,
//     CASE
//         WHEN d.PriorityID = 1 THEN
//             CASE
//                 WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) > 3 THEN 'green'
//                 WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) >= 0 AND (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) <= 3 THEN 'yellow'
//                 ELSE 'red'
//             END
//         WHEN d.PriorityID = 2 THEN
//             CASE
//                 WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) > 3 THEN 'green'
//                 WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) >= 0 AND (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) <= 3 THEN 'yellow'
//                 ELSE 'red'
//             END
//         WHEN d.PriorityID = 3 THEN
//             CASE
//                 WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) = 0 THEN 'green'
//                 ELSE 'red'
//             END
//         ELSE ''
//     END AS flagcolor,
//     d.sla - DATEDIFF(day, a.dtmupd, GETDATE()) AS overdue
// 		FROM [Case] a
// 		INNER JOIN tbltype b ON a.TypeID = b.TypeID
// 				INNER JOIN (
// 					SELECT a.*,
// 						CASE
// 							WHEN b.isactive = 1 AND b.isactive_alternate = 0 THEN b.employeeid
// 							WHEN b.isactive = 0 AND b.isactive_alternate = 1 THEN b.employeeid_alternate
// 							ELSE b.employeeid
// 						END AS employeeid
// 					FROM tblSubtype a
// 					LEFT JOIN tblassignment b ON a.cost_center = b.costcenterid
// 				) c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
// 		INNER JOIN Priority d ON a.PriorityID = d.PriorityID
// 		INNER JOIN status e ON a.statusid = e.statusid
// 		INNER JOIN contact f ON a.contactid = f.contactid
// 		INNER JOIN relation g ON a.relationid = g.relationid
// 		LEFT JOIN [PORTAL_EXT].[dbo].[users] u ON a.usrupd = u.user_name
// 		WHERE %s
// 	`, whereClause)
// }

// // Helper function to build the query for [Case_NewCustomer] table
// func buildCaseNewCustomerQuery(whereClause string) string {
// 	return fmt.Sprintf(`
// 		SELECT
//     ROW_NUMBER() OVER (ORDER BY RIGHT(a.ticketno, 3) DESC) AS RowNumber,
//     a.flagcompany, a.ticketno, a.agreementno, a.applicationid, a.customerid,
//     a.typeid, b.description AS typedescription, a.subtypeid, c.SubDescription AS typesubdescription,
//     a.priorityid, d.Description AS prioritydescription, a.statusid, e.statusname,
//     e.description AS statusdescription, a.customername, a.branchid, a.description, a.phoneno,
//     a.email, a.usrupd, a.dtmupd, a.date_cr, f.contactid, f.Description AS contactdescription,
//     a.relationid, g.description AS relationdescription, a.relationname, a.callerid, a.foragingdays,
//     dbo.FnGetFullBranchName(a.branchid) as cabang,
//     ---dbo.getWaktuPenyelesaian(a.ticketno) AS waktupenyelesaian,
//     CASE
//         WHEN dbo.getTanggalPenyelesaian(a.ticketno) IS NULL THEN NULL
//         WHEN dbo.GetBusinessDays(a.foragingdays, dbo.getTanggalPenyelesaian(a.ticketno)) = 0
//         THEN CONCAT('0 Day(s), ', FORMAT(dbo.getTanggalPenyelesaian(a.ticketno), 'HH:mm:ss'))
//         ELSE CONCAT(
//             dbo.GetBusinessDays(a.foragingdays, dbo.getTanggalPenyelesaian(a.ticketno)),
//             ' Day(s), ',
//             FORMAT(dbo.getTanggalPenyelesaian(a.ticketno), 'HH:mm:ss')
//         )
//     END AS waktupenyelesaian,
//     CASE
//         WHEN xx.statusid = 5
//         THEN FORMAT(xx.dtmupd, 'dd MMM yyyy')
//         ELSE ''
//     END AS tgl_ext,
//     CASE
//         WHEN d.PriorityID = 1 THEN
//             CASE
//                 WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) > 3 THEN 'green'
//                 WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) >= 0 AND (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) <= 3 THEN 'yellow'
//                 ELSE 'red'
//             END
//         WHEN d.PriorityID = 2 THEN
//             CASE
//                 WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) > 3 THEN 'green'
//                 WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) >= 0 AND (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) <= 3 THEN 'yellow'
//                 ELSE 'red'
//             END
//         WHEN d.PriorityID = 3 THEN
//             CASE
//                 WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) = 0 THEN 'green'
//                 ELSE 'red'
//             END
//         ELSE ''
//     END AS flagcolor,
//     d.sla - DATEDIFF(day, a.dtmupd, GETDATE()) AS overdue
// 		FROM [Case_NewCustomer] a
// 		INNER JOIN tbltype b ON a.TypeID = b.TypeID
// 				INNER JOIN (
// 					SELECT a.*,
// 						CASE
// 							WHEN b.isactive = 1 AND b.isactive_alternate = 0 THEN b.employeeid
// 							WHEN b.isactive = 0 AND b.isactive_alternate = 1 THEN b.employeeid_alternate
// 							ELSE b.employeeid
// 						END AS employeeid
// 					FROM tblSubtype a
// 					LEFT JOIN tblassignment b ON a.cost_center = b.costcenterid
// 				) c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
// 		INNER JOIN Priority d ON a.PriorityID = d.PriorityID
// 		INNER JOIN status e ON a.statusid = e.statusid
// 		INNER JOIN contact f ON a.contactid = f.contactid
// 		INNER JOIN relation g ON a.relationid = g.relationid
// 		LEFT JOIN [PORTAL_EXT].[dbo].[users] u ON a.usrupd = u.user_name
// 		WHERE %s
// 	`, whereClause)
// }
