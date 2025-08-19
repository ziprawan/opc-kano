package kanoutils

import (
	"fmt"
	finaiditb "kano/internals/utils/finaid-itb"
	"net/url"
	"strings"
)

func GenerateFinaidScholarshipMessage(data finaiditb.ScholarshipsData) string {
	msg := "*🚨 Finaid ITB - New Scholarship Detected 🚨*\n"

	// Header
	if data.Program.Code == finaiditb.PROGRAM_CODE_KERJA {
		msg += "🗒 *NEW J*B AVAILABLE* ❗️"
	} else {
		msg += "⚠️ *BEASISWA BARU DITEMUKAN* ❗️"
	}
	msg += "\n\n"

	// Some bs that will be used in the details
	sifatBeasiswa := "Unknown"
	switch data.Program.SelectionType {
	case finaiditb.SELECTION_TYPE_INTERNAL:
		sifatBeasiswa = "Internal"
	case finaiditb.SELECTION_TYPE_EXTERNAL:
		sifatBeasiswa = "External"
	}

	stratas := []string{}
	for _, strata := range data.StrataPendidikans {
		stratas = append(stratas, strata.BaseStrataId)
	}

	angkatans := []string{}
	for _, angkatan := range data.Angkatans {
		angkatans = append(angkatans, angkatan.Name)
	}

	partner := data.Partner

	// Scholarship details
	msg += fmt.Sprintf("*[ID: %d] %s*\n", data.ID, data.Name)
	msg += fmt.Sprintf("by \"%s\"\n", data.Partner.Name)

	msg += "\n"
	msg += fmt.Sprintf("[%s] Kuota: *%d*, Jumlah dana: *%s*\n", sifatBeasiswa, data.Quota, FormatNumber(int64(data.Amount)))
	msg += fmt.Sprintf("Periode pendaftaran: *%s*\n", FormatRangeDateOnly(data.RegistrationStartDate, data.RegistrationEndDate))

	msg += "\n"
	msg += data.Description
	msg += "\n"

	if data.Program.Code == finaiditb.PROGRAM_CODE_KERJA {
		jobTitle := "-"
		jobDesc := "-"
		jobPos := "-"
		if data.JobTitle != nil {
			jobTitle = *data.JobTitle
		}
		if data.JobDesc != nil {
			jobDesc = *data.JobDesc
		}
		if data.JobPosition != nil {
			jobPos = *data.JobPosition
		}

		msg += "\n"
		msg += "🤔 *Detail Pekerjaan*\n"
		msg += fmt.Sprintf("- *Judul*: %s\n- *Deskripsi*:\n%s\n- *Posisi*: %s\n", jobTitle, jobDesc, jobPos)
	}

	msg += "\n"
	msg += "✅  *Manfaat yang diperoleh*\n"

	for _, benefit := range data.BenefitScholarships {
		msg += fmt.Sprintf("- *%s* - %s\n", benefit.Benefit.Name, benefit.Description)
	}

	msg += "\n"
	msg += "❓ *Persyaratan*\n"
	msg += fmt.Sprintf("- Strata Pendidikan: *%s*\n", strings.Join(stratas, ", "))
	msg += fmt.Sprintf("- Angkatan: *%s*\n", strings.Join(angkatans, ", "))
	msg += fmt.Sprintf("- Minimal IP *%s* dan Minimal IPK *%s*\n", data.MinIP, data.MinIPK)

	msg += "\n"
	msg += "📌 *Kelengkapan Dokumen*\n"
	for _, requirement := range data.ScholarshipRequirements {
		l := ""
		if requirement.NeedToUploadFile {
			l = "[📂]"
		}
		msg += fmt.Sprintf("- %s %s\n", l, requirement.Name)
	}

	msg += "\n"
	msg += "👥 *Informasi Partner*\n"
	msg += fmt.Sprintf("- *Nama*: %s\n- *Email*: %s\n- *Alamat*: %s\n- *Nomor Alamat*: %s\n- *Contact Person*: %s (%s)\n", partner.Name, partner.Email, partner.Address, partner.Phone, partner.Pic, partner.PicPhone)

	msg += "\n"
	msg += fmt.Sprintf("Periode pendanaan: *%s*\n", FormatRangeDateOnly(data.FundingStartDate, data.FundingEndDate))
	msg += fmt.Sprintf("https://finaid.itb.ac.id/beasiswa/%s", url.PathEscape(data.Slug))

	if data.ExternalUrl != nil {
		msg += fmt.Sprintf("\nURL Eksternal: %s", *data.ExternalUrl)
	}

	return msg
}
