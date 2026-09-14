package schedules

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type BasicSemester struct {
	Semester uint `json:"semester"`
	Year     uint `json:"year"`
}

type BasicSubject struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	SKS  uint   `json:"sks"`
}

type Major struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Faculty string `json:"faculty"`
}

type Constraint struct {
	Faculties []string          `json:"faculties"`
	Majors    []MajorConstraint `json:"majors"`
	Stratas   []Strata          `json:"stratas"`
	Campuses  []Campus          `json:"campuses"`
	Semesters []uint            `json:"semesters"`

	Others []string `json:"others"`
}

type MajorConstraint struct {
	// The major ID
	ID uint `json:"id"`
	// Some additional constraint data to the major
	Addon string `json:"addon"`
}

type Strata string

const (
	StrataS1      Strata = "S1"
	StrataS2      Strata = "S2"
	StrataS3      Strata = "S3"
	StrataPR      Strata = "PR"
	StrataUnknown Strata = ""
)

func toStrata(str string) (Strata, error) {
	str = strings.ToLower(str)

	switch str {
	case "s1", "strata s1":
		return StrataS1, nil
	case "s2", "strata s2":
		return StrataS2, nil
	case "s3", "strata s3":
		return StrataS3, nil
	case "pr", "strata pr":
		return StrataPR, nil
	default:
		return StrataUnknown, fmt.Errorf("unknown strata value %s", str)
	}
}

type Campus string

const (
	CampusJatinangor Campus = "JATINANGOR"
	CampusGanesha    Campus = "GANESHA"
	CampusCirebon    Campus = "CIREBON"
	CampusJakarta    Campus = "JAKARTA"
	CampusUnknown    Campus = ""
)

func toCampus(str string) (Campus, error) {
	str = strings.ToLower(str)

	switch str {
	case "jatinangor":
		return CampusJatinangor, nil
	case "ganesha":
		return CampusGanesha, nil
	case "cirebon":
		return CampusCirebon, nil
	case "jakarta":
		return CampusJakarta, nil
	default:
		return CampusUnknown, fmt.Errorf("unknown campus value %s", str)
	}
}

type NullConstraint struct {
	Constraint Constraint
	Valid      bool
}

type Class struct {
	ID          uint           `json:"id"`
	Number      uint           `json:"number"`
	Quota       sql.NullInt32  `json:"quota"`
	Lecturers   []string       `json:"lecturers"`
	Constraints NullConstraint `json:"constraints"`
	Links       ClassLink      `json:"links"`
	Schedules   []Schedule     `json:"schedules"`

	AvailableAtMajorId uint `json:"available_at_major_id"`
}

type ScheduleRow struct {
	Subject Subject `json:"subject"`
	Class   Class   `json:"class"`
}

type Subject struct {
	BasicSubject

	Classes []Class `json:"classes"`
}

type SemesterSubject struct {
	BasicSemester

	Subjects []Subject `json:"subjects"`
	Majors   []Major   `json:"majors"`
}

type ClassLink struct {
	EdunexClassId sql.NullInt32 `json:"edunex_class_id"`
	Teams         string        `json:"teams"`
}

type Schedule struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Rooms    []string  `json:"rooms"`
	Activity Activity  `json:"activity"`
	Method   Method    `json:"method"`
}

type LastColumn struct {
	ClassLink

	ClassId   uint       `json:"class_id"`
	Schedules []Schedule `json:"schedules"`
}

type Activity string

const (
	ActivityLecture  Activity = "LECTURE"
	ActivityTutorial Activity = "TUTORIAL"
	ActivityLabWork  Activity = "LAB_WORK"
	ActivityQuiz     Activity = "QUIZ"
	ActivityMidterm  Activity = "MIDTERM"
	ActivityFinal    Activity = "FINAL"
	ActivityUnknown  Activity = ""
)

func toActivity(str string) (Activity, error) {
	str = strings.ToLower(str)

	switch str {
	case "kuliah":
		return ActivityLecture, nil
	case "tutorial":
		return ActivityTutorial, nil
	case "praktikum":
		return ActivityLabWork, nil
	case "kuis":
		return ActivityQuiz, nil
	case "uts":
		return ActivityMidterm, nil
	case "uas":
		return ActivityFinal, nil
	default:
		return ActivityUnknown, fmt.Errorf("unknown activity value %s", str)
	}
}

type Method string

const (
	MethodInPerson Method = "IN_PERSON"
	MethodOnline   Method = "ONLINE"
	MethodHybrid   Method = "HYBRID"
	MethodUnknown  Method = ""
)

func toMethod(str string) (Method, error) {
	str = strings.ToLower(str)
	switch str {
	case "tatap muka":
		return MethodInPerson, nil
	case "online / e-learning":
		return MethodOnline, nil
	case "hybrid":
		return MethodHybrid, nil
	default:
		return MethodUnknown, fmt.Errorf("unknown method value %s", str)
	}
}
