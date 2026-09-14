package models

import (
	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

type Community struct {
	gorm.Model

	JID  types.JID `gorm:"not null;type:text;column:jid"`
	Name string    `gorm:"not null"`

	Groups []Group
}

func (_ Community) TableName() string {
	return "community"
}
