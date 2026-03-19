package handles

import (
	"encoding/json"
	"fmt"
	"kano/internal/config"
	"kano/internal/database"
	"kano/internal/database/models"
	"kano/internal/utils/messageutil"
	"kano/internal/utils/word"
	"strconv"
	"strings"

	"go.mau.fi/whatsmeow/types"
)

func Matkul(ctx *messageutil.MessageContext) error {
	db := database.GetInstance()
	args := ctx.Parser.Args
	if len(args) == 0 {
		ctx.QuoteReply("Berikan nama atau kode matkul (Contoh: elektromagnetika, ET2202)")
		return nil
	}
	query := ctx.Parser.GetAllOriginalArg()
	useCode := false
	if len(query) == 6 {
		query = strings.ToUpper(query)
		_, err := strconv.ParseUint(query[2:], 10, 0)
		useCode = word.IsCharUpper(query[0]) && word.IsCharUpper(query[1]) && err == nil
	}

	var subjectScheds []models.SubjectSchedule
	q := db.
		Table("subject_schedule AS ss").
		Joins(`INNER JOIN "subject" s ON s.id = ss."subject_id"`).
		Where(`ss."semester_id" = ?`, 2)

	// conditional part
	if useCode {
		fmt.Println("Pakek kode:", query)
		q = q.Where(`s."code" = ?`, query)
	} else {
		q = q.Where(`s."name" ILIKE ?`, "%"+query+"%")
	}

	q = q.
		Preload("Subject").
		Preload("SubjectClasses.Lecturers.Lecturer").
		Preload("SubjectClasses.AvailableAtMajor")

	tx := q.Find(&subjectScheds)
	if tx.Error != nil {
		ctx.QuoteReply("Internal Error: %s", tx.Error)
		return nil
	}

	if ctx.IsSenderSame(config.GetConfig().OwnerJID) && ctx.GetChat().Server != types.GroupServer {
		mar, _ := json.MarshalIndent(subjectScheds, "", " ")
		ctx.QuoteReply("%s", mar)
	}

	if len(subjectScheds) == 0 {
		ctx.QuoteReply("Tidak ada matkul yang dapat ditemukan")
		return nil
	}

	var res strings.Builder
	for i, subjSched := range subjectScheds {
		fmt.Fprintf(&res,
			"*%s - %s*\n",
			subjSched.Subject.Code, subjSched.Subject.Name,
		)

		if len(subjSched.SubjectClasses) == 0 {
			res.WriteString("- Tidak ada kelas yang dibuka.")
			continue
		}

		for _, class := range subjSched.SubjectClasses {
			lecturers := make([]string, len(class.Lecturers))
			for i, lec := range class.Lecturers {
				lecturers[i] = lec.Lecturer.Name
			}

			var lecStr string
			if len(lecturers) > 0 {
				lecStr = strings.Join(lecturers, ", ")
			} else {
				lecStr = "Somehow ga ada (special case)"
			}

			ket := "(Tidak dibuka)"
			if class.Quota.Valid {
				ket = ""
			}

			fmt.Fprintf(&res,
				"*Kelas %d:*\n- Kuota: %d %s\n- Dosen: %s\n- Ambil di: (%d) %s - %s\n",
				class.Number, class.Quota.Int32, ket, lecStr,
				class.AvailableAtMajor.ID, class.AvailableAtMajor.Faculty, class.AvailableAtMajor.Name,
			)
		}

		if i != len(subjectScheds)-1 {
			res.WriteRune('\n')
		}
	}

	ctx.QuoteReply("Ketemu %d matkul:\n\n%s", len(subjectScheds), res.String())

	return nil
}
