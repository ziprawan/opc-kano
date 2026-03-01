package messageutil

import (
	"kano/internal/config"
	"kano/internal/logger"
	"kano/internal/utils/chatutil/contactutil"
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
	// Contact (database) info
	Contact contactutil.Contact

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

	var err error
	parser := config.GetParser()
	ctx.Parser, err = parser.Parse(ctx.GetText())
	if err != nil {
		ctx.Logger.Errorf("Failed to parse given text: %s", err)
		ctx.QuoteReply("parser: %s", err)
		return nil
	}

	sender := ctx.GetSender()
	if sender.Server != types.HiddenUserServer {
		ctx.Logger.Errorf("Sender JID server is not @lid and not @s.whatsapp.net")
	} else {
		contact, err := contactutil.Init(sender, ctx.Info.PushName)
		if err != nil {
			ctx.Logger.Errorf("Failed to init contact object: %s", err)
			return nil
		}
		ctx.Contact = contact
	}

	return &ctx
}
