package ytdlpbind

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lrstanley/go-ytdlp"
)

func Call(url string) (YtDlpJSON, error) {
	yt := ytdlp.New().DumpJSON().Cookies("secrets/cookies.txt")
	res, err := yt.Run(context.TODO(), url)
	if err != nil {
		if res == nil {
			return YtDlpJSON{}, err
		}
		return YtDlpJSON{}, fmt.Errorf("yt-dlp failed: %s", res.Stderr)
	}
	outJson := res.Stdout

	var ytdlpJson YtDlpJSON
	err = json.Unmarshal([]byte(outJson), &ytdlpJson)
	if err != nil {
		return YtDlpJSON{}, fmt.Errorf("unmarshal failed: %s", err)
	}

	return ytdlpJson, nil
}
