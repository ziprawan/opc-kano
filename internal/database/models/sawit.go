package models

import (
	"database/sql"
	"time"
)

type Sawit struct {
	ParticipantId uint      `gorm:"primaryKey"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`

	LastGrowDate string
	Height       int

	AttackTotal          uint
	AttackWin            uint
	AttackAcquiredHeight uint
	AttackLostHeight     uint

	Participant *Participant `gorm:"foreignKey:ParticipantId;references:ID"`
}

func (_ Sawit) TableName() string {
	return "sawit"
}

type SawitAttack struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	ParticipantId uint
	GroupId       uint
	MessageId     string
	AttackSize    uint
	AcceptedBy    sql.NullInt32
	IsAttackerWin sql.NullBool

	Group       *Group       `gorm:"foreignKey:GroupId;references:ID"`
	Participant *Participant `gorm:"foreignKey:ParticipantId;references:ID"`
	Accepted    *Participant `gorm:"foreignKey:AcceptedBy;references:ID"`
}

func (_ SawitAttack) TableName() string {
	return "sawit_attack"
}
