package handles

import (
	"kano/internal/utils/messageutil"
)

func Ping(c *messageutil.MessageContext) error {
	_, err := c.QuoteReply("Pong!")
	return err
}

var PingMan = CommandMan{
	Name:     "ping - pong",
	Synopsis: []string{"*ping*"},
	Description: []string{
		"A simple command that just makes the bot respond \"pong\". Usually used to check whether a bot is online or not.",
	},
	SourceFilename: "ping.go",
	SeeAlso:        []SeeAlso{},
}
