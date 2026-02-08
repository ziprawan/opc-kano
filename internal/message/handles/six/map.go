package six

import (
	"kano/internal/utils/messageutil"
)

var CommandMap = map[string]func(*messageutil.MessageContext) error{
	"update": updateHandler,
	"u":      updateHandler,

	"help": helpHandler,

	"follow": followHandler,
	"f":      followHandler,

	"reminder": reminderHandler,
	"r":        reminderHandler,
}

var helpMap = map[string]func(*messageutil.MessageContext){
	"r":        reminderHelp,
	"reminder": reminderHelp,
}
