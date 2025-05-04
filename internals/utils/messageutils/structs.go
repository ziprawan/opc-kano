package messageutils

import (
	"fmt"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
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

type ReplyVideoOptions struct {
	MimeType string
	Caption  string
	Quoted   bool

	// Generate the info automatically inside the function
	// By default it is true
	DontGenerateInfo bool

	// These fields won't be processed if DontGenerateInfo is false

	Seconds uint32
	Height  uint32
	Width   uint32
	// Bytes of JPEG Thumbnail, this function will automatically converts them into base64
	JPEGThumbnail []byte

	TargetJID *types.JID
}

type ReplyImageOptions struct {
	MimeType string
	Caption  string
	Quoted   bool

	// Generate the info automatically inside the function
	// By default it is true
	DontGenerateInfo bool

	// These fields won't be processed if DontGenerateInfo is false

	Height uint32
	Width  uint32
	// Bytes of JPEG Thumbnail, this function will automatically converts them into base64
	JPEGThumbnail []byte

	TargetJID *types.JID
}

func CreateMessageInstance(client *whatsmeow.Client, event *events.Message) *Message {
	message := Message{Client: client, Event: event}

	if event.Info.Chat.Server == "g.us" {
		grp, err := saveutils.GetOrSaveGroup(client, &event.Info.Chat)
		if err != nil {
			fmt.Printf("Error saving group: %v\n", err)
			return nil
		} else {
			fmt.Printf("Found group: %+v\n", grp)
			message.Group = grp
		}
	}
	nonADJID := event.Info.Sender.ToNonAD()
	contact, err := saveutils.GetOrSaveContact(&nonADJID)
	if err != nil {
		fmt.Printf("Error saving contact: %v\n", err)
		return nil
	} else {
		fmt.Printf("Found contact: %+v\n", contact)
		message.Contact = contact
	}

	return &message
}

func (m Message) InitParser(prefixes []string) (parser.Parser, error) {
	return parser.InitParser(prefixes, m.Text())
}
