package message

import (
	"context"
	"fmt"

	"go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
)

type MessageHandlerFunc func(ctx *MessageContext)

// Inspired by man page :D
type CommandMan struct {
	Name        string
	Synopsis    []string
	Description []string

	Source  string
	SeeAlso []SeeAlso
}

type SeeAlsoType string

const (
	SeeAlsoTypeCommand      SeeAlsoType = "command"
	SeeAlsoTypeExternalLink SeeAlsoType = "external_link"
)

type SeeAlso struct {
	Content string
	Type    SeeAlsoType
}

type MessageHandler struct {
	Func    MessageHandlerFunc
	Aliases []string
	Man     CommandMan
}

type MessageHandlerFuncMap map[string]MessageHandler

var MESSAGE_HANDLERS MessageHandlerFuncMap = MessageHandlerFuncMap{
	"ping": MessageHandler{
		Func:    PingHandler,
		Man:     PingMan,
		Aliases: []string{"p"},
	},
	"vo": MessageHandler{
		Func: VoHandler,
		Man:  VoMan,
	},
	"confess": MessageHandler{
		Func:    ConfessHandler,
		Man:     ConfessMan,
		Aliases: []string{"fess"},
	},
	"stk": MessageHandler{
		Func:    StkHandler,
		Man:     StkMan,
		Aliases: []string{"s"},
	},
	"stkinfo": MessageHandler{
		Func: StkInfoHandler,
		Man:  StkInfoMan,
	},
	"confesstarget": MessageHandler{
		Func:    ConfessTargetHandler,
		Man:     ConfessTargetMan,
		Aliases: []string{"fesstarget", "targetconfess", "ct"},
	},
	"login": MessageHandler{
		Func: LoginHandler,
		Man:  LoginMan,
	},
	"hsr": MessageHandler{
		Func: HSRHandler,
		Man:  HSRMan,
	},
	"gi": MessageHandler{
		Func: GIHandler,
		Man:  GIMan,
	},
	"zzz": MessageHandler{
		Func: ZZZHandler,
		Man:  ZZZMan,
	},
	"nim": MessageHandler{
		Func: NIMHandler,
		Man:  NimMan,
	},
	"pddikti": MessageHandler{
		Func:    PDDIKTIHandler,
		Man:     DiddyMan,
		Aliases: []string{"diddy"},
	},
	"shipping": MessageHandler{
		Func: ShippingHandler,
		Man:  ShippingMan,
	},
	"wordle": MessageHandler{
		Func:    WorldeHandler,
		Man:     WordleMan,
		Aliases: []string{"worlde", "word"},
	},
	"json": MessageHandler{
		Func: JSONHandler,
		Man:  JsonMan,
	},
	"oxford": MessageHandler{
		Func: OxfordHandler,
		Man:  OxfordMan,
	},
	"leaderboard": MessageHandler{
		Func:    LeaderboardHandler,
		Man:     LeaderboardMan,
		Aliases: []string{"lb"},
	},
	"setname": MessageHandler{
		Func:    SetNameHandler,
		Man:     SetnameMan,
		Aliases: []string{"name"},
	},
	"down": MessageHandler{
		Func: DownloaderHandler,
		Man:  DownloaderMan,
	},
	"test": MessageHandler{
		Man: CommandMan{
			Name:     "test - Test",
			Synopsis: []string{"test"},
			Description: []string{
				"Buat ngetes.",
				"Argumen mungkin diperlukan dalam beberapa kasus.",
			},
			SeeAlso: []SeeAlso{},
			Source:  "COMMANDS.go",
		},
		Func: func(ctx *MessageContext) {
			jid, _ := types.ParseJID("6282112981691@s.whatsapp.net")
			list, err := ctx.Instance.Client.DangerousInternals().Usync(context.TODO(), []types.JID{jid}, "full", "background", []binary.Node{
				{Tag: "business", Content: []binary.Node{{Tag: "verified_name"}}},
				{Tag: "status"},
				{Tag: "picture"},
				{Tag: "devices", Attrs: binary.Attrs{"version": "2"}},
			})
			if err != nil {
				ctx.Instance.Reply(err.Error(), true)
				return
			} else {
				ctx.Instance.Reply(fmt.Sprintf("%+v", list), true)
			}
			// msg := &waE2E.Message{
			// 	ProtocolMessage: &waE2E.ProtocolMessage{
			// 		Key: &waCommon.MessageKey{
			// 			RemoteJID: proto.String(ctx.Instance.ChatJID().String()),
			// 			FromMe:    proto.Bool(true),
			// 		},
			// 		Type: waE2E.ProtocolMessage_LIMIT_SHARING.Enum(),
			// 		LimitSharing: &waCommon.LimitSharing{
			// 			SharingLimited:               proto.Bool(true),
			// 			Trigger:                      waCommon.LimitSharing_CHAT_SETTING.Enum(),
			// 			LimitSharingSettingTimestamp: proto.Int64(time.Now().UnixMilli()),
			// 			InitiatedByMe:                proto.Bool(true),
			// 		},
			// 	},
			// }
			// ctx.Instance.Client.SendMessage(context.Background(), *ctx.Instance.ChatJID(), msg)
		},
	},
}
var mappedCommands map[string]MessageHandler = map[string]MessageHandler{}

func (ctx *MessageContext) Handle() {
	cmd := ctx.Parser.GetCommand()

	if len(mappedCommands) < len(MESSAGE_HANDLERS) {
		for key, val := range MESSAGE_HANDLERS {
			mappedCommands[key] = val
			for _, alias := range val.Aliases {
				mappedCommands[alias] = val
			}
		}
	}

	detectedFunc, ok := mappedCommands[cmd.Command]
	if ok {
		detectedFunc.Func(ctx)
	}
}
