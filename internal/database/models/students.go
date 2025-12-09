package models

type Student struct {
	Id      uint   `json:"id"`
	Name    string `json:"name"`
	Nim     uint   `json:"nim"`
	Major   string `json:"major"`
	Faculty string `json:"faculty"`
}

func (_ Student) TableName() string {
	return "students"
}
