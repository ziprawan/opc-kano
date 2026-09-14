package schedules

import (
	"kano/internal/database"
	"kano/internal/database/models"
)

// Universal
var db = database.GetInstance()

// lecturer.go
var lecturers []models.Lecturer

// room.go
var rooms []models.Room

// major-constraint.go
var majorConstraints []models.ConstraintMajor
