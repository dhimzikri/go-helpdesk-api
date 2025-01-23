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
		props.GET("/getSettingEmail", controllers.GetEmailSetting)
		props.GET("/getBranch_id", controllers.GetBranchID)

		// Props Data
		props.GET("/getRelation", controllers.GetRelation)
		props.GET("/getStatus", controllers.GetTblStatus)
		props.GET("/getAgreement", controllers.AgreementNoHandler())
		props.GET("/getContact", controllers.GetContact)
	}
}
