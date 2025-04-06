package routes

import (
	"kano/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.RouterGroup) {
	userRouter := router.Group("/user")
	userRouter.Use(middleware.AuthMiddleware())
	{
		// /user
		userRouter.GET("", func(c *gin.Context) {
			c.Redirect(302, "/")
		})
		// /user/titles
		userTitleRoutes(userRouter)
	}
}
