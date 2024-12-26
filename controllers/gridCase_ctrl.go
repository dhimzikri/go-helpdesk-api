package controllers

import (
	"fmt"
	"net/http"

	"golang-sqlserver-app/config"

	"github.com/gin-gonic/gin"
)

// GetTblType handles fetching data from the Type table
func GetTblType(c *gin.Context) {
	// Get query parameters
	// query := c.Query("query") // Example: ?query=some_value
	// col := c.Query("col")     // Example: ?col=description

	// Build the dynamic WHERE clause
	// src := "0=0 AND ParentID=0"
	// if query != "" && col != "" {
	// 	src = fmt.Sprintf("%s AND %s LIKE '%%%s%%'", src, col, query)
	// }

	// Construct the SQL query
	sqlStr := fmt.Sprintf(`
		SELECT *
		FROM tblType`)

	// Execute the SQL query
	rows, err := config.DB.Query(sqlStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	defer rows.Close()

	// Parse the results into a slice of maps
	var msg []map[string]interface{}
	for rows.Next() {
		var (
			typeID      int
			description string
			isactive    bool
			usrUpd      string
			dtmUpd      string
		)

		if err := rows.Scan(
			&typeID, &description, &isactive, &usrUpd, &dtmUpd,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}

		msg = append(msg, map[string]interface{}{
			"typeid":      typeID,
			"description": description,
			"isactive":    isactive,
			"usrupd":      usrUpd,
			"dtmupd":      dtmUpd,
		})
	}

	// Check if any data was retrieved
	if len(msg) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false})
		return
	}

	// Return the results as JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   len(msg),
		"data":    msg,
	})
}

func GetCase(c *gin.Context) {
	// Get query parameters
	query := c.DefaultQuery("query", "")
	col := c.DefaultQuery("col", "")
	start, limit := c.DefaultQuery("start", "0"), c.DefaultQuery("limit", "10")

	// For pagination
	startInt := 0
	limitInt := 10

	// Convert start and limit to integers
	if _, err := fmt.Sscanf(start, "%d", &startInt); err != nil {
		startInt = 0
	}
	if _, err := fmt.Sscanf(limit, "%d", &limitInt); err != nil {
		limitInt = 10
	}

	// Assuming the session username is stored somewhere (like in a cookie or token)
	userName := "user_name" // Replace with session-based user name

	// SQL conditions
	src := fmt.Sprintf("0=0 and a.statusid<>1 and a.usrupd='%s'", userName)
	if query != "" && col != "" {
		src += fmt.Sprintf(" AND %s LIKE '%%%s%%'", col, query)
	}

	// SQL query
	sqlStr := fmt.Sprintf(`
		set nocount on
		declare @jml as int
		select @jml=count(a.ticketno)
		from [Case] a
		inner join tbltype b on a.TypeID=b.TypeID
		inner join tblSubtype c on a.SubTypeID=c.SubTypeID and a.TypeID=c.TypeID
		inner join [Priority] d on a.PriorityID=d.PriorityID
		inner join [status] e on a.statusid=e.statusid
		inner join [contact] f on a.contactid=f.contactid
		inner join [relation] g on a.relationid=g.relationid
		where %s
		
		select *
		from (
			SELECT ROW_NUMBER() OVER (ORDER BY RIGHT(a.ticketno, 3) desc) AS 'RowNumber', a.flagcompany, a.ticketno, a.agreementno, a.applicationid, a.customerid, 
			a.typeid, b.description as typedescriontion, a.subtypeid, c.SubDescription as typesubdescriontion, a.priorityid, 
			d.Description as prioritydescription, a.statusid, e.statusname, e.description as statusdescription, a.customername, 
			a.branchid, a.description, a.phoneno, a.email, a.usrupd, a.dtmupd, a.date_cr, @jml as jml, f.contactid, f.Description as contactdescription, 
			a.relationid, g.description as relationdescription, a.relationname, a.callerid, a.email_, a.foragingdays
			from [Case] a
			inner join tbltype b on a.TypeID=b.TypeID
			inner join tblSubtype c on a.SubTypeID=c.SubTypeID and a.TypeID=c.TypeID
			inner join [Priority] d on a.PriorityID=d.PriorityID
			inner join [status] e on a.statusid=e.statusid
			inner join [contact] f on a.contactid=f.contactid
			inner join [relation] g on a.relationid=g.relationid
			where %s
		) as a
		where RowNumber>%d and RowNumber<=%d order by a.foragingdays desc`, src, src, startInt, startInt+limitInt)

	// Execute the SQL query
	rows, err := config.DB.Query(sqlStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	defer rows.Close()

	// Parse the results into a slice of maps
	var msg []map[string]interface{}
	for rows.Next() {
		var (
			FlagCompany         string
			TicketNo            string
			AgreementNo         string
			BranchID            string
			CustomerName        string
			ApplicationID       string
			CustomerID          string
			StatusID            int
			StatusDescription   string
			SubDescription      string
			StatusName          string
			TypeID              int
			TypeDescription     string
			SubTypeID           int
			SubTypeDescription  string
			PriorityID          int
			PriorityDescription string
			Description         string
			PhoneNo             string
			Email               string
			ContactID           string
			ContactDescription  string
			RelationID          string
			RelationDescription string
			RelationName        string
			UserUpd             string
			DtmUpd              string
			CallerID            string
			Email_              string
			DateCR              string
			ForagingDays        int
		)

		if err := rows.Scan(
			&FlagCompany, &TicketNo, &AgreementNo, &BranchID, &CustomerName, &ApplicationID,
			&CustomerID, &StatusID, &StatusDescription, &SubDescription, &StatusName, &TypeID, &TypeDescription,
			&SubTypeID, &SubTypeDescription, &PriorityID, &PriorityDescription, &Description, &PhoneNo, &Email, &ContactID,
			&ContactDescription, &RelationID, &RelationDescription, &RelationName, &UserUpd, &DtmUpd, &CallerID, &Email_, &DateCR, &ForagingDays,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}

		msg = append(msg, map[string]interface{}{
			"FlagCompany":         FlagCompany,
			"TicketNo":            TicketNo,
			"AgreementNo":         AgreementNo,
			"BranchID":            BranchID,
			"CustomerName":        CustomerName,
			"ApplicationID":       ApplicationID,
			"CustomerID":          CustomerID,
			"StatusID":            StatusID,
			"StatusDescription":   StatusDescription,
			"SubDescription":      SubDescription,
			"StatusName":          StatusName,
			"TypeID":              TypeID,
			"TypeDescription":     TypeDescription,
			"SubTypeID":           SubTypeID,
			"SubTypeDescription":  SubTypeDescription,
			"PriorityID":          PriorityID,
			"PriorityDescription": PriorityDescription,
			"Description":         Description,
			"PhoneNo":             PhoneNo,
			"Email":               Email,
			"ContactID":           ContactID,
			"ContactDescription":  ContactDescription,
			"RelationID":          RelationID,
			"RelationDescription": RelationDescription,
			"RelationName":        RelationName,
			"UserUpd":             UserUpd,
			"DtmUpd":              DtmUpd,
			"CallerID":            CallerID,
			"Email_":              Email_,
			"DateCR":              DateCR,
			"ForagingDays":        ForagingDays,
		})
	}

	// Check if any data was retrieved
	if len(msg) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false})
		return
	}

	// Return the results as JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   len(msg),
		"data":    msg,
	})
}
