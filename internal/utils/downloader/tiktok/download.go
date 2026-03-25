package tiktok

import (
	"fmt"
	"io"
	"kano/internal/utils/downloader/types"
	ytdlpbind "kano/internal/utils/downloader/ytdlp-bind"
	"net/url"
)

func Download(ctx *types.DownloaderContext, url *url.URL) error {
	res, err := ytdlpbind.Call(url.String())
	if err != nil {
		return err
	}

	if len(res.Url) != 0 {
		size := res.FileSize
		if size > 200*1024*1024 {
			return fmt.Errorf("file size is too big (>200 MB)")
		}

		resp, err := ytdlpbind.Request(res.YtDlpFormat)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		mergedBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		ctx.AddMedia(mergedBytes, true, res.Height, res.Width, float64(res.Duration))

		caption := fmt.Sprintf(
			"%s\n\nUploaded by: %s (%s)\n%d 👀 %d ❤️ %d 💬",
			res.Description,
			res.Uploader, res.UploaderUrl,
			res.ViewCount, res.LikeCount, res.CommentCount,
		)
		ctx.SetCaption(caption)

	} else {
		return fmt.Errorf("expected 2 requested formats, got %d", len(res.RequestedFormats))
	}

	return nil
}
