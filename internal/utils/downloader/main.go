package downloader

import (
	"fmt"
	"kano/internal/utils/downloader/instagram"
	"kano/internal/utils/downloader/types"
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
	default:
		return ctx, fmt.Errorf("unsupported host %q", u.Host)
	}
}
