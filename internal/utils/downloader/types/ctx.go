package types

type DownloaderContext struct {
	caption string
	medias  []DownloaderMedia
}

type DownloaderMedia struct {
	data    []byte
	isVideo bool

	height   int
	width    int
	duration float64
}

func (d DownloaderMedia) GetData() ([]byte, bool) {
	return d.data, d.isVideo
}

func (d DownloaderMedia) GetDimensions() (uint32, uint32, uint32) {
	return uint32(d.height), uint32(d.width), uint32(d.duration)
}

func (d *DownloaderContext) SetCaption(caption string) {
	d.caption = caption
}

func (d DownloaderContext) GetCaption() string {
	return d.caption
}

func (d *DownloaderContext) AddMedia(data []byte, isVideo bool, height, width int, duration float64) {
	d.medias = append(d.medias, DownloaderMedia{
		data:     data,
		isVideo:  isVideo,
		height:   height,
		width:    width,
		duration: duration,
	})
}

func (d DownloaderContext) GetMedias() []DownloaderMedia {
	return d.medias
}
