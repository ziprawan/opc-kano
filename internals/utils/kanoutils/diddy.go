package kanoutils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var wle = []string{"ian", "pdd", "ps:", "htt", "emd", "int", "ikt", "car", "/en", "pi-", "pen", "ll/", "ek.", "i.k", "//a", "c/a", "isa", "id/", "ikt", "go."}

// Search result
type DiddyMHS struct {
	ID          string `json:"id"`
	Nama        string `json:"nama"`
	NIM         string `json:"nim"`
	NamaPT      string `json:"nama_pt"`
	SingkatanPT string `json:"sinkatan_pt"`
	NamaProdi   string `json:"nama_prodi"`
}

type DiddyPNS struct {
	ID          string `json:"id"`
	Nama        string `json:"nama"`
	NIDN        string `json:"nidn"`
	NUPTK       string `json:"nuptk"`
	NamaPT      string `json:"nama_pt"`
	SingkatanPT string `json:"sinkatan_pt"`
	NamaProdi   string `json:"nama_prodi"`
}

type DiddyPT struct {
	ID          string `json:"id"`
	Kode        string `json:"kode"`
	NamaSingkat string `json:"nama_singkat"`
	Nama        string `json:"nama"`
}

type DiddyStud struct {
	ID        string `json:"id"`
	Nama      string `json:"nama"`
	Jenjang   string `json:"jenjang"`
	PT        string `json:"pt"`
	PTSingkat string `json:"pt_singkat"`
}

type DiddySearchResult struct {
	Mahasiswa []DiddyMHS  `json:"mahasiswa"`
	Dosen     []DiddyPNS  `json:"dosen"`
	PT        []DiddyPT   `json:"pt"`
	Prodi     []DiddyStud `json:"prodi"`
}

// MHS details
type DiddyDetailsMHS struct {
	ID            string `json:"id"`
	NamaPT        string `json:"nama_pt"`
	KodePT        string `json:"kode_pt"`
	KodeProdi     string `json:"kode_prodi"`
	Prodi         string `json:"prodi"`
	Nama          string `json:"nama"`
	NIM           string `json:"nim"`
	JenisDaftar   string `json:"jenis_daftar"`
	PT_ID         string `json:"id_pt"`
	SMS_ID        string `json:"id_sms"`
	JenisKelamin  string `json:"jenis_kelamin"`
	Jenjang       string `json:"jenjang"`
	StatusSaatIni string `json:"status_saat_ini"`
	TanggalMasuk  string `json:"tanggal_masuk"`
}

// PNS details
type DiddyPNSProfile struct {
	ID                  string `json:"id_sdm"`
	NamaDosen           string `json:"nama_dosen"`
	NamaPT              string `json:"nama_pt"`
	NamaProdi           string `json:"nama_prodi"`
	JenisKelamin        string `json:"jenis_kelamin"`
	JabatanAkademik     string `json:"jabatan_akademik"`
	PendidikanTertinggi string `json:"pendidikan_tertinggi"`
	StatusIkatanKerja   string `json:"status_ikatan_kerja"`
	StatusAktivitas     string `json:"status_aktivitas"`
}
type DiddyPNSStudyHistory struct {
	ID             string `json:"id_sdm"`
	NIDN           string `json:"nidn"`
	TahunMasuk     int    `json:"tahun_masuk"`
	TahunLulus     int    `json:"tahun_lulus"`
	NamaProdi      string `json:"nama_prodi"`
	Jenjang        string `json:"jenjang"`
	NamaPT         string `json:"nama_pt"`
	GelarAkademik  string `json:"gelar_akademik"`
	SingkatanGelar string `json:"singkatan_gelar"`
}
type DiddyPNSTeachHistory struct {
	ID           string `json:"id_sdm"`
	NamaSemester string `json:"nama_semester"`
	KodeMatkul   string `json:"kode_matkul"`
	NamaMatkul   string `json:"nama_matkul"`
	NamaKelas    string `json:"nama_kelas"`
	NamaPT       string `json:"nama_pt"`
}
type DiddyPNSPortfolio struct {
	ID            string `json:"id_sdm"`
	JenisKegiatan string `json:"jenis_kegiatan"`
	JudulKegiatan string `json:"judul_kegiatan"`
	TahunKegiatan int    `json:"tahun_kegiatan"`
}
type DiddyDetailsPNS struct {
	Profile           DiddyPNSProfile
	StudyHistories    []DiddyPNSStudyHistory
	TeachingHistories map[string][]DiddyPNSTeachHistory
	Portfolios        map[string][]DiddyPNSPortfolio
	Researches        []DiddyPNSPortfolio
	Devotionals       []DiddyPNSPortfolio
	Creations         []DiddyPNSPortfolio
	Patents           []DiddyPNSPortfolio
}

// Functions

func buildDiddyUrl(path ...string) string {
	return fmt.Sprintf(
		"%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s",
		wle[3],
		wle[2],
		wle[14],
		wle[9],
		wle[1],
		wle[6],
		wle[13],
		wle[4],
		wle[6],
		wle[16],
		wle[5],
		wle[12],
		wle[19],
		wle[17],
		strings.Join(path, "/"),
	)
}

func decryptSearchResult(body []byte, key, iv string) ([]byte, error) {
	cipherText, err := base64.StdEncoding.DecodeString(string(body[1 : len(body)-2]))
	if err != nil {
		return nil, err
	}
	keyByte, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}
	ivByte, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		return nil, err
	}
	if len(cipherText)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("CipherText is not a multiple of block size")
	}
	mode := cipher.NewCBCDecrypter(block, ivByte)
	decrypted := make([]byte, len(cipherText))
	mode.CryptBlocks(decrypted, cipherText)
	decrypted, err = Pkcs7Unpad(decrypted, aes.BlockSize)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

func fetchDiddy(url string) (*http.Response, error) {
	origUrl := strings.Replace(buildDiddyUrl(), "api-", "", 1)
	origUrl = strings.Replace(origUrl, ".go.id/", ".go.id", 1)

	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println(origUrl)

	// Fix 403 status
	req.Header.Set("Origin", origUrl)
	req.Header.Set("Referer", origUrl)

	return client.Do(req)
}

func SearchDiddy(query, key, iv string) (*DiddySearchResult, error) {
	url := fmt.Sprintf(
		"%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s",
		wle[3],
		wle[2],
		wle[14],
		wle[9],
		wle[1],
		wle[6],
		wle[13],
		wle[4],
		wle[6],
		wle[16],
		wle[5],
		wle[12],
		wle[19],
		wle[17],
		wle[10],
		wle[7],
		wle[0],
		wle[8],
		wle[15],
		wle[11],
		query,
	)

	resp, err := fetchDiddy(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("diddy: Expected HTTP code 200, got %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	decrypted, err := decryptSearchResult(body, key, iv)
	if err != nil {
		return nil, err
	}

	var result DiddySearchResult
	err = json.Unmarshal(decrypted, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func GetMHSDetails(mhsID string) (*DiddyDetailsMHS, error) {
	url := buildDiddyUrl("detail", "mhs", mhsID)
	fmt.Println("Fetching:", url)
	resp, err := fetchDiddy(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result DiddyDetailsMHS
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func GetPNSDetails(pnsID string) (*DiddyDetailsPNS, error) {
	detail := DiddyDetailsPNS{}
	detail.TeachingHistories = map[string][]DiddyPNSTeachHistory{}

	// PNS Detail
	profileUrl := buildDiddyUrl("dosen", "profile", pnsID)
	resp, err := fetchDiddy(profileUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&detail.Profile)
	if err != nil {
		return nil, err
	}

	// Study Histories
	sHistoryUrl := buildDiddyUrl("dosen", "study-history", pnsID)
	resp, err = fetchDiddy(sHistoryUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&detail.StudyHistories)
	if err != nil {
		return nil, err
	}

	// Teaching Histories
	var tHistories []DiddyPNSTeachHistory
	tHistoryUrl := buildDiddyUrl("dosen", "teaching-history", pnsID)
	resp, err = fetchDiddy(tHistoryUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&tHistories)
	if err != nil {
		return nil, err
	}
	for _, t := range tHistories {
		detail.TeachingHistories[t.NamaSemester] = append(detail.TeachingHistories[t.NamaSemester], t)
	}

	portfolios := []string{"penelitian", "pengabdian", "karya", "paten"}
	for _, portfolioType := range portfolios {
		portfolioUrl := buildDiddyUrl("dosen", "portofolio", portfolioType, pnsID)
		resp, err = fetchDiddy(portfolioUrl)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		switch portfolioType {
		case "penelitian":
			err = json.NewDecoder(resp.Body).Decode(&detail.Researches)
		case "pengabdian":
			err = json.NewDecoder(resp.Body).Decode(&detail.Devotionals)
		case "karya":
			err = json.NewDecoder(resp.Body).Decode(&detail.Creations)
		case "paten":
			err = json.NewDecoder(resp.Body).Decode(&detail.Patents)
		}
		if err != nil {
			return nil, err
		}
	}

	return &detail, nil
}
