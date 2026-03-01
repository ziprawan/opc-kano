package messageutil

import (
	"fmt"
	"kano/internal/utils/image"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func (c *MessageContext) BuildReplyContextInfo() *waE2E.ContextInfo {
	sender := c.GetNonADSender().String()
	content := c.GetCleanMessage()

	return &waE2E.ContextInfo{
		StanzaID:      proto.String(c.GetID()),
		Participant:   proto.String(sender),
		QuotedMessage: content,
	}
}

func (c *MessageContext) Reply(text string, quoted bool) (whatsmeow.SendResponse, error) {
	var msg waE2E.Message

	if quoted {
		msg.ExtendedTextMessage = &waE2E.ExtendedTextMessage{
			Text:        proto.String(text),
			ContextInfo: c.BuildReplyContextInfo(),
		}
	} else {
		msg.Conversation = proto.String(text)
	}

	return c.SendMessage(&msg)
}

func (c *MessageContext) QuoteReply(text string, args ...any) (whatsmeow.SendResponse, error) {
	return c.Reply(fmt.Sprintf(text, args...), true)
}

func (c *MessageContext) SendText(text string, args ...any) (whatsmeow.SendResponse, error) {
	return c.Reply(fmt.Sprintf(text, args...), false)
}

func (c *MessageContext) ReplySticker(content []byte, quoted bool) (whatsmeow.SendResponse, error) {
	resp, err := c.Client.Upload(content, whatsmeow.MediaImage)
	if err != nil {
		return whatsmeow.SendResponse{}, err
	}

	now := time.Now().Unix()
	stkMsg := &waE2E.StickerMessage{
		StickerSentTS:     proto.Int64(now),
		Mimetype:          proto.String("image/webp"),
		URL:               proto.String(resp.URL),
		DirectPath:        proto.String(resp.DirectPath),
		MediaKey:          resp.MediaKey,
		FileEncSHA256:     resp.FileEncSHA256,
		FileSHA256:        resp.FileSHA256,
		FileLength:        proto.Uint64(resp.FileLength),
		IsLottie:          proto.Bool(false),
		MediaKeyTimestamp: proto.Int64(now),
		// IsAnimated:        proto.Bool(false), // rancu
	}

	if quoted {
		stkMsg.ContextInfo = c.BuildReplyContextInfo()
	}

	return c.SendMessage(&waE2E.Message{
		StickerMessage: stkMsg,
	})
}

func (c *MessageContext) ReplyDocument(content []byte, quoted bool) (whatsmeow.SendResponse, error) {
	resp, err := c.Client.Upload(content, whatsmeow.MediaDocument)
	if err != nil {
		return whatsmeow.SendResponse{}, err
	}

	now := time.Now().Unix()
	docMsg := &waE2E.DocumentMessage{
		Mimetype:          proto.String("plain/text"),
		URL:               proto.String(resp.URL),
		DirectPath:        proto.String(resp.DirectPath),
		MediaKey:          resp.MediaKey,
		FileEncSHA256:     resp.FileEncSHA256,
		FileSHA256:        resp.FileSHA256,
		FileLength:        proto.Uint64(resp.FileLength),
		MediaKeyTimestamp: proto.Int64(now),
	}

	if quoted {
		docMsg.ContextInfo = c.BuildReplyContextInfo()
	}

	return c.SendMessage(&waE2E.Message{
		DocumentMessage: docMsg,
	})
}

func (c *MessageContext) ReplyImage(content []byte, quoted bool, caption string) (whatsmeow.SendResponse, error) {
	resp, err := c.Client.Upload(content, whatsmeow.MediaImage)
	if err != nil {
		return whatsmeow.SendResponse{}, err
	}

	now := time.Now().Unix()
	imgMsg := &waE2E.ImageMessage{
		Mimetype:          proto.String("image/jpeg"),
		URL:               proto.String(resp.URL),
		DirectPath:        proto.String(resp.DirectPath),
		MediaKey:          resp.MediaKey,
		FileEncSHA256:     resp.FileEncSHA256,
		FileSHA256:        resp.FileSHA256,
		FileLength:        proto.Uint64(resp.FileLength),
		MediaKeyTimestamp: proto.Int64(now),
	}

	thumb, err := image.GenerateThumbnail(content)
	if err == nil {
		imgMsg.JPEGThumbnail = thumb
	}

	if quoted {
		imgMsg.ContextInfo = c.BuildReplyContextInfo()
	}

	if caption != "" {
		imgMsg.Caption = &caption
	}

	return c.SendMessage(&waE2E.Message{
		ImageMessage: imgMsg,
	})
}
