package user_titles

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"kano/internals/database"
	"kano/structs"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

// I assume when this function is called, the user is already authenticated
// and the middleware has set the user information in the context.
// Nah, just recheck it, so if the variable doesn't exists, just return 500

type AddTitleRequest struct {
	TitleName string `json:"title_name"`
	Claimable bool   `json:"claimable"`
}

// GET /user/titles/:group_id/add
func AddTitlePage(c *gin.Context) {
	h := gin.H{"title": "Titles"}
	hCtx, exists := c.Get("h")
	if !exists {
		h["navbar"] = gin.H{}
	}
	h["navbar"] = hCtx.(gin.H)["navbar"]

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

	if group.MemberRole == structs.MemberRoleMember {
		c.HTML(401, "components/error.html", gin.H{"error_code": 401, "error_message": "Hanya admin atau manager yang bisa mengakses laman ini"})
		return
	}

	h["group"] = group
	h["IsAdmin"] = group.MemberRole != structs.MemberRoleMember

	c.HTML(200, "titles/add.html", h)
}

// POST /user/titles/:group_id/add
func AddTitleAPI(c *gin.Context) {
	returnJSON := strings.Contains(c.Request.Header.Get("Accept"), "application/json")
	var bind AddTitleRequest
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

	bind.TitleName = strings.ToLower(bind.TitleName)
	specials := []string{"member", "manager", "admin", "superadmin"}

	for _, char := range bind.TitleName {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') {
			if returnJSON {
				c.JSON(422, gin.H{"code": 422, "description": "Title name must be alphanumeric"})
			} else {
				c.HTML(422, "components/error.html", gin.H{"error_code": 422, "error_message": "Title name must be alphanumeric"})
			}
			return
		}
	}

	if slices.Contains(specials, bind.TitleName) {
		if returnJSON {
			c.JSON(400, gin.H{"code": 400, "description": "Cannot use special title name"})
		} else {
			c.HTML(400, "components/error.html", gin.H{"error_code": 400, "error_message": "Cannot use special title name"})
		}
		return
	}

	if len(bind.TitleName) < 3 {
		if returnJSON {
			c.JSON(400, gin.H{"code": 400, "description": "Title length is less than 3"})
		} else {
			c.HTML(400, "components/error.html", gin.H{"error_code": 400, "error_message": "Title length is less than 3"})
		}
		return
	}

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
			c.JSON(403, gin.H{"code": 403, "description": fmt.Sprintf("Role %s is not allowed to do this request", group.MemberRole)})
		} else {
			c.HTML(403, "components/error.html", gin.H{"error_code": 403, "error_message": fmt.Sprintf("Role %s is not allowed to do this request", group.MemberRole)})
		}
		return
	}

	var ID int64
	err = db.QueryRow("SELECT id FROM group_title WHERE title_name = $1 AND group_id = $2", bind.TitleName, group.ID).Scan(&ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			fmt.Println("Failed to query title", err)
			if returnJSON {
				c.JSON(500, gin.H{"code": 500, "description": "Internal server error"})
			} else {
				c.HTML(500, "components/5xx.html", gin.H{})
			}
			return
		}
	} else {
		if returnJSON {
			c.JSON(400, gin.H{"code": 400, "description": "Title already exists"})
		} else {
			c.HTML(400, "components/error.html", gin.H{"error_code": 400, "error_message": "Title already exists"})
		}
		return
	}

	stmt, err := db.Prepare("INSERT INTO group_title VALUES (DEFAULT, $1, $2, $3)")
	if err != nil {
		fmt.Println("Failed to prepare insert title statement", err)
		if returnJSON {
			c.JSON(500, gin.H{"code": 500, "description": "Internal server error"})
		} else {
			c.HTML(500, "components/5xx.html", gin.H{})
		}
		return
	}

	_, err = stmt.Exec(group.ID, bind.TitleName, bind.Claimable)
	if err != nil {
		fmt.Println("Failed to execute insert title statement", err)
		fmt.Printf("%T\n", err)
		if returnJSON {
			c.JSON(500, gin.H{"code": 500, "description": "Internal server error"})
		} else {
			c.HTML(500, "components/5xx.html", gin.H{})
		}
		return
	}

	if returnJSON {
		c.JSON(200, gin.H{"code": 200, "data": gin.H{"title_name": bind.TitleName, "claimable": bind.Claimable}})
	} else {
		h := gin.H{"title": "Titles"}
		h["group"] = group
		h["IsAdmin"] = group.MemberRole != structs.MemberRoleMember
		c.HTML(200, "titles/list.html", h)
	}
}
