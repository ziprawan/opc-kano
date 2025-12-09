package messageutil

import (
	"errors"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

var (
	ErrMissingDirectPath    = errors.New("missing direct path and url field")
	ErrMissingFileEncSHA256 = errors.New("missing file enc sha 256 field")
	ErrMissingFileSHA256    = errors.New("missing file sha 256 field")
	ErrMissingMediaKey      = errors.New("missing media key field")
)

func (c MessageContext) ValidateDownloadableMessage(m whatsmeow.DownloadableMessage) error {
	if len(m.GetDirectPath()) == 0 {
		return ErrMissingDirectPath
	} else if len(m.GetFileEncSHA256()) == 0 {
		return ErrMissingDirectPath
	} else if len(m.GetFileSHA256()) == 0 {
		return ErrMissingDirectPath
	} else if len(m.GetMediaKey()) == 0 {
		return ErrMissingDirectPath
	}

	return nil
}

func (c MessageContext) IsSenderSame(compareJid types.JID) bool {
	nonAD := compareJid.ToNonAD().String()
	return c.GetNonADSender().String() == nonAD || c.GetNonADSenderAlt().String() == nonAD
}

func (c MessageContext) SendMessage(message *waE2E.Message) (whatsmeow.SendResponse, error) {
	chat := c.GetChat(true)
	if chat.Server == types.HiddenUserServer {
		found, err := c.Client.GetPNForLID(chat)
		if err == nil {
			chat = found
		} else {
			c.Logger.Warnf("Couldn't find PN for LID %s, this might cause message not appears in some devices", chat.String())
		}
	}
	return c.Client.SendMessage(chat, message)
}
