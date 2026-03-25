package instagram

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type igDashInfo struct {
	Video []igDashInfo__Video
	Audio []igDashInfo__Audio
}

type igDashInfo__Video struct {
	Id     string
	Url    string
	Length int

	Width        int
	Height       int
	FrameRate    float64
	QualityLabel string
}

type igDashInfo__Audio struct {
	Id     string
	Url    string
	Length int

	Bitrate int
}

func parseDashInfo(dashManifest string) igDashInfo {
	var dash igDashInfo
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(dashManifest))
	if err != nil {
		return dash
	}

	adaptations := doc.Find("AdaptationSet")
	for _, adaptation := range adaptations.EachIter() {
		contentType := adaptation.AttrOr("contenttype", "")

		switch contentType {
		case "video":
			representations := adaptation.Find("Representation")
			for _, repr := range representations.EachIter() {
				id := repr.AttrOr("id", "")
				if id == "" {
					continue
				}
				length, err := strconv.ParseUint(repr.AttrOr("fbcontentlength", ""), 10, 0)
				if err != nil {
					continue
				}
				width, err := strconv.ParseUint(repr.AttrOr("width", ""), 10, 0)
				if err != nil {
					continue
				}
				height, err := strconv.ParseUint(repr.AttrOr("height", ""), 10, 0)
				if err != nil {
					continue
				}
				frameRate := repr.AttrOr("framerate", "")
				if frameRate == "" {
					continue
				}
				spl := strings.Split(frameRate, "/")
				num, err1 := strconv.ParseUint(spl[0], 10, 0)
				dem, err2 := strconv.ParseUint(spl[1], 10, 0)
				if err1 != nil || err2 != nil {
					continue
				}
				fps := float64(num) / float64(dem)
				label := repr.AttrOr("fbqualitylabel", "")
				if label == "" {
					continue
				}
				url := strings.TrimSpace(repr.Text())

				dash.Video = append(dash.Video, igDashInfo__Video{
					Id:           id,
					Url:          url,
					Length:       int(length),
					Width:        int(width),
					Height:       int(height),
					FrameRate:    fps,
					QualityLabel: label,
				})
			}
		case "audio":
			representations := adaptation.Find("Representation")
			for _, repr := range representations.EachIter() {
				id := repr.AttrOr("id", "")
				if id == "" {
					continue
				}
				length, err := strconv.ParseUint(repr.AttrOr("fbcontentlength", ""), 10, 0)
				if err != nil {
					continue
				}
				bitrate, err := strconv.ParseUint(repr.AttrOr("fbavgbitrate", ""), 10, 0)
				if err != nil {
					continue
				}
				url := strings.TrimSpace(repr.Text())

				dash.Audio = append(dash.Audio, igDashInfo__Audio{
					Id:      id,
					Url:     url,
					Length:  int(length),
					Bitrate: int(bitrate),
				})
			}
		default:
			continue
		}
	}

	return dash
}
