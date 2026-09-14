package models

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func NoopJoin(
	db gorm.JoinBuilder,
	joinTable, curTable clause.Table,
) error {
	return nil
}
