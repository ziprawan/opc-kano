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

var PddiktiMan = CommandMan{
	Name: "pddikti - search using pddikti",
	Synopsis: []string{
		"*pddikti* _query_ ...",
	},
	Description: []string{
		"Find information about lecturers, students, departments, or campuses from pddikti (Pangkalan Data Pendidikan Tinggi) site. Found results are limited up to 20 data. The bot will only make a search request to the system and return the value according to what the system returns.",
		"_The system's response to the request was encrypted, but due to a vulnerability on the website, the bot managed to obtain the key to decrypt it. We found no violations of the website's privacy policy regarding this activity. For more information, see https://pddikti.kemdiktisaintek.go.id/privacy-policy._",
		"_query_" +
			"\n{SPACE}Any string query to search in pddikti. The more detailed you search, the fewer results will be returned.",
	},
	SourceFilename: "pddikti.go",
	SeeAlso: []SeeAlso{
		{"https://pddikti.kemdiktisaintek.go.id/", SeeAlsoTypeExternalLink},
	},
}
