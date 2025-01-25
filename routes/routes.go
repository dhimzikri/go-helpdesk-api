package routes

import (
	"golang-sqlserver-app/controllers"
	"golang-sqlserver-app/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Auth endpoints
	router := gin.Default()
	router.Use(middleware.RateLimitMiddleware)
	auth := r.Group("/auth")
	{
		auth.POST("/login", controllers.Login)
		auth.POST("/register", controllers.Register)
		auth.GET("/logout", controllers.Logout)
	}

	// CRUD endpoints
	props := r.Group("/props")
	{
		// Existing Cust
		props.GET("/getCase", controllers.GetCase)
		props.POST("/saveCase", controllers.SaveCase)
		props.POST("/closeCase", controllers.CloseCase)

		// New Cust
		props.GET("/getNewCase", controllers.GetCaseNewCust)
		props.POST("/saveNewCase", controllers.SaveNewCase)
		props.POST("/closeNewCase", controllers.CloseCaseNewCust)

		// Type and SubType
		props.GET("/getSubType", controllers.GetSubType)
		props.POST("/createSubType", controllers.CreateTblSubType)
		props.GET("/gettblType", controllers.GetTblType)
		props.POST("/saveType", controllers.CreateTblType)

		// Props
		props.GET("/gettblPriority", controllers.GetTblPriority)
		props.GET("/getHoliday", controllers.GetHolidays)
		props.GET("/getBranch_id", controllers.GetBranchID)

		//  Set Email
		props.GET("/getSettingEmail", controllers.GetEmailSetting)
		props.POST("/addSetEmail", controllers.AddSetEmail)

		// Props Data
		props.GET("/getRelation", controllers.GetRelation)
		props.POST("/addRelation", controllers.AddRelation)
		props.GET("/getAgreement", controllers.AgreementNoHandler())
		props.GET("/getContact", controllers.GetContact)

		// get Assigment
		props.GET("/getEmp", controllers.GetUsers)
		props.GET("/getAssign", controllers.GetAssignments)
		props.POST("/setAssign", controllers.SaveAssign)

		// SetStatus and Holiday
		props.POST("/addStatus", controllers.SaveStatus)
		props.GET("/getStatus", controllers.GetTblStatus)
		props.POST("/uploadXls", controllers.UploadXLS)
	}
}
