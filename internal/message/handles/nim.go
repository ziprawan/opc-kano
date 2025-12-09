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
			res := db.Where("name % ?", name).Clauses(clause.OrderBy{
				Expression: clause.Expr{SQL: "SIMILARITY(name, ?) DESC", Vars: []any{name}},
			}).Find(&noSpecificNims)
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
		if e := toUint(arg.Content); e != 0 {
			reset()

			nim := nimQuery{}
			nim.lower, nim.upper = findLowerUpperNim(arg.Content)
			queries = append(queries, query{qtype: qnim, nim: nim})
		} else if split := strings.Split(arg.Content, "-"); len(split) == 2 {
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

	builtStr := fmt.Sprintf("Found %d students\nQuery time: %d ms", len(founds), qDiffTime)

	for _, student := range founds {
		builtStr += fmt.Sprintf("\n=====\nName: %s\nNIM: %d\nMajor - Faculty: %s - %s", student.Name, student.Nim, student.Major, student.Faculty)
	}

	c.QuoteReply("%s", builtStr)

	return nil
}
