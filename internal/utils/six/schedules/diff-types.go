package schedules

import (
	"database/sql"
	"kano/internal/database/models"
)

type vars interface {
	~string | uint | sql.NullInt32 | NullConstraint | ClassLink
}

type varDiff[T vars] struct {
	Before T `json:"before"`
	After  T `json:"after"`

	// Must be the value returned by Before != After
	HasDiff bool `json:"has_diff"`
}

type SemesterDiff struct {
	ID uint `json:"id"` // Primary key

	// Would likely to happen, but the chance are low
	AddedSubjects   []Subject `json:"added_subjects"`
	RemovedSubjects []Subject `json:"removed_subjects"`

	// This also includes modified subject classes
	ModifiedSubjects []SubjectDiff `json:"modified_subjects"`
}

type SubjectDiff struct {
	ID uint `json:"id"` // Primary key

	// Not adding code, name, and sks field here as varDiff[T] type
	// Because I already updated them (refer to subject.go)

	AddedClasses   []Class `json:"added_class"`
	RemovedClasses []Class `json:"removed_class"`

	// This also includes modified class schedules
	ModifiedClasses []ClassDiff `json:"modified_class"`
}

type ClassDiff struct {
	ID uint `json:"id"` // Primary key

	Number      varDiff[uint]           `json:"number"` // Unlikely to happen
	Quota       varDiff[sql.NullInt32]  `json:"quota"`
	Constraints varDiff[NullConstraint] `json:"constraints"`
	Links       varDiff[ClassLink]      `json:"links"`

	AddedLecturers   []string `json:"added_lecturers"`
	RemovedLecturers []string `json:"removed_lecturers"`

	AddedSchedules    []Schedule             `json:"added_schedules"`
	RemovedSchedules  []models.ClassSchedule `json:"removed_schedules"`
	ModifiedSchedules []ScheduleDiff         `json:"modified_schedules"`
}

type ScheduleDiff struct {
	ID uint `json:"id"` // Primary key

	// Not adding start and end field
	// Check diff-03-class.go for more info

	Activity varDiff[Activity] `json:"activity"`
	Method   varDiff[Method]   `json:"method"`

	AddedRooms   []string `json:"added_rooms"`
	RemovedRooms []string `json:"removed_rooms"`
}
