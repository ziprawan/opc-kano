package structs

type Title struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Claimable   bool   `json:"claimable"`
	IsHolding   bool   `json:"is_holding"`
	HolderCount int    `json:"holder_count"`
}
