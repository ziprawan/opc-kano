package six

import (
	"errors"
	"fmt"
	"kano/internal/database/models"
	"kano/internal/utils/messageutil"

	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

const OFFSET_MAX = 10080

func reminderHandler(c *messageutil.MessageContext) error {
	jid := c.GetChat()
	if jid.Server == types.DefaultUserServer {
		c.QuoteReply("Gagal mengambil ID pengguna %q", jid)
		return fmt.Errorf("unable to resolve sender jid: %s", jid)
	}
	if jid.Server != types.HiddenUserServer {
		c.QuoteReply("Lakukan di private chat.")
		return nil
	}

	args := c.Parser.Args
	if len(args) == 1 {
		reminderHelp(c)
		return nil
	}

	classCtx := args[1].Content.Data
	classCode, classNum, err := parseClassCtx(classCtx)
	if err != nil {
		c.QuoteReply("Format kelas salah: %s. Contoh yang benar: `ET1201-01`", err)
		return nil
	}

	offset := 0
	anchorAtEnd := false
	if len(args) == 3 {
		offsetStr := args[2].Content.Data
		offset, anchorAtEnd, err = parseOffset(offsetStr)
		if err != nil {
			c.QuoteReply("Format offset salah: %s. Contoh yang benar: `+10`, `^-20`, `-30m`, `^`", err)
			return nil
		}
		if offset > OFFSET_MAX || offset < -OFFSET_MAX {
			c.QuoteReply(
				"Nilai offset terlalu besar atau kecil. Paling kecil -1 pekan (-%d menit) dan paling besar 1 pekan (%d menit). Didapat: %d menit",
				OFFSET_MAX, OFFSET_MAX, offset,
			)
			return nil
		}
	}

	foundSubjectClass := models.SubjectClass{
		Subject: models.Subject{Code: classCode},
		Number:  classNum,
	}
	stmt := db.
		Model(&models.SubjectClass{}).
		InnerJoins("Subject").
		Where("number = ? AND code = ?", classNum, classCode)

	tx := stmt.First(&foundSubjectClass)
	if err = tx.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.QuoteReply("Tidak dapat menemukan matkul %s kelas %02d. Jika ini merupakan kesalahan, coba hubungi pemilik bot.", classCode, classNum)
			return nil
		} else {
			c.QuoteReply("Gagal mengambil ID kelas: Kesalahan internal, harap segera lapor pemilik bot.\nInfo tambahan: %s", err)
			return err
		}
	}

	toInsert := models.ClassReminder{
		Jid:            jid,
		SubjectClassID: foundSubjectClass.ID,
		AnchorAtEnd:    anchorAtEnd,
		OffsetMinutes:  offset,
	}

	tx = db.
		Where(
			"jid = ? AND subject_class_id = ? AND anchor_at_end = ? AND offset_minutes = ?",
			toInsert.Jid, toInsert.SubjectClassID, toInsert.AnchorAtEnd, toInsert.OffsetMinutes,
		).
		Attrs(toInsert).
		FirstOrCreate(&toInsert)
	if tx.Error != nil {
		c.QuoteReply("Gagal menambahkan reminder kelas: Kesalahan internal, harap segera lapor pemilik bot.\nInfo tambahan: %s", tx.Error)
		return tx.Error
	}

	ctx := fmt.Sprintf("%d menit %s kelas %s-%02d (%s) %s", abs(offset), afterBefore(offset), classCode, classNum, foundSubjectClass.Subject.Name, anchor(anchorAtEnd))

	if tx.RowsAffected == 0 {
		c.QuoteReply("Pengingat %q sudah pernah ditambahkan.", ctx)
	} else {
		c.QuoteReply("Berhasil menambahkan pengingat %q.", ctx)
	}

	return nil
}

func afterBefore(offset int) string {
	if offset < 0 {
		return "sebelum"
	} else if offset > 0 {
		return "setelah"
	} else {
		return "saat"
	}
}

func anchor(anchorAtEnd bool) string {
	if anchorAtEnd {
		return "berakhir"
	} else {
		return "dimulai"
	}
}

func abs(e int) int {
	if e < 0 {
		return -e
	}
	return e
}
