package message

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"kano/internals/database"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types"
)

func RandomHexString(length int) (string, error) {
	byteLen := (length + 1) / 2 // 2 hex chars per byte
	bytes := make([]byte, byteLen)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	hexStr := hex.EncodeToString(bytes)
	return hexStr[:length], nil
}

func LoginHandler(ctx *MessageContext) {
	if ctx.Instance.ChatJID().Server != types.DefaultUserServer {
		return
	}

	args := ctx.Parser.GetArgs()
	contact := ctx.Instance.Contact
	if contact == nil {
		ctx.Instance.Reply("Terjadi kesalahan saat mengambil data kontak", true)
		return
	}

	var redirect string
	if len(args) > 0 {
		firstParam := args[0].Content
		decodedRedirect, err := base64.RawURLEncoding.DecodeString(firstParam)
		if err == nil {
			redirect = string(decodedRedirect)
		}
	}

	if redirect != "" {
		splits := strings.Split(redirect, "|")
		if len(splits) < 2 {
			redirect = ""
		}

		if splits[0] == "title" {
			if len(splits) == 2 {
				redirect = fmt.Sprintf("/user/titles/%s", splits[1])
			} else {
				redirect = fmt.Sprintf("/user/titles/%s/%s", splits[1], splits[2])
			}
		}
	}

	randomizedHex, err := RandomHexString(32)
	if err != nil {
		ctx.Instance.Reply("Terjadi kesalahan saat membuat data", true)
		return
	}

	loginRequestID := sql.NullString{String: randomizedHex, Valid: true}
	loginExpirationDate := sql.NullTime{Time: time.Unix(time.Now().Unix()+3600, 0), Valid: true}
	loginRedirect := sql.NullString{Valid: false}

	if redirect != "" {
		loginRedirect.Valid = true
		loginRedirect.String = redirect
	}

	db := database.GetDB()
	stmt, err := db.Prepare("UPDATE contact SET login_request_id = $1, login_expiration_date = $2, login_redirect = $3 WHERE id = $4")
	if err != nil {
		ctx.Instance.Reply("Terjadi kesalahan saat menyiapkan data", true)
		return
	}

	_, err = stmt.Exec(loginRequestID, loginExpirationDate, loginRedirect, contact.ID)
	if err != nil {
		ctx.Instance.Reply("Terjadi kesalahan saat memasukkan data", true)
		return
	}

	// TODO: Since the redirect URL is hardcoded, I might want to add WEB_URL at env.sh
	ctx.Instance.Reply(fmt.Sprintf("https://opc.ajos.my.id/auth/onetaplogin?token=%s", loginRequestID.String), true)
}
