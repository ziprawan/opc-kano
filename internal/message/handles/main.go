package handles

import (
	"errors"
	"fmt"
	"kano/internal/config"
	"kano/internal/database"
	"kano/internal/utils/messageutil"
	"strings"
)

var ErrNotImplemented = errors.New("command not implemented")

type CommandHandlerFuncMap map[string]CommandHandler

var HANDLES CommandHandlerFuncMap = CommandHandlerFuncMap{
	"ping": CommandHandler{
		Func:    Ping,
		Aliases: []string{"p"},
		Man:     PingMan,
	},
	"stk": CommandHandler{
		Func:    Stk,
		Aliases: []string{"s"},
		Man:     StkMan,
	},
	"vo": CommandHandler{
		Func:    Vo,
		Aliases: []string{"v"},
		Man:     VoMan,
	},
	"nim": CommandHandler{
		Func: Nim,
		Man:  NimHelp,
	},
	"pddikti": CommandHandler{
		Func:    Pddikti,
		Aliases: []string{"diddy"},
		Man:     PddiktiMan,
	},
	"confess": CommandHandler{
		Func:    Confess,
		Aliases: []string{"c"},
	},
	"confesstarget": CommandHandler{
		Func:    ConfessTargetHandler,
		Aliases: []string{"ct"},
	},
	"six": CommandHandler{
		Func: Six,
		Man:  SixMan,
	},
	"ta": CommandHandler{
		Func: Ta,
		Man:  TaMan,
	},
	"test": CommandHandler{
		Func: Test,
		Man:  TestMan,
	},
	"redirect": CommandHandler{
		Func:    Redirect,
		Aliases: []string{"r", "getredir", "getloc"},
		Man:     RedirectMan,
	},
	"resolve-subject": CommandHandler{
		Func:    ResolveSubject,
		Aliases: []string{"rs"},
		Man:     ResolveSubjectMan,
	},
	"wordle": CommandHandler{
		Func:    WordleHandler,
		Aliases: []string{"worlde", "w"},
		Man:     WordleMan,
	},
	"sawit": CommandHandler{
		Func: SawitHandler,
		Man:  SawitMan,
	},
	"game": CommandHandler{
		Func: GameHandler,
		Man:  GameMan,
	},
	"help": CommandHandler{
		Func:    HelpHandler,
		Aliases: []string{"man"},
		Man:     HelpMan,
	},
	"enable": CommandHandler{
		Func:    EnableHandler,
		Aliases: []string{"disable"},
		Man:     EnableMan,
	},
	"stkline": CommandHandler{
		Func: StkLineHandler,
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
var db = database.GetInstance()

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

		// Command list string for help message
		var msg strings.Builder
		for cmd := range HANDLES {
			msg.WriteString("- " + cmd)
			aliases := HANDLES[cmd].Aliases
			if len(aliases) > 0 {
				fmt.Fprintf(&msg, " (%s)", strings.Join(aliases, ", "))
			}
			msg.WriteString("\n")
		}

		commandListStr = msg.String()
	}
}

func Handle(c *messageutil.MessageContext) error {
	cmd := c.Parser.Command.Name.Data
	c.Logger.Debugf("Got command: %s", cmd)

	detectedFunc, ok := mappedCommands[cmd]
	if ok {
		c.Logger.Debugf("Command handler found")
		return detectedFunc.Func(c)
	}

	c.Logger.Debugf("Command handler not found, ignore the command")
	return nil
}
