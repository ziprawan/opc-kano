package kanoutils

import (
	"bytes"
	"image"
	"image/jpeg"
	"math"

	"github.com/disintegration/imaging"
)

// GenerateThumbnail melakukan crop + resize sesuai spesifikasi
func GenerateThumbnail(imgBytes []byte) ([]byte, error) {
	img, err := jpeg.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, err
	}

	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	maxDim := float64(max(w, h))
	if maxDim <= 320 {
		// Gambar sudah kecil, return seperti semula
		return imgBytes, nil
	}

	aspect := float64(w) / float64(h)

	// Toleransi untuk dianggap "normal"
	if aspect > 0.5 && aspect < 2.0 {
		// Anggap rasio normal
		resized := resizeToMax100(img)
		return encodeJPEG(resized)
	}

	// Kalau abnormal, crop tengah dulu ke rasio 16:9 atau 9:16
	var cropRatio float64
	if w > h {
		cropRatio = 16.0 / 9.0
	} else {
		cropRatio = 9.0 / 16.0
	}

	cropped := cropCenterToRatio(img, cropRatio)
	resized := resizeToMax100(cropped)

	return encodeJPEG(resized)
}

// cropCenterToRatio memotong bagian tengah gambar dengan rasio target (misal 16:9)
func cropCenterToRatio(img image.Image, ratio float64) image.Image {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	var newW, newH int
	if float64(w)/float64(h) > ratio {
		// Terlalu lebar, potong samping
		newH = h
		newW = int(float64(h) * ratio)
	} else {
		// Terlalu tinggi, potong atas bawah
		newW = w
		newH = int(float64(w) / ratio)
	}

	startX := (w - newW) / 2
	startY := (h - newH) / 2

	return imaging.Crop(img, image.Rect(startX, startY, startX+newW, startY+newH))
}

// resizeToMax100 me-resize gambar supaya panjang maksimal 320px
func resizeToMax100(img image.Image) image.Image {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	scale := 100.0 / float64(max(w, h))
	newW := int(math.Round(float64(w) * scale))
	newH := int(math.Round(float64(h) * scale))

	return imaging.Resize(img, newW, newH, imaging.Lanczos)
}

// encodeJPEG mengubah image.Image menjadi []byte JPEG
func encodeJPEG(img image.Image) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// max helper
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
