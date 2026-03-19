package handles

import (
	"fmt"
	"kano/internal/database"
	"kano/internal/database/models"
	"kano/internal/utils/messageutil"
	"slices"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm/clause"
)

const NIM_LENGTH = 8

type nimQuery struct {
	upper uint
	lower uint
}

type qtype uint

const (
	qnim qtype = iota
	qname
)

type query struct {
	qtype qtype
	nim   nimQuery
	name  string
}

func searchStudents(queries []query) ([]models.Student, error) {
	db := database.GetInstance()
	db = db.Debug()

	if len(queries) == 0 {
		return nil, nil
	}

	nims := []nimQuery{}
	var results []models.Student
	var addedIds []uint

	appendResults := func(adds []models.Student) {
		for _, add := range adds {
			if !slices.Contains(addedIds, add.Id) {
				results = append(results, add)
			}
		}
	}

	for _, query := range queries {
		if query.qtype == qnim {
			nim := query.nim
			nims = append(nims, query.nim)

			var founds []models.Student
			res := db.Where("nim >= ? AND nim < ?", nim.lower, nim.upper).Order("nim ASC").Find(&founds)
			if err := res.Error; err != nil {
				return nil, err
			}
			appendResults(founds)
		} else if name := strings.TrimSpace(query.name); query.qtype == qname && name != "" {
			for _, nim := range nims {
				var specificNims []models.Student
				res := db.Where("nim >= ? AND nim < ?", nim.lower, nim.upper).Where("name % ?", name).Order("nim ASC").Find(&specificNims)
				if err := res.Error; err != nil {
					return nil, err
				}
				appendResults(specificNims)
			}

			var noSpecificNims []models.Student
			res := db.
				Where("name % ?", name).
				Or("custom_name % ?", name).
				Clauses(clause.OrderBy{
					Expression: clause.Expr{SQL: "SIMILARITY(custom_name, ?) DESC", Vars: []any{name}},
				}).
				Clauses(clause.OrderBy{
					Expression: clause.Expr{SQL: "SIMILARITY(name, ?) DESC", Vars: []any{name}},
				}).
				Find(&noSpecificNims)
			if err := res.Error; err != nil {
				return nil, err
			}
			appendResults(noSpecificNims)
		}
	}

	return results, nil
}

func toUint(str string) uint {
	con, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0
	} else {
		return uint(con)
	}
}

// I assume right and left are convertable to uint
func findLowerUpperNim(right string, left ...string) (uint, uint) {
	tmpRight := right

	var tmpLeft string
	if len(left) > 0 {
		tmpLeft = left[0]
	} else {
		tmpLeft = strconv.FormatUint(uint64(toUint(tmpRight)+1), 10)
	}

	// Decrease left and right first
	if len(tmpLeft) > NIM_LENGTH {
		tmpLeft = tmpLeft[:NIM_LENGTH]
	}
	if len(tmpRight) > NIM_LENGTH {
		tmpRight = tmpRight[:NIM_LENGTH]
	}

	// Kalau right lebih pendek dari left
	// Ambil angka awal dari left untuk memenuhi panjang yang sama
	if len(tmpRight) < len(tmpLeft) {
		// Misal 12345-6
		// Berarti ambil 1234 terus dari di kirinya 6, jadinya 12346
		// => 12345-12346
		tmpRight = tmpLeft[:len(tmpLeft)-len(tmpRight)] + tmpRight
	}

	// Increase left and right
	if len(tmpLeft) < NIM_LENGTH {
		tmpLeft += strings.Repeat("0", NIM_LENGTH-len(tmpLeft))
	}
	if len(tmpRight) < NIM_LENGTH {
		tmpRight += strings.Repeat("0", NIM_LENGTH-len(tmpRight))
	}

	if tmpRight < tmpLeft {
		tmpLeft, tmpRight = tmpRight, tmpLeft
	}

	return toUint(tmpLeft), toUint(tmpRight)
}

func Nim(c *messageutil.MessageContext) error {
	first, last := -1, -1
	queries := []query{}

	reset := func() {
		if first == -1 || last == -1 {
			return
		}

		name := c.Parser.GetJoinedArg(first, last)
		queries = append(queries, query{qtype: qname, name: name})
		first, last = -1, -1
	}

	for idx, arg := range c.Parser.Args {
		if e := toUint(arg.Content.Data); e != 0 {
			reset()

			nim := nimQuery{}
			nim.lower, nim.upper = findLowerUpperNim(arg.Content.Data)
			queries = append(queries, query{qtype: qnim, nim: nim})
		} else if split := strings.Split(arg.Content.Data, "-"); len(split) == 2 {
			if toUint(split[0]) != 0 && toUint(split[1]) != 0 {
				reset()

				nim := nimQuery{}
				nim.lower, nim.upper = findLowerUpperNim(split[1], split[0])
				queries = append(queries, query{qtype: qnim, nim: nim})
			}
		} else {
			if first == -1 {
				first = idx
				last = idx
			} else {
				last = idx
			}
		}
	}
	reset()

	qStartTime := time.Now().UnixMilli()
	founds, err := searchStudents(queries)
	qDiffTime := time.Now().UnixMilli() - qStartTime
	if err != nil {
		c.QuoteReply("Failed to query to database: %s\nQuery time: %d ms", err.Error(), qDiffTime)
		return nil
	}

	var builtStr strings.Builder
	fmt.Fprintf(&builtStr, "Found %d students\nQuery time: %d ms", len(founds), qDiffTime)

	for i, student := range founds {
		if i > 100 {
			break
		}
		fmt.Fprintf(&builtStr, "\n=====\nName: %s\nNIM: %d\nMajor - Faculty: %s - %s", student.Name, student.Nim, student.Major, student.Faculty)
	}

	c.QuoteReply("%s", builtStr.String())

	return nil
}

var NimHelp = CommandMan{
	Name: "nim - search for some ITB student info",
	Synopsis: []string{
		"*nim* _nim number_ ...",
		"*nim* _student name_ ...",
	},
	Description: []string{
		"Find a little info about some students at ITB (Institut Teknologi Bandung). These data are sourced from Edunex site Search results are limited to 100 student records. " +
			"The info that will appear as results includes:" +
			"\n- Name" +
			"\n- Student ID (NIM)" +
			"\n- Faculty and Major",
		"ITB's student ID number has an 8-digit format convention. The first digit indicates the level of study (bachelor's, master's, or doctoral), the next two digits (digits 2-3) indicate the major taken, the next two digits (digits 4-5) indicate the year of commencement, and the last three digits (digits 6-8) are unique numbers sequentially starting from 1. In special cases, students who enter through international channels have the last three student ID numbers that are different from the others.",
		"If the argument is a positive number, the search will be performed based on matching the student ID number with the given number. Searching using the student ID number has the following simple rules:" +
			"\n- If the given number is exactly 8 digits, the returned result is guaranteed to only find one student with the same student ID number, otherwise, it will return 0 students." +
			"\n- If the given number is more than 8 digits, the search will only retrieve the first 8 digits of the given number, and the same procedure will be applied for searches with 8-digit numbers." +
			"\n- If the given number is less than 8 digits, the returned result will be all students whose student ID number begins with the given number.",
		"Searching using Student ID supports ranges with the condition: in one given argument, there are at least two numbers separated by HYPEN MINUS U+002D (-), where the number on the left is called the *starting number* and the number on the right is called the *ending number*. For example: `10124001-10124050`, where `10124001` is the *starting number* and `10124050` is the *ending number*. The number rules are still the same as the number argument rules, but with the following additions:" +
			"\n- If the *starting number* has the same length as the *ending number* and the value of the *ending number* is smaller than the value of the *starting number*, the automatic search will swap the positions of the *ending number* and *starting number*. Example: `10124050-10124001` will be treated as `10124001-10124050`" +
			"\n- If the *ending number* has smaller length than the *ending number*, then the *ending number* will be prefixed like the *starting number* until it has the same length as the *starting number*. Example: `10120-4` will be treated as `10120-10124`, which will be treated as `10120000-10124000`." +
			"\n- If the *starting number* has smaller length than the *ending number*, the *starting number* will be multiplied by 10 until it has the same length as the *ending number*. Example: `1-20` will be treated as `10-20`, which will be treated as `10000000-20000000`.",
		"If the given argument cannot be converted to a number, the bot will assume a name search. Name searches are sorted by how well the given name matches the names in the database using a trigram search algorithm.",
	},
	SourceFilename: "nim.go",
	SeeAlso: []SeeAlso{
		{"https://edunex.itb.ac.id/messages/", SeeAlsoTypeExternalLink},
	},
}
