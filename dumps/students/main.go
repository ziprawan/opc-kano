package main

import (
	"encoding/json"
	"fmt"
	"kano/internal/database"
	"math"
	"os"
)

type StudentDump struct {
	Id      uint   `json:"id"`
	Name    string `json:"name"`
	Nim     uint   `json:"nim"`
	Major   string `json:"major"`
	Faculty string `json:"faculty"`
}

func (_ StudentDump) TableName() string {
	return "students"
}

// Please run this script only when this is the first time you are initialized the database
func main() {
	db := database.GetInstance()

	content, err := os.ReadFile("dumps/students/students.json")
	if err != nil {
		panic(err)
	}

	var dump []StudentDump
	err = json.Unmarshal(content, &dump)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found total %d of student data\n", len(dump))

	itemPerLoop := (float64)(1000)
	loops := int(math.Ceil(float64(len(dump)) / itemPerLoop))

	for idx := range loops {
		curStudents := dump[idx*int(itemPerLoop) : int(math.Min(float64(len(dump)-1), float64(idx+1)*itemPerLoop))]
		fmt.Printf("Inserting %d data... ", len(curStudents))

		res := db.Create(curStudents)
		affected := res.RowsAffected
		if err := res.Error; err != nil {
			fmt.Printf("Failed! (%d)\n", affected)
			panic(err)
		} else {
			fmt.Printf("Success! (%d)\n", affected)
		}
	}

	fmt.Printf("Done.\n")
}
