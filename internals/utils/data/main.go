package data

import (
	"encoding/json"
	"os"
	"sync"
)

type L map[string]string

type LocsData struct {
	Genshin L
	HSR     L
	ZZZ     L
}

var (
	data *LocsData
	once sync.Once
)

func LoadAllMaps() {
	once.Do(func() {
		parsed := LocsData{}

		content, err := os.ReadFile("data/gi_chars.json")
		if err != nil {
			panic(err)
		}

		var maps L
		err = json.Unmarshal(content, &maps)
		if err != nil {
			panic(err)
		}

		parsed.Genshin = maps

		content, err = os.ReadFile("data/hsr_chars.json")
		if err != nil {
			panic(err)
		}

		maps = L{}
		err = json.Unmarshal(content, &maps)
		if err != nil {
			panic(err)
		}

		parsed.HSR = maps

		content, err = os.ReadFile("data/zzz_chars.json")
		if err != nil {
			panic(err)
		}

		maps = L{}
		err = json.Unmarshal(content, &maps)
		if err != nil {
			panic(err)
		}

		parsed.ZZZ = maps

		data = &parsed
	})
}

func GetData() LocsData {
	if data == nil {
		LoadAllMaps()
	}

	return *data
}
