package instagram

import (
	"encoding/json"
	"fmt"
	"io"
	"kano/internal/config"
	"kano/internal/utils/downloader/types"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type instagramRuling struct {
	Title       string
	Description string
	Status      string
}

const instagram_encoding = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

var instagram_table = map[uint8]int{}

func instagramIdToPk(id string) uint64 {
	if len(instagram_table) == 0 {
		for i := range len(instagram_encoding) {
			instagram_table[instagram_encoding[i]] = i
		}
	}

	result, base := uint64(0), uint64(len(instagram_table))
	for i := range len(id) {
		result = (result * base) + uint64(instagram_table[id[i]])
	}

	return result
}

func instagramApiCheck(id string) (apiCheck instagramRuling, csrftoken string) {
	// Create request
	cli := http.Client{}
	checkReq, err := instagramCreateReq(
		fmt.Sprintf(
			"https://i.instagram.com/api/v1/web/get_ruling_for_content/?content_type=MEDIA&target_id=%d",
			instagramIdToPk(id),
		),
	)
	if err != nil {
		return
	}

	// Do the request and parse it
	// If it is returned errors, just return empty struct
	checkResp, err := cli.Do(checkReq)
	if err == nil {
		// Parse the response
		json.NewDecoder(checkResp.Body).Decode(&apiCheck)
		checkResp.Body.Close()

		// Get csrf token from Set-Cookie
		setCookies := checkResp.Cookies()
		for _, cookie := range setCookies {
			if cookie == nil {
				continue
			}

			if cookie.Name == "csrftoken" {
				csrftoken = cookie.Value
			}
		}
	}

	return
}

func instagramCreateReq(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return req, err
	}

	req.Header.Set("X-IG-App-ID", "936619743392459")
	req.Header.Set("X-ASBD-ID", "198387")
	req.Header.Set("X-IG-WWW-Claim", "0")
	req.Header.Set("Origin", "https://www.instagram.com")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", config.USER_AGENT)

	return req, nil
}

func Download(ctx *types.DownloaderContext, igUrl *url.URL) error {
	cli := http.Client{}

	id := ""
	paths := strings.Split(igUrl.Path, "/")
	if len(paths) > 2 {
		switch contentType := paths[1]; contentType {
		case "p", "reels", "reel", "tv":
			if paths[2] == "audio" {
				return fmt.Errorf("downloading reels audio only is unsupported")
			} else {
				id = paths[2]
			}
		default:
			return fmt.Errorf("unsupported content type %q", contentType)
		}
	}
	if id == "" {
		return fmt.Errorf("unable to get content id from url, is the format changed?")
	}

	_, csrftoken := instagramApiCheck(id)

	variables := map[string]string{"shortcode": id}
	variablesMar, _ := json.Marshal(variables)

	infoUrl, _ := url.Parse("https://www.instagram.com/graphql/query/")
	q := infoUrl.Query()
	q.Add("doc_id", "8845758582119845")
	q.Add("variables", string(variablesMar))
	infoUrl.RawQuery = q.Encode()

	infoReq, _ := instagramCreateReq(infoUrl.String())
	infoReq.Header.Set("X-Requested-With", "XMLHttpRequest")
	infoReq.Header.Set("Referer", igUrl.String())
	if csrftoken != "" {
		infoReq.Header.Set("X-CSRF-Token", csrftoken)
	}

	infoResp, err := cli.Do(infoReq)
	if err != nil {
		return fmt.Errorf("failed to get general info: %s", err)
	}
	defer infoResp.Body.Close()

	respBody, err := io.ReadAll(infoResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read info response body: %s", err)
	}

	// Response prefixed with for(;;);
	// Just find the first '{'
	startIdx := slices.Index(respBody, byte('{'))
	parsed, err := parseGeneralInfo(respBody[startIdx:])
	if err != nil {
		return fmt.Errorf("failed to parse general info: %s", err)
	}
	ctx.SetCaption(parsed.caption)
	for _, media := range parsed.medias {
		err = processMedia(ctx, media)
		if err != nil {
			return fmt.Errorf("processMedia: %s", err)
		}
	}

	return nil
}

func downloadMedia(url string) ([]byte, error) {
	cli := http.Client{}
	req, err := instagramCreateReq(url)
	if err != nil {
		return nil, fmt.Errorf("failed to create media request: %s", err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do media request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("server responded with status %s", resp.Status)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %s", err)
	}

	return respBytes, nil
}

func mergeVideoAudio(vidBytes, audBytes []byte, vidId, audId string) ([]byte, error) {
	tempdir, err := os.MkdirTemp("", "downloader_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempdir)
	inputVid := fmt.Sprintf("%s/%s.mp4", tempdir, vidId)
	inputAud := fmt.Sprintf("%s/%s.mp4", tempdir, audId)
	output := fmt.Sprintf("%s/%s_%s.mp4", tempdir, vidId, audId)

	defer os.Remove(inputVid)
	defer os.Remove(inputAud)
	defer os.Remove(output)

	err = os.WriteFile(inputVid, vidBytes, 0644)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(inputAud, audBytes, 0644)
	if err != nil {
		return nil, err
	}

	in1 := ffmpeg.Input(inputVid)
	in2 := ffmpeg.Input(inputAud)

	err = ffmpeg.Output(
		[]*ffmpeg.Stream{in1, in2},
		output,
		ffmpeg.KwArgs{"c:v": "copy", "c:a": "copy"},
	).Run()
	if err != nil {
		return nil, err
	}

	mergedBytes, err := os.ReadFile(output)
	if err != nil {
		return nil, err
	}

	return mergedBytes, nil
}

func processMedia(ctx *types.DownloaderContext, media igParsed_Media) error {
	if media.IsVideo {
		vidBytes, err := downloadMedia(media.Url)
		if err != nil {
			return err
		}

		if media.AudioUrl != "" {
			audBytes, err := downloadMedia(media.AudioUrl)
			if err != nil {
				return err
			}

			mergedBytes, err := mergeVideoAudio(vidBytes, audBytes, media.VideoId, media.AudioId)
			ctx.AddMedia(mergedBytes, true, media.Dimensions.Height, media.Dimensions.Width, media.Duration)
		} else {
			ctx.AddMedia(vidBytes, true, media.Dimensions.Height, media.Dimensions.Width, media.Duration)
		}
	} else {
		imgBytes, err := downloadMedia(media.Url)
		if err != nil {
			return err
		}

		ctx.AddMedia(imgBytes, false, media.Dimensions.Height, media.Dimensions.Width, 0)
	}

	return nil
}
