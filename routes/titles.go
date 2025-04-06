package routes

import (
	"fmt"
	"nopi/middleware"
	user_titles "nopi/web_handlers/user/titles"

	"github.com/gin-gonic/gin"
)

func userTitleRoutes(router *gin.RouterGroup) {
	userTitlesGroup := router.Group("/titles")
	{
		userTitlesGroup.GET("", user_titles.GetHome)
		userTitleGroupRoutes(userTitlesGroup)
	}
}

func userTitleGroupRoutes(router *gin.RouterGroup) {
	userTitleGroupGroup := router.Group("/:group_id")
	userTitleGroupGroup.Use(middleware.GroupMiddleware())
	{
		userTitleGroupGroup.GET("", user_titles.GetTitles)
		userTitleGroupGroup.POST("", user_titles.TakeTitles)
		userTitleGroupGroup.GET("/add", user_titles.AddTitlePage)
		userTitleGroupGroup.POST("/add", user_titles.AddTitleAPI)
		userTitleGroupGroup.GET("/:title_id", func(c *gin.Context) {
			groupID := c.Param("group_id")
			c.Redirect(302, fmt.Sprintf("/user/titles/%s", groupID))
		})
		userTitleGroupGroup.GET("/:title_id/holders", user_titles.GetHolders)
		userTitleGroupGroup.POST("/:title_id/holders", user_titles.EditHolders)
	}
}
