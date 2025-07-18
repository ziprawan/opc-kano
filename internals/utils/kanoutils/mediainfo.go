package kanoutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"math"
	"os"
	"strconv"

	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/webp"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

// I DON'T KNOW HOW TO NAME THINGS LMAO

type FFProbeResult struct {
	Streams []Stream `json:"streams,omitempty"`
}

type Stream struct {
	CodecType *string `json:"codec_type,omitempty"`
	Duration  *string `json:"duration,omitempty"`
	Height    *int    `json:"height,omitempty"`
	Width     *int    `json:"width,omitempty"`
}

// Generate video infos that contains video's duration length, height, and thumbnail (not encoded in base64 yet)
func GenerateVideoInfo(videoBytes []byte) (*KanoVideoInfo, error) {
	if len(videoBytes) == 0 {
		return nil, ErrVideoLengthIsZero
	}

	data, err := ffmpeg_go.ProbeReader(bytes.NewReader(videoBytes))
	if err != nil {
		return nil, err
	}

	fmt.Println(data)
	var ffprobeRet FFProbeResult
	err = json.Unmarshal([]byte(data), &ffprobeRet)
	if err != nil {
		return nil, err
	}

	if ffprobeRet.Streams == nil {
		return nil, ErrVideoNoStreams
	}

	var info KanoVideoInfo
	for _, stream := range ffprobeRet.Streams {
		fmt.Printf("%+v\n", stream)
		if stream.CodecType != nil && *stream.CodecType != "video" {
			continue
		}

		if stream.Duration != nil {
			duration, _ := strconv.ParseFloat(*stream.Duration, 32)
			info.Duration = uint32(math.Round(duration))
		}
		if stream.Height != nil {
			info.Height = uint32(*stream.Height)
		}
		if stream.Width != nil {
			info.Width = uint32(*stream.Width)
		}
	}

	fmt.Printf("%+v\n", info)

	buf := bytes.NewBuffer(nil)
	err = ffmpeg_go.
		Input("pipe:0").
		Filter("select", ffmpeg_go.Args{fmt.Sprintf("gte(n,%d)", 0)}).
		Output("pipe:", ffmpeg_go.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithInput(bytes.NewReader(videoBytes)).
		WithOutput(buf, os.Stdout).
		Run()
	if err != nil {
		return nil, err
	}

	if buf != nil {
		thumb, err := GenerateThumbnail(buf.Bytes())
		if err != nil {
			fmt.Println("Something went wrong when generating thumbnail: ", err)
		} else {
			info.JPEGThumbnail = thumb
		}
	}

	return &info, nil
}

func GenerateImageInfo(imgBytes []byte) (*KanoImageInfo, error) {
	decoded, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		decoded, err = webp.Decode(bytes.NewReader(imgBytes), &decoder.Options{})
		if err != nil {
			return nil, err
		}
	}
	bounds := decoded.Bounds()

	var info KanoImageInfo
	info.Width = uint32(bounds.Dx())
	info.Height = uint32(bounds.Dy())
	thumb, err := GenerateThumbnail(imgBytes)
	if err != nil {
		fmt.Println(err)
	} else {
		info.JPEGThumbnail = thumb
	}

	return &info, nil
}
