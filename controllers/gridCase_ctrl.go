package controllers

import (
	"golang-sqlserver-app/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetTblType(c *gin.Context) {
	// Define a slice of maps to hold the query results
	var results []map[string]interface{}

	// Execute the query using GORM's raw SQL method
	if err := config.DB.Table("tblType").Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Check if any data was retrieved
	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "data": []map[string]interface{}{}})
		return
	}

	// Return the results as JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   len(results),
		"data":    results,
	})
}

func GetTblPriority(c *gin.Context) {
	// Define a slice of maps to hold the query results
	var results []map[string]interface{}

	// Execute the query using GORM's raw SQL method
	if err := config.DB.Table("Priority").Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Check if any data was retrieved
	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "data": []map[string]interface{}{}})
		return
	}

	// Return the results as JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   len(results),
		"data":    results,
	})
}

func GetBranch(c *gin.Context) {
	// Parse query parameters
	query := c.Query("query") // e.g., ?query=some_value
	col := c.Query("col")     // e.g., ?col=name

	// Initialize the result structure
	result := gin.H{"success": false}

	// Base query with UNION
	sqlStr := `
		SELECT branchid, branchfullname FROM (
			SELECT id_cost_center AS branchid, name AS branchfullname FROM portal_ext.dbo.cost_centers
			UNION
			SELECT branchid, branchfullname FROM [172.16.4.31].SGFDB.dbo.branch
		) AS a
	`

	// Add filtering if query and col are provided
	if query != "" && col != "" {
		sqlStr += " WHERE " + col + " LIKE ?"
	}

	// Execute the query
	var branches []map[string]interface{}
	var err error

	if query != "" && col != "" {
		err = config.DB.Raw(sqlStr, "%"+query+"%").Scan(&branches).Error
	} else {
		err = config.DB.Raw(sqlStr).Scan(&branches).Error
	}

	if err != nil {
		result["error"] = err.Error()
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	// Check if data is retrieved
	if len(branches) == 0 {
		result["data"] = []map[string]interface{}{} // Return an empty array
		result["total"] = 0
	} else {
		result["success"] = true
		result["data"] = branches
		result["total"] = len(branches)
	}

	// Return the result
	c.JSON(http.StatusOK, result)
}

func GetRelation(c *gin.Context) {
	// Define a slice of maps to hold the query results
	var results []map[string]interface{}

	// Execute the query using GORM's raw SQL method
	if err := config.DB.Table("Relation").Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Check if any data was retrieved
	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "data": []map[string]interface{}{}})
		return
	}

	// Return the results as JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   len(results),
		"data":    results,
	})
}

func GetTblStatus(c *gin.Context) {
	// Define a slice of maps to hold the query results
	var results []map[string]interface{}

	// Execute the query using GORM's raw SQL method
	if err := config.DB.Table("Status").Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Check if any data was retrieved
	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "data": []map[string]interface{}{}})
		return
	}

	// Return the results as JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   len(results),
		"data":    results,
	})
}

func ReadGetAgreementNo(c *gin.Context) {
	// Parse request parameters
	query := c.Query("query")             // ?query=value
	col := c.Query("col")                 // ?col=column_name
	flagCompany := c.Query("flagcompany") // ?flagcompany=value
	agreementNo := strings.TrimSpace(c.Query("agreementno"))
	applicationID := strings.TrimSpace(c.Query("applicationid"))
	customerName := strings.TrimSpace(c.Query("customername"))

	// Build dynamic WHERE clause
	src := "0=0 AND " +
		"a.agreementno LIKE ? AND " +
		"a.applicationid LIKE ? AND " +
		"b.name LIKE ?"

	if query != "" && col != "" {
		src += " AND " + col + " LIKE ?"
	}

	// Build SQL query
	sqlStr := "EXEC sp_getInformation ?, ?"

	// Define parameters for the SQL query
	params := []interface{}{
		flagCompany,
		src,
		"%" + agreementNo + "%",
		"%" + applicationID + "%",
		"%" + customerName + "%",
	}

	if query != "" && col != "" {
		params = append(params, "%"+query+"%")
	}

	// Execute the query
	var results []map[string]interface{}
	err := config.DB.Raw(sqlStr, params...).Scan(&results).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Check if data was retrieved
	if len(results) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"total":   0,
			"data":    []map[string]interface{}{},
		})
		return
	}

	// Return the results
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"total":   len(results),
		"data":    results,
	})
}

// 	// Build the dynamic SQL condition based on provided filters
// 	src := "1=1"
// 	args := []interface{}{flagCompany}

// 	// Handle agreementNo filter
// 	if agreementNo != "" {
// 		src += " AND a.agreementno LIKE @agreementNo"
// 		args = append(args, "%"+agreementNo+"%")
// 	}

// 	// Handle applicationID filter
// 	if applicationID != "" {
// 		src += " AND a.applicationid LIKE @applicationID"
// 		args = append(args, "%"+applicationID+"%")
// 	}

// 	// Handle customerName filter
// 	if customerName != "" {
// 		src += " AND b.name LIKE @customerName"
// 		args = append(args, "%"+customerName+"%")
// 	}

// 	// Handle query and col filter
// 	if query != "" && col != "" {
// 		src += " AND " + col + " LIKE @query"
// 		args = append(args, "%"+query+"%")
// 	}

// 	// Construct the SQL query string with parameter declarations
// 	sqlStr := `
// 		DECLARE @flagCompany VARCHAR(255)
// 		DECLARE @src VARCHAR(MAX)

// 		SET @flagCompany = @flagCompany
// 		SET @src = @src
// 		EXEC sp_getInformation @flagCompany, @src, @agreementNo, @applicationID, @customerName, @query
// 	`

// 	// Prepare and execute the query
// 	rows, err := config.DB.Query(sqlStr, args...)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
// 		return
// 	}
// 	defer rows.Close()

// 	// Parse the results into a slice of maps
// 	var msg []AgreementInfo
// 	for rows.Next() {
// 		var agreement AgreementInfo
// 		// Use the correct field order and types for your query result
// 		if err := rows.Scan(
// 			&agreement.FlagCompany,
// 			&agreement.customerid,
// 			&agreement.BranchID,
// 			&agreement.BranchName,
// 			&agreement.AgreementNo,
// 			&agreement.CustomerType,
// 			&agreement.DOB,
// 			&agreement.NamaIbuKandung,
// 			&agreement.MerkTypeKendaraan,
// 			&agreement.NamaDebitur,
// 			&agreement.Alamat,
// 			&agreement.Phone,
// 			&agreement.Nopol,
// 			&agreement.ContractStatus,
// 			&agreement.ApplicationID,
// 			&agreement.PhoneNumber,
// 			&agreement.Email,
// 		); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
// 			return
// 		}

// 		msg = append(msg, agreement)
// 	}

// 	// Check if any data was retrieved
// 	if len(msg) == 0 {
// 		c.JSON(http.StatusOK, gin.H{"success": false})
// 		return
// 	}

// 	// Return the results as JSON
// 	c.JSON(http.StatusOK, gin.H{
// 		"success": true,
// 		"total":   len(msg),
// 		"data":    msg,
// 	})
// }

// func GetAgreementNo(c *gin.Context) {
// 	query := c.Query("query")
// 	col := c.Query("col")
// 	flagCompany := c.Query("flagcompany")

// 	src := "0=0"
// 	args := []interface{}{}
// 	if query != "" && col != "" {
// 		src += " AND " + col + " LIKE ?"
// 		args = append(args, "%"+query+"%")
// 	}

// 	sqlStr := "EXEC sp_getInformation ?, ?"
// 	rows, err := config.DB.Raw(sqlStr, flagCompany, src)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
// 		return
// 	}
// 	defer rows.Close()

// 	// Handle parsing and returning data
// 	var msg []map[string]interface{}
// 	for rows.Next() {
// 		var flagcompany, branchname, agreementno, customertype, dob, nama_ibu_kandung, merk_type_kendaraan, nama_debitur, alamat, contractstatus, email string
// 		var customerid, branchid, phone, nopol, applicationid, phonenumber int
// 		// var isactive bool

// 		if err := rows.Scan(&flagcompany, &customerid, &branchid, &branchname,
// 			&agreementno, &customertype, &dob, &nama_ibu_kandung, &merk_type_kendaraan,
// 			&nama_debitur, &alamat, &phone, &nopol, &contractstatus, &applicationid, &phonenumber, &email); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
// 			return
// 		}

// 		msg = append(msg, map[string]interface{}{
// 			"flagcompany":         flagcompany,
// 			"customerid":          customerid,
// 			"branchid":            branchid,
// 			"branchname":          branchname,
// 			"agreementno":         agreementno,
// 			"customertype":        customertype,
// 			"dob":                 dob,
// 			"nama_ibu_kandung":    nama_ibu_kandung,
// 			"merk_type_kendaraan": merk_type_kendaraan,
// 			"nama_debitur":        nama_debitur,
// 			"alamat":              alamat,
// 			"phone":               phone,
// 			"nopol":               nopol,
// 			"contractstatus":      contractstatus,
// 			"applicationid":       applicationid,
// 			"phonenumber":         phonenumber,
// 			"email":               email,
// 		})
// 	}

// 	if len(msg) == 0 {
// 		c.JSON(http.StatusOK, gin.H{"success": false})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"success": true,
// 		"total":   len(msg),
// 		"data":    msg,
// 	})
// }
