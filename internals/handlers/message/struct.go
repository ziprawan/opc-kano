package message

import (
	"nopi/internals/utils/messageutils"
	"nopi/internals/utils/parser"
)

type MessageContext struct {
	Instance *messageutils.Message
	Parser   *parser.Parser
}

func InitHandlerContext(instance *messageutils.Message, parser *parser.Parser) *MessageContext {
	return &MessageContext{
		Instance: instance,
		Parser:   parser,
	}
}
