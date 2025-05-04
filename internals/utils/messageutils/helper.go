package messageutils

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"kano/internals/database"
	"kano/internals/utils/account"
	"kano/internals/utils/kanoutils"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// Helper functions

func (m *Message) Upload(content []byte, mediaType whatsmeow.MediaType) (whatsmeow.UploadResponse, error) {
	return m.Client.Upload(context.Background(), content, mediaType)
}

func (m *Message) EditText(text string) (*Message, error) {
	var msg *waE2E.Message = m.Client.BuildEdit(*m.ChatJID(), *m.ID(), &waE2E.Message{
		Conversation: proto.String(text),
	})
	acc, _ := account.GetData()

	jjj, sss := json.MarshalIndent(msg, "", "  ")
	fmt.Println(string(jjj), sss)

	resp, err := m.Client.SendMessage(context.Background(), *m.ChatJID(), msg)
	craftedMsg := Message{Client: m.Client, Group: m.Group, Contact: m.Contact, Event: &events.Message{
		Info: types.MessageInfo{
			MessageSource: types.MessageSource{
				Chat:     *m.ChatJID(),
				Sender:   *acc.JID,
				IsFromMe: true,
			},
			PushName: acc.PushName,
			ID:       resp.ID,
		},
		Message:    msg,
		RawMessage: msg,
	}}
	craftedMsg.SaveToDatabase()
	return &craftedMsg, err
}

func GetFutureProof(m *waE2E.Message) *waE2E.FutureProofMessage {
	if m == nil {
		return nil
	}

	if g := m.EphemeralMessage; g != nil {
		return g
	} else if g := m.ViewOnceMessage; g != nil {
		return g
	} else if g := m.DocumentWithCaptionMessage; g != nil {
		return g
	} else if g := m.ViewOnceMessageV2; g != nil {
		return g
	} else if g := m.ViewOnceMessageV2Extension; g != nil {
		return g
	} else if g := m.EditedMessage; g != nil {
		return g
	} else {
		return nil
	}
}

func removeContextInfo(m *waE2E.Message) {
	if m == nil {
		return
	}
	if k := m.ImageMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.ContactMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.LocationMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.ExtendedTextMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.DocumentMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.AudioMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.VideoMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.ContactsArrayMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.LiveLocationMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.TemplateMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.StickerMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.GroupInviteMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.TemplateButtonReplyMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.ProductMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.ListMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.OrderMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.ButtonsMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.ButtonsResponseMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.InteractiveMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.InteractiveResponseMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.PollCreationMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.RequestPhoneNumberMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.PollCreationMessageV2; k != nil {
		k.ContextInfo = nil
	} else if k := m.PollCreationMessageV3; k != nil {
		k.ContextInfo = nil
	} else if k := m.PtvMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.MessageHistoryBundle; k != nil {
		k.ContextInfo = nil
	} else if k := m.NewsletterAdminInviteMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.AlbumMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.StickerPackMessage; k != nil {
		k.ContextInfo = nil
	} else if k := m.PollResultSnapshotMessage; k != nil {
		k.ContextInfo = nil
	}
}

func (m *Message) getCleanQuotedMessage() (content *waE2E.Message) {
	content = m.Event.RawMessage
	for range 5 {
		inner := GetFutureProof(content)

		if inner == nil {
			break
		}

		content = inner.Message
	}
	removeContextInfo(content)
	return
}

func (m *Message) Reply(text string, quoted bool) (*Message, error) {
	acc, err := account.GetData()
	if err != nil {
		return nil, err
	}

	var msg waE2E.Message
	sender := m.SenderJID().String()

	if quoted {
		content := m.getCleanQuotedMessage()
		msg.ExtendedTextMessage = &waE2E.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:      m.ID(),
				Participant:   &sender,
				QuotedMessage: content,
			},
		}
	} else {
		msg.Conversation = proto.String(text)
	}

	fmt.Printf("==========\nID: %s\nSender: %s\nRaw Message: %+v\nMsg to send: %+v\n==========", *m.ID(), sender, m.Event.RawMessage, &msg)

	sent, err := m.Client.SendMessage(context.Background(), *m.ChatJID(), &msg)
	craftedMsg := Message{Client: m.Client, Group: m.Group, Contact: m.Contact, Event: &events.Message{
		Info: types.MessageInfo{
			MessageSource: types.MessageSource{
				Chat:     *m.ChatJID(),
				Sender:   *acc.JID,
				IsFromMe: true,
			},
			PushName: acc.PushName,
			ID:       sent.ID,
		},
		Message:    &msg,
		RawMessage: &msg,
	}}
	fmt.Printf("%+v\n", sent)
	saveErr := saveToDatabase(&craftedMsg)

	if saveErr != nil {
		fmt.Println("[Reply] Failed to save to database", saveErr)
	}

	return &craftedMsg, err
}

func (m *Message) ReplyWithTags(text string, tags []string) (*whatsmeow.SendResponse, error) {
	acc, err := account.GetData()
	if err != nil {
		return nil, err
	}

	var msg waE2E.Message
	sender := m.SenderJID().String()
	content := m.getCleanQuotedMessage()

	msg.ExtendedTextMessage = &waE2E.ExtendedTextMessage{
		Text: proto.String(text),
		ContextInfo: &waE2E.ContextInfo{
			StanzaID:      m.ID(),
			Participant:   &sender,
			QuotedMessage: content,
			MentionedJID:  tags,
		},
	}

	fmt.Printf("==========\nID: %s\nSender: %s\nRaw Message: %+v\nMsg to send: %+v\n==========", *m.ID(), sender, m.Event.RawMessage, &msg)

	sent, err := m.Client.SendMessage(context.Background(), *m.ChatJID(), &msg)

	saveToDatabase(&Message{Client: m.Client, Group: m.Group, Contact: m.Contact, Event: &events.Message{
		Info: types.MessageInfo{
			MessageSource: types.MessageSource{
				Chat:     *m.ChatJID(),
				Sender:   *acc.JID,
				IsFromMe: true,
			},
			PushName: acc.PushName,
			ID:       sent.ID,
		},
		Message:    &msg,
		RawMessage: &msg,
	}})

	return &sent, err
}

func (m *Message) ReplyImage(content []byte, op ReplyImageOptions) (*whatsmeow.SendResponse, error) {
	resp, err := m.Upload(content, whatsmeow.MediaImage)
	if err != nil {
		return nil, err
	}

	imgMsg := &waE2E.ImageMessage{
		URL:           &resp.URL,
		DirectPath:    &resp.DirectPath,
		MediaKey:      resp.MediaKey,
		FileEncSHA256: resp.FileEncSHA256,
		FileSHA256:    resp.FileSHA256,
		FileLength:    &resp.FileLength,
	}
	sender := m.SenderJID().String()

	if op.Quoted {
		content := m.getCleanQuotedMessage()
		imgMsg.ContextInfo = &waE2E.ContextInfo{
			StanzaID:      m.ID(),
			Participant:   &sender,
			QuotedMessage: content,
		}
	}
	// MimeType option, override the default mimetype field ("image/jpeg")
	if op.MimeType != "" {
		imgMsg.Mimetype = &op.MimeType
	} else {
		imgMsg.Mimetype = proto.String("image/jpeg")
	}
	// Caption option, add caption to the video
	if op.Caption != "" {
		imgMsg.Caption = &op.Caption
	}
	if op.DontGenerateInfo { // If both Height and Width are specified (not 0), set the video height and width
		if op.Height != 0 && op.Width != 0 {
			imgMsg.Height = &op.Height
			imgMsg.Width = &op.Width
		}
		if len(op.JPEGThumbnail) != 0 {
			imgMsg.JPEGThumbnail = op.JPEGThumbnail
		}
	} else {
		info, err := kanoutils.GenerateImageInfo(content)
		if err != nil {
			fmt.Println("reply_video: Failed to generate video info:", err.Error())
		} else {
			imgMsg.Height = &info.Height
			imgMsg.Width = &info.Width
			imgMsg.JPEGThumbnail = info.JPEGThumbnail
		}
	}

	targetJID := *m.ChatJID()
	if op.TargetJID != nil {
		targetJID = *op.TargetJID
	}

	sent, err := m.Client.SendMessage(context.Background(), targetJID, &waE2E.Message{
		ImageMessage: imgMsg,
	})

	return &sent, err
}

func (m *Message) ReplyVideo(content []byte, op ReplyVideoOptions) (*whatsmeow.SendResponse, error) {
	resp, err := m.Upload(content, whatsmeow.MediaVideo)
	if err != nil {
		return nil, err
	}

	var msg waE2E.Message

	sender := m.SenderJID().String()

	msg.VideoMessage = &waE2E.VideoMessage{
		URL:           &resp.URL,
		DirectPath:    &resp.DirectPath,
		MediaKey:      resp.MediaKey,
		FileEncSHA256: resp.FileEncSHA256,
		FileSHA256:    resp.FileSHA256,
		FileLength:    &resp.FileLength,
	}
	msgId := m.Client.GenerateMessageID()

	// Quoted option, reply to context message
	if op.Quoted {
		content := m.getCleanQuotedMessage()
		msg.VideoMessage.ContextInfo = &waE2E.ContextInfo{
			StanzaID:      m.ID(),
			Participant:   &sender,
			QuotedMessage: content,
		}
	}
	// MimeType option, override the default mimetype field ("video/mp4")
	if op.MimeType != "" {
		msg.VideoMessage.Mimetype = &op.MimeType
	} else {
		msg.VideoMessage.Mimetype = proto.String("video/mp4")
	}
	// Caption option, add caption to the video
	if op.Caption != "" {
		msg.VideoMessage.Caption = &op.Caption
	}
	if op.DontGenerateInfo { // If both Height and Width are specified (not 0), set the video height and width
		if op.Height != 0 && op.Width != 0 {
			msg.VideoMessage.Height = &op.Height
			msg.VideoMessage.Width = &op.Width
		}
		// Seconds option, override the detected duration length of the video
		// Eh, but, ..., what if the duration is actually ZERO?
		// If you didn'y specify it, it will use 0 duration tho, nothing changes
		if op.Seconds != 0 {
			msg.VideoMessage.Seconds = &op.Seconds
		}
		if len(op.JPEGThumbnail) != 0 {
			msg.VideoMessage.JPEGThumbnail = op.JPEGThumbnail
		}
	} else {
		info, err := kanoutils.GenerateVideoInfo(content)
		if err != nil {
			fmt.Println("reply_video: Failed to generate video info:", err.Error())
		} else {
			msg.VideoMessage.Seconds = &info.Duration
			msg.VideoMessage.Height = &info.Height
			msg.VideoMessage.Width = &info.Width
			msg.VideoMessage.JPEGThumbnail = info.JPEGThumbnail
		}
	}

	targetJID := *m.ChatJID()
	if op.TargetJID != nil {
		targetJID = *op.TargetJID
	}

	sent, err := m.Client.SendMessage(context.Background(), targetJID, &msg, whatsmeow.SendRequestExtra{
		ID: msgId,
	})
	return &sent, err
}

func (m *Message) ReplyAudio(content []byte, mimeType string, quoted bool) (*whatsmeow.SendResponse, error) {
	resp, err := m.Upload(content, whatsmeow.MediaAudio)
	if err != nil {
		return nil, err
	}

	audMsg := &waE2E.AudioMessage{
		Mimetype:      proto.String(mimeType),
		URL:           &resp.URL,
		DirectPath:    &resp.DirectPath,
		MediaKey:      resp.MediaKey,
		FileEncSHA256: resp.FileEncSHA256,
		FileSHA256:    resp.FileSHA256,
		FileLength:    &resp.FileLength,
	}
	sender := m.SenderJID().String()

	if quoted {
		content := m.getCleanQuotedMessage()
		audMsg.ContextInfo = &waE2E.ContextInfo{
			StanzaID:      m.ID(),
			Participant:   &sender,
			QuotedMessage: content,
		}
	}

	sent, err := m.Client.SendMessage(context.Background(), *m.ChatJID(), &waE2E.Message{
		AudioMessage: audMsg,
	})

	return &sent, err
}

func (m *Message) ReplySticker(content []byte, quoted bool) (*whatsmeow.SendResponse, error) {
	resp, err := m.Upload(content, whatsmeow.MediaImage)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	sender := m.SenderJID().String()

	stkMsg := &waE2E.StickerMessage{
		StickerSentTS:     &now,
		Mimetype:          proto.String("image/webp"),
		URL:               &resp.URL,
		DirectPath:        &resp.DirectPath,
		MediaKey:          resp.MediaKey,
		FileEncSHA256:     resp.FileEncSHA256,
		FileSHA256:        resp.FileSHA256,
		FileLength:        &resp.FileLength,
		IsAnimated:        proto.Bool(false),
		IsLottie:          proto.Bool(false),
		MediaKeyTimestamp: &now,
	}

	if quoted {
		stkMsg.ContextInfo = &waE2E.ContextInfo{
			StanzaID:      m.ID(),
			Participant:   &sender,
			QuotedMessage: m.getCleanQuotedMessage(),
		}
	}

	sent, err := m.Client.SendMessage(context.Background(), *m.ChatJID(), &waE2E.Message{
		StickerMessage: stkMsg,
	})

	return &sent, err
}

func (m *Message) MarkRead() {
	m.Client.MarkRead([]types.MessageID{*m.ID()}, time.Now(), m.Event.Info.Chat, *m.SenderJID())
}

func (m *Message) SaveToDatabase() error {
	return saveToDatabase(m)
}

func (m *Message) GetMessage(msgId string, jid *types.JID) (*Message, error) {
	db := database.GetDB()
	acc, err := account.GetData()
	if err != nil {
		return nil, err
	}

	var raw string
	err = db.QueryRow("SELECT raw FROM message_with_jid WHERE message_id = $1 AND entity_jid = $2 AND account_id = $3", msgId, jid.String(), acc.ID).Scan(&raw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	var ev events.Message
	err = json.Unmarshal([]byte(raw), &ev)
	if err != nil {
		return nil, err
	}

	return &Message{
		Client: m.Client,
		Event:  &ev,
	}, nil
}
