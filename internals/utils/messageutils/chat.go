package messageutils

import (
	"errors"

	"go.mau.fi/whatsmeow/types"
)

type Chat struct {
	JID     *types.JID
	Contact *Contact
	Group   *Group
}

var (
	ErrUnsupportedJidServer    = errors.New("unsupported jid server")
	ErrChatNotFound            = errors.New("chat not found")
	ErrUnexpectedReachableCode = errors.New("unexpected reachable code")
)

func (m Message) SaveGroup(force bool) (*Chat, error) {
	return saveGroup(m.ChatJID(), m.SenderJID(), m.Event, m.Client, force)
}

func (m Message) SaveContact() (*Chat, error) {
	return saveContact(m.SenderJID(), m.Event)
}

// func saveEntities() ([]Chat, error) {}

// It is guaranteed will return [Group, Contact] if no errors occured
func (m Message) SaveEntities() ([]Chat, error) {
	grp, gerr := m.SaveGroup(false)
	if gerr != nil {
		return nil, gerr
	}

	ctc, cerr := m.SaveContact()
	if cerr != nil {
		return nil, cerr
	}

	return []Chat{*grp, *ctc}, nil
}
