package models

import (
	"database/sql"
	"time"
)

type EdunexClassId struct {
	ID          uint      `gorm:"primaryKey"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
	SemsCtx     string    `gorm:"not null"`
	SubjectCode string    `gorm:"not null"`
	ClassNum    uint      `gorm:"not null"`
	ClassId     sql.NullInt32
}

func (_ EdunexClassId) TableName() string {
	return "edunex_class_id"
}
