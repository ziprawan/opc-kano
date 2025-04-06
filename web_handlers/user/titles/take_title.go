package user_titles

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"nopi/internals/database"
	"nopi/structs"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

type TitleIDsRequest struct {
	TitleIDs []int64 `json:"title_ids"`
}

type TitleStatus struct {
	TitleID   int64
	Claimable bool
	HolderID  sql.NullInt64
	IsHolding sql.NullBool
}

// POST /user/titles/:group_id
func TakeTitles(c *gin.Context) {
	returnJSON := strings.Contains(c.Request.Header.Get("Accept"), "application/json")
	var bind TitleIDsRequest
	err := c.ShouldBind(&bind)
	if err != nil {
		h := gin.H{"Error": gin.H{"code": 422}}
		switch e := err.(type) {
		case *json.UnmarshalTypeError:
			h["Error"].(gin.H)["description"] = fmt.Sprintf("Invalid type at path \"%s\" (Expected %s)", e.Field, e.Type)
		case *json.SyntaxError:
			h["Error"].(gin.H)["description"] = fmt.Sprintf("Syntax error at offset %d", e.Offset)
		default:
			if e.Error() == "EOF" {
				h["Error"].(gin.H)["description"] = "EOF"
			} else {
				h["Error"].(gin.H)["description"] = "Internal server error"
				h["Error"].(gin.H)["code"] = 500
			}
		}

		if returnJSON {
			c.JSON(h["Error"].(gin.H)["code"].(int), h["Error"])
		} else {
			c.HTML(h["Error"].(gin.H)["code"].(int), "components/error.html", h)
		}

		fmt.Printf("%T\n", err)
		return
	}

	if bind.TitleIDs == nil {
		h := gin.H{"Error": gin.H{"code": 422, "description": "title_ids is null"}}
		if returnJSON {
			c.JSON(h["Error"].(gin.H)["code"].(int), h["Error"])
		} else {
			c.HTML(h["Error"].(gin.H)["code"].(int), "components/error.html", h)
		}
		return
	}

	uniqueTitleIDs := slices.Compact(bind.TitleIDs)
	// h := gin.H{}
	db := database.GetDB()
	groupCtx, _ := c.Get("group")
	group, ok := groupCtx.(structs.GroupInfo)
	if !ok {
		if returnJSON {
			c.JSON(500, gin.H{"code": 500, "description": "Internal server error"})
		} else {
			c.HTML(500, "components/5xx.html", gin.H{})
		}
		return
	}

	rows, err := db.Query("SELECT gt.id, gt.claimable, gth.id, gth.holding FROM group_title gt LEFT JOIN group_title_holder gth ON gth.group_title_id = gt.id AND gth.participant_id = $1 WHERE gt.group_id = $2", group.MemberID, group.ID)
	if err != nil {
		if returnJSON {
			c.JSON(500, gin.H{"code": 500, "description": "Internal server error"})
		} else {
			c.HTML(500, "components/5xx.html", gin.H{})
		}
		return
	}
	defer rows.Close()

	var updates []TitleStatus
	var inserts []TitleStatus
	var added []int64 = []int64{}
	var removed []int64 = []int64{}
	for rows.Next() {
		var stat TitleStatus
		err := rows.Scan(&stat.TitleID, &stat.Claimable, &stat.HolderID, &stat.IsHolding)
		if err != nil {
			fmt.Println(err)
			if returnJSON {
				c.JSON(500, gin.H{"code": 500, "description": "Internal server error"})
			} else {
				c.HTML(500, "components/5xx.html", gin.H{})
			}
			return
		}

		if !stat.Claimable {
			continue
		}

		if stat.IsHolding.Valid {
			if stat.IsHolding.Bool && !slices.Contains(uniqueTitleIDs, stat.TitleID) {
				stat.IsHolding.Bool = false
				updates = append(updates, stat)
				removed = append(removed, stat.TitleID)
			} else if !stat.IsHolding.Bool && slices.Contains(uniqueTitleIDs, stat.TitleID) {
				stat.IsHolding.Bool = true
				updates = append(updates, stat)
				added = append(added, stat.TitleID)
			}
		} else if !stat.IsHolding.Valid && slices.Contains(uniqueTitleIDs, stat.TitleID) {
			stat.IsHolding.Bool = true
			inserts = append(inserts, stat)
			added = append(added, stat.TitleID)
		}
	}

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		fmt.Println("Failed to initialize transaction", err)
		if returnJSON {
			c.JSON(500, gin.H{"code": 500, "description": "Internal server error"})
		} else {
			c.HTML(500, "components/5xx.html", gin.H{})
		}
		return
	}
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	for _, update := range updates {
		stmt, err := db.Prepare("UPDATE group_title_holder SET holding = $1 WHERE id = $2")
		if err != nil {
			fmt.Println("Failed to prepare update statement", err)
			if returnJSON {
				c.JSON(500, gin.H{"code": 500, "description": "Internal server error"})
			} else {
				c.HTML(500, "components/5xx.html", gin.H{})
			}
			return
		}

		_, err = stmt.Exec(update.IsHolding.Bool, update.HolderID.Int64)
		if err != nil {
			fmt.Println("Failed to execute update statement", err)
			if returnJSON {
				c.JSON(500, gin.H{"code": 500, "description": "Internal server error"})
			} else {
				c.HTML(500, "components/5xx.html", gin.H{})
			}
		}
	}

	for _, insert := range inserts {
		stmt, err := db.Prepare("INSERT INTO group_title_holder VALUES(DEFAULT, $1, $2, true)")
		if err != nil {
			fmt.Println("Failed to prepare insert statement", err)
			if returnJSON {
				c.JSON(500, gin.H{"code": 500, "description": "Internal server error"})
			} else {
				c.HTML(500, "components/5xx.html", gin.H{})
			}
			return
		}

		_, err = stmt.Exec(insert.TitleID, group.MemberID)
		if err != nil {
			fmt.Println("Failed to execute insert statement", err)
			if returnJSON {
				c.JSON(500, gin.H{"code": 500, "description": "Internal server error"})
			} else {
				c.HTML(500, "components/5xx.html", gin.H{})
			}
		}
	}

	if err = tx.Commit(); err != nil {
		fmt.Println("Failed to commit transaction", err)
		if returnJSON {
			c.JSON(500, gin.H{"code": 500, "description": "Internal server error"})
		} else {
			c.HTML(500, "components/5xx.html", gin.H{})
		}
		return
	}

	if returnJSON {
		c.JSON(200, gin.H{"code": 200, "data": gin.H{"added": added, "removed": removed}})
	} else {
		h := gin.H{"title": "Titles"}
		h["group"] = group
		h["IsAdmin"] = group.MemberRole != structs.MemberRoleMember
		c.HTML(200, "titles/list.html", h)
	}
}
