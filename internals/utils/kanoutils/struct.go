package kanoutils

type KanoVideoInfo struct {
	Duration      uint32
	Height        uint32
	Width         uint32
	JPEGThumbnail []byte
}

type KanoImageInfo struct {
	Height        uint32
	Width         uint32
	JPEGThumbnail []byte
}
