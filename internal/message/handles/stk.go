package handles

import (
	"errors"
	"fmt"
	"kano/internal/utils/messageutil"
	"kano/internal/utils/sticker"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

func detectIsAnimatedFromDocument(d *waE2E.DocumentMessage) (bool, error) {
	docType := d.GetMimetype()
	if docType == "" {
		return false, errors.New("mime type is unspecified")
	}
	if strings.HasPrefix(docType, "video/") {
		return true, nil
	}
	if strings.HasPrefix(docType, "image/") {
		if strings.HasPrefix(strings.Replace(docType, "image/", "", 1), "gif") {
			return true, nil // image/gif
		} else {
			return false, nil // image/jpeg, image/png
		}
	}

	return false, fmt.Errorf("unsupported mime type %s", docType)
}

func Stk(c *messageutil.MessageContext) error {
	log := c.Logger.Sub("Stk")
	isViewOnce := false
	isAnimated := false
	download := (whatsmeow.DownloadableMessage)(nil)
	msg := c.RawMessage
	if repl := msg.GetExtendedTextMessage().GetContextInfo().GetQuotedMessage(); repl != nil {
		msg = repl
	}
	if dc := msg.GetDocumentWithCaptionMessage().GetMessage(); dc != nil {
		msg = dc
	}

	if i := msg.GetImageMessage(); i != nil {
		download = i
		isViewOnce = i.GetViewOnce()
	} else if v := msg.GetVideoMessage(); v != nil {
		download = v
		isViewOnce = v.GetViewOnce()
		isAnimated = true
	} else if d := msg.GetDocumentMessage(); d != nil {
		download = d
		mime := d.GetMimetype()
		if mime == "image/gif" || strings.HasPrefix(mime, "video/") {
			isAnimated = true
		}
	} else if ptv := msg.GetPtvMessage(); ptv != nil {
		download = ptv
		isAnimated = true
	}

	if download == nil {
		c.QuoteReply("Given message is not a media. Don't do that again.")
		log.Debugf("Given message is not a media")
		return nil
	}

	if isViewOnce {
		c.QuoteReply("Given media is a view once. Don't do that again.")
		log.Debugf("Given media is a view once")
		return nil
	}

	err := c.ValidateDownloadableMessage(download)
	if err != nil {
		c.QuoteReply("Given media is not downloadable, please resend it.\nDebug info: %s", err.Error())
		log.Infof("Given media is not downloadable: %s", err.Error())
		return nil
	}

	downloadedBytes, err := c.Client.Download(download)
	if err != nil {
		c.QuoteReply("Unable to download media. Please resend it.\nDebug info: %s", err.Error())
		log.Warnf("Unable to download media: %s", err.Error())
		return nil
	}

	createdSticker, err := sticker.MakeSticker(downloadedBytes, sticker.WhatsAppStickerMetadata{
		StickerPackId:        "kano_sticker_packs",
		StickerPackName:      "Kano",
		StickerPackPublisher: "Absolutely Kano",
	}, isAnimated)
	if err != nil {
		c.QuoteReply("Failed to convert media into sticker. This is an internal error, try again later.\nDebug info: %s", err.Error())
		log.Errorf("Sticker creation failed: %s", err.Error())
		return nil
	}

	_, err = c.ReplySticker(createdSticker, true)

	return err
}
