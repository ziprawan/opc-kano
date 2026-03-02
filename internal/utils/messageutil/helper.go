package messageutil

import (
	"errors"
	"fmt"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
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
	return c.GetSender().String() == nonAD || c.GetSenderAlt().String() == nonAD
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

func (c MessageContext) IsMe(jid types.JID) bool {
	if jid.Server != types.HiddenUserServer && jid.Server != types.DefaultUserServer {
		return false
	}

	if jid.Server == types.DefaultUserServer {
		lid, err := c.Client.GetLIDForPN(jid)
		if err == nil {
			jid = lid
		}
	}

	return c.Client.GetJID() == jid
}

func (c MessageContext) GetParticipantID() (uint, error) {
	participant, err := c.Group.GetParticipantByContactId(c.Contact.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			grpInfo, err := c.Client.GetGroupInfo(c.GetChat())
			if err != nil {
				return 0, fmt.Errorf("Failed to get group participants: %s", err)
			}
			err = c.Group.UpdateParticipantList(grpInfo)
			if err != nil {
				return 0, fmt.Errorf("Failed to update group participants: %s", err)
			}

			participant, err = c.Group.GetParticipantByContactId(c.Contact.ID)
			if err != nil {
				return 0, fmt.Errorf("Failed to get participant info: %s", err)
			}
		} else {
			return 0, fmt.Errorf("Failed to get participant info: %s", err)
		}
	}
	return participant.ID, nil
}
