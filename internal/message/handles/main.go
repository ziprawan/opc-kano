package handles

import (
	"errors"
	"kano/internal/config"
	"kano/internal/utils/messageutil"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

var ErrNotImplemented = errors.New("command not implemented")

type CommandHandlerFunc func(ctx *messageutil.MessageContext) error

type CommandHandler struct {
	Func    CommandHandlerFunc
	Aliases []string
	// Man     CommandMan
}

type CommandHandlerFuncMap map[string]CommandHandler

var HANDLES CommandHandlerFuncMap = CommandHandlerFuncMap{
	"ping": CommandHandler{
		Func:    Ping,
		Aliases: []string{"p"},
	},
	"stk": CommandHandler{
		Func:    Stk,
		Aliases: []string{"s"},
	},
	"vo": CommandHandler{
		Func:    Vo,
		Aliases: []string{"v"},
	},
	"nim": CommandHandler{
		Func: Nim,
	},
	"pddikti": CommandHandler{
		Func:    Pddikti,
		Aliases: []string{"diddy"},
	},
	"confess": CommandHandler{
		Func:    Confess,
		Aliases: []string{"c"},
	},
	"six": CommandHandler{
		Func: Six,
	},
	"ta": CommandHandler{
		Func: Ta,
	},
	"test": CommandHandler{
		Func: Test,
	},
	"redirect": CommandHandler{
		Func:    Redirect,
		Aliases: []string{"r", "getredir", "getloc"},
	},
	"resolve-subject": CommandHandler{
		Func:    ResolveSubject,
		Aliases: []string{"rs"},
	},
	"bs": CommandHandler{
		Func: func(ctx *messageutil.MessageContext) error {
			ctx.SendMessage(&waE2E.Message{
				ExtendedTextMessage: &waE2E.ExtendedTextMessage{
					Text: proto.String("Bullshit"),
					ContextInfo: &waE2E.ContextInfo{
						StanzaID:    proto.String(ctx.GetID()),
						Participant: proto.String(ctx.GetSender().String()),
						QuotedMessage: &waE2E.Message{
							Conversation: proto.String("Bullshit"),
						},
					},
				},
			})

			return nil
		},
	},

	// "jadwal": CommandHandler{
	// 	Func:    Jadwal,
	// 	Aliases: []string{"j"},
	// },
	// "matkul": CommandHandler{
	// 	Func:    Matkul,
	// 	Aliases: []string{"m"},
	// },
}

var mappedCommands map[string]CommandHandler = map[string]CommandHandler{}

func init() {
	log := config.GetLogger()

	log.Debugf("Indexing command aliases")
	if len(mappedCommands) == 0 {
		for key, val := range HANDLES {
			mappedCommands[key] = val
			for _, alias := range val.Aliases {
				mappedCommands[alias] = val
			}
		}
	}
}

func Handle(c *messageutil.MessageContext) error {
	cmd := c.Parser.Command.Name.Data
	c.Logger.Debugf("Got command: %s", cmd)

	if len(mappedCommands) == 0 {
		c.Logger.Debugf("Command aliases not initialized yet! Indexing")
		for key, val := range HANDLES {
			mappedCommands[key] = val
			for _, alias := range val.Aliases {
				mappedCommands[alias] = val
			}
		}
	}

	detectedFunc, ok := mappedCommands[cmd]
	if ok {
		c.Logger.Debugf("Command handler found")
		return detectedFunc.Func(c)
	}

	c.Logger.Debugf("Command handler not found, ignore the command")
	return nil
}
