package models

type Student struct {
	Id      uint   `json:"id"`
	Name    string `json:"name"`
	Nim     uint   `json:"nim"`
	Major   string `json:"major"`
	Faculty string `json:"faculty"`

	CustomName string `json:"custom_name" gorm:"column:custom_name"`
}

func (_ Student) TableName() string {
	return "students"
}
