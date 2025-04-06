package user_titles

import (
	"database/sql"
	"fmt"
	"nopi/internals/database"
	"nopi/structs"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Holder struct {
	ParticipantID int64
	Name          string
	IsHolding     sql.NullBool
}

// GET /user/titles/:id/holders
func GetHolders(c *gin.Context) {
	h := gin.H{"title": "Title Holders"}
	hCtx, exists := c.Get("h")
	if !exists {
		h["navbar"] = gin.H{}
	}
	h["navbar"] = hCtx.(gin.H)["navbar"]

	titleIDParam := c.Param("title_id")
	titleID, err := strconv.ParseUint(titleIDParam, 10, 0)
	if err != nil {
		c.HTML(400, "components/error.html", gin.H{"error_code": 400, "error_message": "ID title bukan angka"})
		return
	}

	groupCtx, _ := c.Get("group")
	group, ok := groupCtx.(structs.GroupInfo)
	if !ok {
		c.HTML(500, "components/5xx.html", h)
		return
	}

	var found = false
	var name = ""
	for _, t := range group.Titles {
		if uint64(t.ID) == titleID {
			found = true
			name = t.Name
			break
		}
	}

	if !found {
		c.HTML(404, "components/404.html", h)
		return
	}

	db := database.GetDB()
	rows, err := db.Query("SELECT c.id, c.custom_name, c.push_name, c.jid, p.id, gth.holding FROM participant p LEFT JOIN group_title_holder gth ON gth.participant_id = p.id AND gth.group_title_id = $1 INNER JOIN contact c ON c.id = p.contact_id WHERE p.group_id = $2", titleID, group.ID)
	if err != nil {
		fmt.Println("Failed to initialize query", err)
		c.HTML(500, "components/5xx.html", h)
		return
	}
	defer rows.Close()

	var holders []Holder
	for rows.Next() {
		var user structs.User
		var holder Holder
		err := rows.Scan(&user.ID, &user.CustomName, &user.PushName, &user.JID, &holder.ParticipantID, &holder.IsHolding)
		if err != nil {
			fmt.Println("Failed to scan row", err)
			c.HTML(500, "components/5xx.html", h)
			return
		}

		var name = fmt.Sprintf("+%s", strings.Split(user.JID, "@")[0])
		if user.CustomName.Valid {
			name = fmt.Sprintf("%s (%s)", user.CustomName.String, name)
		} else if user.PushName.Valid {
			name = fmt.Sprintf("%s (%s)", user.PushName.String, name)
		}

		holder.Name = name
		holders = append(holders, holder)
	}

	h["Holders"] = holders
	h["group"] = group
	h["TitleName"] = name
	h["IsAdmin"] = group.MemberRole != structs.MemberRoleMember

	c.HTML(200, "titles/holders.html", h)
}
