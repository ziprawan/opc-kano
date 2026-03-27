package types

import "io"

type DownloaderContext struct {
	caption string
	medias  []DownloaderMedia
}

type DownloaderMedia struct {
	reader      io.Reader
	isVideo     bool
	contentType string

	height   int
	width    int
	duration float64
}

func (d DownloaderMedia) Read(p []byte) (int, error) {
	return d.reader.Read(p)
}

func (d DownloaderMedia) Close() error {
	if closer, ok := d.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (d DownloaderMedia) GetMetadata() (bool, string, uint32, uint32, uint32) {
	return d.isVideo, d.contentType, uint32(d.height), uint32(d.width), uint32(d.duration)
}

func (d *DownloaderContext) SetCaption(caption string) {
	d.caption = caption
}

func (d DownloaderContext) GetCaption() string {
	return d.caption
}

func (d *DownloaderContext) AddMedia(reader io.Reader, isVideo bool, contentType string, height, width int, duration float64) {
	d.medias = append(d.medias, DownloaderMedia{
		reader:      reader,
		isVideo:     isVideo,
		contentType: contentType,
		height:      height,
		width:       width,
		duration:    duration,
	})
}

func (d DownloaderContext) GetMedias() []DownloaderMedia {
	return d.medias
}
