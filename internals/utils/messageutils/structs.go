package messageutils

import (
	"fmt"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"

	"kano/internals/utils/parser"
	"kano/internals/utils/saveutils"
)

type Message struct {
	Client  *whatsmeow.Client
	Event   *events.Message
	Group   *saveutils.Group
	Contact *saveutils.Contact
}

func CreateMessageInstance(client *whatsmeow.Client, event *events.Message) *Message {
	message := Message{Client: client, Event: event}

	if event.Info.Chat.Server == "g.us" {
		grp, err := saveutils.GetOrSaveGroup(client, &event.Info.Chat)
		if err != nil {
			fmt.Printf("Error saving group: %v\n", err)
			return nil
		} else {
			message.Group = grp
		}
	}
	nonADJID := event.Info.Sender.ToNonAD()
	contact, err := saveutils.GetOrSaveContact(&nonADJID)
	if err != nil {
		fmt.Printf("Error saving contact: %v\n", err)
		return nil
	} else {
		message.Contact = contact
	}

	return &message
}

func (m Message) InitParser(prefixes []string) (parser.Parser, error) {
	return parser.InitParser(prefixes, m.Text())
}
