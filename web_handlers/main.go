package webhandlers

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"kano/internals/database"
	projectconfig "kano/internals/project_config"
	"kano/internals/utils/account"
	"kano/middleware"
	"kano/routes"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginRequest struct {
	ContactID           int64
	LoginExpirationDate sql.NullTime
	LoginRedirect       sql.NullString
}

func CustomRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				log.Printf("Recovered from panic: %v\n", err)

				// Return a JSON response instead of crashing
				c.String(500, "Internal Server Error")
				c.Abort()
			}
		}()
		c.Next()
	}
}

func dict(values ...any) map[string]any {
	result := make(map[string]any)
	for i := 0; i < len(values); i += 2 {
		key := values[i].(string)
		result[key] = values[i+1]
	}
	return result
}

func Web() {
	db := database.GetDB()

	r := gin.Default()

	tmpl := template.New("").Funcs(template.FuncMap{"dict": dict})
	tmpl = template.Must(tmpl.ParseGlob("templates/**/*.html"))
	r.SetHTMLTemplate(tmpl)

	r.GET("/css/generated-tailwind.css", func(c *gin.Context) {
		c.File("templates/css/generated-tailwind.css")
	})

	r.GET("/images/favicon.webp", func(c *gin.Context) {
		c.File("templates/images/favicon.webp")
	})

	r.GET("/", func(c *gin.Context) {
		conf := projectconfig.GetConfig()
		h, exists := c.Get("h")

		if !exists {
			h = gin.H{"navbar": gin.H{"name": ""}}
		}

		h.(gin.H)["title"] = "Home"

		auth, _ := c.Cookie("auth")

		if auth != "" {
			token, err := jwt.ParseWithClaims(auth, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
				return conf.JWTSecret, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

			if err != nil {
				c.SetCookie("auth", "", -1, "/", "", false, true)
			} else {
				if token != nil && token.Valid {
					id := token.Claims.(*jwt.RegisteredClaims).Subject

					var pushName, customName sql.NullString
					var jid string
					err := db.QueryRow("SELECT push_name, custom_name, jid FROM contact WHERE id = $1", id).Scan(&pushName, &customName, &jid)
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

						h.(gin.H)["navbar"].(gin.H)["name"] = name
					}
				} else {
					c.SetCookie("auth", "", -1, "/", "", false, true)
				}
			}
		}

		c.HTML(200, "home/index.html", h)
	})

	r.GET("/auth/logout", func(c *gin.Context) {
		c.SetCookie("auth", "", -1, "/", "", false, true)
		c.Redirect(302, "/")
	})

	r.GET("/auth/login", func(c *gin.Context) {
		_, mustLogin := middleware.CheckAuth(c)
		redirect := c.Query("redirect")
		acc, err := account.GetData()
		if err != nil {
			c.String(500, "Internal server error")
			return
		}
		if acc.JID == nil {
			c.String(500, "Internal server error")
			return
		}

		phoneNumber := acc.JID.User

		if mustLogin {
			encodedRedirect := base64.RawURLEncoding.EncodeToString([]byte(redirect))
			c.Redirect(302, fmt.Sprintf("https://wa.me/+%s?text=.login %s", phoneNumber, encodedRedirect))
		} else {
			if redirect == "" {
				redirect = "/"
			}

			c.Redirect(302, redirect)
		}
	})

	r.GET("/auth/onetaplogin", func(c *gin.Context) {
		token := c.Query("token")

		var req LoginRequest
		db := database.GetDB()
		err := db.QueryRow("SELECT c.id, c.login_expiration_date, c.login_redirect FROM contact c WHERE c.login_request_id = $1", token).Scan(&req.ContactID, &req.LoginExpirationDate, &req.LoginRedirect)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.String(403, "Invalid token")
				return
			}

			c.String(500, "Internal server error")
			return
		}

		if !req.LoginExpirationDate.Valid || (req.LoginExpirationDate.Valid && req.LoginExpirationDate.Time.Unix() < time.Now().Unix()) {
			db.Exec("UPDATE contact c SET login_request_id = NULL, login_expiration_date = NULL, login_redirect = NULL WHERE c.id = $1", req.ContactID)
			c.String(403, "Token expired")
			return
		}

		conf := projectconfig.GetConfig()

		claims := jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", req.ContactID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		}

		JWTToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := JWTToken.SignedString(conf.JWTSecret)
		if err != nil {
			c.String(500, "Internal server error")
			return
		}

		c.SetCookie("auth", signedToken, 24*3600, "/", "", false, true)

		if req.LoginRedirect.Valid {
			c.Redirect(302, req.LoginRedirect.String)
		} else {
			c.Redirect(302, "/")
		}
	})

	r.Use(CustomRecovery())

	paths := r.Group("/")
	{
		routes.UserRoutes(paths)
	}

	r.Run(":8080")
}
