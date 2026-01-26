package models

import (
	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

type Contact struct {
	gorm.Model

	JID        types.JID `gorm:"not null;type:text;column:jid"`
	PushName   string
	CustomName string

	Participants []Participant
}

func (_ Contact) TableName() string {
	return "contact"
}
