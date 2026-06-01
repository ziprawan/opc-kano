package twitter

import (
	"fmt"
	"kano/internal/utils/downloader/types"
	"net/url"
)

func Download(ctx *types.DownloaderContext, url *url.URL) error {
	if ctx == nil || url == nil {
		return fmt.Errorf("given ctx or url is nil")
	}

	return nil
}
