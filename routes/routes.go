package routes

import (
	"golang-sqlserver-app/controllers"
	"golang-sqlserver-app/middleware"

	"github.com/gin-gonic/gin"
)

func AuthRoutes() *gin.Engine {
	router := gin.Default()
	router.Use(middleware.RateLimitMiddleware)
	r := router.Group("/user") // Group routes under /user
	{
		r.POST("/login", controllers.Login) // /user/login route
		r.POST("/register", controllers.Register)
		r.GET("/logout", controllers.Logout)
	}

	return router
}

func GetDataHelpDesk() *gin.Engine {
	router := gin.Default()
	router.Use(middleware.RateLimitMiddleware)
	r := router.Group("/get")
	{
		// CRUD endpoints
		r.GET("/getData", controllers.GetData)
		r.GET("/getCase", controllers.GetCase)
		r.POST("/getCase", controllers.GetCase)
		r.GET("/getRelation", controllers.GetRelation)
		r.GET("/getStatus", controllers.GetTblStatus)
		r.GET("/getAgreement", controllers.AgreementNoHandler())
		r.GET("/getContact", controllers.GetContact)
		r.GET("/getSubType", controllers.GetSubType)
		r.POST("/getSaveCase", controllers.SaveCaseHandler)
	}

	return router
}
