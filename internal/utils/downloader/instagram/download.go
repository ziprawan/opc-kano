package instagram

import (
	"encoding/json"
	"fmt"
	"io"
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

func downloadMedia(url string, useTempFile bool) (io.ReadCloser, error) {
	cli := http.Client{}
	req, err := instagramCreateReq(url)
	if err != nil {
		return nil, fmt.Errorf("failed to create media request: %s", err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do media request: %s", err)
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("server responded with status %s", resp.Status)
	}

	if useTempFile {
		tmpFile, err := os.CreateTemp("", "temp_instagram_*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file: %s", err)
		}
		_, err = io.Copy(tmpFile, resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to copy buffer into the temp file: %s", err)
		}
		resp.Body.Close()

		return tmpFile, nil
	} else {
		return resp.Body, nil
	}
}

func mergeVideoAudio(vidReader, audReader io.ReadCloser, vidId, audId string) (io.ReadCloser, error) {
	tempdir, err := os.MkdirTemp("", "downloader_instagram_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempdir)
	inputVid := fmt.Sprintf("%s/%s.mp4", tempdir, vidId)
	inputAud := fmt.Sprintf("%s/%s.mp4", tempdir, audId)
	output := fmt.Sprintf("%s/%s_%s.mp4", tempdir, vidId, audId)

	vidFile, err := os.OpenFile(inputVid, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer vidFile.Close()
	_, err = io.Copy(vidFile, vidReader)
	if err != nil {
		return nil, err
	}
	vidFile.Close()

	audFile, err := os.OpenFile(inputAud, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer audFile.Close()
	_, err = io.Copy(audFile, audReader)
	if err != nil {
		return nil, err
	}
	audFile.Close()

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

	fileReader, err := os.OpenFile(output, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}

	return fileReader, nil
}

func processMedia(ctx *types.DownloaderContext, media igParsed_Media) error {
	if media.IsVideo {
		vidBytes, err := downloadMedia(media.Url, false)
		if err != nil {
			return err
		}

		if media.AudioUrl != "" {
			audBytes, err := downloadMedia(media.AudioUrl, false)
			if err != nil {
				return err
			}

			mergedBytes, err := mergeVideoAudio(vidBytes, audBytes, media.VideoId, media.AudioId)
			ctx.AddMedia(mergedBytes, true, "video/mp4", media.Dimensions.Height, media.Dimensions.Width, media.Duration)
		} else {
			ctx.AddMedia(vidBytes, true, "video/mp4", media.Dimensions.Height, media.Dimensions.Width, media.Duration)
		}
	} else {
		imgReader, err := downloadMedia(media.Url, true)
		if err != nil {
			return err
		}
		defer imgReader.Close()
		imgSeeker, ok := imgReader.(io.ReadSeekCloser)
		if !ok {
			return fmt.Errorf("unable to infer imgReader into io.ReadSeekCloser")
		}
		first512Bytes := make([]byte, 512)
		_, err = imgSeeker.Read(first512Bytes)
		if err != nil {
			return fmt.Errorf("failed to read first 512 bytes for content type detection: %s", err)
		}

		ctx.AddMedia(imgReader, false, http.DetectContentType(first512Bytes), media.Dimensions.Height, media.Dimensions.Width, 0)
	}

	return nil
}
