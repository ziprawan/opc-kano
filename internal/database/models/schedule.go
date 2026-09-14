package models

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Major struct {
	ID      uint   `gorm:"primaryKey;not null"`
	Name    string `gorm:"not null"`
	Faculty string `gorm:"not null"`

	// SubjectClasses   []SubjectClass
	// ConstraintMajors []ConstraintMajor
}

func (_ Major) TableName() string {
	return "major"
}

type Curricula struct {
	Year uint `gorm:"primaryKey;not null"`

	Subjects []Subject
}

func (_ Curricula) TableName() string {
	return "curricula"
}

type Semester struct {
	ID       uint `gorm:"primaryKey;not null"`
	Year     uint `gorm:"not null"`
	Semester uint `gorm:"not null"`

	Start time.Time `gorm:"autoCreateTime"`
	End   time.Time `gorm:"autoCreateTime"`
}

func (_ Semester) TableName() string {
	return "semester"
}

type Subject struct {
	ID   uint   `gorm:"primaryKey;not null"`
	Code string `gorm:"not null"`
	Name string `gorm:"not null"`
	SKS  uint   `gorm:"not null;column:sks"`

	Category NullSubjectCategory

	CurriculaYear uint      `gorm:"not null"`
	Curricula     Curricula `gorm:"foreignKey:CurriculaYear;references:Year"`
}

func (_ Subject) TableName() string {
	return "subject"
}

type Lecturer struct {
	ID   uint   `gorm:"primaryKey;autoIncrement;not null"`
	Name string `gorm:"unique;not null"`

	Classes []SubjectClass `gorm:"many2many:lecturer_in_class"`
}

func (_ Lecturer) TableName() string {
	return "lecturer"
}

type LecturerInClass struct {
	LecturerID uint      `gorm:"primaryKey"`
	Lecturer   *Lecturer `gorm:"foreignKey:LecturerID"`

	SubjectClassID uint          `gorm:"primaryKey"`
	SubjectClass   *SubjectClass `gorm:"foreignKey:SubjectClassID"`
}

func (LecturerInClass) TableName() string {
	return "lecturer_in_class"
}

type SubjectClass struct {
	ID            uint          `gorm:"primaryKey;not null"`
	Number        uint          `gorm:"not null"`
	Quota         sql.NullInt32 `gorm:"not null"`
	EdunexClassID sql.NullInt32
	TeamsLink     sql.NullString

	AvailableAtMajorID uint  `gorm:"not null;column:major_id"` // Yh, biar ga panjang namanya
	AvailableAtMajor   Major `gorm:"foreignKey:AvailableAtMajorID;references:ID"`

	SubjectID  uint     `gorm:"not null"`
	Subject    Subject  `gorm:"foreignKey:SubjectID;references:ID"`
	SemesterID uint     `gorm:"not null"`
	Semester   Semester `gorm:"foreignKey:SemesterID;references:ID"`

	Constraint *ClassConstraint `gorm:"foreignKey:ID;references:SubjectClassID"`

	Lecturers []Lecturer `gorm:"many2many:lecturer_in_class"`
	Schedules []ClassSchedule
}

func (_ SubjectClass) TableName() string {
	return "subject_class"
}

type ClassConstraint struct {
	ID uint `gorm:"primaryKey;autoIncrement;not null"`

	Other     pq.StringArray    `gorm:"not null;type:text[]"`
	Faculties pq.StringArray    `gorm:"not null;type:text[]"`
	Majors    []ConstraintMajor `gorm:"many2many:major_in_constraint"`
	Stratas   StrataArray       `gorm:"not null;type:strata[]"`
	Campuses  CampusArray       `gorm:"not null;type:campus[]"`
	Semester  pq.Int32Array     `gorm:"not null;type:integer[]"`

	SubjectClassID uint         `gorm:"not null;unique"`
	SubjectClass   SubjectClass `gorm:"foreignKey:SubjectClassID;references:ID"`
}

func (_ ClassConstraint) TableName() string {
	return "class_constraint"
}

type ConstraintMajor struct {
	ID        uint   `gorm:"primaryKey;autoIncrement;not null"`
	AddonData string `gorm:"not null"`

	MajorID sql.NullInt32
	Major   *Major `gorm:"foreignKey:MajorID;references:ID"`

	ClassConstraints []ClassConstraint `gorm:"many2many:major_in_constraint"`
}

func (_ ConstraintMajor) TableName() string {
	return "constraint_major"
}

type MajorInConsraint struct {
	ClassConstraintID uint             `gorm:"primaryKey"`
	ClassConstraint   *ClassConstraint `gorm:"foreignKey:ClassConstraintID"`
	ConstraintMajorID uint             `gorm:"primaryKey"`
	ConstraintMajor   *ConstraintMajor `gorm:"foreignKey:ConstraintMajorID"`
}

func (_ MajorInConsraint) TableName() string {
	return "major_in_constraint"
}

type Room struct {
	ID   uint   `gorm:"primaryKey;autoincrement;not null"`
	Name string `gorm:"unique;not null"`

	Classes []ClassSchedule `gorm:"many2many:room_in_class_schedule"`
}

func (_ Room) TableName() string {
	return "room"
}

type RoomInClass struct {
	RoomID          uint  `gorm:"primaryKey"`
	Room            *Room `gorm:"foreignKey:RoomID"`
	ClassScheduleID uint  `gorm:"primaryKey"`
	ClassSchedule   *Room `gorm:"foreignKey:ClassScheduleID"`
}

func (_ RoomInClass) TableName() string {
	return "room_in_class"
}

type ClassSchedule struct {
	ID uint `gorm:"primaryKey;autoIncrement;not null"`

	Start    time.Time `gorm:"type:timestamptz"`
	End      time.Time `gorm:"type:timestamptz"`
	Activity ScheduleActivity
	Method   ScheduleMethod
	Rooms    []Room `gorm:"many2many:room_in_class"`

	UnixStart uint
	UnixEnd   uint

	SubjectClassID uint          `gorm:"not null"`
	SubjectClass   *SubjectClass `gorm:"foreignKey:SubjectClassID;references:ID"`
}

func (_ ClassSchedule) TableName() string {
	return "class_schedule"
}

type SubjectCategory string

const (
	SubjectCategoryLecture    SubjectCategory = "LECTURE"
	SubjectCategoryResearch   SubjectCategory = "RESEARCH"
	SubjectCategoryInternship SubjectCategory = "INTERNSHIP"
	SubjectCategoryThesis     SubjectCategory = "THESIS"
	SubjectCategoryUnknown    SubjectCategory = ""
)

type ScheduleActivity string

const (
	ScheduleActivityLecture  ScheduleActivity = "LECTURE"
	ScheduleActivityTutorial ScheduleActivity = "TUTORIAL"
	ScheduleActivityQuiz     ScheduleActivity = "QUIZ"
	ScheduleActivityMidterm  ScheduleActivity = "MIDTERM"
	ScheduleActivityFinal    ScheduleActivity = "FINAL"
)

type ScheduleMethod string

const (
	ScheduleMethodInPerson ScheduleMethod = "IN_PERSON"
	ScheduleMethodOnline   ScheduleMethod = "ONLINE"
	ScheduleMethodHybrid   ScheduleMethod = "HYBRID"
)

type Strata string

const (
	StrataS1 Strata = "S1"
	StrataS2 Strata = "S2"
	StrataS3 Strata = "S3"
	StrataPR Strata = "PR"
)

func isStaraValid(str string) bool {
	s := Strata(str)
	return s == StrataS1 || s == StrataS2 || s == StrataS3 || s == StrataPR
}

type Campus string

const (
	CampusJatinangor Campus = "JATINANGOR"
	CampusGanesha    Campus = "GANESHA"
	CampusCirebon    Campus = "CIREBON"
	CampusJakarta    Campus = "JAKARTA"
)

func isCampusValid(str string) bool {
	c := Campus(str)
	return c == CampusJatinangor || c == CampusGanesha || c == CampusCirebon || c == CampusJakarta
}
