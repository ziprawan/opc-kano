package message

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kano/internals/utils/kanoutils"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"

	ffmpeg "github.com/u2takey/ffmpeg-go"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

var DownloaderMan = CommandMan{
	Name:     "downloader - Pengunduh",
	Synopsis: []string{"downloader LINK"},
	Description: []string{
		"Unduh media dari platform Instagram (dukungan platform lain akan mendatang).",
		"*LINK* (Wajib)\n{SPACE}URL atau link media yang ingin diunduh.",
	},
	Source:  "downloader.go",
	SeeAlso: []SeeAlso{},
}

type DownloaderResponseData struct {
	AudioUrl     *string  `json:"audio_url"`
	Author       string   `json:"author"`
	Caption      string   `json:"caption"`
	EncryptedUrl string   `json:"encryptedUrl"`
	HDUrl        []string `json:"hd_url"`
	Latency      float64  `json:"latency"`
	MediaCount   int      `json:"media_count"`
	Platform     string   `json:"platform"`
	Thumbnail    string   `json:"thumbnail"`
	Url          []string `json:"url"`

	Message string // When the response is not 200
}
type DownloaderResponse struct {
	Data      DownloaderResponseData `json:"data"`
	IsCached  bool                   `json:"is_cached"`
	Raw       any                    `json:"raw"`
	Status    int                    `json:"status"`
	Timestamp string                 `json:"timestamp"`
}

type MediaType string

var MediaImage MediaType = "image"
var MediaVideo MediaType = "video"

type DownloadMedia struct {
	Type MediaType
	Url  string
}

// Instagram utils
func instagramParseMapType(input map[string]any) ([]DownloadMedia, error) {
	fmt.Println("instagramParseMapType")
	typename, ok := input["__typename"]
	if ok {
		// This one come from GraphQL
		if typename == "GraphSidecar" {
			edge, ok := input["edge_sidecar_to_children"].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("instagram: Failed to get raw.edge_sidecar_to_children, type: %T", input["edge_sidecar_to_children"])
			}

			edges, ok := edge["edges"].([]any)
			if !ok {
				return nil, fmt.Errorf("instagram: Failed to get raw.edge_sidecar_to_children.edges, type: %T", edge["edges"])
			}

			var medias []DownloadMedia
			for _, e := range edges {
				if edg, ok := e.(map[string]any); ok {
					if node, ok := edg["node"].(map[string]any); ok {
						res, err := instagramParseMapType(node)
						if err != nil {
							fmt.Println("Errored at instagramParseMapType(node)", err)
							continue
						}

						medias = append(medias, res...)
					}
				}
			}

			return medias, nil
		} else if typename == "GraphImage" {
			var media DownloadMedia
			displayUrl, ok := input["display_url"].(string)
			if !ok {
				return nil, fmt.Errorf("instagram: Failed to get raw.display_url, type: %T", displayUrl)
			}
			media.Url = displayUrl
			media.Type = MediaImage

			return []DownloadMedia{media}, nil
		} else if typename == "GraphVideo" {
			var media DownloadMedia
			videoUrl, ok := input["video_url"].(string)
			if !ok {
				return nil, fmt.Errorf("instagram: Failed to get raw.video_url, type: %T", videoUrl)
			}
			media.Type = MediaVideo
			media.Url = videoUrl

			return []DownloadMedia{media}, nil
		} else {
			return nil, fmt.Errorf("instagram: Unknown __typename: %s", typename)
		}
	} else {
		// This should come from another API
		var medias []DownloadMedia

		urls, ok := input["url"].([]any)
		if !ok {
			return nil, fmt.Errorf("instagram: Failed to get input.url")
		}
		for _, u := range urls {
			var media DownloadMedia

			if url, ok := u.(map[string]any); ok {
				if uuu, ok1 := url["url"].(string); ok1 {
					media.Url = uuu

					ttt, ok := url["type"].(string)
					if ok {
						if ttt == "mp4" {
							media.Type = MediaVideo
						} else {
							media.Type = MediaImage
						}
					}

					medias = append(medias, media)
				}
			}
		}
		return medias, nil
	}
}
func instagramParseRaw(raw any) ([]DownloadMedia, error) {
	if rawMap, ok := raw.(map[string]any); ok {
		return instagramParseMapType(rawMap)
	} else if rawArr, ok := raw.([]any); ok {
		var medias []DownloadMedia
		for _, a := range rawArr {
			if arr, ok := a.(map[string]any); ok {
				res, err := instagramParseMapType(arr)
				if err != nil {
					return nil, err
				}
				medias = append(medias, res...)
			}
		}

		return medias, nil
	}

	return nil, fmt.Errorf("instagram: Invalid raw data type: %T", raw)
}

// API Call
func getRawDownloaderData(requestUrl string) (*DownloaderResponse, error) {
	API_URL, _ := url.Parse("https://tikapi11-e3d106ab50c7.herokuapp.com/api") // This shouldn't produce any error since our url is constant
	query := API_URL.Query()
	query.Set("url", requestUrl)
	API_URL.RawQuery = query.Encode()

	request, err := http.NewRequest("GET", API_URL.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("downloader: Server responded with status %d", response.StatusCode)
	}

	responseBytes, _ := io.ReadAll(response.Body)
	file, _ := os.CreateTemp("tmp", "*.json")
	if file != nil {
		file.Write(responseBytes)
		file.Close()
	}

	var parsedResponse DownloaderResponse
	err = json.Unmarshal(responseBytes, &parsedResponse)
	if err != nil {
		return nil, err
	}

	return &parsedResponse, nil
}

// Uploader
func uploadFromUrl(ctx *MessageContext, reqUrl string, media MediaType) (whatsmeow.UploadResponse, []byte, error) {
	fmt.Println("Uploading", reqUrl)
	if ctx == nil {
		return whatsmeow.UploadResponse{}, nil, fmt.Errorf("ctx is nil")
	}

	var mediaType whatsmeow.MediaType
	if media == MediaImage {
		mediaType = whatsmeow.MediaImage
	} else if media == MediaVideo {
		mediaType = whatsmeow.MediaVideo
	}

	resp, err := http.Get(reqUrl)
	if err != nil {
		return whatsmeow.UploadResponse{}, nil, err
	}
	defer resp.Body.Close()

	contentBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return whatsmeow.UploadResponse{}, nil, err
	}

	upResp, err := ctx.Instance.Upload(contentBytes, mediaType)
	return upResp, contentBytes, err
}

// Message builder
func buildMessageFromMedia(ctx *MessageContext, medias []DownloadMedia) ([]*waE2E.Message, error) {
	messages := []*waE2E.Message{}

	for _, media := range medias {
		resp, contentBytes, err := uploadFromUrl(ctx, media.Url, media.Type)
		if err != nil {
			return messages, err
		}

		if media.Type == MediaImage {
			imgMsg := waE2E.ImageMessage{
				URL:           &resp.URL,
				DirectPath:    &resp.DirectPath,
				MediaKey:      resp.MediaKey,
				FileEncSHA256: resp.FileEncSHA256,
				FileSHA256:    resp.FileSHA256,
				FileLength:    &resp.FileLength,
				Mimetype:      proto.String("image/jpeg"),
				ContextInfo: &waE2E.ContextInfo{
					PairedMediaType: waE2E.ContextInfo_NOT_PAIRED_MEDIA.Enum(),
				},
			}

			imgInfo, err := kanoutils.GenerateImageInfo(contentBytes)
			if err != nil {
				fmt.Println("Failed to generate image info:", err)
			} else {
				imgMsg.Height = &imgInfo.Height
				imgMsg.Width = &imgInfo.Width
				imgMsg.JPEGThumbnail = imgInfo.JPEGThumbnail
			}

			messages = append(messages, &waE2E.Message{
				ImageMessage: &imgMsg,
			})
		} else if media.Type == MediaVideo {
			vidMsg := waE2E.VideoMessage{
				URL:           &resp.URL,
				DirectPath:    &resp.DirectPath,
				MediaKey:      resp.MediaKey,
				FileEncSHA256: resp.FileEncSHA256,
				FileSHA256:    resp.FileSHA256,
				FileLength:    &resp.FileLength,
				Mimetype:      proto.String("video/mp4"),
				ContextInfo: &waE2E.ContextInfo{
					PairedMediaType: waE2E.ContextInfo_NOT_PAIRED_MEDIA.Enum(),
				},
			}

			// Write temp file
			file, err := os.CreateTemp("tmp", "*.mp4")
			if err != nil {
				fmt.Println("Failed to create temp file:", err)
			} else {
				defer os.Remove(file.Name())
				file.Write(contentBytes)
				file.Close()

				buf := bytes.NewBuffer(nil)
				err := ffmpeg.Input(file.Name()).
					Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", 0)}).
					Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
					WithOutput(buf, os.Stdout).
					Run()
				if err != nil {
					fmt.Println("Failed to run ffmpeg:", err)
				} else {
					imgBytes := buf.Bytes()
					imgInfo, err := kanoutils.GenerateImageInfo(imgBytes)
					if err != nil {
						fmt.Println("Failed to generate video thumb:", err)
					} else {
						vidMsg.Height = &imgInfo.Height
						vidMsg.Width = &imgInfo.Width
						vidMsg.JPEGThumbnail = imgInfo.JPEGThumbnail
					}
				}

				data, err := ffmpeg.ProbeReader(bytes.NewReader(contentBytes))
				if err != nil {
					fmt.Println("instagram: Failed to get video info: ", err)
				} else {
					var ffprobeRet *FFProbeResult
					err = json.Unmarshal([]byte(data), &ffprobeRet)
					if err != nil {
						fmt.Println("Failed to parse ffprobe JSON result", err)
					}

					for _, stream := range ffprobeRet.Streams {
						if stream.CodecType != nil && *stream.CodecType == "video" {
							if stream.Duration != nil {
								duration, _ := strconv.ParseFloat(*stream.Duration, 32)
								duration = math.Ceil(duration)
								vidMsg.Seconds = proto.Uint32(uint32(duration))
							}
						}
					}
				}
			}

			messages = append(messages, &waE2E.Message{
				VideoMessage: &vidMsg,
			})
		}
	}

	return messages, nil
}

// MAIN FUNCTION
func DownloaderHandler(ctx *MessageContext) {
	args := ctx.Parser.GetArgs()
	if len(args) == 0 {
		ctx.Instance.Reply("Berikan URL", true)
		return
	}
	queryUrl := args[0].Content
	data, err := getRawDownloaderData(queryUrl)
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Failed to get data:\n%s", err), true)
		return
	}

	if data.Data.Platform != "instagram" {
		ctx.Instance.Reply(fmt.Sprintf("Unsupported platform: %s", data.Data.Platform), true)
		return
	}

	parsed, err := instagramParseRaw(data.Raw)
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Error wak: %s", err), true)
		return
	}

	ctx.Instance.Client.SendMessage(context.Background(), *ctx.Instance.ChatJID(), ctx.Instance.Client.BuildReaction(
		*ctx.Instance.ChatJID(),
		*ctx.Instance.SenderJID(),
		*ctx.Instance.ID(),
		"✅",
	))

	albumMessage := &waE2E.AlbumMessage{
		ExpectedImageCount: proto.Uint32(0),
		ExpectedVideoCount: proto.Uint32(0),
	}
	for _, p := range parsed {
		if p.Type == MediaImage {
			// *albumMessage.ExpectedImageCount += 1
		} else if p.Type == MediaVideo {
			// *albumMessage.ExpectedVideoCount += 1
		}
	}

	msgs, err := buildMessageFromMedia(ctx, parsed)
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Failed to build message: %s", err), true)
		return
	}

	sent, err := ctx.Instance.Client.SendMessage(context.Background(), *ctx.Instance.ChatJID(), &waE2E.Message{
		AlbumMessage: albumMessage,
	})
	if err != nil {
		fmt.Println("Failed to send album message")
		return
	}

	if len(msgs) == 0 {
		ctx.Instance.Reply("Kosong", true)
	}

	for idx, msg := range msgs {
		msg.MessageContextInfo = &waE2E.MessageContextInfo{
			MessageAssociation: &waE2E.MessageAssociation{
				AssociationType: waE2E.MessageAssociation_MEDIA_ALBUM.Enum(),
				ParentMessageKey: &waCommon.MessageKey{
					RemoteJID: proto.String(sent.Sender.String()),
					FromMe:    proto.Bool(true),
					ID:        proto.String(sent.ID),
				},
			},
		}

		if idx == 0 {
			caption := fmt.Sprintf("%s\n\nAuthor: @%s", data.Data.Caption, data.Data.Author)
			if msg.VideoMessage != nil {
				msg.VideoMessage.Caption = &caption
			} else if msg.ImageMessage != nil {
				msg.ImageMessage.Caption = &caption
			}
		}
		ctx.Instance.Client.SendMessage(context.Background(), *ctx.Instance.ChatJID(), msg)
	}
}
