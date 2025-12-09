package handles

import "kano/internal/utils/messageutil"

func Confess(c *messageutil.MessageContext) error {
	c.QuoteReply("Not implemented yet. Wait for future update.")
	return ErrNotImplemented
}
