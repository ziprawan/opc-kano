package messageutil

import (
	"go.mau.fi/whatsmeow/types"
)

func (c *MessageContext) GetID() types.MessageID {
	return c.Info.ID
}

func (c *MessageContext) GetSender() types.JID {
	if c.cache.sender != nil {
		return *c.cache.sender
	}

	sender := c.Info.Sender

	c.Logger.Debugf("Sender server is %s", sender.Server)
	if sender.Server == types.DefaultUserServer {
		c.Logger.Debugf("Resolving LID")
		lid, err := c.Client.GetLIDForPN(sender)
		if err == nil {
			c.Logger.Debugf("Got LID %s for PN %s", lid.User, sender.User)
			sender = lid
		} else {
			c.Logger.Warnf("Failed to resolve LID from given PN: %s", err.Error())
		}
	}

	c.cache.sender = &sender
	return sender
}

func (c *MessageContext) GetNonADSender() types.JID {
	return c.GetSender().ToNonAD()
}

func (c *MessageContext) GetSenderAlt() types.JID {
	if c.cache.senderAlt != nil {
		return *c.cache.senderAlt
	}

	senderAlt := c.Info.SenderAlt

	c.Logger.Debugf("SenderAlt server is %s", senderAlt.Server)
	if senderAlt.Server == types.DefaultUserServer {
		c.Logger.Debugf("Resolving LID")
		lid, err := c.Client.GetLIDForPN(senderAlt)
		if err == nil {
			c.Logger.Debugf("Got LID %s for PN %s", lid.User, senderAlt.User)
			senderAlt = lid
		} else {
			c.Logger.Warnf("Failed to resolve LID from given PN: %s", err.Error())
		}
	}

	c.cache.senderAlt = &senderAlt
	return senderAlt
}

func (c *MessageContext) GetNonADSenderAlt() types.JID {
	return c.GetSenderAlt().ToNonAD()
}

func (c *MessageContext) GetAddressingMode() types.AddressingMode {
	return c.Info.AddressingMode
}

func (c *MessageContext) GetChat() types.JID {
	if c.cache.chat != nil {
		return *c.cache.chat
	}

	chat := c.Info.Chat

	c.Logger.Debugf("Chat server is %s", chat.Server)
	if chat.Server == types.DefaultUserServer {
		c.Logger.Debugf("Resolving LID")
		lid, err := c.Client.GetLIDForPN(chat)
		if err == nil {
			c.Logger.Debugf("Got LID %s for PN %s", lid.User, chat.User)
			chat = lid
		} else {
			c.Logger.Warnf("Failed to resolve LID from given PN: %s", err.Error())
		}
	}

	c.cache.chat = &chat
	return chat
}
