package models

import "gorm.io/gorm"

type Participant struct {
	gorm.Model

	GroupID   uint            `gorm:"not null"`
	ContactID uint            `gorm:"not null"`
	Role      ParticipantRole `gorm:"not null"`

	Group   *Group   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Contact *Contact `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (_ Participant) TableName() string {
	return "participant"
}
