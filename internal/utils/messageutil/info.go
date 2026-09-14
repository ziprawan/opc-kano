package messageutil

import (
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/types"
)

func (c *MessageContext) GetID() types.MessageID {
	return c.Info.ID
}

func (c *MessageContext) GetSender() types.JID {
	return c.GetADSender().ToNonAD()
}

func (c *MessageContext) GetADSender() types.JID {
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

func (c *MessageContext) GetSenderAlt() types.JID {
	return c.GetADSenderAlt().ToNonAD()
}

func (c *MessageContext) GetADSenderAlt() types.JID {
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

func (c *MessageContext) GetAddressingMode() types.AddressingMode {
	return c.Info.AddressingMode
}

func (c *MessageContext) IsReactedToMe() bool {
	return c.GetReactionKey().GetFromMe()
}

func (c *MessageContext) GetReactedMsgId() types.MessageID {
	return c.GetReactionKey().GetID()
}

func (c *MessageContext) GetReactionKey() *waCommon.MessageKey {
	return c.Message.GetReactionMessage().GetKey()
}

func (c *MessageContext) GetReaction() string {
	return c.Message.GetReactionMessage().GetText()
}

func (c *MessageContext) GetChat(leaveAsIs ...bool) types.JID {
	if len(leaveAsIs) == 1 && leaveAsIs[0] {
		return c.Info.Chat
	}

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
