package sticker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strconv"

	imageutil "kano/internal/utils/image"

	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/webp"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"golang.org/x/image/draw"
)

func toKnownImageStruct(b []byte) (img image.Image, err error) {
	img, err = jpeg.Decode(bytes.NewReader(b))
	if err == nil {
		return
	}

	img, err = png.Decode(bytes.NewReader(b))
	if err == nil {
		return
	}

	img, err = webp.Decode(bytes.NewReader(b), &decoder.Options{})
	return
}

func makeStaticSticker(docBytes []byte) ([]byte, error) {
	img, err := toKnownImageStruct(docBytes)
	if err != nil {
		return nil, ErrUnsupportedFile
	}

	// Get current image info
	imgBounds := img.Bounds()
	imgWidth := imgBounds.Dx()
	imgHeight := imgBounds.Dy()

	// Resizing the image into 512x512
	// First, calculate the scale first
	scale := float64(STICKER_SIZE_PX) / float64(max(imgWidth, imgHeight)) // Get the bigger value between image height and width
	// Then calculate the resized width and height
	newWidth := int(float64(imgWidth) * scale)
	newHeight := int(float64(imgHeight) * scale)

	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.BiLinear.Scale(resized, resized.Bounds(), img, imgBounds, draw.Over, nil)

	// New stiker image with RGBA color.
	// If the new width or height is less than STICKER_SIZE_PX, remaining pixel will be transparent
	stk := image.NewRGBA(image.Rect(0, 0, STICKER_SIZE_PX, STICKER_SIZE_PX))

	// Calculate the offset and draw over the transparent rect image
	offsetX := (STICKER_SIZE_PX - newWidth) / 2
	offsetY := (STICKER_SIZE_PX - newHeight) / 2
	draw.Draw(stk, image.Rect(offsetX, offsetY, offsetX+newWidth, offsetY+newHeight), resized, image.Point{}, draw.Over)

	// Encode using WebP
	var byteBuffer bytes.Buffer
	webp.Encode(&byteBuffer, stk, nil)
	stkBytes := byteBuffer.Bytes()

	return stkBytes, nil
}

func makeAnimatedSticker(docBytes []byte) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	reader := bytes.NewReader(docBytes)

	data, err := ffmpeg.ProbeReader(bytes.NewReader(docBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to probe file: %s", err.Error())
	}

	var ffprobeRet *FFProbeResult
	err = json.Unmarshal([]byte(data), &ffprobeRet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse probe file result: %s", err.Error())
	}
	if ffprobeRet.Streams == nil {
		return nil, ErrNoStream
	}

	for _, stream := range ffprobeRet.Streams {
		if stream.CodecType == "video" {
			duration, _ := strconv.ParseFloat(stream.Duration, 32)
			if duration > 10 {
				return nil, ErrDurationTooLong
			}
		}
	}

	cmd := ffmpeg.
		Input("pipe:0").
		Filter("scale", ffmpeg.Args{"w=512:h=512:force_original_aspect_ratio=decrease"}).
		Filter("pad", ffmpeg.Args{"512:512:(ow-iw)/2:(oh-ih)/2:color=0x00000000"}).
		Output("pipe:1", ffmpeg.KwArgs{"c:v": "libwebp", "q:v": "1", "f": "webp", "pix_fmt": "yuva420p"}).
		WithInput(reader).
		WithOutput(buffer, os.Stdout).
		ErrorToStdOut()

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("video conversion into sticker failed: %s", err.Error())
	}
	if buffer == nil {
		return nil, ErrNoResult
	}

	stkBytes := make([]byte, buffer.Len())
	buffer.Read(stkBytes)

	return stkBytes, nil
}

func appendMetadataToSticker(stickerByte []byte, metadata WhatsAppStickerMetadata) ([]byte, error) {
	mar, _ := json.Marshal(metadata)
	stickerChunks, err := imageutil.ExtractChunksFromWebP(stickerByte)
	if err != nil {
		return nil, err
	}

	exif, err := imageutil.BuildTIFF([]imageutil.IFD{
		{NumberOfEntries: 1, Entries: []imageutil.DirectoryEntry{
			{Tag: 0x4157, Type: imageutil.EntryTypeUndefined, Count: uint32(len(mar)), Value: mar},
		}},
	}, false)
	if err != nil {
		return nil, err
	}

	stickerChunks["EXIF"] = exif
	stickerChunks = imageutil.FixWebPExtendedChunks(stickerChunks, STICKER_SIZE_PX, STICKER_SIZE_PX)
	newStickerByte, err := imageutil.BuildWebPFromChunks(stickerChunks)
	if err != nil {
		return nil, err
	}

	return newStickerByte, nil
}

func MakeSticker(docBytes []byte, metadata WhatsAppStickerMetadata, isAnimated bool) ([]byte, error) {
	var err error
	var sticker []byte
	if isAnimated {
		sticker, err = makeAnimatedSticker(docBytes)
	} else {
		sticker, err = makeStaticSticker(docBytes)
	}
	if err != nil {
		return nil, err
	}

	return appendMetadataToSticker(sticker, metadata)
}
