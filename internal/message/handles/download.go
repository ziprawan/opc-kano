package handles

import (
	"fmt"
	"kano/internal/utils/downloader"
	"kano/internal/utils/downloader/types"
	"kano/internal/utils/messageutil"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func uploadMedias(c *messageutil.MessageContext, media types.DownloaderMedia) (*waE2E.Message, error) {
	defer media.Close()

	msg := &waE2E.Message{}

	isVideo, contentType, height, width, duration := media.GetMetadata()

	mediaType := whatsmeow.MediaImage
	if isVideo {
		mediaType = whatsmeow.MediaVideo
	}

	upResp, err := c.Client.UploadReader(media, mediaType)
	if err != nil {
		return nil, err
	}

	if isVideo {
		msg.VideoMessage = &waE2E.VideoMessage{
			URL:           &upResp.URL,
			Mimetype:      &contentType,
			FileSHA256:    upResp.FileSHA256,
			FileLength:    &upResp.FileLength,
			Seconds:       &duration,
			MediaKey:      upResp.MediaKey,
			GifPlayback:   proto.Bool(false),
			Height:        &height,
			Width:         &width,
			FileEncSHA256: upResp.FileEncSHA256,
			DirectPath:    &upResp.DirectPath,
		}
	} else {
		msg.ImageMessage = &waE2E.ImageMessage{
			URL:           &upResp.URL,
			Mimetype:      &contentType,
			FileSHA256:    upResp.FileSHA256,
			FileLength:    &upResp.FileLength,
			MediaKey:      upResp.MediaKey,
			Height:        &height,
			Width:         &width,
			FileEncSHA256: upResp.FileEncSHA256,
			DirectPath:    &upResp.DirectPath,
		}
	}

	return msg, nil
}

func DownloadHandler(c *messageutil.MessageContext) error {
	url := c.Parser.RawArg.Content.Data
	if len(url) == 0 {
		c.QuoteReply("Give url (currently supports: instagram, tiktok, youtube[pls don't])")
		return nil
	}

	if strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be") {
		c.QuoteReply("Youtube support is currently dropped")
		return nil
	}

	rectMsg := &waE2E.Message{
		ReactionMessage: &waE2E.ReactionMessage{
			Key: &waCommon.MessageKey{
				RemoteJID: proto.String(c.GetChat().String()),
				FromMe:    proto.Bool(false),
				ID:        proto.String(c.GetID()),
			},
			Text: proto.String("⏳"),
		},
	}
	if c.Group != nil {
		rectMsg.ReactionMessage.Key.Participant = proto.String(c.GetSender().String())
	}
	c.SendMessage(rectMsg)

	downloaded, err := downloader.Download(url)
	if err != nil {
		c.QuoteReply("%s", err)
		return nil
	}

	caption := downloaded.GetCaption()
	medias := downloaded.GetMedias()

	parentMsgId := c.Client.GetClient().GenerateMessageID()

	msgs := make([]*waE2E.Message, len(medias))
	imageCount := uint32(0)
	videoCount := uint32(0)
	captionGiven := false
	for i, med := range medias {
		msg, err := uploadMedias(c, med)
		if err != nil {
			c.QuoteReply("failed to upload media: %s", err)
			return err
		}

		msg.MessageContextInfo = &waE2E.MessageContextInfo{
			MessageAssociation: &waE2E.MessageAssociation{
				AssociationType: waE2E.MessageAssociation_MEDIA_ALBUM.Enum(),
				ParentMessageKey: &waCommon.MessageKey{
					ID:        proto.String(parentMsgId),
					FromMe:    proto.Bool(true),
					RemoteJID: proto.String(c.GetChat().String()),
				},
			},
		}

		if c.Group != nil {
			msg.MessageContextInfo.MessageAssociation.ParentMessageKey.Participant = proto.String(c.GetSender().String())
		}

		if msg.VideoMessage != nil {
			videoCount++

			if !captionGiven {
				msg.VideoMessage.Caption = &caption
				captionGiven = true
			}
		} else if msg.ImageMessage != nil {
			imageCount++

			if !captionGiven {
				msg.ImageMessage.Caption = &caption
				captionGiven = true
			}
		}

		msgs[i] = msg
	}

	if len(msgs) == 0 {
		c.QuoteReply("%s", caption)
	} else if len(msgs) == 1 {
		msgs[0].MessageContextInfo.MessageAssociation = nil
		c.SendMessage(msgs[0])
	} else {
		c.SendMessage(&waE2E.Message{
			AlbumMessage: &waE2E.AlbumMessage{
				ExpectedImageCount: &imageCount,
				ExpectedVideoCount: &videoCount,
				ContextInfo:        c.BuildReplyContextInfo(),
			},
		}, whatsmeow.SendRequestExtra{ID: parentMsgId})

		for _, msg := range msgs {
			fmt.Println(c.SendMessage(msg))
		}
	}

	return nil
}
