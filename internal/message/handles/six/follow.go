package six

import (
	"errors"
	"fmt"
	"kano/internal/database/models"
	"kano/internal/utils/messageutil"

	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
)

func followHandler(c *messageutil.MessageContext) error {
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
		followHelp(c)
		return nil
	}

	classCtx := args[1].Content.Data
	classCode, classNum, err := parseClassCtx(classCtx)
	if err != nil {
		c.QuoteReply("Format kelas salah: %s. Contoh yang benar: `ET1201-01`", err)
		return nil
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
			c.QuoteReply("Kesalahan internal, harap segera lapor pemilik bot.\nInfo tambahan: %s", err)
			return err
		}
	}

	toInsert := models.ClassFollower{
		Jid:            jid,
		SubjectClassID: foundSubjectClass.ID,
	}

	tx = db.
		Where(
			"jid = ? AND subject_class_id = ?",
			toInsert.Jid, toInsert.SubjectClassID,
		).
		Attrs(toInsert).
		FirstOrCreate(&toInsert)
	if tx.Error != nil {
		c.QuoteReply("Gagal menambahkan status mengikuti kelas: Kesalahan internal, harap segera lapor pemilik bot.\nInfo tambahan: %s", tx.Error)
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		c.QuoteReply("Sudah pernah mengikuti %s-%02d (%s).", classCode, classNum, foundSubjectClass.Subject.Name)
	} else {
		c.QuoteReply("Berhasil mengikuti %s-%02d (%s).", classCode, classNum, foundSubjectClass.Subject.Name)
	}

	return nil
}
