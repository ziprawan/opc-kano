package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"kano/internals/utils/data"
	"net/http"
	"strconv"
)

var kne = []string{"nka", "lte", "a.j", "htt", "//e", "=01", "/__", ".ne", "dat", "sve", "kit", "?x-", "ps:", "val", "ted", "rk/", "ida", "son", "two", "-in"}

type MapResponse map[string]any

type HSRAvatarDetail struct {
	AvatarID int `json:"avatarId"`
	Level    int `json:"level"`
}

type HSRRecordInfo struct {
	AchievementCount int `json:"achievementCount"`
	BookCount        int `json:"bookCount"`
	AvatarCount      int `json:"avatarCount"`
	EquipmentCount   int `json:"equipmentCount"`
	MusicCount       int `json:"musicCount"`
	RelicCount       int `json:"relicCount"`
}

type HSRDetailInfo struct {
	WorldLevel    int               `json:"worldLevel"`
	Signature     string            `json:"signature"`
	Nickname      string            `json:"nickname"`
	Level         int               `json:"level"`
	UID           int               `json:"uid"`
	RecordInfo    HSRRecordInfo     `json:"recordInfo"`
	AvatarDetails []HSRAvatarDetail `json:"avatarDetailList"`
}

type HSRInfo struct {
	DetailInfo HSRDetailInfo `json:"detailInfo"`
}

type GenshinAvatarInfo struct {
	AvatarID int `json:"avatarId"`
	Level    int `json:"level"`
}

type GenshinPlayerInfo struct {
	Nickname         string              `json:"nickname"`
	Level            int                 `json:"level"`
	Signature        string              `json:"signature"`
	WorldLevel       int                 `json:"worldLevel"`
	AchievementTotal int                 `json:"finishAchievementNum"`
	AvatarInfoList   []GenshinAvatarInfo `json:"showAvatarInfoList"`
}

type GenshinInfo struct {
	PlayerInfo GenshinPlayerInfo `json:"playerInfo"`
	UID        string            `json:"uid"`
}

type ZZZPlayerDetail struct {
	Nickname string
	AvatarID int `json:"AvatarId"`
	UID      int `json:"Uid"`
	Level    int
}

type ZZZSocialDetail struct {
	ProfileDetail ZZZPlayerDetail
	Desc          string
}

type ZZZAvatar struct {
	ID    int `json:"Id"`
	Level int
}

type ZZZShowcaseDetail struct {
	AvatarList []ZZZAvatar
}

type ZZZPlayerInfo struct {
	SocialDetail   ZZZSocialDetail
	ShowcaseDetail ZZZShowcaseDetail
}

type ZZZInfo struct {
	PlayerInfo ZZZPlayerInfo
}

func processObjects(root []any, idx int) (any, error) {
	currObj := root[idx]

	if arrCurrObj, ok := currObj.([]any); ok {
		arr := []any{}

		for _, val := range arrCurrObj {
			val, ok := val.(float64)
			if !ok {
				return nil, errors.New("val is not an integer (in array)")
			}
			processed, err := processObjects(root, int(val))
			if err != nil {
				return nil, err
			}
			arr = append(arr, processed)
		}

		return arr, nil
	} else if objCurrObj, ok := currObj.(map[string]any); ok {
		obj := map[string]any{}

		for key, val := range objCurrObj {
			val, ok := val.(float64)
			if !ok {
				return nil, fmt.Errorf("val of %s is not an integer", key)
			}
			processed, err := processObjects(root, int(val))
			if err != nil {
				return nil, err
			}
			obj[key] = processed
		}

		return obj, nil
	} else {
		return currObj, nil
	}
}

func getHSRData(UID int) (*HSRInfo, error) {
	url := fmt.Sprintf(
		"%s%s%s%s%s%s%s%s%d%s%s%s%s%s%s%s%s%s%s%s%s%s",
		kne[3],
		kne[12],
		kne[4],
		kne[0],
		kne[7],
		kne[18],
		kne[15],
		"hsr/",
		UID,
		kne[6],
		kne[8],
		kne[2],
		kne[17],
		kne[11],
		kne[9],
		kne[1],
		kne[10],
		kne[19],
		kne[13],
		kne[16],
		kne[14],
		kne[5],
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("hsr: Error fetching data: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hsr: Error: HTTP status code %d", resp.StatusCode)
	}

	var data MapResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("hsr: Error decoding JSON: %s", err)
	}

	nodes, ok := data["nodes"].([]any)
	if !ok {
		return nil, errors.New("hsr: Nodes is not an array")
	}

	var raw any

	for _, node := range nodes {
		node, ok := node.(map[string]any)
		if !ok {
			return nil, errors.New("node is not a map of any")
		}
		nodeType, ok := node["type"].(string)
		if !ok || nodeType != "data" {
			if nodeType == "error" {
				errorObj := node["error"].(map[string]any)
				errorMsg := errorObj["message"].(string)

				return nil, errors.New(errorMsg)
			}
			continue
		}
		data, ok := node["data"].([]any)
		if !ok {
			return nil, errors.New("data field is not an array of any")
		}
		nig, err := processObjects(data, 0)
		if err != nil {
			return nil, err
		}
		raw = nig
	}

	jsonData, err := json.Marshal(raw)
	if err != nil {
		panic(err)
	}

	var hsrInfo HSRInfo
	err = json.Unmarshal(jsonData, &hsrInfo)
	if err != nil {
		panic(err)
	}

	return &hsrInfo, nil
}

func getGIData(UID int) (*GenshinInfo, error) {
	fmt.Printf("Get GENSHIN UID: %d\n", UID)
	url := fmt.Sprintf(
		"%s%s%s%s%s%s%s%s%d%s%s%s%s%s%s%s%s%s%s%s%s%s",
		kne[3],
		kne[12],
		kne[4],
		kne[0],
		kne[7],
		kne[18],
		kne[15],
		"u/",
		UID,
		kne[6],
		kne[8],
		kne[2],
		kne[17],
		kne[11],
		kne[9],
		kne[1],
		kne[10],
		kne[19],
		kne[13],
		kne[16],
		kne[14],
		kne[5],
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %d", resp.StatusCode)
	}

	var data MapResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON: %s", err)
	}

	nodes, ok := data["nodes"].([]any)
	if !ok {
		return nil, errors.New("nodes is not an array")
	}

	var raw any

	for _, node := range nodes {
		node, ok := node.(map[string]any)
		if !ok {
			return nil, errors.New("node is not a map of any")
		}
		nodeType, ok := node["type"].(string)
		if !ok || nodeType != "data" {
			if nodeType == "error" {
				errorObj := node["error"].(map[string]any)
				errorMsg := errorObj["message"].(string)

				return nil, errors.New(errorMsg)
			}
			continue
		}
		data, ok := node["data"].([]any)
		if !ok {
			return nil, errors.New("data field is not an array of any")
		}
		nig, err := processObjects(data, 0)
		if err != nil {
			return nil, err
		}
		raw = nig
	}

	jsonData, err := json.Marshal(raw)
	if err != nil {
		panic(err)
	}

	var genshinInfo GenshinInfo
	err = json.Unmarshal(jsonData, &genshinInfo)
	if err != nil {
		panic(err)
	}

	return &genshinInfo, nil
}

func getZZZData(UID int) (*ZZZInfo, error) {
	fmt.Printf("Get GENSHIN UID: %d\n", UID)
	url := fmt.Sprintf(
		"%s%s%s%s%s%s%s%s%d%s%s%s%s%s%s%s%s%s%s%s%s%s",
		kne[3],
		kne[12],
		kne[4],
		kne[0],
		kne[7],
		kne[18],
		kne[15],
		"zzz/",
		UID,
		kne[6],
		kne[8],
		kne[2],
		kne[17],
		kne[11],
		kne[9],
		kne[1],
		kne[10],
		kne[19],
		kne[13],
		kne[16],
		kne[14],
		kne[5],
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: HTTP status code %d", resp.StatusCode)
	}

	var data MapResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON: %s", err)
	}

	nodes, ok := data["nodes"].([]any)
	if !ok {
		return nil, errors.New("nodes is not an array")
	}

	var raw any

	for _, node := range nodes {
		node, ok := node.(map[string]any)
		if !ok {
			return nil, errors.New("node is not a map of any")
		}
		nodeType, ok := node["type"].(string)
		if !ok || nodeType != "data" {
			if nodeType == "error" {
				errorObj := node["error"].(map[string]any)
				errorMsg := errorObj["message"].(string)

				return nil, errors.New(errorMsg)
			}
			continue
		}
		data, ok := node["data"].([]any)
		if !ok {
			return nil, errors.New("data field is not an array of any")
		}
		nig, err := processObjects(data, 0)
		if err != nil {
			return nil, err
		}
		raw = nig
	}

	jsonData, err := json.Marshal(raw)
	if err != nil {
		panic(err)
	}

	var zzzInfo ZZZInfo
	err = json.Unmarshal(jsonData, &zzzInfo)
	if err != nil {
		panic(err)
	}

	return &zzzInfo, nil
}

func HSRHandler(ctx *MessageContext) {
	maps := data.GetData()
	args := ctx.Parser.GetArgs()
	if len(args) == 0 {
		ctx.Instance.Reply("Berikan UID HSR", true)
		return
	}

	uid, err := strconv.ParseInt(args[0].Content, 10, 0)
	if err != nil {
		ctx.Instance.Reply("UID mungkin bukan bilangan yang valid", true)
		return
	}

	info, err := getHSRData(int(uid))
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Terjadi kesalahan saat mengambil data: %s", err), true)
		return
	}

	msg := "*Info HSR*\n==========\n"
	msg += fmt.Sprintf("UID: %d\n", info.DetailInfo.UID)
	msg += fmt.Sprintf("Nickname: %s\n", info.DetailInfo.Nickname)
	msg += fmt.Sprintf("Signature: %s\n", info.DetailInfo.Signature)
	msg += fmt.Sprintf("Level: %d\n", info.DetailInfo.Level)
	msg += fmt.Sprintf("World Level: %d\n==========\nRecords\n==========\n", info.DetailInfo.WorldLevel)
	msg += fmt.Sprintf("Achievement: %d\n", info.DetailInfo.RecordInfo.AchievementCount)
	msg += fmt.Sprintf("Book: %d\n", info.DetailInfo.RecordInfo.BookCount)
	msg += fmt.Sprintf("Avatar: %d\n", info.DetailInfo.RecordInfo.AvatarCount)
	msg += fmt.Sprintf("Equipment: %d\n", info.DetailInfo.RecordInfo.EquipmentCount)
	msg += fmt.Sprintf("Music: %d\n", info.DetailInfo.RecordInfo.MusicCount)
	msg += fmt.Sprintf("Relic: %d\n==========\n", info.DetailInfo.RecordInfo.RelicCount)

	if len(info.DetailInfo.AvatarDetails) == 0 {
		msg += "No characters in the showcase 💔"
	} else {
		msg += "*Character showcase(s)*\n==========\n"
		for _, avatar := range info.DetailInfo.AvatarDetails {
			avatarName := maps.HSR[strconv.FormatInt(int64(avatar.AvatarID), 10)]
			msg += fmt.Sprintf("%s (Level %d)\n", avatarName, avatar.Level)
		}
		msg += "=========="
	}

	ctx.Instance.Reply(msg, true)
}

func GIHandler(ctx *MessageContext) {
	maps := data.GetData()
	args := ctx.Parser.GetArgs()
	if len(args) == 0 {
		ctx.Instance.Reply("Berikan UID Genshin", true)
		return
	}

	uid, err := strconv.ParseInt(args[0].Content, 10, 0)
	if err != nil {
		ctx.Instance.Reply("UID mungkin bukan bilangan yang valid", true)
		return
	}

	info, err := getGIData(int(uid))
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Terjadi kesalahan saat mengambil data: %s", err), true)
		return
	}

	msg := "*Info Genshin*\n==========\n"
	msg += fmt.Sprintf("UID: %s\n", info.UID)
	msg += fmt.Sprintf("Nickname: %s\n", info.PlayerInfo.Nickname)
	msg += fmt.Sprintf("Level: %d\n", info.PlayerInfo.Level)
	msg += fmt.Sprintf("Signature: %s\n", info.PlayerInfo.Signature)
	msg += fmt.Sprintf("World level: %d\n", info.PlayerInfo.WorldLevel)
	msg += fmt.Sprintf("Achievement total: %d\n==========\n", info.PlayerInfo.AchievementTotal)

	if len(info.PlayerInfo.AvatarInfoList) == 0 {
		msg += "No characters in the showcase 💔"
	} else {
		msg += "*Character showcase(s)*\n==========\n"
		for _, avatar := range info.PlayerInfo.AvatarInfoList {
			avatarName := maps.Genshin[strconv.FormatInt(int64(avatar.AvatarID), 10)]
			msg += fmt.Sprintf("%s (Level %d)\n", avatarName, avatar.Level)
		}
		msg += "=========="
	}

	ctx.Instance.Reply(msg, true)
}

func ZZZHandler(ctx *MessageContext) {
	maps := data.GetData()
	args := ctx.Parser.GetArgs()
	if len(args) == 0 {
		ctx.Instance.Reply("Berikan UID ZZZ", true)
		return
	}

	uid, err := strconv.ParseInt(args[0].Content, 10, 0)
	if err != nil {
		ctx.Instance.Reply("UID mungkin bukan bilangan yang valid", true)
		return
	}

	info, err := getZZZData(int(uid))
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Terjadi kesalahan saat mengambil data: %s", err), true)
		return
	}

	msg := "*Info ZZZ*\n==========\n"
	msg += fmt.Sprintf("UID: %d\n", info.PlayerInfo.SocialDetail.ProfileDetail.UID)
	msg += fmt.Sprintf("Nickname: %s\n", info.PlayerInfo.SocialDetail.ProfileDetail.Nickname)
	msg += fmt.Sprintf("Level: %d\n", info.PlayerInfo.SocialDetail.ProfileDetail.Level)
	msg += fmt.Sprintf("Desc: %s\n==========\n", info.PlayerInfo.SocialDetail.Desc)

	if len(info.PlayerInfo.ShowcaseDetail.AvatarList) == 0 {
		msg += "No characters in the showcase 💔"
	} else {
		msg += "*Character showcase(s)*\n==========\n"
		for _, avatar := range info.PlayerInfo.ShowcaseDetail.AvatarList {
			avatarName := maps.ZZZ[strconv.FormatInt(int64(avatar.ID), 10)]
			msg += fmt.Sprintf("%s (Level %d)\n", avatarName, avatar.Level)
		}
		msg += "=========="
	}

	ctx.Instance.Reply(msg, true)
}
