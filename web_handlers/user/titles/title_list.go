package user_titles

import (
	"fmt"
	"kano/internals/database"
	"kano/structs"

	"github.com/gin-gonic/gin"
)

// I assume when this function is called, the user is already authenticated
// and the middleware has set the user information in the context.
// Nah, just recheck it, so if the variable doesn't exists, just return 500

// GET /user/titles
func GetHome(c *gin.Context) {
	h := gin.H{"title": "Titles"}
	hCtx, exists := c.Get("h")
	if !exists {
		h["navbar"] = gin.H{}
	}
	h["navbar"] = hCtx.(gin.H)["navbar"]

	userInfo, exists := c.Get("user")
	if !exists {
		c.HTML(500, "components/5xx.html", h)
		return
	}
	user, ok := userInfo.(structs.User)
	if !ok {
		c.HTML(500, "components/5xx.html", h)
		return
	}

	db := database.GetDB()
	var groups []structs.BasicGroupInfo
	rows, err := db.Query("SELECT g.id, name, topic FROM participant INNER JOIN \"group\" g ON g.id = participant.group_id AND g.is_incognito = false WHERE contact_id = $1", user.ID)
	if err != nil {
		fmt.Println(err)
		c.HTML(500, "components/5xx.html", h)
		return
	}

	for rows.Next() {
		var group structs.BasicGroupInfo
		err := rows.Scan(&group.ID, &group.Name, &group.Topic)
		if err != nil {
			fmt.Println(err)
			c.HTML(500, "components/5xx.html", h)
			return
		}
		groups = append(groups, group)
	}

	h["groups"] = groups

	c.HTML(200, "titles/index.html", h)
}

// GET /user/titles/:id
func GetTitles(c *gin.Context) {
	h := gin.H{"title": "Titles"}
	hCtx, exists := c.Get("h")
	if !exists {
		h["navbar"] = hCtx.(gin.H)["navbar"].(gin.H)
	}
	h["navbar"] = hCtx.(gin.H)["navbar"]

	userInfo, exists := c.Get("user")
	if !exists {
		c.HTML(500, "components/5xx.html", h)
		return
	}
	user, ok := userInfo.(structs.User)
	if !ok {
		c.HTML(500, "components/5xx.html", h)
		return
	}

	groupCtx, exists := c.Get("group")
	if !exists {
		c.HTML(500, "components/5xx.html", h)
		return
	}
	group, ok := groupCtx.(structs.GroupInfo)
	if !ok {
		c.HTML(500, "components/5xx.html", h)
		return
	}

	h["user"] = user
	h["group"] = group
	h["IsAdmin"] = group.MemberRole != structs.MemberRoleMember

	c.HTML(200, "titles/list.html", h)
}
