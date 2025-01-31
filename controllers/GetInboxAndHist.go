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

// GetCases retrieves cases with pagination, search, and total count
var caseCache = cache.New(5*time.Minute, 10*time.Minute)

func GetJoinCases(c *gin.Context) {
	// Get page number and limit from query parameters
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "30")

	// Convert page and limit to integers
	pageNum, _ := strconv.Atoi(page)
	limitNum, _ := strconv.Atoi(limit)

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

	// Add condition for statusid
	conditions = append(conditions, "a.statusid <> 1")

	// Combine conditions with "AND"
	whereClause := strings.Join(conditions, " AND ")

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
		LEFT JOIN [PORTAL_EXT].[dbo].[users] u ON a.usrupd = u.user_name
		WHERE %s
	`, whereClause)

	var totalCount int64
	if err := config.DB.Raw(sqlCount).Scan(&totalCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count total records"})
		return
	}

	// Build SQL query for fetching paginated data
	sqlQuery := fmt.Sprintf(`
		WITH CombinedCases AS (
			%s
			UNION ALL
			%s
		)
		SELECT *
		FROM CombinedCases
		ORDER BY RowNumber
		OFFSET %d ROWS FETCH NEXT %d ROWS ONLY
	`, buildCaseQuery(whereClause), buildCaseNewCustomerQuery(whereClause), offset, limitNum)

	var cases []map[string]interface{}
	if err := config.DB.Raw(sqlQuery).Scan(&cases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch case data"})
		return
	}

	// Filter out the email_ field from the response
	filteredCases := make([]map[string]interface{}, 0)
	for _, caseData := range cases {
		delete(caseData, "email_") // Remove the email_ field
		filteredCases = append(filteredCases, caseData)
	}

	// Prepare the response
	response := gin.H{
		"total": totalCount,
		"cases": filteredCases,
	}

	// Return the results in JSON format
	c.JSON(http.StatusOK, response)
}

// Helper function to build the query for [Case] table
func buildCaseQuery(whereClause string) string {
	return fmt.Sprintf(`
		SELECT 
			ROW_NUMBER() OVER (ORDER BY RIGHT(a.ticketno, 3) DESC) AS RowNumber,
			a.flagcompany, a.ticketno, a.agreementno, a.applicationid, a.customerid,
			a.typeid, b.description AS typedescription, a.subtypeid, c.SubDescription AS typesubdescription,
			a.priorityid, d.Description AS prioritydescription, a.statusid, e.statusname,
			e.description AS statusdescription, a.customername, a.branchid, a.description, a.phoneno,
			a.email, a.usrupd, a.dtmupd, a.date_cr, f.contactid, f.Description AS contactdescription,
			a.relationid, g.description AS relationdescription, a.relationname, a.callerid, a.foragingdays,
			dbo.FnGetFullBranchName(a.branchid) as cabang, u.real_name,
			CASE 
				WHEN d.PriorityID = 1 THEN
					CASE 
						WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) > 3 THEN 'green' 
						WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) >= 0 AND (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) <= 3 THEN 'yellow' 
						ELSE 'red'
					END
				WHEN d.PriorityID = 2 THEN
					CASE 
						WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) > 3 THEN 'green' 
						WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) >= 0 AND (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) <= 3 THEN 'yellow' 
						ELSE 'red'
					END
				WHEN d.PriorityID = 3 THEN
					CASE 
						WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) = 0 THEN 'green' 
						ELSE 'red'
					END	
				ELSE ''	 											 
			END AS flagcolor,
			d.sla - DATEDIFF(day, a.dtmupd, GETDATE()) AS overdue
		FROM [Case] a
		INNER JOIN tbltype b ON a.TypeID = b.TypeID
		INNER JOIN tblSubtype c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
		INNER JOIN Priority d ON a.PriorityID = d.PriorityID
		INNER JOIN status e ON a.statusid = e.statusid
		INNER JOIN contact f ON a.contactid = f.contactid
		INNER JOIN relation g ON a.relationid = g.relationid
		LEFT JOIN [PORTAL_EXT].[dbo].[users] u ON a.usrupd = u.user_name
		WHERE %s
	`, whereClause)
}

// Helper function to build the query for [Case_NewCustomer] table
func buildCaseNewCustomerQuery(whereClause string) string {
	return fmt.Sprintf(`
		SELECT 
			ROW_NUMBER() OVER (ORDER BY RIGHT(a.ticketno, 3) DESC) AS RowNumber,
			a.flagcompany, a.ticketno, a.agreementno, a.applicationid, a.customerid,
			a.typeid, b.description AS typedescription, a.subtypeid, c.SubDescription AS typesubdescription,
			a.priorityid, d.Description AS prioritydescription, a.statusid, e.statusname,
			e.description AS statusdescription, a.customername, a.branchid, a.description, a.phoneno,
			a.email, a.usrupd, a.dtmupd, a.date_cr, f.contactid, f.Description AS contactdescription,
			a.relationid, g.description AS relationdescription, a.relationname, a.callerid, a.foragingdays,
			dbo.FnGetFullBranchName(a.branchid) as cabang, u.real_name,
			CASE 
				WHEN d.PriorityID = 1 THEN
					CASE 
						WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) > 3 THEN 'green' 
						WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) >= 0 AND (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) <= 3 THEN 'yellow' 
						ELSE 'red'
					END
				WHEN d.PriorityID = 2 THEN
					CASE 
						WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) > 3 THEN 'green' 
						WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) >= 0 AND (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) <= 3 THEN 'yellow' 
						ELSE 'red'
					END
				WHEN d.PriorityID = 3 THEN
					CASE 
						WHEN (d.sla - DATEDIFF(day, a.dtmupd, GETDATE())) = 0 THEN 'green' 
						ELSE 'red'
					END	
				ELSE ''	 											 
			END AS flagcolor,
			d.sla - DATEDIFF(day, a.dtmupd, GETDATE()) AS overdue
		FROM [Case_NewCustomer] a
		INNER JOIN tbltype b ON a.TypeID = b.TypeID
		INNER JOIN tblSubtype c ON a.SubTypeID = c.SubTypeID AND a.TypeID = c.TypeID
		INNER JOIN Priority d ON a.PriorityID = d.PriorityID
		INNER JOIN status e ON a.statusid = e.statusid
		INNER JOIN contact f ON a.contactid = f.contactid
		INNER JOIN relation g ON a.relationid = g.relationid
		LEFT JOIN [PORTAL_EXT].[dbo].[users] u ON a.usrupd = u.user_name
		WHERE %s
	`, whereClause)
}
