package handles

import (
	"kano/internal/message/handles/six"
	"kano/internal/utils/messageutil"
	"strings"
)

func Six(c *messageutil.MessageContext) error {
	args := c.Parser.Args
	if len(args) == 0 {
		return six.CommandMap["help"](c)
	}

	sixCommand := strings.ToLower(args[0].Content.Data)
	theFunc, ok := six.CommandMap[sixCommand]
	if !ok {
		c.QuoteReply("Perintah SIX tidak valid: %s", sixCommand)
		return nil
	} else {
		return theFunc(c)
	}
}
