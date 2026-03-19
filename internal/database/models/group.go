package models

import (
	"database/sql"

	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

type Group struct {
	gorm.Model

	JID            types.JID `gorm:"not null;type:text;column:jid"`
	Name           string    `gorm:"not null"`
	CommunityID    sql.NullInt64
	IsAnnouncement bool

	Settings     *GroupSettings `gorm:"foreignKey:ID;references:ID"`
	Community    *Community     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Participants []Participant
}

func (_ Group) TableName() string {
	return "group"
}

type GroupSettings struct {
	ID               uint `gorm:"primaryKey"`
	IsGameAllowed    bool
	IsConfessAllowed bool

	Group *Group `gorm:"foreignKey:ID;references:ID"`
}

func (_ GroupSettings) TableName() string {
	return "group_settings"
}
