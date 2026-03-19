package handles

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"kano/internal/config"
	"kano/internal/utils/image/png"
	"kano/internal/utils/messageutil"
	"kano/internal/utils/sticker"
	"net/http"
	"net/url"
	"os"
	"strings"

	imagepng "image/png"
	imageutil "kano/internal/utils/image"

	"github.com/PuerkitoBio/goquery"
	"github.com/kettek/apng"
	"github.com/kolesa-team/go-webp/webp"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

type LineStickerData struct {
	Type              string `json:"type"`
	ID                string `json:"id"`
	StaticUrl         string `json:"staticUrl"`
	FallbackStaticUrl string `json:"fallbackStaticUrl"`
	AnimationUrl      string `json:"animationUrl"`
	PopupUrl          string `json:"popupUrl,omitempty"`
	SoundUrl          string `json:"soundUrl,omitempty"`
}

func StkLineHandler(c *messageutil.MessageContext) error {
	args := c.Parser.Args
	if len(args) == 0 {
		c.QuoteReply("Give the sticker store url")
		return nil
	}

	givenUrl := args[0].Content.Data
	u, err := url.Parse(givenUrl)
	if err != nil {
		c.QuoteReply("Given url is not parsable")
		return nil
	}

	if u.Scheme != "https" {
		c.QuoteReply(`Given url scheme is not "https"`)
		return nil
	}
	if u.Host != "store.line.me" {
		c.QuoteReply(`Given url host is not "store.line.me"`)
		return nil
	}
	paths := strings.Split(u.Path, "/")
	if len(paths) <= 1 {
		c.QuoteReply("Given url is the store home")
		return nil
	}
	if paths[1] != "stickershop" && paths[1] != "emojishop" {
		c.QuoteReply("Shop type is unsupported, ensure the given url is a stickershop or emojishop. Got: %s", paths[1])
		return nil
	}

	itemType := strings.Replace(paths[1], "shop", "", 1)
	if len(paths) < 4 {
		c.QuoteReply("Unable to get sticker/emoji id, is your given url valid?")
		return nil
	}
	itemId := paths[3]

	req, _ := http.NewRequest("GET", givenUrl, nil)
	req.Header.Set("User-Agent", config.USER_AGENT)
	cli := http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		c.QuoteReply("Failed to fetch store page")
		return nil
	}
	defer resp.Body.Close()

	page, err := goquery.NewDocumentFromReader(resp.Body)
	stickerTitle := strings.TrimSpace(page.Find(`[data-test="sticker-name-title"]`).Text())
	if stickerTitle == "" {
		stickerTitle = strings.TrimSpace(page.Find(`[data-test="emoji-name-title"]`).Text())
	}
	if stickerTitle == "" {
		c.QuoteReply("Failed to get sticker/emoji name")
		return nil
	}

	stickerAuthor := strings.TrimSpace(page.Find(`[data-test="sticker-author"]`).Text())
	if stickerAuthor == "" {
		stickerAuthor = strings.TrimSpace(page.Find(`[data-test="emoji-author"]`).Text())
	}
	if stickerAuthor == "" {
		c.QuoteReply("Failed to get sticker/emoji author")
		return nil
	}

	// Something2 about zip file
	var zipBytes bytes.Buffer
	z := zip.NewWriter(&zipBytes)
	addFile := func(filename string, content []byte) error {
		w, err := z.Create(filename)
		if err != nil {
			return err
		}

		if _, err := io.Copy(w, bytes.NewReader(content)); err != nil {
			return err
		}

		return nil
	}
	defer z.Close()

	// Persiapan ya ya ya
	stickerPackId := fmt.Sprintf("line_%s_%s", itemType, itemId)
	stkPackMsg := &waE2E.StickerPackMessage{
		StickerPackID:    proto.String(stickerPackId),
		Name:             proto.String(stickerTitle),
		PackDescription:  proto.String(""),
		StickerPackSize:  proto.Uint64(0),
		TrayIconFileName: proto.String(stickerPackId + ".webp"),
	}

	// Cover image
	cover := page.Find(`[ref="mainImage"]`)
	if cover.Length() != 1 {
		c.QuoteReply("Unable to find sticker/emoji cover: expected 1, got %d", cover.Length())
		return nil
	}
	rawCoverData, ok := cover.Attr("data-preview")
	if !ok {
		c.QuoteReply("Unable to find sticker/emoji cover: cannot get data-preview attribute")
		return nil
	}
	rawCoverData = strings.ReplaceAll(rawCoverData, "&quot;", `"`)
	var coverData LineStickerData
	err = json.Unmarshal([]byte(rawCoverData), &coverData)
	if !ok {
		c.QuoteReply("Unable to find sticker/emoji cover: invalid data")
		return nil
	}

	req2, _ := http.NewRequest("GET", coverData.StaticUrl, nil)
	req2.Header.Set("User-Agent", config.USER_AGENT)
	resp, err = cli.Do(req2)
	if err != nil {
		c.QuoteReply("Unable to find sticker/emoji cover: failed to fetch cover image: %s", err)
		return nil
	}
	defer resp.Body.Close()
	pngImg, err := imagepng.Decode(resp.Body)
	if err != nil {
		c.QuoteReply("Unable to find sticker/emoji cover: failed to read cover image as png: %s", err)
		return nil
	}
	var res bytes.Buffer
	err = webp.Encode(&res, pngImg, nil)
	if err != nil {
		c.QuoteReply("Unable to find sticker/emoji cover: failed to convert cover image to webp: %s", err)
		return nil
	}
	err = addFile(stickerPackId+".webp", res.Bytes())
	if err != nil {
		c.QuoteReply("Failed to add cover image to the zip: %s", err)
		return nil
	}
	*stkPackMsg.StickerPackSize += uint64(res.Len())

	// All stickers
	previews := page.Find(".FnStickerPreviewItem")
	c.QuoteReply("Sticker name: %q\nStickers found: %d\n\nPlease wait...", stickerTitle, previews.Length())
	stkPackMsg.Stickers = make([]*waE2E.StickerPackMessage_Sticker, previews.Length())

	for i, preview := range previews.EachIter() {
		data, ok := preview.Attr("data-preview")
		if !ok {
			c.QuoteReply("Failed to get sticker/emoji data at index %d", i)
			return nil
		}
		data = strings.ReplaceAll(data, "&quot;", "\"")

		var parsed LineStickerData
		err = json.Unmarshal([]byte(data), &parsed)
		if err != nil {
			c.QuoteReply("Failed to parse sticker/emoji data at index %d: %s", i, err)
			return nil
		}

		isAnimated := parsed.Type == "animation"
		fmt.Println(i, isAnimated)
		stkUrl := parsed.StaticUrl
		if isAnimated {
			stkUrl = parsed.AnimationUrl
		}
		if stkUrl == "" {
			c.QuoteReply("Unexpected empty sticker/emoji url at index %d", i)
			return nil
		}

		req, _ := http.NewRequest("GET", stkUrl, nil)
		req.Header.Set("User-Agent", config.USER_AGENT)
		resp, err := cli.Do(req)
		if err != nil {
			c.QuoteReply("Failed to download the sticker/emoji at index %d", i)
			return err
		}

		metadata := sticker.WhatsAppStickerMetadata{
			StickerPackId:        stickerPackId,
			StickerPackName:      stickerTitle,
			StickerPackPublisher: stickerAuthor,
		}

		var stkBytes []byte
		imgBytes, _ := io.ReadAll(resp.Body)
		if isAnimated {
			a, err := apng.DecodeAll(bytes.NewBuffer(imgBytes))
			if err != nil {
				c.QuoteReply("Failed to read sticker/emoji as animated PNG file: %s", err)
				return err
			}
			header, err := png.GetHeader(imgBytes)
			if err != nil {
				c.QuoteReply("Failed to get PNG header: %s", err)
				return err
			}
			rendered := png.RenderAPNGFrames(a.Frames, int(header.Width), int(header.Height))

			webpChunk := imageutil.WebPChunk{
				VP8X: &imageutil.VP8X{HasAnimation: true},
				ANIM: &imageutil.ANIM{
					BackgroundColor: color.RGBA{0, 0, 0, 0},
					LoopCount:       0,
				},
				ANMF: make([]imageutil.ANMF, len(rendered)),
			}

			for i, render := range rendered {
				var w bytes.Buffer
				err = webp.Encode(&w, render.Image, nil)
				if err != nil {
					c.QuoteReply("Failed to encode as webp: %s", err)
					return err
				}
				chunk, err := imageutil.ExtractChunksFromWebP(w.Bytes())
				if err != nil {
					c.QuoteReply("Failed to modify the webp: %s", err)
					return err
				}
				webpChunk.ANMF[i].BlendingMethod = true
				webpChunk.ANMF[i].DisposalMethod = false
				webpChunk.ANMF[i].X = 0
				webpChunk.ANMF[i].Y = 0
				webpChunk.ANMF[i].W = uint(header.Width)
				webpChunk.ANMF[i].H = uint(header.Height)
				webpChunk.ANMF[i].D = uint(render.DelayNum * 1000 / render.DelayDen)
				// Blindly copying EVERYthing
				webpChunk.ANMF[i].ALPH = chunk.ALPH
				webpChunk.ANMF[i].VP8 = chunk.VP8
				webpChunk.ANMF[i].VP8L = chunk.VP8L
			}

			stkBytes, err = imageutil.BuildWebPFromChunks(webpChunk)
			if err != nil {
				c.QuoteReply("Failed to build webp after modification: %s", err)
				return err
			}

			stkBytes, err = imageutil.FixRIFFHeader(stkBytes)
			if err != nil {
				c.QuoteReply("Failed: %s", err)
				return err
			}

			stkBytes, err = sticker.AppendMetadataToSticker(stkBytes, metadata)
			if err != nil {
				c.QuoteReply("Failed to append metadata: %s", err)
				return err
			}
		} else {
			stkBytes, err = sticker.MakeSticker(imgBytes, metadata, false)
			if err != nil {
				c.QuoteReply("Failed to make sticker: %s", err)
				return err
			}
		}

		token := make([]byte, 32)
		rand.Read(token)
		name := base64.StdEncoding.EncodeToString(token) + ".webp"
		name = strings.ReplaceAll(name, "/", "-")

		err = addFile(name, stkBytes)
		if err != nil {
			c.QuoteReply("Failed to add sticker/emoji to a zip: %s", err)
			return err
		}
		*stkPackMsg.StickerPackSize += uint64(len(stkBytes))
		stkPackMsg.Stickers[i] = &waE2E.StickerPackMessage_Sticker{
			FileName:           proto.String(name),
			IsAnimated:         proto.Bool(isAnimated),
			Emojis:             []string{""},
			AccessibilityLabel: proto.String(""),
			IsLottie:           proto.Bool(false),
			Mimetype:           proto.String("image/webp"),
		}

		resp.Body.Close()
	}

	z.Flush()
	z.Close()
	os.WriteFile("test.zip", zipBytes.Bytes(), 0644)

	upResp, err := c.Client.Upload(zipBytes.Bytes(), whatsmeow.MediaStickerPack)
	if err != nil {
		c.QuoteReply("Failed to upload sticker pack into the server: %s", err)
		return err
	}

	stkPackMsg.StickerPackID = proto.String(stickerPackId)
	stkPackMsg.FileLength = &upResp.FileLength
	stkPackMsg.FileSHA256 = upResp.FileSHA256
	stkPackMsg.FileEncSHA256 = upResp.FileEncSHA256
	stkPackMsg.MediaKey = upResp.MediaKey
	stkPackMsg.DirectPath = &upResp.DirectPath

	c.SendMessage(&waE2E.Message{
		StickerPackMessage: stkPackMsg,
	})

	return nil
}
