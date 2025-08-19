package finaiditb

type ScholarshipAngkatan struct {
	Name string `json:"name"`
}

type ScholarshipBenefitMedia struct {
	OriginalUrl string `json:"original_url"`
}

type ScholarshipBenefit struct {
	Code string `json:"code"`
	Name string `json:"name"`

	Media []ScholarshipBenefitMedia `json:"media"`
}

type ScholarshipBenefitScholarship struct {
	Description string             `json:"description"`
	Benefit     ScholarshipBenefit `json:"benefit"`
}

type ScholarshipSastraPendidikan struct {
	BaseStrataId string `json:"base_strata_id"`
}

type ScholarshipProgram struct {
	Code        ScholarshipProgramCode `json:"code"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`

	SelectionType ScholarshipProgramSelectionType `json:"selection_type"` // As far as I know, 1 = "Internal", 2 = "External"
}

type ScholarshipRequirement struct {
	Name             string `json:"name"`
	NeedToUploadFile bool   `json:"need_to_upload_file"`
}

type ScholarshipPartner struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	Pic      string `json:"pic"`
	PicPhone string `json:"pic_phone"`
}

type ScholarshipsData struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	ExternalUrl *string `json:"external_url"`
	Description string  `json:"description"`
	Slug        string  `json:"slug"`

	RegistrationStartDate string `json:"registration_start_date"`
	RegistrationEndDate   string `json:"registration_end_date"`
	FundingStartDate      string `json:"funding_start_date"`
	FundingEndDate        string `json:"funding_end_date"`

	Quota  int `json:"quota"`
	Amount int `json:"amount"`

	MinIP  string `json:"min_ip"`
	MinIPK string `json:"min_ipk"`

	// Only exists if program_id is 2
	JobTitle    *string `json:"job_title"`
	JobDesc     *string `json:"job_desc"`
	JobPosition *string `json:"job_position"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`

	// Relationships
	Partner ScholarshipPartner `json:"partner"`
	Program ScholarshipProgram `json:"program"`
	// Arr
	StrataPendidikans       []ScholarshipSastraPendidikan   `json:"strata_pendidikans"`
	BenefitScholarships     []ScholarshipBenefitScholarship `json:"benefit_scholarships"`
	Angkatans               []ScholarshipAngkatan           `json:"angkatans"`
	ScholarshipRequirements []ScholarshipRequirement        `json:"scholarship_requirements"`
}

type FinaidScholarshipsResponse struct {
	Data []ScholarshipsData `json:"data"`
}

type ScholarshipProgramCode string

var (
	PROGRAM_CODE_BPPUKT    ScholarshipProgramCode = "P00001"
	PROGRAM_CODE_KERJA     ScholarshipProgramCode = "P00002"
	PROGRAM_CODE_GTA       ScholarshipProgramCode = "P00003"
	PROGRAM_CODE_FASILITAS ScholarshipProgramCode = "P00004"
	PROGRAM_CODE_KIPK      ScholarshipProgramCode = "P00005"
	PROGRAM_CODE_LPDP      ScholarshipProgramCode = "P00006"
	PROGRAM_CODE_SWASTA    ScholarshipProgramCode = "P00007"
)

type ScholarshipProgramSelectionType int

var (
	SELECTION_TYPE_INTERNAL ScholarshipProgramSelectionType = 1
	SELECTION_TYPE_EXTERNAL ScholarshipProgramSelectionType = 2
)
