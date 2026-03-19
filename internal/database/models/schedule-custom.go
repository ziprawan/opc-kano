package models

import (
	"database/sql/driver"
	"fmt"

	"github.com/lib/pq"
)

type StrataArray []Strata

func (a *StrataArray) Scan(src interface{}) error {
	q := pq.StringArray{}
	err := q.Scan(src)
	if err != nil {
		return err
	}

	if q == nil {
		*a = nil
	} else {
		r := make(StrataArray, len(q))
		for i := range q {
			if !isStaraValid(q[i]) {
				return fmt.Errorf("invalid Strata value %s", q[i])
			}
			r[i] = Strata(q[i])
		}
		*a = r
	}

	return nil
}

func (a StrataArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}

	arr := make(pq.StringArray, len(a))
	for i := range a {
		arr[i] = string(a[i])
	}
	return arr.Value()
}

type CampusArray []Campus

func (a *CampusArray) Scan(src interface{}) error {
	q := pq.StringArray{}
	err := q.Scan(src)
	if err != nil {
		return err
	}

	if q == nil {
		*a = nil
	} else {
		r := make(CampusArray, len(q))
		for i := range q {
			if !isCampusValid(q[i]) {
				return fmt.Errorf("invalid Campus value %s", q[i])
			}
			r[i] = Campus(q[i])
		}
		*a = r
	}

	return nil
}

func (a CampusArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}

	arr := make(pq.StringArray, len(a))
	for i := range a {
		arr[i] = string(a[i])
	}
	return arr.Value()
}

type NullSubjectCategory struct {
	SubjectCategory SubjectCategory
	Valid           bool
}

func (sc *NullSubjectCategory) Scan(value any) error {
	s, ok := value.(string)
	if ok {
		q := SubjectCategory(s)
		switch q {
		case SubjectCategoryInternship:
		case SubjectCategoryLecture:
		case SubjectCategoryResearch:
		case SubjectCategoryThesis:
		case SubjectCategoryUnknown:
			break
		default:
			return fmt.Errorf("invalid subject category value")
		}

		n := NullSubjectCategory{
			SubjectCategory: q,
			Valid:           true,
		}
		if q == SubjectCategoryUnknown {
			n.Valid = false
		}

		*sc = n
		return nil
	}
	return fmt.Errorf("given value is not a string")
}

func (sc NullSubjectCategory) Value() (driver.Value, error) {
	if sc.Valid {
		return string(sc.SubjectCategory), nil
	} else {
		return nil, nil
	}
}
