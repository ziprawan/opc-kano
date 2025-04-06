package user_titles

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"kano/internals/database"
	"kano/structs"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ParticipantIDsRequest struct {
	ParticipantIDs []int64 `json:"participant_ids"`
}

type HolderStatus struct {
	ParticipantID int64
	HolderID      sql.NullInt64
	IsHolding     sql.NullBool
}

// POST /user/titles/:group_id/:title_id/holders
func EditHolders(c *gin.Context) {
	returnJSON := strings.Contains(c.Request.Header.Get("Accept"), "application/json")
	var bind ParticipantIDsRequest
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

	if bind.ParticipantIDs == nil {
		h := gin.H{"Error": gin.H{"code": 422, "description": "title_ids is null"}}
		if returnJSON {
			c.JSON(h["Error"].(gin.H)["code"].(int), h["Error"])
		} else {
			c.HTML(h["Error"].(gin.H)["code"].(int), "components/error.html", h)
		}
		return
	}

	uniqueParticipantIDs := slices.Compact(bind.ParticipantIDs)
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

	if group.MemberRole == structs.MemberRoleMember {
		if returnJSON {
			c.JSON(401, gin.H{"code": 401, "description": fmt.Sprintf("Role %s is not allowed to do this request", group.MemberRole)})
		} else {
			c.HTML(401, "components/error.html", gin.H{"error_code": 401, "error_message": fmt.Sprintf("Role %s is not allowed to do this request", group.MemberRole)})
		}
		return
	}

	titleIDParam := c.Param("title_id")
	titleID, err := strconv.ParseUint(titleIDParam, 10, 0)
	if err != nil {
		if returnJSON {
			c.JSON(422, gin.H{"code": 422, "description": "Title ID is not a number"})
		} else {
			c.HTML(422, "components/error.html", gin.H{"error_code": 422, "error_message": "Title ID is not a number"})
		}
		return
	}

	var found = false
	for _, t := range group.Titles {
		if uint64(t.ID) == titleID {
			found = true
			break
		}
	}
	if !found {
		if returnJSON {
			c.JSON(404, gin.H{"code": 404, "description": "Title ID is not found"})
		} else {
			c.HTML(404, "components/404.html", gin.H{})
		}
		return
	}

	rows, err := db.Query("SELECT p.id, gth.id, gth.holding FROM participant p LEFT JOIN group_title_holder gth ON gth.participant_id = p.id AND gth.group_title_id = $1 WHERE p.group_id = $2", titleID, group.ID)
	if err != nil {
		fmt.Println("Failed to query participant and holder status", err)
		if returnJSON {
			c.JSON(500, gin.H{"code": 500, "description": "Internal server error"})
		} else {
			c.HTML(500, "components/5xx.html", gin.H{})
		}
		return
	}
	defer rows.Close()

	var updates []HolderStatus
	var inserts []HolderStatus
	var added []int64 = []int64{}
	var removed []int64 = []int64{}
	for rows.Next() {
		var stat HolderStatus
		err := rows.Scan(&stat.ParticipantID, &stat.HolderID, &stat.IsHolding)
		if err != nil {
			fmt.Println("Failed to scan participant and holder status row", err)
			if returnJSON {
				c.JSON(500, gin.H{"code": 500, "description": "Internal server error"})
			} else {
				c.HTML(500, "components/5xx.html", gin.H{})
			}
			return
		}

		if stat.IsHolding.Valid {
			if stat.IsHolding.Bool && !slices.Contains(uniqueParticipantIDs, stat.ParticipantID) {
				stat.IsHolding.Bool = false
				updates = append(updates, stat)
				removed = append(removed, stat.ParticipantID)
			} else if !stat.IsHolding.Bool && slices.Contains(uniqueParticipantIDs, stat.ParticipantID) {
				stat.IsHolding.Bool = true
				updates = append(updates, stat)
				added = append(added, stat.ParticipantID)
			}
		} else if !stat.IsHolding.Valid && slices.Contains(uniqueParticipantIDs, stat.ParticipantID) {
			stat.IsHolding.Bool = true
			inserts = append(inserts, stat)
			added = append(added, stat.ParticipantID)
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

		_, err = stmt.Exec(titleID, insert.ParticipantID)
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
