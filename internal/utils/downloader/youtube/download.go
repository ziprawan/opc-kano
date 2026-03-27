package youtube

import (
	"fmt"
	"io"
	"kano/internal/utils/downloader/types"
	ytdlpbind "kano/internal/utils/downloader/ytdlp-bind"
	"net/url"
	"os"

	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

func Download(ctx *types.DownloaderContext, url *url.URL) error {
	res, err := ytdlpbind.Call(url.String())
	if err != nil {
		return err
	}

	tempdir, err := os.MkdirTemp("tmp/", "downloader_yt_*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempdir)

	if len(res.RequestedFormats) == 2 {
		size := res.RequestedFormats[0].FileSize + res.RequestedFormats[1].FileSize
		if size > 200*1024*1024 {
			return fmt.Errorf("file size is too big (>200 MB)")
		}

		input1 := fmt.Sprintf("%s/%s_%s.%s", tempdir, res.Id, res.RequestedFormats[0].FormatId, res.RequestedFormats[0].Ext)
		input2 := fmt.Sprintf("%s/%s_%s.%s", tempdir, res.Id, res.RequestedFormats[1].FormatId, res.RequestedFormats[1].Ext)
		output := fmt.Sprintf("%s/out_%s.mp4", tempdir, res.Id)

		// Opening the files
		file1, err := os.OpenFile(input1, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to open input1 file: %w", err)
		}
		defer file1.Close()

		file2, err := os.OpenFile(input2, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to open input2 file: %w", err)
		}
		defer file2.Close()

		// Downloading and copying the files
		resp1, err := ytdlpbind.Request(res.RequestedFormats[0])
		if err != nil {
			return err
		}
		defer resp1.Body.Close()
		_, err = io.Copy(file1, resp1.Body)
		if err != nil {
			return err
		}
		file1.Close()

		resp2, err := ytdlpbind.Request(res.RequestedFormats[1])
		if err != nil {
			return err
		}
		defer resp2.Body.Close()
		_, err = io.Copy(file2, resp2.Body)
		if err != nil {
			return err
		}
		file2.Close()

		// Generating the output
		err = ffmpeg_go.Output(
			[]*ffmpeg_go.Stream{ffmpeg_go.Input(input1), ffmpeg_go.Input(input2)},
			output,
			ffmpeg_go.KwArgs{"c:v": "copy", "c:a": "copy"},
		).Run()
		if err != nil {
			return err
		}

		fileReader, err := os.OpenFile(output, os.O_RDONLY, 0600)
		if err != nil {
			return err
		}

		ctx.AddMedia(fileReader, true, "video/mp4", res.Height, res.Width, float64(res.Duration))

		caption := fmt.Sprintf(
			"%s\n\n%s\n\nUploaded by: %s (%s) - %d 🫂\n%d 👀 %d ❤️ %d 💬",
			res.Title, res.Description,
			res.Channel, res.ChannelUrl, res.ChannelFollowerCount,
			res.ViewCount, res.LikeCount, res.CommentCount,
		)
		ctx.SetCaption(caption)

	} else {
		return fmt.Errorf("expected 2 requested formats, got %d", len(res.RequestedFormats))
	}

	return nil
}
