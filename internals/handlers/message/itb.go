package message

import (
	"database/sql"
	"errors"
	"fmt"
	"kano/internals/database"
	"math"
	"strconv"
	"strings"
	"time"
)

var NimMan = CommandMan{
	Name:     "nim - Cari mahasiswa ITB",
	Synopsis: []string{"nim NIM_OR_NAME ...", "nim NIM_OR_NAME ... [PAGE_NUMBER]"},
	Description: []string{
		"Menampilkan daftar informasi sederhana mahasiswa ITB yang ditemukan berdasarkan pencarian NIM atau nama mahasiswa. Daftar hasil yang ditemukan akan diurutkan berdasarkan angka NIM nya.",
		"*NIM_OR_NAME* (Wajib)\n{SPACE}NIM atau nama mahasiswa yang ingin dicari. ",
		"*PAGE_NUMBER*\n{SPACE}Nomor halaman daftar mahasiswa yang ditemukan. Argumen ini hanya dianggap ada ketika argumen NIM_OR_NAME terpenuhi dan argumen paling akhir berupa bilangan bulat selain nol. Gunakan bilangan negatif untuk mengurutkan hasil dari belakang.",
	},

	SeeAlso: []SeeAlso{
		{Content: "pddikti", Type: SeeAlsoTypeCommand},
	},
	Source: "itb.go",
}

const QUERY_NIM_MAHASISWA = `SELECT
	"s"."name" AS "nama",
	"s"."nim" AS "NIM",
	"m"."name" AS "jurusan",
	"f"."name" AS "fakultas"
FROM
	"student" "s"
	INNER JOIN "major" "m" ON "m"."id" = "s"."major_id"
	INNER JOIN "faculty" "f" ON "f"."id" = "m"."faculty_id"
WHERE
	"s"."nim" >= $1 AND "s"."nim" < $2
ORDER BY
	"s"."nim" %s
LIMIT
	10
OFFSET
	$3`
const QUERY_TOTAL_MAHASISWA_BY_NIM = `SELECT
	COUNT(*)
FROM
	"student" "s"
WHERE
	"s"."nim" >= $1 AND "s"."nim" < $2`

const QUERY_NAMA_MAHASISWA = `SELECT
	"s"."name" AS "nama",
	"s"."nim" AS "NIM",
	"m"."name" AS "jurusan",
	"f"."name" AS "fakultas"
FROM
	"student" "s"
	INNER JOIN "major" "m" ON "m"."id" = "s"."major_id"
	INNER JOIN "faculty" "f" ON "f"."id" = "m"."faculty_id"
WHERE
	"s"."name" ILIKE $1
ORDER BY
	"s"."nim" %s
LIMIT
	10
OFFSET
	$2`
const QUERY_TOTAL_MAHASISWA_BY_NAMA = `SELECT
	COUNT(*)
FROM
	"student" "s"
WHERE
	"s"."name" ILIKE $1`

type MahasiswaITB struct {
	Nama     string
	NIM      int
	Jurusan  string
	Fakultas string
}

type MahasiswaFoundResult struct {
	Mahasiswa  []MahasiswaITB
	TotalCount uint
}

func filterString(inp string) string {
	newStr := ""
	for _, char := range inp {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == ' ' {
			newStr += string(char)
		}
	}
	return strings.TrimSpace(newStr)
}

func cariMahasiswa(query string, page int) (res MahasiswaFoundResult, errMsg string) {
	var offset int = 0
	sortType := "ASC"
	if page < 0 {
		sortType = "DESC"
		offset = ((-page) - 1) * 10
	} else if page > 0 {
		offset = (page - 1) * 10
	}
	db := database.GetDB()
	nim, err := strconv.ParseUint(query, 10, 0)
	if err != nil {
		filt := fmt.Sprintf("%%%s%%", filterString(query))
		fmt.Printf("[NIM] Entered string query, query: %s - %s - %s, offset: %d\n", query, filterString(query), filt, offset)
		if len(filt) == 2 {
			errMsg = "Berikan NIM atau nama mahasiswa"
			return
		}

		dbQuery := fmt.Sprintf(QUERY_NAMA_MAHASISWA, sortType)
		rows, err := db.Query(dbQuery, filt, offset)
		if err != nil {
			fmt.Println(err)
			errMsg = "Terjadi kesalahan saat membuat kueri ke database"
			return
		}
		defer rows.Close()

		for rows.Next() {
			var mahasiswa MahasiswaITB
			err := rows.Scan(&mahasiswa.Nama, &mahasiswa.NIM, &mahasiswa.Jurusan, &mahasiswa.Fakultas)
			if err != nil {
				fmt.Println(err)
				errMsg = "Terjadi kesalahan saat mengolah data"
				return
			}
			res.Mahasiswa = append(res.Mahasiswa, mahasiswa)
		}

		err = db.QueryRow(QUERY_TOTAL_MAHASISWA_BY_NAMA, filt).Scan(&res.TotalCount)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				fmt.Println(err)
				errMsg = "Terjadi kesalahan saat menghitung total data yang ditemukan"
				return
			}
		}
	} else {
		fmt.Printf("[NIM] Entered number query")
		nimLength := len(query)
		if nimLength > 8 || nim == 0 {
			errMsg = "Panjang NIM yang diizinkan ada di rentang 1 sampai 8 digit saja."
			if nim == 0 {
				errMsg += " (Jangan dikasih 0 juga dong...)"
			}
			return
		}

		multiplier := math.Pow10(8 - nimLength)
		min := nim * uint64(multiplier)
		max := (nim + 1) * uint64(multiplier)
		dbQuery := fmt.Sprintf(QUERY_NIM_MAHASISWA, sortType)
		rows, err := db.Query(dbQuery, min, max, offset)
		if err != nil {
			fmt.Println(err)
			errMsg = "Terjadi kesalahan saat membuat kueri ke database"
			return
		}
		defer rows.Close()

		for rows.Next() {
			var mahasiswa MahasiswaITB
			err := rows.Scan(&mahasiswa.Nama, &mahasiswa.NIM, &mahasiswa.Jurusan, &mahasiswa.Fakultas)
			if err != nil {
				fmt.Println(err)
				errMsg = "Terjadi kesalahan saat mengolah data"
				return
			}
			res.Mahasiswa = append(res.Mahasiswa, mahasiswa)
		}

		err = db.QueryRow(QUERY_TOTAL_MAHASISWA_BY_NIM, min, max).Scan(&res.TotalCount)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				fmt.Println(err)
				errMsg = "Terjadi kesalahan saat menghitung total data yang ditemukan"
				return
			}
		}
	}

	return
}

func NIMHandler(ctx *MessageContext) {
	var page int64 = 0
	args := ctx.Parser.GetArgs()
	query := ""
	if len(args) == 0 {
		ctx.Instance.Reply("Berikan NIM atau nama mahasiswa", true)
		return
	}

	if len(args) > 1 {
		lastArg := args[len(args)-1].Content
		parsed, err := strconv.ParseInt(lastArg, 10, 0)
		if err == nil {
			page = parsed
			query = ctx.Parser.Text[args[0].Start : args[len(args)-2].End+1]
		} else {
			query = ctx.Parser.Text[args[0].Start:]
		}
	} else {
		query = ctx.Parser.Text[args[0].Start:]
	}

	fmt.Printf("[NIM] got query: %s, offset: %d\n", query, page)
	start := time.Now()
	searchResult, errMsg := cariMahasiswa(query, int(page))
	queryTimeMilli := float64(time.Since(start).Microseconds()) / 1000

	if errMsg != "" {
		errMsg += fmt.Sprintf("\nQuery time: %.3f ms", queryTimeMilli)
		ctx.Instance.Reply(errMsg, true)
	} else {
		addons := fmt.Sprintf("\nQuery time: %.3f ms", queryTimeMilli)
		if len(searchResult.Mahasiswa) == 0 {
			ctx.Instance.Reply("Data tidak ditemukan"+addons, true)
		} else {
			var msgs []string
			for _, mahasiswa := range searchResult.Mahasiswa {
				msg := fmt.Sprintf("Nama: %s\n", mahasiswa.Nama)
				msg += fmt.Sprintf("NIM: %d\n", mahasiswa.NIM)
				msg += fmt.Sprintf("Jurusan: %s\n", mahasiswa.Jurusan)
				msg += fmt.Sprintf("Fakultas: %s", mahasiswa.Fakultas)
				msgs = append(msgs, msg)
			}
			currPage := page
			totalPage := uint64(math.Ceil(float64(searchResult.TotalCount) / 10))
			if currPage == 0 {
				currPage = 1
			} else if currPage < 0 {
				currPage = int64(totalPage) + currPage + 1
			}
			ctx.Instance.Reply(fmt.Sprintf("Ketemu: %d (Halaman %d dari %d)%s\n\n%s", searchResult.TotalCount, currPage, totalPage, addons, strings.Join(msgs, "\n======\n")), true)
		}
	}
}
