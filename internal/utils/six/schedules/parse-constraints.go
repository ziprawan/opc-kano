package schedules

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var faculties = []string{
	"FMIPA",
	"SITH",
	"SF",
	"FTTM",
	"FITB",
	"FTI",
	"STEI",
	"FTMD",
	"FTSL",
	"SAPPK",
	"FSRD",
	"SBM",
	"SPITM", "SPS",
	"NONFS",
}

func parseConstraints(col *goquery.Selection) (NullConstraint, error) {
	res := NullConstraint{}

	paragraphs := col.Find("p")
	if paragraphs.Length() == 0 {
		return res, nil
	}

	constraint := Constraint{}
	for _, p := range paragraphs.EachIter() {
		childs := p.Contents()
		firstChild := childs.First()

		if strings.HasPrefix(strings.ToLower(text(firstChild)), "batasan:") {
			for i, pChild := range childs.EachIter() {
				if i == 0 {
					continue // Skip "Batasan:" text
				}
				nodeName := goquery.NodeName(pChild)

				// Filter non-text node
				if nodeName != "#text" {
					if nodeName != "br" {
						return res, fmt.Errorf("p %d: unexpected another element besides <br/>: %s", i, nodeName)
					}

					continue
				}

				data := text(pChild)
				if slices.Contains(faculties, data) {
					constraint.Faculties = []string{data}
				} else if strings.HasPrefix(data, "Strata") {
					data = strings.Replace(data, "Strata", "", 1)
					stratas, err := idkWhatIsThis(data, toStrata)
					if err != nil {
						return res, fmt.Errorf("p %d: %s", i, err)
					}
					constraint.Stratas = append(constraint.Stratas, stratas...)
				} else if strings.HasPrefix(data, "Kampus") {
					data = strings.Replace(data, "Kampus", "", 1)
					campuses, err := idkWhatIsThis(data, toCampus)
					if err != nil {
						return res, fmt.Errorf("p %d: %s", i, err)
					}
					constraint.Campuses = append(constraint.Campuses, campuses...)
				} else if strings.HasPrefix(data, "Prodi") {
					data = strings.Replace(data, "Prodi", "", 1)
					majorConstraints, err := idkWhatIsThis(data, dataToMajorConstraint)
					if err != nil {
						return res, fmt.Errorf("p %d: %s", i, err)
					}
					constraint.Majors = append(constraint.Majors, majorConstraints...)
				} else if strings.HasPrefix(data, "Semester") {
					data = strings.Replace(data, "Semester", "", 1)
					semesters, err := idkWhatIsThis(data, func(a string) (uint, error) { b, c := strconv.ParseUint(a, 10, 0); return uint(b), c })
					if err != nil {
						return res, fmt.Errorf("p %d: %s", i, err)
					}
					constraint.Semesters = append(constraint.Semesters, semesters...)
				} else if strings.ContainsRune(data, '/') {
					splits := splitTrim(data, "/")
					for i := range splits {
						if !slices.Contains(faculties, splits[i]) {
							return res, fmt.Errorf("p %d: invalid faculty value: %s", i, splits[i])
						}
					}
					constraint.Faculties = append(constraint.Faculties, splits...)
				} else {
					return res, fmt.Errorf("p %d: unhandled constraint data %s", i, data)
				}
			}
		} else {
			for _, pChild := range childs.EachIter() {
				if !isText(pChild) {
					continue
				}

				constraint.Others = append(constraint.Others, text(pChild))
			}
		}
	}

	res.Constraint = constraint
	res.Valid = true
	return res, nil
}

func idkWhatIsThis[T any](data string, converter func(a string) (T, error)) ([]T, error) {
	splits := splitTrim(data, "/")

	converted := make([]T, len(splits))
	for i := range splits {
		p, err := converter(splits[i])
		if err != nil {
			return nil, fmt.Errorf("invalid value: %q: %s", splits[i], err)
		}
		converted[i] = p
	}

	return converted, nil
}
