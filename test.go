package main

import (
	"encoding/json"
	"fmt"
	"kano/internal/database"
	"kano/internal/database/models"
	"kano/internal/utils/parser"
	"kano/internal/utils/six"
	"kano/internal/utils/six/schedules"
	"net/url"
	"os"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types"
)

func testaja() {
	mystr := `            there is        a space
between us???                                            even
                      at the end too!                  `
	fmt.Printf("%q\n", strings.Join(strings.Fields(mystr), " "))

	mypath := "/search?q=my+query"
	u, _ := url.Parse("https://google.com")
	p, _ := url.Parse(mypath)
	u.Path = p.Path
	u.RawQuery = p.RawQuery

	fmt.Println(u.String())

	quotestr := "“その音が鳴るなら (feat. 星乃一歌 & 天馬咲希 & 望月穂波 & 日野森志歩 & 巡音ルカ) · Leo/need”"
	for i, r := range quotestr {
		fmt.Printf("%d\t: %c\n", i, r)
	}

	for i := range 4 {
		fmt.Println(i, ^i)
	}

	prs := parser.Init([]string{"."})
	// res, err := prs.Parse(".test filter=normal other='spaced value' other=other empty='' empty= normal_arg 'normal arg' ''")
	res, err := prs.Parse(".stk name='' ''")
	if err != nil {
		fmt.Println(err)
	} else {
		m, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(m))
	}

	mymaps := map[string]string{"a": "b", "c": "d", "e": "f", "g": "h", "i": "j"}
	for key, val := range mymaps {
		fmt.Println(key, val)
		delete(mymaps, key)
	}

	fmt.Printf("%+v\n", mymaps)
	fmt.Println("Hai"[0:0])
}

func dbgw() {
	db := database.GetInstance()
	jid, _ := types.ParseJID("1234567890@g.us")
	grp := models.Group{
		JID:  jid,
		Name: "Test",
		Participants: []models.Participant{
			{ContactID: 1, Role: models.ParticipantRoleLeft},
		},
	}
	tx := db.Create(&grp)
	fmt.Println(tx.Error)
}

func mysix() {
	fmt.Println("Start")
	a := time.Now()
	subjs, err := six.GetAllSchedules()
	if err != nil {
		panic(err)
	}

	b := time.Now()
	diff := (b.UnixMilli() - a.UnixMilli())
	fmt.Println("Schedule Parser took", diff, "milliseconds")

	m, _ := json.MarshalIndent(subjs, "", "  ")
	os.WriteFile("schedules.json", m, 0644)

	c := time.Now()
	diff = c.UnixMilli() - b.UnixMilli()
	fmt.Println("Marshal took", diff, "milliseconds")
}

func mysixDiff() {
	fmt.Println("Reading schedules.json")
	schedBytes, err := os.ReadFile("schedules.json")
	if err != nil {
		panic(err)
	}

	fmt.Println("Parsing schedules.json")
	var scheds []schedules.SemesterSubject
	err = json.Unmarshal(schedBytes, &scheds)
	if err != nil {
		panic(err)
	}

	fmt.Println("Generating diff")
	diff, err := schedules.GetScheduleDiff(scheds)
	if err != nil {
		panic(err)
	}

	fmt.Println("Marshaling generated diff")
	mar, err := json.MarshalIndent(diff, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println("Writing diff into diff.json")
	err = os.WriteFile("diff.json", mar, 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("Applying diff")
	err = schedules.ApplyDiff(diff)
	if err != nil {
		panic(err)
	}
	schedules.CleanupTmpFiles()
}

// func main() {
// 	mysix()
// 	mysixDiff()
// }
