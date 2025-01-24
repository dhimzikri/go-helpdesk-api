package controllers

import (
	"fmt"
	"golang-sqlserver-app/config"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	// Query parameters for filtering and pagination
	query := c.Query("query")
	col := c.Query("col")
	start, _ := strconv.Atoi(c.DefaultQuery("start", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	// Base SQL query
	baseSQL := `
		SELECT user_name, real_name, b.name AS divisi
		FROM portal_ext.dbo.users a
		INNER JOIN portal_ext.dbo.cost_centers b ON a.cost_center_id = b.id_cost_center
		WHERE a.user_name NOT IN ('admin', 'opr') AND a.is_active = 1
	`

	// Add filtering
	if query != "" && col != "" {
		baseSQL += " AND " + col + " LIKE ?"
	}

	// Total count query
	var total int64
	countSQL := `
		SELECT COUNT(1) 
		FROM portal_ext.dbo.users a
		INNER JOIN portal_ext.dbo.cost_centers b ON a.cost_center_id = b.id_cost_center
		WHERE a.user_name NOT IN ('admin', 'opr') AND a.is_active = 1
	`
	if query != "" && col != "" {
		countSQL += " AND " + col + " LIKE ?"
		config.DB.Raw(countSQL, "%"+query+"%").Scan(&total)
	} else {
		config.DB.Raw(countSQL).Scan(&total)
	}

	// Paginated query
	paginatedSQL := baseSQL + " ORDER BY a.user_id OFFSET ? ROWS FETCH NEXT ? ROWS ONLY"
	var users []map[string]interface{}
	if query != "" && col != "" {
		config.DB.Raw(paginatedSQL, "%"+query+"%", start, limit).Scan(&users)
	} else {
		config.DB.Raw(paginatedSQL, start, limit).Scan(&users)
	}

	// Return JSON response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   total,
		"data":    users,
	})
}

func GetAssignments(c *gin.Context) {

	// Query parameters for pagination
	start, _ := strconv.Atoi(c.DefaultQuery("start", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	query := c.Query("query")
	col := c.Query("col")

	// Base SQL query
	baseSQL := `SELECT * FROM v_tblAssignment WHERE 1=1`

	// Add filtering
	if query != "" && col != "" {
		baseSQL += " AND " + col + " LIKE ?"
	}

	// Total count query
	var total int64
	countSQL := `SELECT COUNT(1) FROM v_tblAssignment WHERE 1=1`
	if query != "" && col != "" {
		countSQL += " AND " + col + " LIKE ?"
		config.DB.Raw(countSQL, "%"+query+"%").Scan(&total)
	} else {
		config.DB.Raw(countSQL).Scan(&total)
	}

	// Paginated query
	paginatedSQL := baseSQL + " ORDER BY id OFFSET ? ROWS FETCH NEXT ? ROWS ONLY"
	var assignments []map[string]interface{}
	if query != "" && col != "" {
		config.DB.Raw(paginatedSQL, "%"+query+"%", start, limit).Scan(&assignments)
	} else {
		config.DB.Raw(paginatedSQL, start, limit).Scan(&assignments)
	}

	// Return JSON response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   total,
		"data":    assignments,
	})
}

func SaveAssign(c *gin.Context) {
	// Parse JSON input
	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "Invalid input"})
		return
	}

	// Extract values from the input
	id := request["id"].(string)
	costcenterid := request["costcenterid"].(string)
	employeeid := request["employeeid"].(string)
	isactive := 0
	if request["isactive"] == "on" {
		isactive = 1
	}
	employeeid_alternate := request["employeeid_alternate"].(string)
	isactive_alternate := 0
	if request["isactive_alternate"] == "on" {
		isactive_alternate = 1
	}

	// Assuming `userid` is retrieved from a session or request context
	userid := "8023" // Replace this with your session-based or token-based user retrieval logic

	// Construct the SQL query using fmt.Sprintf
	sqlQuery := fmt.Sprintf(`
		set nocount on;
		exec sp_insertAssignment '%s', '%s', '%s', '%d', '%s', '%d', '%s';

		insert into tbllog ([tgl], [table_name], [menu], [script])
		select GETDATE(), '[tblAssignment]', 'assignment',
			   'exec sp_insertAssignment %s, %s, %s, %d, %s, %d, %s';

		select '1' as data;
	`, id, costcenterid, employeeid, isactive, employeeid_alternate, isactive_alternate, userid,
		id, costcenterid, employeeid, isactive, employeeid_alternate, isactive_alternate, userid)

	// Execute the SQL query
	err := config.DB.Exec(sqlQuery).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": err.Error()})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": ""})
}
