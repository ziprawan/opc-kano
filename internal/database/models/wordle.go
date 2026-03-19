package models

import "github.com/lib/pq"

type Wordle struct {
	ID       uint `gorm:"primaryKey"`
	Word     string
	Point    uint
	Lang     string
	IsWordle bool
}

func (_ Wordle) TableName() string {
	return "wordle"
}

type UserWordle struct {
	ID      uint           `gorm:"primaryKey"`
	Guesses pq.StringArray `gorm:"type:text[]"`
	DateStr string

	TargetId uint
	Target   Wordle `gorm:"foreignKey:TargetId"`
	UserId   uint
	User     Contact `gorm:"foreignKey:UserId"`
}

func (_ UserWordle) TableName() string {
	return "user_wordle"
}
