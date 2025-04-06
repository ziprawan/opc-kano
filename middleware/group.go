package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	"kano/internals/database"
	"kano/structs"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GroupMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userCtx, exists := c.Get("user")
		if !exists {
			c.HTML(403, "components/error.html", gin.H{"error_code": 403, "error_message": "Forbidden"})
			c.Abort()
			return
		}
		user, ok := userCtx.(structs.User)
		if !ok {
			c.HTML(500, "components/5xx.html", gin.H{})
			c.Abort()
			return
		}

		db := database.GetDB()
		paramGroupID := c.Param("group_id")
		groupID, e := strconv.ParseUint(paramGroupID, 10, 0)
		if e != nil {
			c.HTML(400, "components/error.html", gin.H{
				"error_code":    "400",
				"error_message": "ID grup tidak valid",
			})
			c.Abort()
			return
		}

		var group structs.GroupInfo
		err := db.QueryRow("SELECT g.id, g.name, g.topic, p.id, p.role FROM participant p INNER JOIN \"group\" g ON g.id = p.group_id WHERE g.id = $1 AND p.contact_id = $2", groupID, user.ID).Scan(&group.ID, &group.Name, &group.Topic, &group.MemberID, &group.MemberRole)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.HTML(404, "components/404.html", gin.H{})
				c.Abort()
				return
			}

			fmt.Println(err)
			c.HTML(500, "components/5xx.html", gin.H{})
			c.Abort()
			return
		}

		rows, err := db.Query("SELECT gt.id, gt.title_name, gt.claimable, COUNT(gth.id) AS holder_count, BOOL_OR(CASE WHEN gth.participant_id = $2 AND gth.holding = true THEN true ELSE false END) FROM group_title gt LEFT JOIN group_title_holder gth ON gth.group_title_id = gt.id AND gth.holding = true WHERE gt.group_id = $1 GROUP BY gt.id, gt.title_name, gt.claimable", group.ID, group.MemberID)
		if err != nil {
			fmt.Println(err)
			c.HTML(500, "components/5xx.html", gin.H{})
			c.Abort()
			return
		}
		defer rows.Close()
		for rows.Next() {
			var title structs.Title
			err := rows.Scan(&title.ID, &title.Name, &title.Claimable, &title.HolderCount, &title.IsHolding)
			if err != nil {
				fmt.Println(err)
				c.HTML(500, "components/5xx.html", gin.H{})
				c.Abort()
				return
			}
			group.Titles = append(group.Titles, title) // Is this fine?
		}

		c.Set("group", group)
		c.Next()
	}
}
