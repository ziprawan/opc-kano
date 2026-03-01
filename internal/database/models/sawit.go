package models

import "time"

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
	ParticipantId uint      `gorm:"primaryKey"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`

	MessageId  string
	BetSize    uint
	AcceptedBy uint

	Participant *Participant `gorm:"foreignKey:ParticipantId;references:ID"`
	Accepted    *Participant `gorm:"foreignKey:AcceptedBy;references:ID"`
}

func (_ SawitAttack) TableName() string {
	return "sawit_attack"
}
