package pddikti

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

func (d DiddySearchResult) TotalLength() uint {
	return uint(len(d.Mahasiswa) + len(d.Dosen) + len(d.PT) + len(d.Prodi))
}

func (d DiddySearchResult) IsEmpty() bool {
	return d.TotalLength() == 0
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
