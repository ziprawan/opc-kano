package handles

import (
	"kano/internal/config"
	"kano/internal/utils/messageutil"
)

type MessageContext messageutil.MessageContext
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
		Aliases: []string{"stk"},
	},
	"vo": CommandHandler{
		Func: Vo,
	},
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
	cmd := c.Parser.Command.Name
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
