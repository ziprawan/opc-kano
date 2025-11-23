package messageutil

import (
	"kano/internal/config"
	"kano/internal/logger"
	"kano/internal/utils/client"
	"kano/internal/utils/parser"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type MessageContextCache struct {
	sender    *types.JID
	senderAlt *types.JID
	chat      *types.JID
}

type MessageContext struct {
	// Whole message event
	Event *events.Message
	// Client context
	Client *client.ClientContext
	// Text parser, so you wouldn't call config.GetParser()
	Parser parser.ParseResult
	// Logger, so you wouldn't call config.GetLogger()
	Logger *logger.Logger

	// Shorthand of Event.Message
	Message *waE2E.Message
	// Shorthand of Event.Message
	RawMessage *waE2E.Message
	// Shorthand of Event.Info
	Info types.MessageInfo

	cache MessageContextCache
}

func CreateContext(cli *whatsmeow.Client, ev *events.Message) *MessageContext {
	ctx := MessageContext{
		Event:  ev,
		Client: client.CreateContext(cli),
		Logger: config.GetLogger().Sub("Message"),

		Message:    ev.Message,
		RawMessage: ev.RawMessage,
		Info:       ev.Info,
	}

	parser := config.GetParser()
	ctx.Parser = parser.Parse(ctx.GetText())

	return &ctx
}
