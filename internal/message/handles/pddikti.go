package handles

import (
	"errors"
	"fmt"
	"kano/internal/utils/messageutil"
	"kano/internal/utils/pddikti"
	"math"
	"strings"
)

func Pddikti(c *messageutil.MessageContext) error {
	query := c.Parser.GetAllJoinedArg()

	res, err := pddikti.Search(query)
	if err != nil {
		if errors.Is(err, pddikti.ErrNoKeyOrIv) {
			c.QuoteReply("This command is not initialized by the owner.\nDebug: Missing PDDIKTI_KEY or PDDIKTI_IV")
		} else {
			c.QuoteReply("Something went wrong\nDebug: %s", err.Error())
		}
		return err
	}

	if res.IsEmpty() {
		c.QuoteReply("Not found")
		return nil
	}

	message := ""
	if len(res.Mahasiswa) != 0 {
		message += "*List Mahasiswa*\n==============\n\n"

		hehe := []string{}
		for i := range int(math.Min(10, float64(len(res.Mahasiswa)))) {
			mhs := res.Mahasiswa[i]
			hehe = append(hehe, fmt.Sprintf("Nama (NIM): %s (%s)\nProdi - PT: %s - %s (%s)", mhs.Nama, mhs.NIM, mhs.NamaProdi, mhs.NamaPT, mhs.SingkatanPT))
		}

		message += strings.Join(hehe, "\n----------\n") + "\n"
	}
	if len(res.Dosen) != 0 {
		message += "*List Dosen*\n==========\n\n"

		hehe := []string{}
		for i := range int(math.Min(10, float64(len(res.Dosen)))) {
			dsn := res.Dosen[i]
			hehe = append(hehe, fmt.Sprintf("Nama - NIDN: %s - %s\nNUPTK: %s\nProdi - PT: %s - %s (%s)", dsn.Nama, dsn.NIDN, dsn.NUPTK, dsn.NamaProdi, dsn.NamaPT, dsn.SingkatanPT))
		}

		message += strings.Join(hehe, "\n----------\n") + "\n"
	}
	if len(res.PT) != 0 {
		message += "*List PT*\n=======\n\n"

		hehe := []string{}
		for i := range int(math.Min(10, float64(len(res.PT)))) {
			pt := res.PT[i]
			hehe = append(hehe, fmt.Sprintf("Nama: %s (%s)\nKode: %s", pt.Nama, pt.NamaSingkat, pt.Kode))
		}

		message += strings.Join(hehe, "\n----------\n") + "\n"
	}
	if len(res.Prodi) != 0 {
		message += "*List Prodi*\n==========\n\n"

		hehe := []string{}
		for i := range int(math.Min(10, float64(len(res.Prodi)))) {
			prodi := res.Prodi[i]
			hehe = append(hehe, fmt.Sprintf("Nama: %s - %s\nPerguruan Tinggi: %s (%s)", prodi.Jenjang, prodi.Nama, prodi.PT, prodi.PTSingkat))
		}

		message += strings.Join(hehe, "\n----------\n") + "\n"
	}

	c.QuoteReply("%s", message)

	return nil
}
