package handles

import (
	"kano/internal/utils/messageutil"
)

func Ping(c *messageutil.MessageContext) error {
	_, err := c.QuoteReply("Pong!")
	return err
}
