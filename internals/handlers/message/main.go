package message

import (
	"context"
)

type MessageHandlerFunc func(ctx *MessageContext)

type MessageHandler struct {
	Func        MessageHandlerFunc
	Aliases     []string
	HelpMessage string
}

type MessageHandlerFuncMap map[string]MessageHandler

var MESSAGE_HANDLERS MessageHandlerFuncMap = MessageHandlerFuncMap{
	"ping": MessageHandler{
		Func: PingHandler,
	},
	"vo": MessageHandler{
		Func: VoHandler,
	},
	"confess": MessageHandler{
		Func: ConfessHandler,
	},
	"stk": MessageHandler{
		Func: StkHandler,
	},
	"stkinfo": MessageHandler{
		Func: StkInfoHandler,
	},
	"confesstarget": MessageHandler{
		Func: ConfessTargetHandler,
	},
	"login": MessageHandler{
		Func: LoginHandler,
	},
	"hsr": MessageHandler{
		Func: HSRHandler,
	},
	"gi": MessageHandler{
		Func: GIHandler,
	},
	"zzz": MessageHandler{
		Func: ZZZHandler,
	},
	"nim": MessageHandler{
		Func: NIMHandler,
	},
	"pddikti": MessageHandler{
		Func:    PDDIKTIHandler,
		Aliases: []string{"diddy"},
	},
	"shipping": MessageHandler{
		Func: ShippingHandler,
	},
	"wordle": MessageHandler{
		Func:    WorldeHandler,
		Aliases: []string{"worlde", "word"},
	},
	"json": MessageHandler{
		Func: JSONHandler,
	},
	"test": MessageHandler{
		Func: func(ctx *MessageContext) {
			ctx.Instance.Client.SendMessage(context.Background(), *ctx.Instance.ChatJID(), ctx.Instance.Client.BuildReaction(*ctx.Instance.ChatJID(), *ctx.Instance.SenderJID(), *ctx.Instance.ID(), "😂"))
		},
	},
	"oxford": MessageHandler{
		Func: OxfordHandler,
	},
}
var funcs map[string]MessageHandlerFunc = map[string]MessageHandlerFunc{}

func (ctx *MessageContext) Handle() {
	cmd := ctx.Parser.GetCommand()

	if len(funcs) < len(MESSAGE_HANDLERS) {
		for key, val := range MESSAGE_HANDLERS {
			funcs[key] = val.Func
			for _, alias := range val.Aliases {
				funcs[alias] = val.Func
			}
		}
	}

	detectedFunc := funcs[cmd.Command]
	if detectedFunc != nil {
		detectedFunc(ctx)
	}
}
