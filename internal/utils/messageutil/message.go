package messageutil

import (
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type MessageWithContextInfo interface {
	GetContextInfo() *waE2E.ContextInfo
}

func (c *MessageContext) GetConversation() string {
	var text string

	if conv := c.Message.GetConversation(); conv != "" {
		text = conv
	} else if ext := c.Message.GetExtendedTextMessage(); ext != nil {
		text = ext.GetText()
	}

	return strings.TrimSpace(text)
}

func (c *MessageContext) GetCaption() string {
	var caption string

	if d := c.RawMessage.GetDocumentWithCaptionMessage(); d != nil {
		if e := d.GetMessage(); e != nil {
			if f := e.GetDocumentMessage(); f != nil {
				caption = f.GetCaption()
			}
		}
	} else if i := c.RawMessage.GetImageMessage(); i != nil {
		caption = i.GetCaption()
	} else if v := c.RawMessage.GetVideoMessage(); v != nil {
		caption = v.GetCaption()
	}

	return strings.TrimSpace(caption)
}

func (c *MessageContext) GetText() string {
	var text string
	if conv := c.GetConversation(); conv != "" {
		return conv
	} else if cap := c.GetCaption(); cap != "" {
		text = cap
	}
	return text
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

func RemoveContextInfo(m *waE2E.Message) {
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

func (c *MessageContext) GetCleanMessage() (content *waE2E.Message) {
	content = c.RawMessage
	for range 5 {
		inner := GetFutureProof(content)

		if inner == nil {
			break
		}

		content = inner.Message
	}
	// RemoveContextInfo(content)
	return
}

func (c *MessageContext) GetRepliedMessage() (types.MessageID, types.JID, *waE2E.Message) {
	msgReflect := c.RawMessage.ProtoReflect()
	msgDescriptor := msgReflect.Descriptor()
	msgFields := msgDescriptor.Fields()

	target := protoreflect.Name("contextInfo")

	// Iterate all message fields
	for i := 0; i < msgFields.Len(); i++ {
		field := msgFields.Get(i)
		if !msgReflect.Has(field) {
			continue
		}

		// Ensure current field is a protobuf message
		if field.Kind() != protoreflect.MessageKind {
			continue
		}

		subMsg := msgReflect.Get(field).Message() // Get the field message

		// Check if this sub-field has contextInfo
		subDesc := subMsg.Descriptor().Fields().ByName(target)
		if subDesc == nil {
			continue
		}

		// Check if the contextInfo is actually exists
		if !subMsg.Has(subDesc) {
			continue
		}

		// Extract the contextInfo
		msg, ok := subMsg.Get(subDesc).Message().Interface().(*waE2E.ContextInfo)
		if ok {
			senderJid, _ := types.ParseJID(msg.GetParticipant())
			if senderJid.Server == types.DefaultUserServer {
				theLid, err := c.Client.GetLIDForPN(senderJid)
				if err == nil {
					senderJid = theLid
				}
			}
			return msg.GetStanzaID(), senderJid, msg.GetQuotedMessage()
		}
	}

	return "", types.JID{}, nil
}

func (c *MessageContext) EditMessageWithID(msgId types.MessageID, msg *waE2E.Message) (whatsmeow.SendResponse, error) {
	return c.SendMessage(
		c.Client.BuildEdit(c.GetChat(), msgId, msg),
	)
}
