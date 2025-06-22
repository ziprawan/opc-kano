package message

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strconv"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"golang.org/x/image/draw"
	"google.golang.org/protobuf/proto"

	"github.com/forPelevin/gomoji"
	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/webp"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type StickerMetadata struct {
	StickerPackID        *string  `json:"sticker-pack-id,omitempty"`
	StickerPackName      *string  `json:"sticker-pack-name,omitempty"`
	StickerPackPublisher *string  `json:"sticker-pack-publisher,omitempty"`
	AndroidAppStoreLink  *string  `json:"android-app-store-link,omitempty"`
	IOSAppStoreLink      *string  `json:"ios-app-store-link,omitempty"`
	Emoji                []string `json:"emojis"`
}

type FFProbeResult struct {
	Streams []Stream `json:"streams,omitempty"`
}

type Stream struct {
	CodecType *string `json:"codec_type,omitempty"`
	Duration  *string `json:"duration,omitempty"`
}

func appendEXIF(src []byte, metadata StickerMetadata) []byte {
	mar, _ := json.Marshal(metadata)

	length := make([]byte, 4)
	binary.LittleEndian.PutUint32(length, uint32(len(mar)))

	// TIFF header, II 2A 00 08 00 00 00
	// IFD0 only contains 1 entry (01 00)
	// That entry tag is WA (41 57) with type undefined (07 00)
	tiff := []byte("II\x2A\x00\x08\x00\x00\x00\x01\x00\x41\x57\x07\x00") // Always fixed like this
	tiff = append(tiff, length...)                                       // Give tiff length
	tiff = append(tiff, []byte("\x16\x00\x00\x00")...)                   // The address is fixed and always at 0x16
	tiff = append(tiff, mar...)                                          // Append the data

	// Calculate the tiff length for EXIF chunk
	tiffLength := make([]byte, 4)
	binary.LittleEndian.PutUint32(tiffLength, uint32(len(tiff)))

	if len(tiff)%2 != 0 {
		// I just realized I was wrong here
		// I thought \x00 is counted as chunk length too smh
		tiff = append(tiff, '\x00')
	}

	// Create EXIF bytes
	toAppend := []byte("EXIF")                 // Chunk FourCC
	toAppend = append(toAppend, tiffLength...) // Chunk length
	toAppend = append(toAppend, tiff...)       // Chunk data for EXIF (alias TIFF)

	src = append(src, toAppend...) // Blindly append EXIF to image source

	var srcFixed []byte = []byte{}

	if idx := bytes.Index(src, []byte("VP8X")); idx != -1 {
		// Recalculate image data length
		stkDataLength := make([]byte, 4)
		binary.LittleEndian.PutUint32(stkDataLength, uint32(len(src)-8)) // -8 bytes of RIFF header

		currentAttribute := fmt.Sprintf("%08b", src[idx+8])

		if currentAttribute[4] != '1' {
			currentAttribute = currentAttribute[:4] + "1" + currentAttribute[5:]
		}

		newVal, _ := strconv.ParseUint(currentAttribute, 2, 0)
		src[idx+8] = byte(newVal)
		srcFixed = append(srcFixed, src[:4]...)
		srcFixed = append(srcFixed, stkDataLength...)
		srcFixed = append(srcFixed, src[8:]...)
	} else {
		// Recalculate image data length
		stkDataLength := make([]byte, 4)
		binary.LittleEndian.PutUint32(stkDataLength, uint32(len(src)+10)) // -8 bytes of RIFF header, +10 bytes of VP8X chunk

		srcFixed = append(srcFixed, src[:4]...)
		srcFixed = append(srcFixed, stkDataLength...)
		srcFixed = append(srcFixed, []byte("WEBPVP8X\x0A\x00\x00\x00\x08\x00\x00\x00\xff\x01\x00\xff\x01\x00")...)
		srcFixed = append(srcFixed, src[12:]...)
	}

	// os.WriteFile("test.webp", srcFixed, 0644)

	return srcFixed
}

func stkImage(ctx *MessageContext, downloadableMessage whatsmeow.DownloadableMessage, metadata StickerMetadata) {
	downloaded_bytes, err := ctx.Instance.Client.Download(context.Background(), downloadableMessage)
	if err != nil {
		ctx.Instance.Reply("Terjadi kesalahan saat mengunduh media", true)
		return
	}

	src, err := jpeg.Decode(bytes.NewReader(downloaded_bytes))
	if err != nil {
		fmt.Println(err)

		src, err = png.Decode(bytes.NewReader(downloaded_bytes))
		if err != nil {
			fmt.Println(err)

			src, err = webp.Decode(bytes.NewReader(downloaded_bytes), &decoder.Options{UseThreads: true})
			if err != nil {
				ctx.Instance.Reply("Terjadi kesalahan saat memproses gambar", true)
				return
			}
		}
	}

	// Lets resize!
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	// Calculate the scale
	scale := float64(512) / float64(max(srcWidth, srcHeight))
	newWidth := int(float64(srcWidth) * scale)
	newHeight := int(float64(srcHeight) * scale)

	// Do the resize
	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.BiLinear.Scale(resized, resized.Bounds(), src, srcBounds, draw.Over, nil)

	// Put the resized image to the center of empty image
	dst := image.NewRGBA(image.Rect(0, 0, 512, 512))

	offsetX := (512 - newWidth) / 2
	offsetY := (512 - newHeight) / 2
	draw.Draw(dst, image.Rect(offsetX, offsetY, offsetX+newWidth, offsetY+newHeight), resized, image.Point{}, draw.Over)

	var byteResult bytes.Buffer
	webp.Encode(&byteResult, dst, nil)
	resBytes := byteResult.Bytes()

	resBytes = appendEXIF(resBytes, metadata)

	// 500 KB
	if len(resBytes) > 500*1024 {
		ctx.Instance.Reply(fmt.Sprintf("Stiker yang dibuat melebihi 500 KB (that's weird...) (Didapat: %.2f KB)", float32(len(resBytes))/1024), true)
		return
	}

	ctx.Instance.ReplySticker(resBytes, false)
}

func stkVideo(ctx *MessageContext, downloadableMsg whatsmeow.DownloadableMessage, metadata StickerMetadata) {
	downloaded_bytes, err := ctx.Instance.Client.Download(context.Background(), downloadableMsg)
	if err != nil {
		ctx.Instance.Reply("Terjadi kesalahan saat mengunduh media!", true)
		return
	}

	buf := bytes.NewBuffer(nil)
	reader := bytes.NewReader(downloaded_bytes)

	data, err := ffmpeg.ProbeReader(bytes.NewReader(downloaded_bytes))
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Gagal mendapatkan info video, mungkin video rusak? Ket: %s", err), true)
		return
	}

	var ffprobeRet *FFProbeResult
	err = json.Unmarshal([]byte(data), &ffprobeRet)
	if err != nil {
		ctx.Instance.Reply("Failed to parse ffprobe JSON result", true)
		return
	}

	if ffprobeRet.Streams == nil {
		ctx.Instance.Reply("Given video has no streams, which is weird.", true)
		return
	}

	for _, stream := range ffprobeRet.Streams {
		if stream.CodecType != nil && *stream.CodecType == "video" {
			if stream.Duration != nil {
				duration, _ := strconv.ParseFloat(*stream.Duration, 32)

				if duration > 10 {
					ctx.Instance.Reply(fmt.Sprintf("Durasi video melebihi 10 detik. (Didapat: %.2f detik)", duration), true)
					return
				}
			}
		}
	}

	cmd := ffmpeg.
		Input("pipe:0").
		Filter("scale", ffmpeg.Args{"w=512:h=512:force_original_aspect_ratio=decrease"}).
		Filter("pad", ffmpeg.Args{"512:512:(ow-iw)/2:(oh-ih)/2:color=0x00000000"}).
		Output("pipe:1", ffmpeg.KwArgs{"c:v": "libwebp", "q:v": "1", "f": "webp", "pix_fmt": "yuva420p"}).
		WithInput(reader).
		WithOutput(buf, os.Stdout).
		ErrorToStdOut()

	fmt.Println("NEGAAA", cmd.Compile())

	err = cmd.Run()
	if err != nil {
		ctx.Instance.Reply(fmt.Sprintf("Terjadi kesalahan saat mengkonversi video menjadi stiker: %s", err), true)
		return
	}

	if buf == nil {
		ctx.Instance.Reply("Expected buffer, got <nil>", true)
		return
	}

	bbb := make([]byte, buf.Len())
	buf.Read(bbb)
	bbb = appendEXIF(bbb, metadata)

	// 1 MB
	if len(bbb) > 1024*1024 {
		ctx.Instance.Reply(fmt.Sprintf("Stiker yang dibuat melebihi 1 MB, mungkin coba kurangi durasi video? (Didapat: %.2f KB)", float32(len(bbb))/1024), true)
		return
	}

	ctx.Instance.ReplySticker(bbb, true)
}

func StkHandler(ctx *MessageContext) {
	repliedMsg, _ := ctx.Instance.ResolveReplyMessage(false)

	var img *waE2E.ImageMessage
	var vid *waE2E.VideoMessage
	var doc *waE2E.DocumentMessage

	if i := ctx.Instance.Event.RawMessage.GetImageMessage(); i != nil {
		img = i
	} else if v := ctx.Instance.Event.RawMessage.GetVideoMessage(); v != nil {
		vid = v
	} else if d := ctx.Instance.Event.RawMessage.GetDocumentMessage(); d != nil {
		doc = d
	} else if dc := ctx.Instance.Event.RawMessage.GetDocumentWithCaptionMessage(); dc != nil {
		if dc.Message != nil && dc.Message.DocumentMessage != nil {
			doc = dc.Message.DocumentMessage
		}
	} else if repliedMsg != nil {
		if i := repliedMsg.Event.RawMessage.GetImageMessage(); i != nil {
			img = i
		} else if v := repliedMsg.Event.RawMessage.GetVideoMessage(); v != nil {
			vid = v
		} else if d := repliedMsg.Event.RawMessage.GetDocumentMessage(); d != nil {
			doc = d
		} else if dc := repliedMsg.Event.RawMessage.GetDocumentWithCaptionMessage(); dc != nil {
			if dc.Message != nil && dc.Message.DocumentMessage != nil {
				doc = dc.Message.DocumentMessage
			}
		} else if ptv := repliedMsg.Event.RawMessage.GetPtvMessage(); ptv != nil {
			vid = ptv
		}
	}

	publisher := ctx.Instance.Event.Info.PushName
	if publisher == "" {
		publisher = "Kano"
	}
	metadata := StickerMetadata{
		StickerPackID:        proto.String("kano_sticker_packs"),
		StickerPackName:      proto.String("Kano Bot"),
		StickerPackPublisher: proto.String(publisher),
		Emoji:                []string{},
		AndroidAppStoreLink:  proto.String("https://play.google.com/store/apps/details?id=com.github.android"),
		IOSAppStoreLink:      proto.String("https://apps.apple.com/us/app/github/id1477376905"),
	}

	args := ctx.Parser.GetArgs()
	if len(args) >= 5 {
		metadata.AndroidAppStoreLink = &args[3].Content
		metadata.IOSAppStoreLink = &args[4].Content
	}
	if len(args) == 4 {
		ctx.Instance.Reply("Berikan link untuk iOS juga!", true)
		return
	}
	if len(args) >= 3 {
		found := gomoji.FindAll(args[2].Content)
		emojis := []string{}
		for _, emoji := range found {
			emojis = append(emojis, emoji.Character)
		}
		metadata.Emoji = emojis
	}
	if len(args) >= 2 {
		metadata.StickerPackPublisher = &args[1].Content
	}
	if len(args) >= 1 {
		metadata.StickerPackName = &args[0].Content
	}

	if img != nil {
		if img.ViewOnce != nil && *img.ViewOnce {
			ctx.Instance.Reply("Gambarnya sekali lihat dawg", true)
			return
		}
		stkImage(ctx, img, metadata)
		return
	} else if vid != nil {
		if vid.ViewOnce != nil && *vid.ViewOnce {
			ctx.Instance.Reply("Videonya sekali lihat dawg", true)
			return
		}
		if vid.FileLength != nil && *vid.FileLength > 10*1024*1024 {
			ctx.Instance.Reply("Kegedean 🥵 (Jangan lebih dari 10 MB ya mas)", true)
			return
		}

		stkVideo(ctx, vid, metadata)
		return
	} else if doc != nil {
		if doc.Mimetype == nil {
			ctx.Instance.Reply("Tidak dapat memastikan tipe dokumen", true)
			return
		}

		mimeType := *doc.Mimetype

		fmt.Println("Got mimetype:", mimeType)

		if strings.HasPrefix(mimeType, "image/") {
			if strings.HasSuffix(mimeType, "gif") {
				stkVideo(ctx, doc, metadata)
			} else {
				stkImage(ctx, doc, metadata)
			}
		} else if strings.HasPrefix(mimeType, "video/") {
			stkVideo(ctx, doc, metadata)
		} else {
			ctx.Instance.Reply(fmt.Sprintf("Tidak dapat memproses tipe media: %s", mimeType), true)
		}
	} else {
		ctx.Instance.Reply("Berikan atau reply gambar/video/dokumen. \nFormat: .stk <nama pack> <nama publisher> <emoji> <link pack android> <link pack ios>", true)
		return
	}
}
