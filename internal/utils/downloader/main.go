package downloader

import (
	"fmt"
	"kano/internal/utils/downloader/instagram"
	"kano/internal/utils/downloader/tiktok"
	"kano/internal/utils/downloader/types"
	"kano/internal/utils/downloader/youtube"
	"net/url"
)

func Download(urlStr string) (types.DownloaderContext, error) {
	ctx := types.DownloaderContext{}
	u, err := url.Parse(urlStr)
	if err != nil {
		return ctx, fmt.Errorf("url is not parsable")
	}
	if u.Scheme != "https" {
		return ctx, fmt.Errorf("url scheme is not https")
	}

	switch u.Host {
	case "instagram.com", "www.instagram.com":
		err = instagram.Download(&ctx, u)
		if err != nil {
			return ctx, fmt.Errorf("instagram: %s", err)
		}
		return ctx, nil
	case "www.youtube.com", "youtube.com", "youtu.be":
		err = youtube.Download(&ctx, u)
		if err != nil {
			return ctx, fmt.Errorf("youtube: %s", err)
		}
		return ctx, nil
	case "www.tiktok.com", "vt.tiktok.com":
		err = tiktok.Download(&ctx, u)
		if err != nil {
			return ctx, fmt.Errorf("tiktok: %s", err)
		}
		return ctx, nil
	default:
		return ctx, fmt.Errorf("unsupported host %q", u.Host)
	}
}
