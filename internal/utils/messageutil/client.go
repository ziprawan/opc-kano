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
	sender := c.GetSender().String()
	content := c.GetCleanMessage()

	return &waE2E.ContextInfo{
		StanzaID:      proto.String(c.GetID()),
		Participant:   proto.String(sender),
		QuotedMessage: content,
	}
}

type ReplyConfig struct {
	Quoted           bool
	ContextInfo      *waE2E.ContextInfo
	SendRequestExtra whatsmeow.SendRequestExtra
}

func (c *MessageContext) Reply(text string, configs ...ReplyConfig) (whatsmeow.SendResponse, error) {
	var msg waE2E.Message

	var config ReplyConfig
	if len(configs) > 0 {
		config = configs[0]
	}

	if config.Quoted {
		msg.ExtendedTextMessage = &waE2E.ExtendedTextMessage{
			Text:        proto.String(text),
			ContextInfo: c.BuildReplyContextInfo(),
		}
		if config.ContextInfo != nil {
			proto.Merge(msg.ExtendedTextMessage.ContextInfo, config.ContextInfo)
		}
	} else {
		if config.ContextInfo != nil {
			msg.ExtendedTextMessage = &waE2E.ExtendedTextMessage{
				Text:        proto.String(text),
				ContextInfo: config.ContextInfo,
			}
		} else {
			msg.Conversation = proto.String(text)
		}
	}

	return c.SendMessage(&msg, config.SendRequestExtra)
}

func (c *MessageContext) QuoteReply(text string, args ...any) (whatsmeow.SendResponse, error) {
	return c.Reply(fmt.Sprintf(text, args...), ReplyConfig{Quoted: true})
}

func (c *MessageContext) ReplySticker(content []byte, configs ...ReplyConfig) (whatsmeow.SendResponse, error) {
	resp, err := c.Client.Upload(content, whatsmeow.MediaImage)
	if err != nil {
		return whatsmeow.SendResponse{}, err
	}

	var config ReplyConfig
	if len(configs) > 0 {
		config = configs[0]
	}

	now := time.Now().Unix()
	stkMsg := &waE2E.StickerMessage{
		StickerSentTS: proto.Int64(now),
		Mimetype:      proto.String("image/webp"),
		URL:           proto.String(resp.URL),
		DirectPath:    proto.String(resp.DirectPath),
		MediaKey:      resp.MediaKey,
		FileEncSHA256: resp.FileEncSHA256,
		FileSHA256:    resp.FileSHA256,
		FileLength:    proto.Uint64(resp.FileLength),
		IsLottie:      proto.Bool(false),
	}

	if config.Quoted {
		stkMsg.ContextInfo = c.BuildReplyContextInfo()
		if config.ContextInfo != nil {
			proto.Merge(stkMsg.ContextInfo, config.ContextInfo)
		}
	}

	return c.SendMessage(&waE2E.Message{
		StickerMessage: stkMsg,
	}, config.SendRequestExtra)
}

func (c *MessageContext) ReplyDocument(content []byte, mimetype string, configs ...ReplyConfig) (whatsmeow.SendResponse, error) {
	resp, err := c.Client.Upload(content, whatsmeow.MediaDocument)
	if err != nil {
		return whatsmeow.SendResponse{}, err
	}

	var config ReplyConfig
	if len(configs) > 0 {
		config = configs[0]
	}

	now := time.Now().Unix()
	docMsg := &waE2E.DocumentMessage{
		Mimetype:          proto.String(mimetype),
		URL:               proto.String(resp.URL),
		DirectPath:        proto.String(resp.DirectPath),
		MediaKey:          resp.MediaKey,
		FileEncSHA256:     resp.FileEncSHA256,
		FileSHA256:        resp.FileSHA256,
		FileLength:        proto.Uint64(resp.FileLength),
		MediaKeyTimestamp: proto.Int64(now),
	}

	if config.Quoted {
		docMsg.ContextInfo = c.BuildReplyContextInfo()
		if config.ContextInfo != nil {
			proto.Merge(docMsg.ContextInfo, config.ContextInfo)
		}
	}

	return c.SendMessage(&waE2E.Message{
		DocumentMessage: docMsg,
	}, config.SendRequestExtra)
}

func (c *MessageContext) ReplyImage(content []byte, caption string, configs ...ReplyConfig) (whatsmeow.SendResponse, error) {
	resp, err := c.Client.Upload(content, whatsmeow.MediaImage)
	if err != nil {
		return whatsmeow.SendResponse{}, err
	}

	var config ReplyConfig
	if len(configs) > 0 {
		config = configs[0]
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

	if config.Quoted {
		imgMsg.ContextInfo = c.BuildReplyContextInfo()
		if config.ContextInfo != nil {
			proto.Merge(imgMsg.ContextInfo, config.ContextInfo)
		}
	}

	if caption != "" {
		imgMsg.Caption = &caption
	}

	return c.SendMessage(&waE2E.Message{
		ImageMessage: imgMsg,
	}, config.SendRequestExtra)
}
