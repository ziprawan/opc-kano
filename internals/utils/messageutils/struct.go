package messageutils

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"

	"nopi/internals/utils/parser"
)

type Message struct {
	Client *whatsmeow.Client
	Event  *events.Message
}

func CreateMessageInstance(client *whatsmeow.Client, event *events.Message) Message {
	return Message{Client: client, Event: event}
}

func (m Message) InitParser(prefixes []string) (parser.Parser, error) {
	return parser.InitParser(prefixes, m.Text())
}
