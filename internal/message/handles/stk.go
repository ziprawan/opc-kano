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

	packName := ""
	if argnames, ok := c.Parser.NamedArgs["name"]; ok && len(argnames) > 0 {
		packName = argnames[0].Content.Data
	}

	publisher := ""
	if arg := c.Parser.GetAllOriginalArg(); len(c.Parser.Args) > 0 {
		publisher = arg
	}
	createdSticker, err := sticker.MakeSticker(downloadedBytes, sticker.WhatsAppStickerMetadata{
		StickerPackId:        "kano_sticker_packs",
		StickerPackName:      packName,
		StickerPackPublisher: publisher,
	}, isAnimated)
	if err != nil {
		c.QuoteReply("Failed to convert media into sticker. This is an internal error, try again later.\nDebug info: %s", err.Error())
		log.Errorf("Sticker creation failed: %s", err.Error())
		return nil
	}

	_, err = c.ReplySticker(createdSticker, messageutil.ReplyConfig{Quoted: true})

	return err
}

var StkMan = CommandMan{
	Name: "stk - make a sticker",
	Synopsis: []string{
		"*stk* [ *name=*_pack_name_ ] [ _pack_publisher_ ]",
	},
	Description: []string{
		"Creates a sticker from the provided media. This command can be invoked by sending media with a caption containing the command, or by replying to a media message with .stk.",
		"Currently supported media types include images, videos, and documents.",
		"- If the media is an image, a static sticker will be generated." +
			"\n- If the media is a video, an animated sticker will be generated." +
			"\n- If the media is a document, the bot will first inspect its type:" +
			"\n{SPACE}- If the document is an image, a static sticker will be generated." +
			"\n{SPACE}- If the document is a video, an animated sticker will be generated." +
			"\n{SPACE}- If the document is neither an image nor a video, the bot will return an error.",
		"_pack_name_" +
			"\n{SPACE}The name of the sticker pack. In the application, this text appears in bold and is typically positioned on the left if a pack publisher is also defined.",
		"_pack_publisher_" +
			"\n{SPACE}The name of the sticker pack publisher. In the application, this text is not bold and appears to the right of the pack name, separated by a small dot when both are present.",
		"By default, generated stickers do not include visible metadata (the text shown beneath the sticker when it is opened). To include this visual metadata, you must define `pack_name` and/or `pack_publisher`.",
		"*Additional Informations:*" +
			"\n- Generated stickers are always in WebP format." +
			"\n- Output size is always 512x512, even if the original media is smaller." +
			"\n- All stickers are assigned to the pack metadata ID kano_sticker_packs." +
			"\n- If the media does not have a 1:1 aspect ratio, the remaining space will be padded with transparency." +
			"\n- If the media is a video, the bot will return an error if the duration exceeds 10 seconds.",
	},
	SourceFilename: "stk.go",
	SeeAlso:        []SeeAlso{},
}
