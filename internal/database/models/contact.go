package models

import (
	"database/sql"

	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

type Contact struct {
	gorm.Model

	JID        types.JID `gorm:"not null;type:text;column:jid"`
	PushName   string
	CustomName string

	ConfessTarget      sql.NullInt32
	ConfessTargetGroup *Group `gorm:"foreignKey:ConfessTarget;references:ID"`

	Participants []Participant
}

func (_ Contact) TableName() string {
	return "contact"
}
