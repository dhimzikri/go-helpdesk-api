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

func GetTblPriority(c *gin.Context) {
	query := c.Query("query")
	col := c.Query("col")

	src := "0=0"
	args := []interface{}{}
	if query != "" && col != "" {
		src += " AND " + col + " LIKE ?"
		args = append(args, "%"+query+"%")
	}

	sqlStr := "SELECT PriorityID, Description, Seq, Usrupd, Dtmupd FROM Priority WHERE " + src
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
			isactive    int
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

func GetBranch(c *gin.Context) {
	query := c.Query("query")
	col := c.Query("col")

	src := "0=0"
	args := []interface{}{}
	if query != "" && col != "" {
		src += " AND " + col + " LIKE ?"
		args = append(args, "%"+query+"%")
	}

	// sqlStr := "SELECT PriorityID, Description, Seq, Usrupd, Dtmupd FROM Priority WHERE " + src
	sqlStr := "SELECT * from (select id_cost_center as branchid,name as branchfullname from portal_ext.dbo.cost_centers union select branchid,branchfullname from [172.16.4.31].SGFDB.dbo.branch) as a WHERE " + src
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
			branchid       int
			branchfullname string
			// isactive    int
			// usrUpd      string
			// dtmUpd      string
		)

		if err := rows.Scan(
			&branchid, &branchfullname,
			// &typeID, &description, &isactive, &usrUpd, &dtmUpd,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}

		msg = append(msg, map[string]interface{}{
			"branchid":       branchid,
			"branchfullname": branchfullname,
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

func GetRelation(c *gin.Context) {
	query := c.Query("query")
	col := c.Query("col")

	src := "0=0"
	args := []interface{}{}
	if query != "" && col != "" {
		src += " AND " + col + " LIKE ?"
		args = append(args, "%"+query+"%")
	}

	sqlStr := "select * from Relation WHERE " + src
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

func GetTblStatus(c *gin.Context) {
	query := c.Query("query")
	col := c.Query("col")

	src := "0=0"
	args := []interface{}{}
	if query != "" && col != "" {
		src += " AND " + col + " LIKE ?"
		args = append(args, "%"+query+"%")
	}

	sqlStr := "SELECT * FROM Status WHERE " + src + " ORDER BY statusid"
	// Execute the SQL query
	rows, err := config.DB.Query(sqlStr, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	defer rows.Close()

	// Parse the results into a slice of maps
	var msg []map[string]interface{}
	for rows.Next() {
		var (
			statusID    int
			statusName  string
			usrUpd      *string // Pointer to handle NULL
			dtmUpd      *string // Pointer to handle NULL
			value       *string // Pointer to handle NULL
			isActive    *string
			description *string // Pointer to handle NULL
		)

		if err := rows.Scan(
			&statusID, &statusName, &usrUpd, &dtmUpd, &value, &isActive, &description,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}

		msg = append(msg, map[string]interface{}{
			"statusid":    statusID,
			"statusname":  statusName,
			"usrupd":      getStringOrNil(usrUpd), // Helper function for NULL
			"dtmupd":      getStringOrNil(dtmUpd), // Helper function for NULL
			"value":       getStringOrNil(value),  // Helper function for NULL
			"isactive":    getStringOrNil(isActive),
			"description": getStringOrNil(description), // Helper function for NULL
		})
	}

	if len(msg) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   len(msg),
		"data":    msg,
	})
}

// Helper function to handle NULL values
func getStringOrNil(value *string) string {
	if value == nil {
		return "" // Default to empty string if NULL
	}
	return *value
}
