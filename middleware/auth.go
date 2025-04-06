package middleware

import (
	"database/sql"
	"fmt"
	"nopi/internals/database"
	projectconfig "nopi/internals/project_config"
	"nopi/structs"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func CheckAuth(c *gin.Context) (name string, shouldRedirect bool) {
	db := database.GetDB()
	conf := projectconfig.GetConfig()
	auth, _ := c.Cookie("auth")
	shouldRedirect = false

	if auth == "" {
		shouldRedirect = true
	} else {
		token, err := jwt.ParseWithClaims(auth, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
			return conf.JWTSecret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

		if err != nil {
			shouldRedirect = true
		} else {
			if token == nil || !token.Valid {
				shouldRedirect = true
			} else {
				id := token.Claims.(*jwt.RegisteredClaims).Subject

				var pushName, customName sql.NullString
				var jid string
				err := db.QueryRow("SELECT push_name, custom_name, jid FROM contact WHERE id = $1", id).Scan(&pushName, &customName, &jid)
				if err != nil {
					c.SetCookie("auth", "", -1, "/", "", false, true)
					fmt.Println(err)
				} else {
					if customName.Valid {
						name = customName.String
					} else if pushName.Valid {
						name = pushName.String
					} else {
						name = fmt.Sprintf("User +%s", strings.Split(jid, "@")[0])
					}
				}
			}
		}
	}

	return
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		conf := projectconfig.LoadConfig()
		db := database.GetDB()
		auth, _ := c.Cookie("auth")

		h := gin.H{"navbar": gin.H{"name": ""}}
		shouldRedirect := false

		if auth == "" {
			shouldRedirect = true
		} else {
			token, err := jwt.ParseWithClaims(auth, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
				return conf.JWTSecret, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

			if err != nil {
				shouldRedirect = true
			} else {
				if token == nil || !token.Valid {
					shouldRedirect = true
				} else {
					id := token.Claims.(*jwt.RegisteredClaims).Subject

					var pushName, customName sql.NullString
					var jid string
					var realID int64
					err := db.QueryRow("SELECT id, push_name, custom_name, jid FROM contact WHERE id = $1", id).Scan(&realID, &pushName, &customName, &jid)
					if err != nil {
						c.SetCookie("auth", "", -1, "/", "", false, true)
						fmt.Println(err)
					} else {
						var name string

						if customName.Valid {
							name = customName.String
						} else if pushName.Valid {
							name = pushName.String
						} else {
							name = fmt.Sprintf("User +%s", strings.Split(jid, "@")[0])
						}

						h["navbar"].(gin.H)["name"] = name

						c.Set("user", structs.User{
							ID:         realID,
							PushName:   pushName,
							CustomName: customName,
							JID:        jid,
						})
					}
				}
			}
		}

		c.Set("h", h)

		if shouldRedirect {
			c.Redirect(302, "/auth/login")
			c.Abort()
		} else {
			c.Next()
		}
	}
}
