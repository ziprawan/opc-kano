package models

import (
	"database/sql"

	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

type VoRequest struct {
	gorm.Model

	ChatJid   types.JID
	MessageId types.MessageID

	RequesterJid      types.JID
	MessageOwnerJid   types.JID
	ApprovalMessageId types.MessageID
	Accepted          sql.NullBool

	Url           sql.NullString
	DirectPath    sql.NullString
	MediaKey      string
	FileSha256    string
	FileEncSha256 string
	MediaType     string
}

func (_ VoRequest) TableName() string {
	return "vo_request"
}
