package message

import (
	"fmt"
	projectconfig "kano/internals/project_config"
	"kano/internals/utils/kanoutils"
	"net/url"
	"slices"
	"strings"
	"time"
)

func generateDetail(id, category string) string {
	str := ""
	layout := "2006-01-02T15:04:05Z"
	dateFormat := "Monday, 02 January 2006"

	switch category {
	case "dosen":
		detail, err := kanoutils.GetPNSDetails(id)
		if err != nil {
			return err.Error()
		}

		profile := detail.Profile

		str = "*Detail Dosen*\n\n"

		// Profile
		str += fmt.Sprintf("Nama: %s (%s)\nPendidikan Terakhir: %s\n", profile.NamaDosen, profile.JenisKelamin, profile.PendidikanTertinggi)
		str += fmt.Sprintf("PT - Prodi: %s - %s\n", profile.NamaPT, profile.NamaProdi)
		str += fmt.Sprintf("Jabatan: %s\n", profile.JabatanAkademik)
		str += fmt.Sprintf("Status Ikatan Kerja: %s (%s)\n", profile.StatusIkatanKerja, profile.StatusAktivitas)

		// Study Histories
		str += "\n*Riwayat Pendidikan Dosen*\n\n"
		for idx, study := range detail.StudyHistories {
			year := fmt.Sprintf("%d", study.TahunLulus)
			if study.TahunMasuk != 0 {
				year = fmt.Sprintf("%d - %d", study.TahunMasuk, study.TahunLulus)
			}

			str += fmt.Sprintf("%d. %s %s (%s - %s) di %s (%s)\n", idx+1, study.Jenjang, study.NamaProdi, study.GelarAkademik, study.SingkatanGelar, study.NamaPT, year)
		}

		// Teaching Histories
		str += "\n*Riwayat Mengajar Dosen*\n\n"
		for semester, teachings := range detail.TeachingHistories {
			str += fmt.Sprintf("%s\n", semester)

			for idx, t := range teachings {
				str += fmt.Sprintf("%d. %s (%s ~ %s) di %s\n", idx+1, t.NamaMatkul, t.KodeMatkul, t.NamaKelas, t.NamaPT)
			}
		}

		// Portfolios
		str += "\n*Portofolio*\n\n"

		str += "Penelitian\n"
		for idx, data := range detail.Researches {
			str += fmt.Sprintf("%d. %s (%d)\n", idx+1, data.JudulKegiatan, data.TahunKegiatan)
		}
		if len(detail.Researches) == 0 {
			str += "Tidak ada.\n"
		}

		str += "Pengabdian Masyarakat\n"
		for idx, data := range detail.Devotionals {
			str += fmt.Sprintf("%d. %s (%d)\n", idx+1, data.JudulKegiatan, data.TahunKegiatan)
		}
		if len(detail.Devotionals) == 0 {
			str += "Tidak ada.\n"
		}

		str += "Karya\n"
		for idx, data := range detail.Creations {
			str += fmt.Sprintf("%d. [%s] %s (%d)\n", idx+1, data.JenisKegiatan, data.JudulKegiatan, data.TahunKegiatan)
		}
		if len(detail.Creations) == 0 {
			str += "Tidak ada.\n"
		}

		str += "HKI/Paten\n"
		for idx, data := range detail.Patents {
			str += fmt.Sprintf("%d. %s (%d)\n", idx+1, data.JudulKegiatan, data.TahunKegiatan)
		}
		if len(detail.Patents) == 0 {
			str += "Tidak ada.\n"
		}
	case "mhs":
		mhs, err := kanoutils.GetMHSDetails(id)
		if err != nil {
			return err.Error()
		}

		enterDate, err := time.Parse(layout, mhs.TanggalMasuk)
		if err != nil {
			return err.Error()
		}
		enterDateFormatted := enterDate.Format(dateFormat)

		gender := "Tidak diketahui"
		switch mhs.JenisKelamin {
		case "L":
			gender = "Laki-laki"
		case "P":
			gender = "Perempuan"
		default:
			gender += fmt.Sprintf("(%s)", mhs.JenisKelamin)
		}

		str = "*Detail Mahasiswa*\n\n"
		str += fmt.Sprintf("- Nama: %s\n- Jenis Kelamin: %s\n- NIM: %s\n", mhs.Nama, gender, mhs.NIM)
		str += fmt.Sprintf("- Perguruan Tinggi: %s\n- Tanggal Masuk: %s\n- Jenjang - Program Studi: %s - %s\n", mhs.NamaPT, enterDateFormatted, mhs.Jenjang, mhs.Prodi)
		str += fmt.Sprintf("- Status Awal Mahasiswa: %s\n- Status Terakhir Mahasiswa: %s", mhs.JenisDaftar, mhs.StatusSaatIni)
	case "pt":
	case "prodi":
	}

	return strings.TrimSpace(str)
}

func PDDIKTIHandler(ctx *MessageContext) {
	conf := projectconfig.GetConfig()
	if !conf.PDDiktiKey.Valid || !conf.PDDiktiIV.Valid {
		ctx.Instance.Reply("This command is disabled, ask the owner to fix it!", true)
		return
	}

	MAX_RESULT := 24
	allowedQueryTypes := []string{"dosen", "mahasiswa", "pt", "prodi", "mhs", "all"}
	fullCmd := ctx.Parser.GetCommand().FullCommand
	args := ctx.Parser.GetArgs()
	if len(args) < 2 {
		ctx.Instance.Reply(fmt.Sprintf("*Pangkalan Data Pendidikan Tinggi (PDDikti) API Wrapper*\nAmbil informasi seputar perguruan tinggi dari PDDikti (yang diambil hanya 3 teratas saja)\n\nPenggunaan:\n%s %s nama/nim\n\nContoh:\n%s dosen Achmad Munir\n%s mahasiswa 2406437994 UI\n%s mhs,prodi,pt nuklir\n%s all telekomunikasi", fullCmd, strings.Join(allowedQueryTypes, ","), fullCmd, fullCmd, fullCmd, fullCmd), true)
		return
	}

	enabledTypes := []bool{false, false, false, false}
	queryType := strings.Split(args[0].Content, ",")
	for _, qType := range queryType {
		idx := slices.Index(allowedQueryTypes, qType)
		if idx == -1 {
			ctx.Instance.Reply(fmt.Sprintf("Tipe kueri yang diizinkan hanya: %s\nDidapat: %s", strings.Join(allowedQueryTypes, ", "), qType), true)
			return
		}

		if idx < 4 {
			enabledTypes[idx] = true
		} else if idx == 4 {
			enabledTypes[1] = true
		} else if idx == 5 {
			enabledTypes[0] = true
			enabledTypes[1] = true
			enabledTypes[2] = true
			enabledTypes[3] = true
		}
	}

	div := 0
	for i := range enabledTypes {
		if enabledTypes[i] {
			div++
		}
	}
	if div == 0 {
		return
	}
	MAX_RESULT /= div

	textRunes := []rune(ctx.Parser.Text)
	queryString := url.QueryEscape(string(textRunes[args[1].Start:]))
	result, err := kanoutils.SearchDiddy(queryString, conf.PDDiktiKey.String, conf.PDDiktiIV.String)
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Terjadi kesalahan: %s", err), true)
		return
	}

	// Create msg
	msgs := []string{}
	cp := map[string]string{} // Value must be one of "dosen", "mhs", "pt", or "prodi"

	if len(result.Dosen) > 0 && enabledTypes[0] {
		adds := "*List Dosen*\n==========\n\n"
		dosens := []string{}

		for idx, dosen := range result.Dosen {
			if idx >= MAX_RESULT {
				break
			}
			cp[dosen.ID] = "dosen"
			dosenAdds := fmt.Sprintf("Nama: %s\n", dosen.Nama)
			dosenAdds += fmt.Sprintf("Nomor Induk Dosen Nasional: %s\n", dosen.NIDN)
			if dosen.NUPTK != "" {
				dosenAdds += fmt.Sprintf("Nomor Unik Pendidik dan Tenaga Kependidikan : %s\n", dosen.NUPTK)
			}
			dosenAdds += fmt.Sprintf("Perguruan Tinggi: %s", dosen.NamaPT)
			if dosen.SingkatanPT != "" {
				dosenAdds += fmt.Sprintf(" (%s)", dosen.SingkatanPT)
			}
			dosenAdds += fmt.Sprintf("\nProgram Studi: %s", dosen.NamaProdi)

			dosens = append(dosens, dosenAdds)
		}

		adds += strings.Join(dosens, "\n----------\n")
		msgs = append(msgs, adds)
	}
	if len(result.Mahasiswa) > 0 && enabledTypes[1] {
		adds := "*List Mahasiswa*\n==========\n\n"
		mhss := []string{}

		for idx, mhs := range result.Mahasiswa {
			if idx >= MAX_RESULT {
				break
			}

			cp[mhs.ID] = "mhs"

			mhsAdds := fmt.Sprintf("Nama: %s\n", mhs.Nama)
			mhsAdds += fmt.Sprintf("Nomor Induk Mahasiswa: %s\n", mhs.NIM)
			mhsAdds += fmt.Sprintf("Perguruan Tinggi: %s", mhs.NamaPT)
			if mhs.SingkatanPT != "" {
				mhsAdds += fmt.Sprintf(" (%s)", mhs.SingkatanPT)
			}
			mhsAdds += fmt.Sprintf("\nProgram Studi: %s", mhs.NamaProdi)

			mhss = append(mhss, mhsAdds)
		}

		adds += strings.Join(mhss, "\n----------\n")
		msgs = append(msgs, adds)
	}
	if len(result.PT) > 0 && enabledTypes[2] {
		adds := "*List Perguruan Tinggi*\n==========\n\n"
		pts := []string{}

		for idx, pt := range result.PT {
			if idx >= MAX_RESULT {
				break
			}

			cp[pt.ID] = "pt"

			ptAdds := fmt.Sprintf("Nama: %s", pt.Nama)
			if pt.NamaSingkat != "" {
				ptAdds += fmt.Sprintf(" (%s)", pt.NamaSingkat)
			}
			ptAdds += fmt.Sprintf("\nKode: %s", pt.Kode)

			pts = append(pts, ptAdds)
		}

		adds += strings.Join(pts, "\n----------\n")
		msgs = append(msgs, adds)
	}
	if len(result.Prodi) > 0 && enabledTypes[3] {
		adds := "*List Program Studi*\n==========\n\n"
		prodis := []string{}

		for idx, prodi := range result.Prodi {
			if idx >= MAX_RESULT {
				break
			}

			cp[prodi.ID] = "prodi"

			mhsAdds := fmt.Sprintf("Nama: %s - %s\nPerguruan Tinggi: %s", prodi.Jenjang, prodi.Nama, prodi.PT)
			if prodi.PTSingkat != "" {
				mhsAdds += fmt.Sprintf(" (%s)", prodi.PTSingkat)
			}

			prodis = append(prodis, mhsAdds)
		}

		adds += strings.Join(prodis, "\n----------\n")
		msgs = append(msgs, adds)
	}

	fmt.Println("CP LENGTH:", len(cp))

	if len(cp) == 1 {
		str := ""
		for k, v := range cp {
			str += generateDetail(k, v)
		}

		if len(str) != 0 {
			ctx.Instance.Reply(str, true)
			return
		}
	}

	if len(msgs) == 0 {
		ctx.Instance.Reply("Server returned an empty data", true)
	} else {
		ctx.Instance.Reply(strings.Join(msgs, "\n\n"), true)
	}
}
